package healthz

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// Server holds the data for the HTTPServer struct
type Server struct {
	ListenAddr   string
	Instance     *Instance
	ServerLogger *log.Logger
}

var done = make(chan bool)
var quit = make(chan os.Signal, 1)

// Handle returns a http handler
func (h *Server) Handle() (*http.Server, error) {
	if h.Instance == nil {
		return nil, errors.New("No healthz configuration passed to the server")
	}
	mux := http.NewServeMux()

	// Add the webserver to the list of healthz providers?
	mux.Handle("/healthz", h.Instance.Healthz())
	mux.Handle("/liveness", h.Instance.Liveness())

	server := &http.Server{
		Addr:         h.ListenAddr,
		Handler:      mux,
		ErrorLog:     h.ServerLogger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	return server, nil
}

// Start starts the HTTP server
func (h *Server) Start() (chan bool, error) {
	server, err := h.Handle()
	if err != nil {
		close(done)
		return nil, err
	}
	signal.Notify(quit, os.Interrupt)

	go h.gracefulShutdown(server)
	h.Instance.Logger.Info("[Healthz-Server]: listening on ", h.ListenAddr)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		h.Instance.Logger.Fatal("[Healthz-Server]: Could not listen on ", h.ListenAddr, err)
	}

	return done, nil
}

func (h *Server) gracefulShutdown(server *http.Server) {
	<-quit
	h.Instance.Logger.Info("[Healthz-Server]: shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	server.SetKeepAlivesEnabled(false)
	if err := server.Shutdown(ctx); err != nil {
		h.Instance.Logger.Fatal("[Healthz-Server]: Could not gracefully shutdown the server: %v\n", err)
	}
	close(done)
	h.Instance.Logger.Info("[Healthz-Server]: Shut down")
}
