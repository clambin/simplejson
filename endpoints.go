package grafana_json

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
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

func (server *Server) search(w http.ResponseWriter, _ *http.Request) {
	targets := server.handler.Endpoints().Search()
	output, err := json.Marshal(targets)

	if err != nil {
		http.Error(w, "failed to create search response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(output)
}

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
	if server.handler.Endpoints().Query == nil {
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
	return server.handler.Endpoints().Query(ctx, target, &args)

}

func (server *Server) handleTableQueryRequest(ctx context.Context, target string, request *QueryRequest) (*TableQueryResponse, error) {
	if server.handler.Endpoints().TableQuery == nil {
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
	return server.handler.Endpoints().TableQuery(ctx, target, &args)
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
	if annotations, err = server.handler.Endpoints().Annotations(request.Annotation.Name, request.Annotation.Query, &args); err == nil {
		for index, annotation := range annotations {
			annotation.request = request.Annotation
			annotations[index] = annotation
		}
		bytes, err = json.Marshal(annotations)
	}

	if err != nil {
		http.Error(w, "failed to process annotations request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(bytes)
}
