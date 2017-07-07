package builder

import (
	"errors"
	"mesos-framework-sdk/include/mesos_v1"
	resourcebuilder "mesos-framework-sdk/resources"
	"mesos-framework-sdk/task"
	"mesos-framework-sdk/task/command"
	"mesos-framework-sdk/task/container"
	"mesos-framework-sdk/task/healthcheck"
	"mesos-framework-sdk/task/labels"
	"mesos-framework-sdk/task/resources"
	"mesos-framework-sdk/utils"
)

var NoNameError = errors.New("A name is required for the application. Please set the name field.")
var NoResourcesError = errors.New("Application requested with no resources. Please set some resources.")

// Parses a JSON request and turns it into a TaskInfo.
func Application(t *task.ApplicationJSON) (*mesos_v1.TaskInfo, error) {
	// Check for required name.
	if t.Name == "" {
		// Fail
		return nil, NoNameError
	}
	if t.Resources == nil {
		// Fail
		return nil, NoResourcesError
	}

	// Agent and TaskID are required but are set by the Resource Manager in the scheduler.

	// Executor or CommandInfo must be set.
	// An end user will never be allowed to set an executor, only other frameworks.
	// So here we assume a commandInfo.
	cmd, err := command.ParseCommandInfo(t.Command)
	if err != nil {
		// If we don't have a commandInfo, it's invalid.
		return nil, err
	}

	// Parse resources
	// These are required, fail if no resources are specified.
	res, err := resources.ParseResources(t.Resources)
	if err != nil {
		return nil, err
	}

	// Container parse
	image, err := container.ParseContainer(t.Container)
	if err != nil {
		return nil, err
	}

	lbls, err := labels.ParseLabels(t.Labels)
	if err != nil {
		return nil, err
	}

	hc, err := healthcheck.ParseHealthCheck(t.HealthCheck)
	if err != nil {
		return nil, err
	}

	name := t.Name
	taskId := &mesos_v1.TaskID{Value: utils.ProtoString(name)}

	return resourcebuilder.CreateTaskInfo(
		utils.ProtoString(name),
		taskId,
		cmd,
		res,
		image,
		hc,
		lbls,
	), nil
}
