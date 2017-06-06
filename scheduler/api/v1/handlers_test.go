package v1

import (
	"io"
	"net/http"
	"net/http/httptest"
	mockApiManager "sprint/scheduler/api/manager/test"
	"strings"
	"testing"
)

var (
	apiMgr                = new(mockApiManager.MockApiManager)
	brokenApiMgr          = new(mockApiManager.MockBrokenApiManager)
	validJSON             = `{"name": "test", "resources": {"cpu": 0.5, "mem": 128.0}, "command": {"cmd": "echo hello"}}`
	killJSON              = `{"name": "test"}`
	junkJSON              = `not even json, how did this even get here`
	filtersJSON           = `{"name": "test", "filters": [{"type": "TEXT", "value": ["tester"]}], "resources": {"cpu": 0.5, "mem": 128.0}, "command": {"cmd": "echo hello"}}`
	badFiltersJSON        = `{"name": "test", "filters": [{"type": "not real", "value": "tester"}], "resources": {"cpu": 0.5, "mem": 128.0}, "command": {"cmd": "echo hello"}}`
	invalidFilterTypeJSON = `{"name": "test", "filters": [{"type": "fake news", "value": ["tester"]}], "resources": {"cpu": 0.5, "mem": 128.0}, "command": {"cmd": "echo hello"}}`
)

// Common request-response handling for tests to share.
func requestFixture(f http.HandlerFunc, method, endpoint string, r io.Reader) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, endpoint, r)

	rr := httptest.NewRecorder()
	http.HandlerFunc(f).ServeHTTP(rr, req)

	return rr
}

// Verifies that the handlers have the correct state.
func TestNewHandlers(t *testing.T) {
	h := NewHandlers(apiMgr)
	if h.manager != apiMgr {
		t.Fatal("API does not contain the correct components")
	}
}

// Validates the deployment endpoint.
func TestHandlers_Deploy(t *testing.T) {
	h := NewHandlers(apiMgr)
	rr := requestFixture(h.Deploy, "POST", "/deploy", strings.NewReader(validJSON))
	if rr.Code != http.StatusOK {
		t.Fatalf("Wrong status code: want %d but got %d", http.StatusOK, rr.Code)
	}
}

// Makes sure the deployment endpoint gives an error when it should.
func TestHandlers_DeployError(t *testing.T) {
	h := NewHandlers(brokenApiMgr)
	rr := requestFixture(h.Deploy, "POST", "/deploy", strings.NewReader(junkJSON))
	if rr.Code == http.StatusOK {
		t.Fatalf("Wrong status code: want %d but got %d", rr.Code, http.StatusOK)
	}
}

// Validates the endpoint to kill tasks.
func TestHandlers_Kill(t *testing.T) {
	h := NewHandlers(apiMgr)
	rr := requestFixture(h.Kill, "DELETE", "/kill", strings.NewReader(killJSON))
	if rr.Code != http.StatusOK {
		t.Fatalf("Wrong status code: want %d but got %d", http.StatusOK, rr.Code)
	}
}

// Makes sure the endpoint to kill tasks gives an error when it should.
func TestHandlers_KillError(t *testing.T) {
	h := NewHandlers(brokenApiMgr)
	rr := requestFixture(h.Kill, "DELETE", "/kill", strings.NewReader(killJSON))
	if rr.Code == http.StatusOK {
		t.Fatalf("Wrong status code: want %d but got %d", rr.Code, http.StatusOK)
	}
}

// Validates the endpoint to get task state.
func TestHandlers_State(t *testing.T) {
	h := NewHandlers(apiMgr)
	rr := requestFixture(h.State, "GET", "/status?name=test", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("Wrong status code: want %d but got %d", http.StatusOK, rr.Code)
	}
}

// Makes sure the endpoint to get task state gives an error when it should.
func TestHandlers_StateError(t *testing.T) {
	h := NewHandlers(brokenApiMgr)
	rr := requestFixture(h.State, "GET", "/status?name=test", nil)
	if rr.Code == http.StatusOK {
		t.Fatalf("Wrong status code: want %d but got %d", rr.Code, http.StatusOK)
	}

	h = NewHandlers(apiMgr)
	rr = requestFixture(h.State, "GET", "/status", nil)
	if rr.Code == http.StatusOK {
		t.Fatalf("Wrong status code: want %d but got %d", rr.Code, http.StatusOK)
	}
}

// Validates the endpoint to get all tasks.
func TestHandlers_Tasks(t *testing.T) {
	h := NewHandlers(apiMgr)
	rr := requestFixture(h.Tasks, "GET", "/tasks", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("Wrong status code: want %d but got %d", http.StatusOK, rr.Code)
	}
}

// Makes sure that the endpoint to get all tasks gives an error when it should.
func TestHandlers_TasksError(t *testing.T) {
	h := NewHandlers(brokenApiMgr)
	rr := requestFixture(h.Tasks, "GET", "/tasks", nil)
	if rr.Code == http.StatusOK {
		t.Fatalf("Wrong status code: want %d but got %d", rr.Code, http.StatusOK)
	}
}

// Validates the endpoint to update a task.
func TestHandlers_Update(t *testing.T) {
	h := NewHandlers(apiMgr)
	rr := requestFixture(h.Update, "PUT", "/update", strings.NewReader(validJSON))
	if rr.Code != http.StatusOK {
		t.Fatalf("Wrong status code: want %d but got %d", http.StatusOK, rr.Code)
	}
}

// Makes sure our endpoint to update a task gives an error when it should.
func TestHandlers_UpdateError(t *testing.T) {
	h := NewHandlers(brokenApiMgr)
	rr := requestFixture(h.Update, "PUT", "/update", strings.NewReader(junkJSON))
	if rr.Code == http.StatusOK {
		t.Fatalf("Wrong status code: want %d but got %d", rr.Code, http.StatusOK)
	}
}