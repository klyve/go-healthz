package healthz_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/klyve/healthz"
)

var instance healthz.Instance
var server healthz.Server

const errMsg = "provider_failed"

func providerError(name string) string {
	return fmt.Sprintf("%s-%s", name, errMsg)
}

type testLog struct {
	t *testing.T
}

func (t testLog) Info(args ...interface{}) {
	t.t.Log("Info", args)
}
func (t testLog) Error(args ...interface{}) {
	t.t.Log("Error", args)
}
func (t testLog) Fatal(args ...interface{}) {
	t.t.Log("Fatal", args)
}

func init() {
	instance = healthz.Instance{}

	logger := log.New(os.Stdout, "", 0)
	instance = healthz.Instance{
		Providers: []healthz.Provider{},
		Detailed:  true,
	}

	server = healthz.Server{
		ListenAddr:   ":3000",
		Instance:     &instance,
		ServerLogger: logger,
	}
}

func TestNilInstance(t *testing.T) {
	server.Instance = nil
	_, err := server.Handle()
	expectError(t, err)
}

func TestNoLogger(t *testing.T) {
	instance.Detailed = false
	server.Instance = &instance
	instance.Providers = []healthz.Provider{}

	sh, err := server.Handle()
	expectNil(t, err)

	req, err := http.NewRequest("GET", "/liveness", nil)
	expectNil(t, err)

	w := httptest.NewRecorder()
	sh.Handler.ServeHTTP(w, req)

	expectStatusCode(t, http.StatusOK, w)
}
func TestNoServerLogger(t *testing.T) {
	s := healthz.Server{
		ListenAddr:   ":3000",
		Instance:     &instance,
		ServerLogger: nil,
	}
	s.Instance = &instance

	sh, err := s.Handle()
	expectNil(t, err)

	req, err := http.NewRequest("GET", "/liveness", nil)
	expectNil(t, err)

	w := httptest.NewRecorder()
	sh.Handler.ServeHTTP(w, req)

	expectStatusCode(t, http.StatusOK, w)
}
func TestLiveness(t *testing.T) {
	instance.Detailed = false
	server.Instance = &instance
	instance.Providers = []healthz.Provider{}
	instance.Logger = testLog{t: t}

	sh, err := server.Handle()
	expectNil(t, err)

	req, err := http.NewRequest("GET", "/liveness", nil)
	expectNil(t, err)

	w := httptest.NewRecorder()
	sh.Handler.ServeHTTP(w, req)

	expectStatusCode(t, http.StatusOK, w)
}
func TestNoProviders(t *testing.T) {
	instance.Detailed = false
	server.Instance = &instance
	instance.Providers = nil
	instance.Logger = testLog{t: t}

	sh, err := server.Handle()
	expectNil(t, err)

	req, err := http.NewRequest("GET", "/healthz", nil)
	expectNil(t, err)

	w := httptest.NewRecorder()
	sh.Handler.ServeHTTP(w, req)

	expectStatusCode(t, http.StatusOK, w)
	testBasicReply(t, w.Body.Bytes(), true)
}

func TestHealthySimple(t *testing.T) {
	instance.Detailed = false
	server.Instance = &instance
	instance.Logger = testLog{t: t}
	instance.Providers = []healthz.Provider{
		healthProvider{
			Name:    "Test2",
			Healthy: true,
		}.provider(),
	}

	sh, err := server.Handle()
	expectNil(t, err)

	req, err := http.NewRequest("GET", "/healthz", nil)
	expectNil(t, err)

	w := httptest.NewRecorder()
	sh.Handler.ServeHTTP(w, req)

	expectStatusCode(t, http.StatusOK, w)
	testBasicReply(t, w.Body.Bytes(), true)
}

func TestFailingSimple(t *testing.T) {
	instance.Detailed = false
	server.Instance = &instance
	instance.Logger = testLog{t: t}
	instance.Providers = []healthz.Provider{
		healthProvider{
			Name:    "Test2",
			Healthy: false,
		}.provider(),
	}

	sh, err := server.Handle()
	expectNil(t, err)

	req, err := http.NewRequest("GET", "/healthz", nil)
	expectNil(t, err)

	w := httptest.NewRecorder()
	sh.Handler.ServeHTTP(w, req)
	expectStatusCode(t, http.StatusServiceUnavailable, w)
	testBasicReply(t, w.Body.Bytes(), false)
}
func TestCustomErrorCode(t *testing.T) {
	instance.Detailed = false
	server.Instance = &instance
	instance.Logger = testLog{t: t}
	instance.FailCode = http.StatusOK
	instance.Providers = []healthz.Provider{
		healthProvider{
			Name:    "Test2",
			Healthy: false,
		}.provider(),
	}

	sh, err := server.Handle()
	expectNil(t, err)

	req, err := http.NewRequest("GET", "/healthz", nil)
	expectNil(t, err)

	w := httptest.NewRecorder()
	sh.Handler.ServeHTTP(w, req)
	expectStatusCode(t, http.StatusOK, w)
	testBasicReply(t, w.Body.Bytes(), false)
	instance.FailCode = 0
}

