package simplejson

import (
	"github.com/clambin/go-common/httpserver"
	"github.com/clambin/go-common/httpserver/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"time"
)

// Server receives SimpleJSON requests from Grafana and dispatches them to the handler that serves the specified target.
type Server struct {
	chi.Router
	Handlers          map[string]Handler
	prometheusMetrics *middleware.PrometheusMetrics
	queryMetrics      *QueryMetrics
	httpServerOptions []httpserver.Option
	httpServer        *httpserver.Server
}

var _ prometheus.Collector = &Server{}
var _ http.Handler = &Server{}

func New(handlers map[string]Handler, options ...Option) (*Server, error) {
	s := Server{Handlers: handlers}
	for _, o := range options {
		o.apply(&s)
	}

	s.httpServerOptions = append(s.httpServerOptions, httpserver.WithHandlers{
		Handlers: []httpserver.Handler{
			{Path: "/", Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })},
			{Path: "/search", Handler: http.HandlerFunc(s.Search), Methods: []string{http.MethodPost}},
			{Path: "/query", Handler: http.HandlerFunc(s.Query), Methods: []string{http.MethodPost}},
			{Path: "/annotations", Handler: http.HandlerFunc(s.Annotations), Methods: []string{http.MethodPost, http.MethodOptions}},
			{Path: "/tag-keys", Handler: http.HandlerFunc(s.TagKeys), Methods: []string{http.MethodPost}},
			{Path: "/tag-values", Handler: http.HandlerFunc(s.TagValues), Methods: []string{http.MethodPost}},
		},
	})

	var err error
	s.httpServer, err = httpserver.New(s.httpServerOptions...)

	return &s, err
}

// Serve starts the SimpleJSon Server.
func (s *Server) Serve() error {
	return s.httpServer.Serve()
}

// Shutdown stops a running Server.
func (s *Server) Shutdown(timeout time.Duration) error {
	return s.httpServer.Shutdown(timeout)
}

// Describe implements the prometheus.Collector interface
func (s *Server) Describe(descs chan<- *prometheus.Desc) {
	if s.httpServer != nil {
		s.httpServer.Describe(descs)
	}
	if s.prometheusMetrics != nil {
		s.prometheusMetrics.Describe(descs)
	}
	if s.queryMetrics != nil {
		s.queryMetrics.Describe(descs)
	}
}

// Collect implements the prometheus.Collector interface
func (s *Server) Collect(metrics chan<- prometheus.Metric) {
	if s.httpServer != nil {
		s.httpServer.Collect(metrics)
	}
	if s.prometheusMetrics != nil {
		s.prometheusMetrics.Collect(metrics)
	}
	if s.queryMetrics != nil {
		s.queryMetrics.Collect(metrics)
	}
}
