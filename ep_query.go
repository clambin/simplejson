package simplejson

import (
	"context"
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"net/http"
	"time"
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
	// fixme: table query uses TableQueryRequest
	var request QueryRequest
	handleEndpoint(w, req, &request, func() (interface{}, error) {
		var err error
		responses := make([]interface{}, 0, len(request.Targets))
		for _, target := range request.Targets {
			start := time.Now()
			switch target.Type {
			case "timeserie", "":
				var response *TimeSeriesResponse
				if response, err = server.handleQueryRequest(req.Context(), target.Target, &request); err == nil {
					responses = append(responses, response)
				}
			case "table":
				var response *TableQueryResponse
				if response, err = server.handleTableQueryRequest(req.Context(), target.Target, &request); err == nil {
					responses = append(responses, response)
				}
			}
			queryDuration.WithLabelValues(server.Name, target.Type, target.Target).Observe(time.Now().Sub(start).Seconds())
			if err != nil {
				queryFailure.WithLabelValues(server.Name, target.Type, target.Target).Add(1.0)
				break
			}
		}
		return responses, err
	})
}

func (server *Server) handleQueryRequest(ctx context.Context, target string, request *QueryRequest) (*TimeSeriesResponse, error) {
	handler, ok := server.Handlers[target]

	if ok == false {
		return nil, fmt.Errorf("no handler found for target '%s'", target)
	}

	q := handler.Endpoints().Query

	if q == nil {
		return nil, errors.New("query endpoint not implemented")
	}

	return q(ctx, &request.TimeSeriesQueryArgs)
}

func (server *Server) handleTableQueryRequest(ctx context.Context, target string, request *QueryRequest) (*TableQueryResponse, error) {
	handler, ok := server.Handlers[target]

	if ok == false {
		return nil, fmt.Errorf("no handler found for target '%s'", target)
	}

	q := handler.Endpoints().TableQuery

	if q == nil {
		return nil, errors.New("table query endpoint not implemented")
	}

	args := &TableQueryArgs{Args: request.Args}

	return q(ctx, args)
}