func TestManyHealthySimple(t *testing.T) {
	instance.Detailed = false
	server.Instance = &instance
	instance.Logger = testLog{t: t}
	instance.Providers = []healthz.Provider{
		healthProvider{
			Name:    "test1",
			Healthy: true,
		}.provider(),
		healthProvider{
			Name:    "Test2",
			Healthy: true,
		}.provider(),
		healthProvider{
			Name:    "Test3",
			Healthy: true,
		}.provider(),
		healthProvider{
			Name:    "Test4",
			Healthy: true,
		}.provider(),
	}

	sh, err := server.Handle()
	expectNil(t, err)

	req, err := http.NewRequest("GET", "/healthz", nil)
	expectNil(t, err)

	w := httptest.NewRecorder()
	sh.Handler.ServeHTTP(w, req)

	expectStatusCode(t, http.StatusOK, w)
	testBasicReply(t, w.Body.Bytes(), true)
}

func TestManyHealthyDetailed(t *testing.T) {
	instance.Detailed = true
	server.Instance = &instance
	instance.Logger = testLog{t: t}
	instance.Providers = []healthz.Provider{
		healthProvider{
			Name:    "test1",
			Healthy: true,
		}.provider(),
		healthProvider{
			Name:    "Test2",
			Healthy: true,
		}.provider(),
		healthProvider{
			Name:    "Test3",
			Healthy: true,
		}.provider(),
		healthProvider{
			Name:    "Test4",
			Healthy: true,
		}.provider(),
	}

	sh, err := server.Handle()
	expectNil(t, err)

	req, err := http.NewRequest("GET", "/healthz", nil)
	expectNil(t, err)

	w := httptest.NewRecorder()
	sh.Handler.ServeHTTP(w, req)

	expectStatusCode(t, http.StatusOK, w)
	testDetailedReply(t, w.Body.Bytes(), true)
}

func TestHealthyDetailed(t *testing.T) {
	instance.Detailed = true
	server.Instance = &instance
	instance.Logger = testLog{t: t}
	instance.Providers = []healthz.Provider{
		healthProvider{
			Name:    "Test2",
			Healthy: true,
		}.provider(),
	}

	sh, err := server.Handle()
	expectNil(t, err)

	req, err := http.NewRequest("GET", "/healthz", nil)
	expectNil(t, err)

	w := httptest.NewRecorder()
	sh.Handler.ServeHTTP(w, req)

	expectStatusCode(t, http.StatusOK, w)
	testDetailedReply(t, w.Body.Bytes(), true)
}

func TestFailingDetailed(t *testing.T) {
	instance.Detailed = true
	server.Instance = &instance
	instance.Logger = testLog{t: t}
	instance.Providers = []healthz.Provider{
		healthProvider{
			Name:    "Test2",
			Healthy: false,
		}.provider(),
	}

	sh, err := server.Handle()
	expectNil(t, err)

	req, err := http.NewRequest("GET", "/healthz", nil)
	expectNil(t, err)

	w := httptest.NewRecorder()
	sh.Handler.ServeHTTP(w, req)

	expectStatusCode(t, http.StatusServiceUnavailable, w)
	testDetailedReply(t, w.Body.Bytes(), false)
}

func TestMixedSimple(t *testing.T) {
	instance.Detailed = false
	server.Instance = &instance
	instance.Logger = testLog{t: t}
	instance.Providers = []healthz.Provider{
		healthProvider{
			Name:    "test1",
			Healthy: true,
		}.provider(),
		healthProvider{
			Name:    "Test2",
			Healthy: false,
		}.provider(),
		healthProvider{
			Name:    "Test3",
			Healthy: true,
		}.provider(),
		healthProvider{
			Name:    "Test4",
			Healthy: false,
		}.provider(),
	}

	sh, err := server.Handle()
	expectNil(t, err)

	req, err := http.NewRequest("GET", "/healthz", nil)
	expectNil(t, err)

	w := httptest.NewRecorder()
	sh.Handler.ServeHTTP(w, req)
	expectStatusCode(t, http.StatusServiceUnavailable, w)
	testBasicReply(t, w.Body.Bytes(), false)
}

