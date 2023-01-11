package simplejson

import (
	"github.com/clambin/go-common/httpserver/middleware"
	"github.com/go-chi/chi/v5"
	middleware2 "github.com/go-chi/chi/v5/middleware"
	"github.com/go-http-utils/headers"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/exp/slog"
	"net/http"
)

// Server receives SimpleJSON requests from Grafana and dispatches them to the handler that serves the specified target.
type Server struct {
	chi.Router
	Handlers          map[string]Handler
	prometheusMetrics *middleware.PrometheusMetrics
	queryMetrics      *QueryMetrics
	logger            *slog.Logger
}

var _ prometheus.Collector = &Server{}
var _ http.Handler = &Server{}

func New(handlers map[string]Handler, options ...Option) *Server {
	s := Server{
		Handlers: handlers,
		Router:   chi.NewRouter(),
		logger:   slog.Default(),
	}
	for _, o := range options {
		o.apply(&s)
	}

	s.Router.Use(middleware2.Heartbeat("/"))
	s.Router.Group(func(r chi.Router) {
		r.Use(middleware.Logger(s.logger))
		if s.prometheusMetrics != nil {
			r.Use(s.prometheusMetrics.Handle)
		}
		r.Post("/search", s.Search)
		r.Post("/query", s.Query)
		r.Post("/annotations", s.Annotations)
		r.Options("/annotations", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set(headers.AccessControlAllowOrigin, "*")
			w.Header().Set(headers.AccessControlAllowMethods, "POST")
			w.Header().Set(headers.AccessControlAllowHeaders, "accept, content-type")
		})
		r.Post("/tag-keys", s.TagValues)
		r.Post("/tag-values", s.TagValues)
	})

	return &s
}

// Describe implements the prometheus.Collector interface
func (s *Server) Describe(descs chan<- *prometheus.Desc) {
	if s.prometheusMetrics != nil {
		s.prometheusMetrics.Describe(descs)
	}
	if s.queryMetrics != nil {
		s.queryMetrics.Describe(descs)
	}
}

// Collect implements the prometheus.Collector interface
func (s *Server) Collect(metrics chan<- prometheus.Metric) {
	if s.prometheusMetrics != nil {
		s.prometheusMetrics.Collect(metrics)
	}
	if s.queryMetrics != nil {
		s.queryMetrics.Collect(metrics)
	}
}
