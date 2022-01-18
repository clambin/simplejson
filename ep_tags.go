package simplejson

import (
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
	handleEndpoint(w, req, nil, func() (interface{}, error) {
		var keys []tagKey
		for _, handler := range server.Handlers {
			if handler.Endpoints().TagKeys != nil {
				for _, newKey := range handler.Endpoints().TagKeys(req.Context()) {
					keys = append(keys, tagKey{Type: "string", Text: newKey})
				}
			}
		}
		return keys, nil
	})
}

func (server *Server) tagValues(w http.ResponseWriter, req *http.Request) {
	var key tagValueKey
	handleEndpoint(w, req, &key, func() (interface{}, error) {
		var tagValues []tagValue
		for _, handler := range server.Handlers {
			if handler.Endpoints().TagValues != nil {
				values, err := handler.Endpoints().TagValues(req.Context(), key.Key)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return nil, err
				}

				for _, value := range values {
					tagValues = append(tagValues, tagValue{Text: value})
				}
			}
		}
		return tagValues, nil
	})
}
