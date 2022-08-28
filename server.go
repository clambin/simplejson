package simplejson

import (
	"context"
	"fmt"
	"github.com/clambin/go-metrics/server"
	"github.com/gorilla/mux"
	"net/http"
	"sort"
	"time"
)

// Server receives SimpleJSON requests from Grafana and dispatches them to the handler that serves the specified target.
type Server struct {
	Name       string
	Handlers   map[string]Handler
	httpServer *http.Server
}

// Run starts the SimpleJSon Server.
func (s *Server) Run(port int) error {
	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: s.GetRouter(),
	}
	return s.httpServer.ListenAndServe()
}

// Shutdown stops a running Server.
func (s *Server) Shutdown(ctx context.Context, timeout time.Duration) (err error) {
	if s.httpServer != nil {
		newCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		err = s.httpServer.Shutdown(newCtx)
	}
	return
}

// GetRouter sets up an HTTP router. Useful if you want to hook other handlers to the HTTP Server.
func (s *Server) GetRouter() (r *mux.Router) {
	r = server.GetRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.HandleFunc("/search", s.search).Methods(http.MethodPost)
	r.HandleFunc("/query", s.query).Methods(http.MethodPost)
	r.HandleFunc("/annotations", s.annotations).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/tag-keys", s.tagKeys).Methods(http.MethodPost)
	r.HandleFunc("/tag-values", s.tagValues).Methods(http.MethodPost)
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
