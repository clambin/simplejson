package simplejson

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/clambin/simplejson/v3/query"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"net/http"
)

var queryDuration = promauto.NewSummaryVec(prometheus.SummaryOpts{
	Name: prometheus.BuildFQName("simplejson", "query", "duration_seconds"),
	Help: "Grafana SimpleJSON server duration of query requests by target",
}, []string{"app", "type", "target"})

var queryFailure = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: prometheus.BuildFQName("simplejson", "query", "failed_count"),
	Help: "Grafana SimpleJSON server count of failed requests",
}, []string{"app", "type", "target"})

func (s *Server) query(w http.ResponseWriter, req *http.Request) {
	var request query.Request
	handleEndpoint(w, req, &request, func() ([]json.Marshaler, error) {
		return s.handleQuery(req.Context(), request)
	})
}

func (s *Server) handleQuery(ctx context.Context, request query.Request) (responses []json.Marshaler, err error) {
	responses = make([]json.Marshaler, 0, len(request.Targets))
	for _, target := range request.Targets {
		timer := prometheus.NewTimer(queryDuration.WithLabelValues(s.Name, target.Type, target.Name))

		var response query.Response
		response, err = s.handleQueryRequest(ctx, target, request)
		timer.ObserveDuration()
		if err == nil {
			responses = append(responses, response)
		} else {
			queryFailure.WithLabelValues(s.Name, target.Type, target.Name).Add(1.0)
			break
		}
	}
	return responses, err
}

func (s *Server) handleQueryRequest(ctx context.Context, target query.Target, request query.Request) (query.Response, error) {
	handler, ok := s.Handlers[target.Name]

	if ok == false {
		return nil, fmt.Errorf("no handler found for target '%s'", target)
	}

	q := handler.Endpoints().Query

	if q == nil {
		return nil, fmt.Errorf("timeseries query not implemented for target '%s'", target)
	}

	return q(ctx, request)
}
