package simplejson

import (
	"net/http"
)

type tagKey struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type valueKey struct {
	Key string `json:"key"`
}

type value struct {
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
	var key valueKey
	handleEndpoint(w, req, &key, func() (interface{}, error) {
		var tagValues []value
		for _, handler := range server.Handlers {
			if handler.Endpoints().TagValues != nil {
				values, err := handler.Endpoints().TagValues(req.Context(), key.Key)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return nil, err
				}

				for _, v := range values {
					tagValues = append(tagValues, value{Text: v})
				}
			}
		}
		return tagValues, nil
	})
}
