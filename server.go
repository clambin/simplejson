package grafana_json

import (
	"context"
	"fmt"
	"github.com/clambin/go-metrics"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

// Server receives SimpleJSON requests from Grafana and dispatches them to the handler that serves the specified target.
// If multiple handlers serve the same target name, the first handler in the list will receive the request.
type Server struct {
	Name       string
	Handlers   []Handler
	httpServer *http.Server
}

// Handler implements the different Grafana SimpleJSON endpoints.  The interface only contains a single Endpoints()  function,
// so that a handler only has to implement the endpoint functions (query, tablequery, annotations, etc.) that it needs.
type Handler interface {
	Endpoints() Endpoints
}

// QueryFunc handles timeseries queries
type QueryFunc func(ctx context.Context, target string, args *TimeSeriesQueryArgs) (*QueryResponse, error)

// TableQueryFunc handles for table queries
type TableQueryFunc func(ctx context.Context, target string, args *TableQueryArgs) (*TableQueryResponse, error)

// TagKeysFunc returns supported tag names
type TagKeysFunc func(ctx context.Context) []string

// TagValuesFunc returns supported values for the specified tag name
type TagValuesFunc func(ctx context.Context, key string) ([]string, error)

// AnnotationsFunc handles requests for annotations
type AnnotationsFunc func(name, query string, args *AnnotationRequestArgs) ([]Annotation, error)

// Endpoints contains the functions that implement each of the SimpleJson endpoints
type Endpoints struct {
	Search      func() []string // /search endpoint: it returns the list of supported targets
	Query       QueryFunc       // /query endpoint: handles timeSeries queries
	TableQuery  TableQueryFunc  // /query endpoint: handles table queries
	Annotations AnnotationsFunc // /annotations endpoint: handles requests for annotations
	TagKeys     TagKeysFunc     // /tag-keys endpoint: returns all supported tag names
	TagValues   TagValuesFunc   // /tag-values endpoint: returns all supported values for the specified tag name
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
func (server *Server) Shutdown(ctx context.Context, timeout time.Duration) error {
	if server.httpServer == nil {
		return nil
	}
	newCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return server.httpServer.Shutdown(newCtx)
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
