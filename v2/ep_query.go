package simplejson

import (
	"context"
	"fmt"
	"github.com/clambin/simplejson/v2/query"
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

func (server *Server) query(w http.ResponseWriter, req *http.Request) {
	var request query.Request
	handleEndpoint(w, req, &request, func() (interface{}, error) {
		return server.handleQuery(req.Context(), &request)
	})
}

func (server *Server) handleQuery(ctx context.Context, request *query.Request) (responses []interface{}, err error) {
	responses = make([]interface{}, 0, len(request.Targets))
	for _, target := range request.Targets {
		timer := prometheus.NewTimer(queryDuration.WithLabelValues(server.Name, target.Type, target.Target))
		switch target.Type {
		case "timeserie", "":
			var response *query.TimeSeriesResponse
			if response, err = server.handleQueryRequest(ctx, target.Target, request); err == nil {
				responses = append(responses, response)
			}
		case "table":
			var response *query.TableResponse
			if response, err = server.handleTableQueryRequest(ctx, target.Target, request); err == nil {
				responses = append(responses, response)
			}
		}
		timer.ObserveDuration()
		if err != nil {
			queryFailure.WithLabelValues(server.Name, target.Type, target.Target).Add(1.0)
			break
		}
	}
	return responses, err
}

func (server *Server) handleQueryRequest(ctx context.Context, target string, request *query.Request) (*query.TimeSeriesResponse, error) {
	handler, ok := server.Handlers[target]

	if ok == false {
		return nil, fmt.Errorf("no handler found for target '%s'", target)
	}

	q := handler.Endpoints().Query

	if q == nil {
		return nil, fmt.Errorf("timeseries query not implemented for target '%s'", target)
	}

	return q(ctx, &request.Args)
}

func (server *Server) handleTableQueryRequest(ctx context.Context, target string, request *query.Request) (*query.TableResponse, error) {
	handler, ok := server.Handlers[target]

	if ok == false {
		return nil, fmt.Errorf("no handler found for target '%s'", target)
	}

	q := handler.Endpoints().TableQuery

	if q == nil {
		return nil, fmt.Errorf("table query not implemented for target '%s'", target)
	}

	return q(ctx, &request.Args)
}
