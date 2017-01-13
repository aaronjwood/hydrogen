package scheduler

import (
	"fmt"
	"mesos-sdk"
	"mesos-sdk/backoff"
	"mesos-sdk/encoding"
	"mesos-sdk/extras"
	ctrl "mesos-sdk/extras/scheduler/controller"
	"mesos-sdk/httpcli"
	"mesos-sdk/httpcli/httpsched"
	"mesos-sdk/scheduler/calls"
	"net/http"
	"sprint"
	"strconv"
	"time"
)

// Base implementation of a scheduler.
type scheduler interface {
	Config() configuration
	Run(c ctrl.Controller, config *ctrl.Config) error
	State() *state
	Caller() *calls.Caller
	FrameworkInfo() *mesos.FrameworkInfo
	ExecutorInfo() *mesos.ExecutorInfo
	SuppressOffers() error
	ReviveOffers() error
}

// Scheduler state.
type state struct {
	frameworkId   string
	tasksLaunched int
	tasksFinished int
	totalTasks    int
	taskResources mesos.Resources
	role          string
	done          bool
	reviveTokens  <-chan struct{}
}

// Holds all necessary information for our scheduler to function.
type sprintScheduler struct {
	config    configuration
	framework *mesos.FrameworkInfo
	executor  *mesos.ExecutorInfo
	http      calls.Caller
	shutdown  chan struct{}
	state     state
}

// Returns a new scheduler using user-supplied configuration.
func NewScheduler(cfg configuration, shutdown chan struct{}) *sprintScheduler {
	uuid := extras.Uuid()
	id := fmt.Sprintf("%X-%X-%X-%X-%X", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])

	return &sprintScheduler{
		config: cfg,
		framework: &mesos.FrameworkInfo{
			User:       cfg.User(),
			Name:       cfg.Name(),
			Checkpoint: cfg.Checkpointing(),
		},
		executor: &mesos.ExecutorInfo{
			ExecutorID: mesos.ExecutorID{Value: id},
			Name:       cfg.ExecutorName(),
			Command: mesos.CommandInfo{
				Value: cfg.ExecutorCmd(),
				URIs: []mesos.CommandInfo_URI{
					{
						// TODO make this configurable
						// Expose more configuration for the server which can be reused here
						Value:      "http://10.0.2.2:" + strconv.Itoa(cfg.ExecutorSrvCfg().ExecutorSrvPort()) + "/executor",
						Executable: sprint.ProtoBool(true),
					},
				},
			},
			// TODO parameterize this
			Resources: []mesos.Resource{
				sprint.Resource("cpus", 0.5),
				sprint.Resource("mem", 1024.0),
			},
			Container: &mesos.ContainerInfo{
				Type: mesos.ContainerInfo_MESOS.Enum(),
			},
		},
		http: httpsched.NewCaller(httpcli.New(
			httpcli.Endpoint(cfg.Endpoint()),
			httpcli.Codec(&encoding.ProtobufCodec),
			httpcli.Do(
				httpcli.With(
					httpcli.Timeout(cfg.Timeout()),
					httpcli.Transport(func(t *http.Transport) {
						t.ResponseHeaderTimeout = 15 * time.Second
						t.MaxIdleConnsPerHost = 2
					}),
				),
			),
		)),
		shutdown: shutdown,
		state: state{
			totalTasks:   5, // TODO For testing, we need to allow POST'ing of tasks to the framework.
			reviveTokens: backoff.BurstNotifier(cfg.ReviveBurst(), cfg.ReviveWait(), cfg.ReviveWait(), nil),
		},
	}
}

// Returns the scheduler's configuration.
func (s *sprintScheduler) Config() configuration {
	return s.config
}

// Returns the internal state of the scheduler
func (s *sprintScheduler) State() *state {
	return &s.state
}

// Returns the caller that we use for communication.
func (s *sprintScheduler) Caller() *calls.Caller {
	return &s.http
}

// Returns the FrameworkInfo that is sent to Mesos.
func (s *sprintScheduler) FrameworkInfo() *mesos.FrameworkInfo {
	return s.framework
}

// Returns the ExecutorInfo that's associated with the scheduler.
func (s *sprintScheduler) ExecutorInfo() *mesos.ExecutorInfo {
	return s.executor
}

// Runs our scheduler with some applied configuration.
func (s *sprintScheduler) Run(c ctrl.Controller, config *ctrl.Config) error {
	return c.Run(*config)
}

// This call suppresses our offers received from Mesos.
func (s *sprintScheduler) SuppressOffers() error {
	return calls.CallNoData(s.http, calls.Suppress())
}

// This call revives our offers received from Mesos.
func (s *sprintScheduler) ReviveOffers() error {
	select {
	// Rate limit offer revivals.
	case <-s.state.reviveTokens:
		return calls.CallNoData(s.http, calls.Revive())
	default:
		return nil
	}
}
