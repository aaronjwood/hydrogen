package main

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"flag"
	"mesos-framework-sdk/client"
	"mesos-framework-sdk/include/mesos"
	"mesos-framework-sdk/include/scheduler"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/persistence/drivers/etcd"
	"mesos-framework-sdk/resources/manager"
	sched "mesos-framework-sdk/scheduler"
	"mesos-framework-sdk/server"
	"mesos-framework-sdk/server/file"
	"mesos-framework-sdk/structures"
	sdkTaskManager "mesos-framework-sdk/task/manager"
	"mesos-framework-sdk/utils"
	"net"
	"net/http"
	"sprint/scheduler"
	"sprint/scheduler/api"
	"sprint/scheduler/events"
	sprintTaskManager "sprint/task/manager"
	"strconv"
	"strings"
	"time"
)

// NOTE: This should be refactored out of the main file.
func CreateFrameworkInfo(config *scheduler.SchedulerConfiguration) *mesos_v1.FrameworkInfo {
	return &mesos_v1.FrameworkInfo{
		User:            &config.User,
		Name:            &config.Name,
		FailoverTimeout: &config.Failover,
		Checkpoint:      &config.Checkpointing,
		Role:            &config.Role,
		Hostname:        &config.Hostname,
		Principal:       &config.Principal,
	}
}

// NOTE: This should be in the event manager.
// Keep our state in check by periodically reconciling.
// This is recommended by Mesos.
func periodicReconcile(c *scheduler.SchedulerConfiguration, e *events.SprintEventController) {
	ticker := time.NewTicker(c.ReconcileInterval)

	for {
		select {
		case <-ticker.C:

			recon, err := e.TaskManager().GetState(sdkTaskManager.RUNNING)
			if err != nil {
				// log here.
				continue
			}
			e.Scheduler().Reconcile(recon)
		}
	}
}

// NOTE: This should be in the event manager.
// Get all of our persisted tasks, convert them back into TaskInfo's, and add them to our task manager.
// If no tasks exist in the data store then we can consider this a fresh run and safely move on.
func restoreTasks(kv *etcd.Etcd, t *sprintTaskManager.SprintTaskManager, logger logging.Logger) error {
	tasks, err := kv.ReadAll("/tasks")
	if err != nil {
		return err
	}

	for _, value := range tasks {
		var task sdkTaskManager.Task
		data, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			logger.Emit(logging.ERROR, err.Error())
		}

		var b bytes.Buffer
		b.Write(data)
		d := gob.NewDecoder(&b)
		err = d.Decode(&task)
		if err != nil {
			logger.Emit(logging.ERROR, err.Error())
		}

		t.Set(task.State, task.Info)
	}

	return nil
}

// Handles connections from other framework instances that try and determine the state of the leader.
// Used in coordination with determining if and when we need to perform leader election.
func leaderServer(c *scheduler.SchedulerConfiguration, logger logging.Logger) {
	ips, err := utils.GetIPs(c.NetworkInterface)
	if err != nil {
		logger.Emit(logging.ERROR, "Leader server exiting: %s", err.Error())
	}

	addr, err := net.ResolveTCPAddr(c.LeaderAddressFamily, "["+ips[c.LeaderAddressFamily]+"]:"+strconv.Itoa(c.LeaderServerPort))
	if err != nil {
		logger.Emit(logging.ERROR, "Leader server exiting: %s", err.Error())
		return
	}

	tcp, err := net.ListenTCP(c.LeaderAddressFamily, addr)
	if err != nil {
		logger.Emit(logging.ERROR, "Leader server exiting: %s", err.Error())
		return
	}

	for {

		// Block here until we get a new connection.
		// We don't want to do anything with the stream so move on without spawning a thread to handle the connection.
		conn, err := tcp.AcceptTCP()
		if err != nil {
			logger.Emit(logging.ERROR, "Failed to accept client: %s", err.Error())
			time.Sleep(1 * time.Second) // TODO should this be configurable?
			continue
		}

		// TODO build out some config to use for setting the keep alive period here
		if err := conn.SetKeepAlive(true); err != nil {
			logger.Emit(logging.ERROR, "Failed to set keep alive: %s", err.Error())
		}
	}
}

// Connects to the leader and determines if and when we should start the leader election process.
func leaderClient(c *scheduler.SchedulerConfiguration, leader string) error {
	conn, err := net.DialTimeout(c.LeaderAddressFamily, "["+leader+"]:"+strconv.Itoa(c.LeaderServerPort), 2*time.Second) // TODO make this configurable?
	if err != nil {
		return err
	}

	// TODO build out some config to use for setting the keep alive period here
	tcp := conn.(*net.TCPConn)
	if err := tcp.SetKeepAlive(true); err != nil {
		return err
	}

	buffer := make([]byte, 1)
	for {
		_, err := tcp.Read(buffer)
		if err != nil {
			return err
		}
	}
}