func TestMixedDetailed(t *testing.T) {
	instance.Detailed = true
	server.Instance = &instance
	instance.Logger = testLog{t: t}
	instance.Providers = []healthz.Provider{
		healthProvider{
			Name:    "test1",
			Healthy: true,
		}.provider(),
		healthProvider{
			Name:    "Test2",
			Healthy: false,
		}.provider(),
		healthProvider{
			Name:    "Test3",
			Healthy: true,
		}.provider(),
		healthProvider{
			Name:    "Test4",
			Healthy: false,
		}.provider(),
	}

	sh, err := server.Handle()
	expectNil(t, err)

	req, err := http.NewRequest("GET", "/healthz", nil)
	expectNil(t, err)

	w := httptest.NewRecorder()
	sh.Handler.ServeHTTP(w, req)
	expectStatusCode(t, http.StatusServiceUnavailable, w)
	testDetailedReply(t, w.Body.Bytes(), false)
}

func testBasicReply(t *testing.T, body []byte, healthy bool) {
	var data healthz.Response
	err := json.Unmarshal(body, &data)
	expectNil(t, err)
	if data.Healthy != healthy {
		t.Fatalf("Health status is invalid expected %t got %t", healthy, data.Healthy)
	}

	if len(data.Services) != 0 {
		t.Fatal("Returned services should be empty")
	}

	testErrors(t, len(data.Errors), healthy)
	testErrorResponse(t, data.Errors)
}

func testDetailedReply(t *testing.T, body []byte, healthy bool) {
	var data healthz.Response
	err := json.Unmarshal(body, &data)
	expectNil(t, err)
	if data.Healthy != healthy {
		t.Fatalf("Health status is invalid expected %t got %t", healthy, data.Healthy)
	}
	sLen := len(data.Services)
	pLen := len(instance.Providers)

	if sLen == 0 && pLen > 0 {
		t.Fatalf("No services returned expected %d got %d", sLen, pLen)
	}
	testErrors(t, len(data.Errors), healthy)

	for _, provider := range instance.Providers {
		s := findService(data.Services, provider.Name)
		if s == nil {
			t.Fatalf("Provider %s not found in response ", provider.Name)
		}

		// Make sure the reply is what the providers tell us
		if s.Healthy {
			expectNil(t, provider.Handle.Healthz())
		} else {
			expectError(t, provider.Handle.Healthz())
		}
	}
	testErrorResponse(t, data.Errors)
}

func testErrorResponse(t *testing.T, errors []healthz.Error) {
	for _, provider := range instance.Providers {
		err := findError(errors, provider.Name)
		// Make sure the reply is what the providers tell us
		if provider.Handle.Healthz() == nil {
			if err != nil {
				t.Fatal("Provider was healthy but found in errors")
			}
		} else {
			if err == nil {
				t.Fatal("Provider is unhealthy but was not in errors")
			}
			if err.Message != providerError(err.Name) {
				t.Fatal("Provider error is invalid")
			}
		}
	}
}

func testErrors(t *testing.T, len int, healthy bool) {
	if len != 0 {
		if healthy {
			t.Fatal("Errors returned should be empty")
		}
	} else {
		if !healthy {
			t.Fatal("Expected errors to be returned but it was empty")
		}
	}
}

type healthProvider struct {
	Name    string
	Healthy bool
}

func (h healthProvider) Healthz() error {
	if h.Healthy {
		return nil
	}
	return errors.New(providerError(h.Name))
}

func (h healthProvider) provider() healthz.Provider {
	return healthz.Provider{
		Handle: h,
		Name:   h.Name,
	}
}

func expectStatusCode(t *testing.T, code int, w *httptest.ResponseRecorder) {
	if code != w.Code {
		t.Fatalf("Expected status code %d found %d", code, w.Code)
	}
}

func expectNil(t *testing.T, err error) {
	if err != nil {
		t.Fatal("Expected error to be nil ", err)
	}
}
func expectError(t *testing.T, err error) {
	if err == nil {
		t.Fatal("Expected error got nil")
	}
}

func findError(a []healthz.Error, name string) *healthz.Error {
	for _, n := range a {
		if name == n.Name {
			return &n
		}
	}
	return nil
}

func findService(a []healthz.Service, name string) *healthz.Service {
	for _, n := range a {
		if name == n.Name {
			return &n
		}
	}
	return nil
}
