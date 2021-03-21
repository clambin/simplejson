package grafana_json

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

// endpoints

func (server *Server) hello(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "Hello")
}

func (server *Server) search(w http.ResponseWriter, _ *http.Request) {
	targets := server.handler.Search()
	output, err := json.Marshal(targets)

	if err == nil {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(output)
	} else {
		log.WithField("err", err).Warning("failed to create search response")
		http.Error(w, "failed to create search response", http.StatusInternalServerError)
	}
}

func (server *Server) query(w http.ResponseWriter, req *http.Request) {
	var (
		err     error
		bytes   []byte
		request QueryRequest
		output  []byte
	)
	defer req.Body.Close()

	if bytes, err = ioutil.ReadAll(req.Body); err == nil {
		err = json.Unmarshal(bytes, &request)
	}

	if err != nil {
		log.WithField("err", err).Warning("failed to parse request")
		http.Error(w, "failed to parse request", http.StatusBadRequest)
		return
	}

	responses := make([]interface{}, 0, len(request.Targets))

	for _, target := range request.Targets {
		switch target.Type {
		case "timeserie", "":
			var response *QueryResponse
			if response, err = server.handleQueryRequest(target.Target, &request); err == nil {
				responses = append(responses, response)
			} else {
				break
			}
		case "table":
			var response *TableQueryResponse
			if response, err = server.handleTableQueryRequest(target.Target, &request); err == nil {
				responses = append(responses, response)
			} else {
				break
			}
		}
	}

	if err == nil {
		if output, err = json.Marshal(responses); err == nil {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(output)
		} else {
			log.WithField("err", err).Warning("unable to create response")
			http.Error(w, "unable to create response", http.StatusInternalServerError)
		}
	} else {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (server *Server) handleQueryRequest(target string, request *QueryRequest) (*QueryResponse, error) {
	if server.handler.Query == nil {
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
	return server.handler.Query(target, &args)

}

func (server *Server) handleTableQueryRequest(target string, request *QueryRequest) (*TableQueryResponse, error) {
	if server.handler.TableQuery == nil {
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
	return server.handler.TableQuery(target, &args)
}

func (server *Server) annotations(w http.ResponseWriter, req *http.Request) {
	var (
		err         error
		bytes       []byte
		request     AnnotationRequest
		annotations []Annotation
	)
	defer req.Body.Close()

	if req.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Headers", "accept, content-type")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		return
	}

	if bytes, err = ioutil.ReadAll(req.Body); err == nil {
		err = json.Unmarshal(bytes, &request)
	}

	if err != nil {
		log.WithField("err", err).Warning("failed to parse request")
		http.Error(w, "failed to parse request", http.StatusBadRequest)
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

	if annotations, err = server.handler.Annotations(request.Annotation.Name, &args); err == nil {
		for index, annotation := range annotations {
			annotation.request = request.Annotation
			annotations[index] = annotation
		}

		if bytes, err = json.Marshal(annotations); err == nil {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(bytes)
		}
	}

	if err != nil {
		http.Error(w, "failed to process annotation request", http.StatusInternalServerError)
	}

}
