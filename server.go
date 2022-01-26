package simplejson

import (
	"context"
	"fmt"
	"github.com/clambin/go-metrics"
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
func (server *Server) Run(port int) error {
	server.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: server.GetRouter(),
	}
	return server.httpServer.ListenAndServe()
}

// Shutdown stops a running Server.
func (server *Server) Shutdown(ctx context.Context, timeout time.Duration) (err error) {
	if server.httpServer != nil {
		newCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		err = server.httpServer.Shutdown(newCtx)
	}
	return
}

// GetRouter sets up an HTTP router. Useful if you want to hook other handlers to the HTTP Server.
func (server *Server) GetRouter() (r *mux.Router) {
	r = metrics.GetRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.HandleFunc("/search", server.search).Methods(http.MethodPost)
	r.HandleFunc("/query", server.query).Methods(http.MethodPost)
	r.HandleFunc("/annotations", server.annotations).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/tag-keys", server.tagKeys).Methods(http.MethodPost)
	r.HandleFunc("/tag-values", server.tagValues).Methods(http.MethodPost)
	return
}

// Targets returns a sorted list of supported targets
func (server Server) Targets() (targets []string) {
	for target := range server.Handlers {
		targets = append(targets, target)
	}
	sort.Strings(targets)
	return
}