// Entry point for the scheduler.
// Parses configuration from user-supplied flags and prepares the scheduler for execution.
func main() {
	logger := logging.NewDefaultLogger()

	// Executor/API server configuration.
	cert := flag.String("server.cert", "", "TLS certificate")
	key := flag.String("server.key", "", "TLS key")
	path := flag.String("server.executor.path", "executor", "Path to the executor binary")
	port := flag.Int("server.executor.port", 8081, "Executor server listen port")
	apiPort := flag.Int("server.api.port", 8080, "API server listen port")

	// Define our framework here
	schedulerConfig := new(scheduler.SchedulerConfiguration).Initialize()
	frameworkInfo := CreateFrameworkInfo(schedulerConfig)

	flag.Parse()

	// Executor Server
	srvConfig := server.NewConfiguration(*cert, *key, *path, *port)
	executorSrv := file.NewExecutorServer(srvConfig, logger)

	// API server
	apiSrv := api.NewApiServer(srvConfig, http.NewServeMux(), apiPort, "v1", logger)

	logger.Emit(logging.INFO, "Starting executor file server")

	// Executor server serves up our custom executor binary, if any.
	go executorSrv.Serve()

	// Used to listen for events coming from mesos master to our scheduler.
	eventChan := make(chan *mesos_v1_scheduler.Event)

	// Wire up dependencies for the event controller
	kv := etcd.NewClient(
		strings.Split(schedulerConfig.StorageEndpoints, ","),
		schedulerConfig.StorageTimeout,
	) // Storage client
	m := sprintTaskManager.NewTaskManager(structures.NewConcurrentMap(100)) // Manages our tasks
	r := manager.NewDefaultResourceManager()                                // Manages resources from the cluster
	c := client.NewClient(schedulerConfig.MesosEndpoint, logger)            // Manages HTTP calls
	s := sched.NewDefaultScheduler(c, frameworkInfo, logger)                // Manages how to route and schedule tasks.

	// Event controller manages scheduler events and how they are handled.
	e := events.NewSprintEventController(schedulerConfig, s, m, r, eventChan, kv, logger)

	logger.Emit(logging.INFO, "Starting leader election socket server")
	go leaderServer(schedulerConfig, logger)

	for {
		e.SetLeader()

		leader, err := e.GetLeader()
		if err != nil {
			logger.Emit(logging.ERROR, "Couldn't get leader: %s", err.Error())
			time.Sleep(schedulerConfig.LeaderRetryInterval)
			continue
		}

		ips, err := utils.GetIPs(schedulerConfig.NetworkInterface)
		if err != nil {
			logger.Emit(logging.ERROR, "Couldn't determine IPs for interface: %s", err.Error())
			time.Sleep(schedulerConfig.LeaderRetryInterval)
			continue
		}

		if leader != ips[schedulerConfig.LeaderAddressFamily] {
			logger.Emit(logging.INFO, "Connecting to leader to determine when we need to wake up and perform leader election")

			// Block here until we lose connection to the leader.
			// Once the connection has been lost elect a new leader.
			err := leaderClient(schedulerConfig, leader)

			// Only delete the key if we've lost the connection, not timed out.
			// This conditional requires Go 1.6+
			if err, ok := err.(net.Error); ok && err.Timeout() {
				logger.Emit(logging.ERROR, "Timed out connecting to leader")
			} else {
				logger.Emit(logging.ERROR, "Lost connection to leader")
				kv.Delete("/leader")
			}
		} else {

			// We are the leader, exit the loop and start the scheduler/API.
			break
		}
	}

	logger.Emit(logging.INFO, "Starting API server")

	// Run our API in a go routine to listen for user requests.
	go apiSrv.RunAPI(e, nil) // nil means to use default handlers.

	// Recover our state (if any) in the event we (or the server) go down.
	logger.Emit(logging.INFO, "Restoring any persisted state from data store")
	if err := restoreTasks(kv, m, logger); err != nil {
		logger.Emit(logging.ERROR, "Failed to restore tasks from persistent data store")
	}

	// Kick off our scheduled reconciling.
	logger.Emit(logging.INFO, "Starting periodic reconciler thread with a %g minute interval", schedulerConfig.ReconcileInterval.Minutes())
	go periodicReconcile(schedulerConfig, e)

	// Run our event controller to subscribe to mesos master and start listening for events.
	e.Run()
}
