package simplejson

import (
	"context"
	"fmt"
	middleware2 "github.com/clambin/go-metrics/server/middleware"
	"github.com/clambin/simplejson/v3/annotation"
	"github.com/clambin/simplejson/v3/pkg/middleware"
	"github.com/clambin/simplejson/v3/query"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"net"
	"net/http"
	"sort"
	"strconv"
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

// Targets returns a sorted list of supported targets
func (s *Server) Targets() (targets []string) {
	for target := range s.Handlers {
		targets = append(targets, target)
	}
	sort.Strings(targets)
	return
}

// GetRouter sets up an HTTP router for the SimpleJSON endpoints
func (s *Server) GetRouter() (m *http.ServeMux) {
	m = http.NewServeMux()
	m.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	m.Handle("/search", middlewareChain(s.Search, http.MethodPost))
	m.Handle("/query", middlewareChain(s.Query, http.MethodPost))
	m.Handle("/annotations", middlewareChain(s.Annotations, http.MethodPost, http.MethodOptions))
	m.Handle("/tag-keys", middlewareChain(s.TagKeys, http.MethodPost))
	m.Handle("/tag-values", middlewareChain(s.TagValues, http.MethodPost))
	return
}

// HTTPMetrics metrics
var (
	httpDuration = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name: "simplejson_request_duration_seconds",
		Help: "Duration of HTTP requests",
	}, []string{"path", "method", "status_code"})
)

func middlewareChain(next http.HandlerFunc, methods ...string) http.Handler {
	return middleware2.HTTPMetricsWithRecorder(
		middleware.HandleForMethods(next, methods...),
		func(path, method string, statusCode int, duration time.Duration) {
			httpDuration.WithLabelValues(path, method, strconv.Itoa(statusCode)).Observe(duration.Seconds())
		})
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
type QueryFunc func(ctx context.Context, req query.Request) (query.Response, error)

// AnnotationsFunc handles requests for annotation
type AnnotationsFunc func(req annotation.Request) ([]annotation.Annotation, error)

// TagKeysFunc returns supported tag names
type TagKeysFunc func(ctx context.Context) []string

// TagValuesFunc returns supported values for the specified tag name
type TagValuesFunc func(ctx context.Context, key string) ([]string, error)
