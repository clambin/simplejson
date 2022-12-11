package simplejson

import "github.com/prometheus/client_golang/prometheus"

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

func newQueryMetrics(name string) *QueryMetrics {
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
