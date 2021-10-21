package grafana_json

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

// endpoints

func (server *Server) hello(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "Hello")
}

func (server *Server) getTargets() (targets []string) {
	for _, h := range server.Handlers {
		if h.Endpoints().Search != nil {
			targets = append(targets, h.Endpoints().Search()...)
		}
	}
	return
}

func (server *Server) search(w http.ResponseWriter, _ *http.Request) {
	targets := server.getTargets()
	output, err := json.Marshal(targets)

	if err != nil {
		http.Error(w, "failed to create search response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(output)
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

var queryDuration = promauto.NewSummaryVec(prometheus.SummaryOpts{
	Name: "grafana_api_query_duration_seconds",
	Help: "Grafana API duration of query requests by target",
}, []string{"type", "target"})

func (server *Server) query(w http.ResponseWriter, req *http.Request) {
	defer func(body io.ReadCloser) {
		_ = body.Close()
	}(req.Body)

	bytes, err := ioutil.ReadAll(req.Body)

	var request QueryRequest
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
			var response *QueryResponse
			if response, err = server.handleQueryRequest(req.Context(), target.Target, &request); err == nil {
				responses = append(responses, response)
			} else {
				break
			}
		case "table":
			var response *TableQueryResponse
			if response, err = server.handleTableQueryRequest(req.Context(), target.Target, &request); err == nil {
				responses = append(responses, response)
			} else {
				break
			}
		}
		queryDuration.WithLabelValues(target.Type, target.Target).Observe(time.Now().Sub(start).Seconds())
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

func (server *Server) handleQueryRequest(ctx context.Context, target string, request *QueryRequest) (*QueryResponse, error) {
	h := server.findHandler(target)

	if h == nil {
		return nil, fmt.Errorf("no handler found for target '%s'", target)
	}

	q := h.Endpoints().Query

	if q == nil {
		return nil, errors.New("query endpoint not implemented")
	}

	args := TimeSeriesQueryArgs{
		CommonQueryArgs: CommonQueryArgs{
			Range: QueryRequestRange{
				From: request.Range.From,
				To:   request.Range.To,
			},
		},
		MaxDataPoints: request.MaxDataPoints,
	}

	return q(ctx, target, &args)
}

func (server *Server) handleTableQueryRequest(ctx context.Context, target string, request *QueryRequest) (*TableQueryResponse, error) {
	h := server.findHandler(target)

	if h == nil {
		return nil, fmt.Errorf("no handler found for target '%s'", target)
	}

	q := h.Endpoints().TableQuery

	if q == nil {
		return nil, errors.New("table query endpoint not implemented")
	}
	args := TableQueryArgs{
		CommonQueryArgs: CommonQueryArgs{
			Range: QueryRequestRange{
				From: request.Range.From,
				To:   request.Range.To,
			},
		},
	}
	return q(ctx, target, &args)
}

func (server *Server) annotations(w http.ResponseWriter, req *http.Request) {
	defer func(body io.ReadCloser) {
		_ = body.Close()
	}(req.Body)

	if req.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Headers", "accept, content-type")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		return
	}

	var request AnnotationRequest
	bytes, err := ioutil.ReadAll(req.Body)
	if err == nil {
		err = json.Unmarshal(bytes, &request)
	}

	if err != nil {
		http.Error(w, "failed to parse request: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.WithFields(log.Fields{
		"name":   request.Annotation.Name,
		"enable": request.Annotation.Enable,
		"query":  request.Annotation.Query,
	}).Debug("annotation received")

	args := AnnotationRequestArgs{
		CommonQueryArgs{
			Range: request.Range,
		},
	}

	var annotations []Annotation
	for _, h := range server.Handlers {
		if h.Endpoints().Annotations == nil {
			continue
		}

		var newAnnotations []Annotation
		newAnnotations, err = h.Endpoints().Annotations(request.Annotation.Name, request.Annotation.Query, &args)

		if err != nil {
			log.WithError(err).Warning("failed to get annotations from handler")
			continue
		}

		for _, annotation := range newAnnotations {
			annotation.request = request.Annotation
			annotations = append(annotations, annotation)
		}
	}

	var output []byte
	output, err = json.Marshal(annotations)

	if err != nil {
		http.Error(w, "failed to process annotations request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(output)
}
