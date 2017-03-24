package executor

/*
Executor interface and default executor implementation is defined here.
*/
import (
	"mesos-framework-sdk/client"
	exec "mesos-framework-sdk/include/executor"
	"mesos-framework-sdk/include/mesos"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/recordio"
)

const (
	subscribeRetry = 2
)

type SprintExecutor struct {
	FrameworkID *mesos_v1.FrameworkID
	ExecutorID  *mesos_v1.ExecutorID
	client      *client.Client
	logger      logging.Logger
}

// Creates a new default executor
func NewSprintExecutor(
	f *mesos_v1.FrameworkID,
	e *mesos_v1.ExecutorID,
	c *client.Client,
	lgr logging.Logger) *SprintExecutor {

	return &SprintExecutor{
		FrameworkID: f,
		ExecutorID:  e,
		client:      c,
		logger:      lgr,
	}

}

func (d *SprintExecutor) Subscribe(eventChan chan *exec.Event) error {
	subscribe := &exec.Call{
		FrameworkId: d.FrameworkID,
		ExecutorId:  d.ExecutorID,
		Type:        exec.Call_SUBSCRIBE.Enum(),
	}

	// If we disconnect we need to reset the stream ID. For this reason always start with a fresh stream ID.
	// Otherwise we'll never be able to reconnect.
	d.client.StreamID = ""

	resp, err := d.client.Request(subscribe)
	if err != nil {
		return err
	} else {
		return recordio.Decode(resp.Body, eventChan)
	}
}

func (d *SprintExecutor) Update(taskStatus *mesos_v1.TaskStatus) {
	update := exec.Call{
		FrameworkId: d.FrameworkID,
		ExecutorId:  d.ExecutorID,
		Type:        exec.Call_UPDATE.Enum(),
		Update: &exec.Call_Update{
			Status: taskStatus,
		},
	}
	d.client.Request(update)
}

func (d *SprintExecutor) Message(data []byte) {
	message := exec.Call{
		FrameworkId: d.FrameworkID,
		ExecutorId:  d.ExecutorID,
		Type:        exec.Call_MESSAGE.Enum(),
		Message: &exec.Call_Message{
			Data: data,
		},
	}
	d.client.Request(message)
}