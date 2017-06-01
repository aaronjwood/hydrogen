package test

import (
	"errors"
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/task/manager"
)

type (
	MockApiManager       struct{}
	MockBrokenApiManager struct{}
)

func (m MockApiManager) Deploy([]byte) (*mesos_v1.TaskInfo, error) { return &mesos_v1.TaskInfo{}, nil }
func (m MockApiManager) Kill([]byte) (string, error)               { return "", nil }
func (m MockApiManager) Update([]byte) (*mesos_v1.TaskInfo, error) { return &mesos_v1.TaskInfo{}, nil }
func (m MockApiManager) Status(string) (mesos_v1.TaskState, error) {
	return mesos_v1.TaskState_TASK_RUNNING, nil
}
func (m MockApiManager) AllTasks() ([]manager.Task, error) {
	return []manager.Task{{
		&mesos_v1.TaskInfo{},
		mesos_v1.TaskState_TASK_RUNNING,
	}}, nil
}

func (m MockBrokenApiManager) Deploy([]byte) (*mesos_v1.TaskInfo, error) {
	return nil, errors.New("Broken")
}
func (m MockBrokenApiManager) Kill([]byte) (string, error) { return "", errors.New("Broken") }
func (m MockBrokenApiManager) Update([]byte) (*mesos_v1.TaskInfo, error) {
	return nil, errors.New("Broken")
}
func (m MockBrokenApiManager) Status(string) (mesos_v1.TaskState, error) {
	return mesos_v1.TaskState_TASK_UNKNOWN, errors.New("Broken")
}
func (m MockBrokenApiManager) AllTasks() ([]manager.Task, error) {
	return nil, errors.New("Broken")
}
