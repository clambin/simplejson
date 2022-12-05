package simplejson

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
)

func (s *Server) handleQuery(ctx context.Context, request QueryRequest) ([]json.Marshaler, error) {
	responses := make([]json.Marshaler, 0, len(request.Targets))
	for _, target := range request.Targets {
		var timer *prometheus.Timer
		if s.queryMetrics != nil {
			timer = prometheus.NewTimer(s.queryMetrics.duration.WithLabelValues(target.Name, target.Type))
		}

		response, err := s.handleQueryRequest(ctx, target, request)

		if timer != nil {
			timer.ObserveDuration()
		}
		if err != nil {
			if s.queryMetrics != nil {
				s.queryMetrics.errors.WithLabelValues(target.Name, target.Type).Add(1.0)
			}
			return nil, err
		}
		responses = append(responses, response)
	}
	return responses, nil
}

type Response interface {
	MarshalJSON() ([]byte, error)
}

func (s *Server) handleQueryRequest(ctx context.Context, target Target, request QueryRequest) (Response, error) {
	handler, ok := s.Handlers[target.Name]
	if !ok {
		return nil, fmt.Errorf("no handler found for target '%s'", target)
	}

	q := handler.Endpoints().Query
	if q == nil {
		return nil, fmt.Errorf("query not implemented for target '%s'", target)
	}

	return q(ctx, request)
}
