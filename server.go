package simplejson

import (
	"context"
	"fmt"
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

// GetRouter sets up an HTTP router with the requested SimpleJSON endpoints
func (s *Server) GetRouter() (m *http.ServeMux) {
	m = http.NewServeMux()
	m.Handle("/", httpMetrics(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	m.Handle("/search", httpMetrics(methodFilter(s.Search, http.MethodPost)))
	m.Handle("/query", httpMetrics(methodFilter(s.Query, http.MethodPost)))
	m.Handle("/annotations", httpMetrics(methodFilter(s.Annotations, http.MethodPost, http.MethodOptions)))
	m.Handle("/tag-keys", httpMetrics(methodFilter(s.TagKeys, http.MethodPost)))
	m.Handle("/tag-values", httpMetrics(methodFilter(s.TagValues, http.MethodPost)))
	return
}

func methodFilter(next func(w http.ResponseWriter, r *http.Request), methods ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isValidMethod(r.Method, methods) {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		next(w, r)
	})
}

func isValidMethod(method string, methods []string) bool {
	for _, m := range methods {
		if m == method {
			return true
		}
	}
	return false
}

func httpMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lrw := &loggingResponseWriter{
			responseWriter: w,
			statusCode:     http.StatusOK, // if the handler doesn't call WriteHeader(), default to HTTP 200
		}
		start := time.Now()
		next.ServeHTTP(lrw, r)
		httpDuration.WithLabelValues(r.URL.Path, r.Method, strconv.Itoa(lrw.statusCode)).Observe(time.Since(start).Seconds())
	})
}

// HTTPMetrics metrics
var (
	httpDuration = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name: "simplejson_request_duration_seconds",
		Help: "Duration of HTTP requests",
	}, []string{"path", "method", "status_code"})
)

// loggingResponseWriter records the HTTP status code of a ResponseWriter, so we can use it to log response times for
// individual status codes.
type loggingResponseWriter struct {
	responseWriter http.ResponseWriter
	statusCode     int
}

// WriteHeader implements the http.ResponseWriter interface.
func (w *loggingResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.responseWriter.WriteHeader(code)
}

// Write implements the http.ResponseWriter interface.
func (w *loggingResponseWriter) Write(body []byte) (int, error) {
	return w.responseWriter.Write(body)
}

// Header implements the http.ResponseWriter interface
func (w *loggingResponseWriter) Header() http.Header {
	return w.responseWriter.Header()
}
