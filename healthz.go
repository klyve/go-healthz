package healthz

import (
	"encoding/json"
	"net/http"
)

type logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
}

// Checkable Makes sure the object has the Healthz() function
type Checkable interface {
	Healthz() error
}

// Provider is a provder we can check for healthz
type Provider struct {
	Handle Checkable
	Name   string
}

// Error the structure of the Error object
type Error struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

// Service struct reprecenting a healthz provider and it's status
type Service struct {
	Name    string `json:"name"`
	Healthy bool   `json:"healthy"`
}

// Response type, we return a json object with {healthy:bool, errors:[]}
type Response struct {
	Services []Service `json:"services,omitempty"`
	Errors   []Error   `json:"errors,omitempty"`
	Healthy  bool      `json:"healthy"`
}

// Instance contains the healthz instance
type Instance struct {
	Providers []Provider
	Logger    logger
	Detailed  bool
	FailCode  int
}

type noLog struct{}

func (t noLog) Info(args ...interface{})  {}
func (t noLog) Error(args ...interface{}) {}
func (t noLog) Fatal(args ...interface{}) {}

// Healthz returns a http.HandlerFunc for the healthz service
func (h *Instance) Healthz() http.HandlerFunc {
	if h.Logger == nil {
		h.Logger = noLog{}
	}
	h.Logger.Info("[Healthz] health service started")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var errs []Error
		var srvs []Service

		// Let's check if we have any providers
		// If we don't we should just return 200 OK
		// As long as the web server is running we will assume it's all good
		if h.Providers != nil {
			for _, provider := range h.Providers {
				service := Service{
					Name:    provider.Name,
					Healthy: true,
				}
				if err := provider.Handle.Healthz(); err != nil {
					errs = append(errs, Error{
						Name:    provider.Name,
						Message: err.Error(),
					})
					service.Healthy = false
				}
				srvs = append(srvs, service)
			}
		}

		response := Response{
			Errors:  errs,
			Healthy: true,
		}

		// Detailed will add the services to the JSON Object
		if h.Detailed {
			response.Services = srvs
		}

		if len(errs) > 0 {
			response.Healthy = false
			if h.FailCode != 0 {
				w.WriteHeader(h.FailCode)
			} else {
				w.WriteHeader(http.StatusServiceUnavailable)
			}
		} else {
			w.WriteHeader(http.StatusOK)
		}

		json, err := json.Marshal(response)
		if err != nil {
			h.Logger.Error("Unable to marshal healthz errors:", err)
		}

		w.Write(json)
	})
}

// Liveness returns a http.HandlerFunc for the liveness probe
func (h *Instance) Liveness() http.HandlerFunc {
	h.Logger.Info("[Healthz] Liveness service started")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}
