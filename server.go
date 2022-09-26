package simplejson

import (
	"context"
	"fmt"
	"github.com/clambin/go-metrics/server/middleware"
	"github.com/gorilla/mux"
	"net"
	"net/http"
	"sort"
	"sync"
	"time"
)

// Server receives SimpleJSON requests from Grafana and dispatches them to the handler that serves the specified target.
type Server struct {
	Name       string
	Handlers   map[string]Handler
	HTTPServer *http.Server
	lock       sync.RWMutex
}

// Run starts the SimpleJSon Server.
func (s *Server) Run(port int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	s.lock.Lock()
	s.HTTPServer = &http.Server{
		Handler: s.GetRouter(),
	}
	s.lock.Unlock()
	return s.HTTPServer.Serve(listener)
}

// Shutdown stops a running Server.
func (s *Server) Shutdown(ctx context.Context, timeout time.Duration) (err error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if s.HTTPServer != nil {
		newCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		err = s.HTTPServer.Shutdown(newCtx)
	}
	return
}

// Running returns true if the HTTP Server is running
func (s *Server) Running() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.HTTPServer != nil
}

// GetRouter sets up an HTTP router with the requested SimpleJSON endpoints
func (s *Server) GetRouter() (r *mux.Router) {
	r = mux.NewRouter()
	r.Use(middleware.HTTPMetrics)
	r.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.HandleFunc("/search", s.Search).Methods(http.MethodPost)
	r.HandleFunc("/query", s.Query).Methods(http.MethodPost)
	r.HandleFunc("/annotations", s.Annotations).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/tag-keys", s.TagKeys).Methods(http.MethodPost)
	r.HandleFunc("/tag-values", s.TagValues).Methods(http.MethodPost)
	return
}

// Targets returns a sorted list of supported targets
func (s *Server) Targets() (targets []string) {
	for target := range s.Handlers {
		targets = append(targets, target)
	}
	sort.Strings(targets)
	return
}
