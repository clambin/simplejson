package grafana_json

import (
	"context"
	"fmt"
	"github.com/clambin/gotools/metrics"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

// Server implements a generic frameworks for the Grafana simpleJson API datasource
type Server struct {
	Handlers   []Handler
	httpServer *http.Server
}

// Handler implements the business logic of the Grafana API datasource so that
// Server can be limited to providing the generic search/query framework
type Handler interface {
	Endpoints() Endpoints
}

// Endpoints contains the functions that implements each of the SimpleJson endpoints
type Endpoints struct {
	// Search implements the /search endpoint: it returns the list of supported targets
	Search func() []string
	// Query implements the /query endpoint for dataSeries targets
	Query func(ctx context.Context, target string, args *TimeSeriesQueryArgs) (*QueryResponse, error)
	// TableQuery implements the /query endpoint for table targets
	TableQuery func(ctx context.Context, target string, args *TableQueryArgs) (*TableQueryResponse, error)
	// Annotations implements the /annotations endpoint
	Annotations func(name, query string, args *AnnotationRequestArgs) ([]Annotation, error)
}

// Run the API Server. Convenience function.
func (server *Server) Run(port int) error {
	server.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: server.GetRouter(),
	}
	return server.httpServer.ListenAndServe()
}

// Shutdown a running API server
func (server *Server) Shutdown(ctx context.Context, timeout time.Duration) error {
	if server.httpServer == nil {
		return nil
	}
	newCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return server.httpServer.Shutdown(newCtx)
}

// GetRouter sets up an HTTP router.  Useful if you want to hook other handlers to the HTTP Server
func (server *Server) GetRouter() (r *mux.Router) {
	r = metrics.GetRouter()
	r.HandleFunc("/", server.hello)
	if server.hasSearch() {
		r.HandleFunc("/search", server.search).Methods(http.MethodPost)
	}
	if server.hasQuery() {
		r.HandleFunc("/query", server.query).Methods(http.MethodPost)
	}
	if server.hasAnnotations() {
		r.HandleFunc("/annotations", server.annotations).Methods(http.MethodPost, http.MethodOptions)
	}
	return
}

func (server *Server) hasSearch() bool {
	for _, h := range server.Handlers {
		if h.Endpoints().Search != nil {
			return true
		}
	}
	return false
}

func (server *Server) hasQuery() bool {
	for _, h := range server.Handlers {
		if h.Endpoints().Query != nil || h.Endpoints().TableQuery != nil {
			return true
		}
	}
	return false
}

func (server *Server) hasAnnotations() bool {
	for _, h := range server.Handlers {
		if h.Endpoints().Annotations != nil {
			return true
		}
	}
	return false
}
