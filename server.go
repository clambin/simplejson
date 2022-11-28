package simplejson

import (
	"context"
	"github.com/clambin/httpserver"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"sort"
	"time"
)

// Server receives SimpleJSON requests from Grafana and dispatches them to the handler that serves the specified target.
type Server struct {
	Handlers     map[string]Handler
	queryMetrics QueryMetrics
	httpServer   *httpserver.Server
}

func New(name string, handlers map[string]Handler, options ...httpserver.Option) (s *Server, err error) {
	return NewWithRegisterer(name, handlers, prometheus.DefaultRegisterer, options...)
}

func NewWithRegisterer(name string, handlers map[string]Handler, r prometheus.Registerer, options ...httpserver.Option) (s *Server, err error) {
	s = &Server{
		Handlers:     handlers,
		queryMetrics: NewQueryMetrics(name),
	}
	s.queryMetrics.Register(r)

	options = append(options,
		httpserver.WithHandlers{Handlers: []httpserver.Handler{
			{Path: "/", Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })},
			{Path: "/search", Handler: http.HandlerFunc(s.Search), Methods: []string{http.MethodPost}},
			{Path: "/query", Handler: http.HandlerFunc(s.Query), Methods: []string{http.MethodPost}},
			{Path: "/annotations", Handler: http.HandlerFunc(s.Annotations), Methods: []string{http.MethodPost, http.MethodOptions}},
			{Path: "/tag-keys", Handler: http.HandlerFunc(s.TagKeys), Methods: []string{http.MethodPost}},
			{Path: "/tag-values", Handler: http.HandlerFunc(s.TagValues), Methods: []string{http.MethodPost}},
		}})

	s.httpServer, err = httpserver.New(options...)

	return s, err
}

// Run starts the SimpleJSon Server.
func (s *Server) Run() error {
	return s.httpServer.Run()
}

// Shutdown stops a running Server.
func (s *Server) Shutdown(timeout time.Duration) (err error) {
	return s.httpServer.Shutdown(timeout)
}

// Targets returns a sorted list of supported targets
func (s *Server) Targets() (targets []string) {
	for target := range s.Handlers {
		targets = append(targets, target)
	}
	sort.Strings(targets)
	return
}

// Handler implements the different Grafana SimpleJSON endpoints.  The interface only contains a single Endpoints() function,
// so that a handler only has to implement the endpoint functions (query, annotation, etc.) that it needs.
type Handler interface {
	Endpoints() Endpoints
}

// Endpoints contains the functions that implement each of the SimpleJson endpoints
type Endpoints struct {
	Query       QueryFunc       // /query endpoint: handles queries
	Annotations AnnotationsFunc // /annotation endpoint: handles requests for annotation
	TagKeys     TagKeysFunc     // /tag-keys endpoint: returns all supported tag names
	TagValues   TagValuesFunc   // /tag-values endpoint: returns all supported values for the specified tag name
}

// QueryFunc handles queries
type QueryFunc func(ctx context.Context, req QueryRequest) (Response, error)

// AnnotationsFunc handles requests for annotation
type AnnotationsFunc func(req AnnotationRequest) ([]Annotation, error)

// TagKeysFunc returns supported tag names
type TagKeysFunc func(ctx context.Context) []string

// TagValuesFunc returns supported values for the specified tag name
type TagValuesFunc func(ctx context.Context, key string) ([]string, error)
