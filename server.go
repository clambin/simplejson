package grafana_json

import (
	"context"
	"fmt"
	"github.com/clambin/go-metrics"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

// Server implements a generic frameworks for the Grafana simpleJson API datasource
type Server struct {
	Name       string
	Handlers   []Handler
	httpServer *http.Server
}

// Handler implements the business logic of the Grafana API datasource so that
// Server can be limited to providing the generic search/query framework
type Handler interface {
	Endpoints() Endpoints
}

type QueryFunc func(ctx context.Context, target string, args *TimeSeriesQueryArgs) (*QueryResponse, error)
type TableQueryFunc func(ctx context.Context, target string, args *TableQueryArgs) (*TableQueryResponse, error)
type TagKeysFunc func(ctx context.Context) []string
type TagValuesFunc func(ctx context.Context, key string) ([]string, error)

// Endpoints contains the functions that implements each of the SimpleJson endpoints
type Endpoints struct {
	// Search implements the /search endpoint: it returns the list of supported targets
	Search func() []string
	// Query implements the /query endpoint for dataSeries targets
	Query QueryFunc
	// TableQuery implements the /query endpoint for table targets
	TableQuery TableQueryFunc
	// Annotations implements the /annotations endpoint
	Annotations func(name, query string, args *AnnotationRequestArgs) ([]Annotation, error)
	// TagKeys implements the /tag-keys endpoint
	TagKeys TagKeysFunc
	// TagValues implements the /tag-values endpoint
	TagValues TagValuesFunc
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
