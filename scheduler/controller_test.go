package scheduler

import (
	"mesos-sdk"
	ctrl "mesos-sdk/extras/scheduler/controller"
	"mesos-sdk/scheduler/calls"
	"reflect"
	"testing"
)

// Mocked controller.
type mockController struct{}

func (m *mockController) SchedulerCtrl() ctrl.Controller {
	return ctrl.New()
}

func (m *mockController) BuildContext() *ctrl.ContextAdapter {
	return new(ctrl.ContextAdapter)
}

func (m *mockController) BuildFrameworkInfo(cfg configuration) *mesos.FrameworkInfo {
	return &mesos.FrameworkInfo{}
}

func (m *mockController) BuildConfig(ctx *ctrl.ContextAdapter, http *calls.Caller, shutdown <-chan struct{}, h *handlers) *ctrl.Config {
	return &ctrl.Config{}
}

var c controller

//Prepare common data for our tests.
func init() {
	cfg = new(mockConfiguration).Initialize(nil)
	s = &mockScheduler{
		cfg: cfg,
	}
	c = NewController(s, make(<-chan struct{}))
}

// Ensures that we get the correct type from creating a new controller.
func TestNewController(t *testing.T) {
	t.Parallel()

	switch c.(type) {
	case *sprintController:
		return
	default:
		t.Fatal("Controller is not of the right type")
	}
}

// Ensures that we get the correct type from getting the internal scheduler controller.
func TestController_GetSchedulerCtrl(t *testing.T) {
	t.Parallel()

	switch c.SchedulerCtrl().(type) {
	case ctrl.Controller:
		return
	default:
		t.Fatal("Scheduler controller is not of the right type")
	}
}

// Ensures we have the right types after building the context.
func TestController_BuildContext(t *testing.T) {
	t.Parallel()

	ctx := c.BuildContext()

	if reflect.TypeOf(ctx) != reflect.TypeOf(new(ctrl.ContextAdapter)) {
		t.Fatal("Controller context is not of the right type")
	}
	if reflect.TypeOf(ctx.DoneFunc).Kind() != reflect.Func {
		t.Fatal("Context does not have a valid done function")
	}
	if reflect.TypeOf(ctx.FrameworkIDFunc).Kind() != reflect.Func {
		t.Fatal("Context does not have a valid FrameworkID function")
	}
	if reflect.TypeOf(ctx.ErrorFunc).Kind() != reflect.Func {
		t.Fatal("Context does not have a valid error function")
	}

	if reflect.TypeOf(ctx.FrameworkIDFunc()).Kind() != reflect.String {
		t.Fatal("FrameworkID function does not return the correct type")
	}
	if reflect.TypeOf(ctx.DoneFunc()).Kind() != reflect.Bool {
		t.Fatal("FrameworkID function does not return the correct type")
	}
}

// Ensures that we have correctly build the FrameworkInfo that will be sent to Mesos.
func TestController_BuildFrameworkInfo(t *testing.T) {
	t.Parallel()

	info := c.BuildFrameworkInfo(cfg)
	if info.GetName() != cfg.Name() {
		t.Fatal("FrameworkInfo has the wrong name")
	}
	if info.GetCheckpoint() != true {
		t.Fatal("FrameworkInfo does not have checkpointing set correctly")
	}
}

// Ensures that we build the controller's configuration correctly.
func TestController_BuildConfig(t *testing.T) {
	t.Parallel()

	ctx := c.BuildContext()
	http := new(mockScheduler).Caller()
	shutdown := make(<-chan struct{})
	handlers := NewHandlers(s)

	config := c.BuildConfig(ctx, http, shutdown, handlers)
	if reflect.TypeOf(config) != reflect.TypeOf(new(ctrl.Config)) {
		t.Fatal("Controller configuration is not of the right type")
	}
	if config.Context != ctx {
		t.Fatal("Configuration contexts don't match")
	}
	if reflect.TypeOf(config.Framework) != reflect.TypeOf(s.FrameworkInfo()) {
		t.Fatal("Configuration FrameworkInfo does not match")
	}
	if config.Caller != *http {
		t.Fatal("Configuration caller does not match")
	}
}