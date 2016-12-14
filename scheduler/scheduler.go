package scheduler

import (
	"mesos-sdk"
	"mesos-sdk/backoff"
	"mesos-sdk/encoding"
	ctrl "mesos-sdk/extras/scheduler/controller"
	"mesos-sdk/httpcli"
	"mesos-sdk/httpcli/httpsched"
	"mesos-sdk/scheduler/calls"
	"net/http"
	"time"
)

// Base implementation of a scheduler.
type scheduler interface {
	Run(c ctrl.Controller, config *ctrl.Config) error
	State() *state
	Caller() *calls.Caller
	FrameworkInfo() *mesos.FrameworkInfo
}

// Scheduler state.
type state struct {
	frameworkId   string
	tasksLaunched uint
	tasksFinished uint
	totalTasks    uint
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
	var executorName = new(string)
	*executorName = "Sprinter"

	return &sprintScheduler{
		config: cfg,
		framework: &mesos.FrameworkInfo{
			Name:       cfg.Name(),
			Checkpoint: cfg.Checkpointing(),
		},
		executor: &mesos.ExecutorInfo{
			ExecutorID: mesos.ExecutorID{
				Value: "default",
			},
			Name: executorName,
			Command: mesos.CommandInfo{
				Value: cfg.Command(),
				URIs:  cfg.Uris(),
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
			reviveTokens: backoff.BurstNotifier(cfg.ReviveBurst(), cfg.ReviveWait(), cfg.ReviveWait(), nil),
		},
	}
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

// Runs our scheduler with some applied configuration.
func (s *sprintScheduler) Run(c ctrl.Controller, config *ctrl.Config) error {
	return c.Run(*config)
}
