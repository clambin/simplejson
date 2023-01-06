package simplejson

import (
	"github.com/clambin/go-common/httpserver"
	"github.com/clambin/go-common/httpserver/middleware"
)

// Option specified configuration options for Server
type Option interface {
	apply(server *Server)
}

// WithQueryMetrics will collect the specified metrics to instrument the Server's Handlers.
type WithQueryMetrics struct {
	Name string
}

func (o WithQueryMetrics) apply(s *Server) {
	if o.Name == "" {
		o.Name = "simplejson"
	}
	s.queryMetrics = newQueryMetrics(o.Name)
}

// WithHTTPServerOption will pass the provided option to the underlying HTTP Server
type WithHTTPServerOption struct {
	Option httpserver.Option
}

func (o WithHTTPServerOption) apply(s *Server) {
	s.httpServerOptions = append(s.httpServerOptions, o.Option)
}

// WithHTTPMetrics will configure the http router to gather statistics on SimpleJson endpoint calls and record them as Prometheus metrics
type WithHTTPMetrics struct {
	Option middleware.PrometheusMetricsOptions
}

func (o WithHTTPMetrics) apply(s *Server) {
	s.prometheusMetrics = middleware.NewPrometheusMetrics(o.Option)
}
