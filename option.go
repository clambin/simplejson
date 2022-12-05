package simplejson

import (
	"github.com/clambin/httpserver"
	"github.com/prometheus/client_golang/prometheus"
)

// Option specified configuration options for Server
type Option interface {
	apply(server *Server)
}

type QueryMetrics struct {
	duration *prometheus.HistogramVec
	errors   *prometheus.CounterVec
}

func (qm QueryMetrics) Describe(ch chan<- *prometheus.Desc) {
	qm.duration.Describe(ch)
	qm.errors.Describe(ch)
}

func (qm QueryMetrics) Collect(ch chan<- prometheus.Metric) {
	qm.duration.Collect(ch)
	qm.errors.Collect(ch)
}

func NewQueryMetrics(name string) *QueryMetrics {
	qm := QueryMetrics{
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:        prometheus.BuildFQName("simplejson", "query", "duration_seconds"),
			Help:        "Grafana SimpleJSON server duration of query requests in seconds",
			ConstLabels: prometheus.Labels{"app": name},
			Buckets:     prometheus.DefBuckets,
		}, []string{"target", "type"}),
		errors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        prometheus.BuildFQName("simplejson", "query", "failed_count"),
			Help:        "Grafana SimpleJSON server count of failed requests",
			ConstLabels: prometheus.Labels{"app": name},
		}, []string{"target", "type"}),
	}
	return &qm
}

// WithQueryMetrics will collect the specified metrics to instrument the Server's Handlers.
type WithQueryMetrics struct {
	QueryMetrics *QueryMetrics
}

func (o WithQueryMetrics) apply(s *Server) {
	if o.QueryMetrics == nil {
		o.QueryMetrics = NewQueryMetrics("simplejson")
	}
	s.queryMetrics = o.QueryMetrics
}

// WithHTTPServerOption will pass the provided option to the underlying HTTP Server
type WithHTTPServerOption struct {
	Option httpserver.Option
}

func (o WithHTTPServerOption) apply(s *Server) {
	s.httpServerOptions = append(s.httpServerOptions, o.Option)
}
