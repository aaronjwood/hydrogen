package test

import (
	"errors"
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/structures"
	"mesos-framework-sdk/structures/test"
	"mesos-framework-sdk/task"
	"mesos-framework-sdk/utils"
	"sprint/task/retry"
)

type MockTaskManager struct{}

func (m MockTaskManager) AddPolicy(*task.TimeRetry, *mesos_v1.TaskInfo) error {
	return nil
}
func (m MockTaskManager) CheckPolicy(*mesos_v1.TaskInfo) (retry.TaskRetry, error) {
	return retry.TaskRetry{}, nil
}
func (m MockTaskManager) ClearPolicy(*mesos_v1.TaskInfo) error {
	return nil
}
func (m MockTaskManager) RunPolicy(retry.TaskRetry, func() error) error {
	return nil
}

func (m MockTaskManager) Add(*mesos_v1.TaskInfo) error {
	return nil
}

func (m MockTaskManager) Delete(*mesos_v1.TaskInfo) error {
	return nil
}

func (m MockTaskManager) Get(*string) (*mesos_v1.TaskInfo, error) {
	return &mesos_v1.TaskInfo{}, nil
}

func (m MockTaskManager) GetById(id *mesos_v1.TaskID) (*mesos_v1.TaskInfo, error) {
	return &mesos_v1.TaskInfo{}, nil
}

func (m MockTaskManager) HasTask(*mesos_v1.TaskInfo) bool {
	return false
}

func (m MockTaskManager) Set(mesos_v1.TaskState, *mesos_v1.TaskInfo) error {
	return nil
}

func (m MockTaskManager) GetState(state mesos_v1.TaskState) ([]*mesos_v1.TaskInfo, error) {
	return []*mesos_v1.TaskInfo{
		{},
	}, nil
}

func (m MockTaskManager) TotalTasks() int {
	return 0
}

func (m MockTaskManager) Tasks() structures.DistributedMap {
	return &test.MockDistributedMap{}
}

func (m MockTaskManager) RetryPolicy() retry.TaskRetry {
	return retry.TaskRetry{}
}

//
// Mock Broken Task Manager
//
type MockBrokenTaskManager struct{}

func (m MockBrokenTaskManager) AddPolicy(*task.TimeRetry, *mesos_v1.TaskInfo) error {
	return nil
}
func (m MockBrokenTaskManager) CheckPolicy(*mesos_v1.TaskInfo) (retry.TaskRetry, error) {
	return retry.TaskRetry{}, nil
}
func (m MockBrokenTaskManager) ClearPolicy(*mesos_v1.TaskInfo) error {
	return nil
}
func (m MockBrokenTaskManager) RunPolicy(retry.TaskRetry, func() error) error {
	return nil
}

func (m MockBrokenTaskManager) Add(*mesos_v1.TaskInfo) error {
	return errors.New("Broken.")
}

func (m MockBrokenTaskManager) Delete(*mesos_v1.TaskInfo) error {
	return errors.New("Broken.")
}

func (m MockBrokenTaskManager) Get(*string) (*mesos_v1.TaskInfo, error) {
	return nil, errors.New("Broken.")
}

func (m MockBrokenTaskManager) GetById(id *mesos_v1.TaskID) (*mesos_v1.TaskInfo, error) {
	return nil, errors.New("Broken.")
}

func (m MockBrokenTaskManager) HasTask(*mesos_v1.TaskInfo) bool {
	return false
}

func (m MockBrokenTaskManager) Set(mesos_v1.TaskState, *mesos_v1.TaskInfo) error {
	return errors.New("Broken.")
}

func (m MockBrokenTaskManager) GetState(state mesos_v1.TaskState) ([]*mesos_v1.TaskInfo, error) {
	return nil, errors.New("Broken.")
}

func (m MockBrokenTaskManager) TotalTasks() int {
	return 0
}

func (m MockBrokenTaskManager) Tasks() structures.DistributedMap {
	return &test.MockBrokenDistributedMap{}
}

func (m MockBrokenTaskManager) RetryPolicy() retry.TaskRetry {
	return retry.TaskRetry{}
}

type MockTaskManagerQueued struct{}

func (m MockTaskManagerQueued) AddPolicy(*task.TimeRetry, *mesos_v1.TaskInfo) error {
	return nil
}
func (m MockTaskManagerQueued) CheckPolicy(*mesos_v1.TaskInfo) (retry.TaskRetry, error) {
	return retry.TaskRetry{}, nil
}
func (m MockTaskManagerQueued) ClearPolicy(*mesos_v1.TaskInfo) error {
	return nil
}
func (m MockTaskManagerQueued) RunPolicy(retry.TaskRetry, func() error) error {
	return nil
}

func (m MockTaskManagerQueued) Add(*mesos_v1.TaskInfo) error {
	return nil
}

func (m MockTaskManagerQueued) Delete(*mesos_v1.TaskInfo) error {
	return nil
}

func (m MockTaskManagerQueued) Get(*string) (*mesos_v1.TaskInfo, error) {
	return &mesos_v1.TaskInfo{}, nil
}

func (m MockTaskManagerQueued) GetById(id *mesos_v1.TaskID) (*mesos_v1.TaskInfo, error) {
	return &mesos_v1.TaskInfo{}, nil
}

func (m MockTaskManagerQueued) HasTask(*mesos_v1.TaskInfo) bool {
	return false
}

func (m MockTaskManagerQueued) Set(mesos_v1.TaskState, *mesos_v1.TaskInfo) error {
	return nil
}

func (m MockTaskManagerQueued) GetState(state mesos_v1.TaskState) ([]*mesos_v1.TaskInfo, error) {
	return []*mesos_v1.TaskInfo{
		{
			Name:    utils.ProtoString("Name"),
			TaskId:  &mesos_v1.TaskID{Value: utils.ProtoString("1")},
			AgentId: &mesos_v1.AgentID{Value: utils.ProtoString("agent")},
		},
	}, nil
}

func (m MockTaskManagerQueued) TotalTasks() int {
	return 1
}

func (m MockTaskManagerQueued) Tasks() structures.DistributedMap {
	return &test.MockDistributedMap{}
}

func (m MockTaskManagerQueued) RetryPolicy() retry.TaskRetry {
	return retry.TaskRetry{}
}
