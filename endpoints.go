package grafana_json

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

// endpoints

func (apiServer *APIServer) hello(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "Hello")
}

func (apiServer *APIServer) search(w http.ResponseWriter, _ *http.Request) {
	targets := apiServer.apiHandler.Search()
	if output, err := json.Marshal(targets); err == nil {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(output)
	} else {
		log.WithField("err", err).Warning("failed to create search response")
		http.Error(w, "failed to create search response", http.StatusInternalServerError)
	}
}

func (apiServer *APIServer) query(w http.ResponseWriter, req *http.Request) {
	var (
		err            error
		bytes          []byte
		request        QueryRequest
		responses      []*QueryResponse
		tableResponses []*QueryTableResponse
		output         []byte
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

	responses, tableResponses, err = apiServer.handleQueryRequest(&request)

	if err == nil {
		packed := make([]interface{}, 0)
		for _, response := range responses {
			packed = append(packed, response)
		}
		for _, response := range tableResponses {
			packed = append(packed, response)
		}

		output, err = json.Marshal(packed)
	}

	if err == nil {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(output)
	} else {
		log.WithField("err", err).Warning("unable to create response")
		http.Error(w, "unable to create response", http.StatusInternalServerError)
	}
}

func (apiServer *APIServer) handleQueryRequest(request *QueryRequest) (responses []*QueryResponse, tableResponses []*QueryTableResponse, err error) {
	for _, target := range request.Targets {
		switch target.Type {
		case "timeserie", "":
			var response *QueryResponse
			if response, err = apiServer.apiHandler.Query(target.Target, request); err == nil {
				response.Target = target.Target
				responses = append(responses, response)
			}
		case "table":
			var response *QueryTableResponse
			if response, err = apiServer.apiHandler.QueryTable(target.Target, request); err == nil {
				tableResponses = append(tableResponses, response)
			}
		default:
			log.WithFields(log.Fields{"target": target.Target, "type": target.Type}).Warning("invalid target type")
			err = fmt.Errorf("unsupport target type: %s", target.Type)
			break
		}
	}
	return
}
