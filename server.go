package grafana_json

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

// Server implements a generic frameworks for the Grafana simpleJson API datasource
type Server struct {
	handler Handler
	port    int
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

// Create creates a Server object for the specified Handler.
// Panics if handler is nil
func Create(handler Handler) *Server {
	if handler == nil {
		panic("no handler provided")
	}
	return &Server{handler: handler}
}

// Run the API Server. Convenience function. This is the same as:
//   s := Create(handler)
//   http.ListenAndServe(":8080", s.GetRouter())
func (server *Server) Run(port int) error {
	return http.ListenAndServe(fmt.Sprintf(":%d", port), server.GetRouter())
}

// GetRouter sets up an HTTP router.  Useful if you want to hook other handlers to the HTTP Server
func (server *Server) GetRouter() (r *mux.Router) {
	r = mux.NewRouter()
	r.Use(prometheusMiddleware)
	r.Path("/metrics").Handler(promhttp.Handler())
	r.HandleFunc("/", server.hello)
	if server.handler.Endpoints().Search != nil {
		r.HandleFunc("/search", server.search).Methods(http.MethodPost)
	}
	if server.handler.Endpoints().Query != nil || server.handler.Endpoints().TableQuery != nil {
		r.HandleFunc("/query", server.query).Methods(http.MethodPost)
	}
	if server.handler.Endpoints().Annotations != nil {
		r.HandleFunc("/annotations", server.annotations).Methods(http.MethodPost, http.MethodOptions)
	}
	return
}

// Prometheus metrics
var (
	httpDuration = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name: "grafana_api_duration_seconds",
		Help: "Grafana API duration of HTTP requests.",
	}, []string{"path"})
	queryDuration = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name: "grafana_api_query_duration_seconds",
		Help: "Grafana API duration of query requests by target",
	}, []string{"type", "target"})
)

func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()
		timer := prometheus.NewTimer(httpDuration.WithLabelValues(path))
		next.ServeHTTP(w, r)
		timer.ObserveDuration()
	})
}
