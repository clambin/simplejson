package grafana_json

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)

type tagKey struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type tagValueKey struct {
	Key string `json:"key"`
}

type tagValue struct {
	Text string `json:"text"`
}

func (server *Server) tagKeys(w http.ResponseWriter, req *http.Request) {
	defer func(body io.ReadCloser) {
		_ = body.Close()
	}(req.Body)

	var keys []tagKey
	for _, handler := range server.Handlers {
		if handler.Endpoints().TagKeys != nil {
			for _, newKey := range handler.Endpoints().TagKeys(req.Context()) {
				keys = append(keys, tagKey{Type: "string", Text: newKey})
			}
		}
	}

	output, err := json.Marshal(keys)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(output)
}

func (server *Server) tagValues(w http.ResponseWriter, req *http.Request) {
	defer func(body io.ReadCloser) {
		_ = body.Close()
	}(req.Body)

	bytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var key tagValueKey
	if err = json.Unmarshal(bytes, &key); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var tagValues []tagValue
	for _, handler := range server.Handlers {
		if handler.Endpoints().TagValues != nil {
			var values []string
			values, err = handler.Endpoints().TagValues(req.Context(), key.Key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			for _, value := range values {
				tagValues = append(tagValues, tagValue{Text: value})
			}
		}
	}

	if bytes, err = json.Marshal(tagValues); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(bytes)
}
