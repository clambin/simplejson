package simplejson

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"io"
	"net/http"
	"time"
)

var queryDuration = promauto.NewSummaryVec(prometheus.SummaryOpts{
	Name: "grafana_api_query_duration_seconds",
	Help: "Grafana SimpleJSON server duration of query requests by target",
}, []string{"app", "type", "target"})

var queryFailure = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "grafana_api_query_failed_count",
	Help: "Grafana SimpleJSON server count of failed requests",
}, []string{"app", "type", "target"})

func (server *Server) query(w http.ResponseWriter, req *http.Request) {
	defer func(body io.ReadCloser) {
		_ = body.Close()
	}(req.Body)

	bytes, err := io.ReadAll(req.Body)

	var request TimeSeriesRequest
	if err == nil {
		err = json.Unmarshal(bytes, &request)
	}

	if err != nil {
		http.Error(w, "failed to parse request: "+err.Error(), http.StatusBadRequest)
		return
	}

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

	if err != nil {
		http.Error(w, "failed to create response: "+err.Error(), http.StatusBadRequest)
		return
	}

	var output []byte
	output, err = json.Marshal(responses)

	if err != nil {
		http.Error(w, "unable to create response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(output)
}

func (server *Server) handleQueryRequest(ctx context.Context, target string, request *TimeSeriesRequest) (*TimeSeriesResponse, error) {
	h := server.findHandler(target)

	if h == nil {
		return nil, fmt.Errorf("no handler found for target '%s'", target)
	}

	q := h.Endpoints().Query

	if q == nil {
		return nil, errors.New("query endpoint not implemented")
	}

	args := TimeSeriesQueryArgs{
		Args: Args{
			Range: Range{
				From: request.Range.From,
				To:   request.Range.To,
			},
			AdHocFilters: request.AdHocFilters,
		},
		MaxDataPoints: request.MaxDataPoints,
	}

	return q(ctx, target, &args)
}

func (server *Server) handleTableQueryRequest(ctx context.Context, target string, request *TimeSeriesRequest) (*TableQueryResponse, error) {
	h := server.findHandler(target)

	if h == nil {
		return nil, fmt.Errorf("no handler found for target '%s'", target)
	}

	q := h.Endpoints().TableQuery

	if q == nil {
		return nil, errors.New("table query endpoint not implemented")
	}
	args := TableQueryArgs{
		Args: Args{
			Range: Range{
				From: request.Range.From,
				To:   request.Range.To,
			},
			AdHocFilters: request.AdHocFilters,
		},
	}
	return q(ctx, target, &args)
}

func (server *Server) findHandler(target string) Handler {
	for _, h := range server.Handlers {
		if h.Endpoints().Search == nil {
			continue
		}

		for _, t := range h.Endpoints().Search() {
			if t == target {
				return h
			}
		}
	}

	return nil
}
