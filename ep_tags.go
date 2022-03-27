package simplejson

import (
	"encoding/json"
	"net/http"
)

type tagKey struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (t tagKey) MarshalJSON() (b []byte, err error) {
	type tagKey2 tagKey
	var t2 = tagKey2(t)
	return json.Marshal(t2)
}

type valueKey struct {
	Key string `json:"key"`
}

func (r *valueKey) UnmarshalJSON(b []byte) (err error) {
	type valueKey2 valueKey
	var c valueKey2
	err = json.Unmarshal(b, &c)
	if err != nil {
		return
	}
	*r = valueKey(c)
	return nil
}

type value struct {
	Text string `json:"text"`
}

func (v value) MarshalJSON() (b []byte, err error) {
	type value2 value
	var v2 = value2(v)
	return json.Marshal(v2)
}

func (server *Server) tagKeys(w http.ResponseWriter, req *http.Request) {
	handleEndpoint(w, req, nil, func() (keys []json.Marshaler, err error) {
		for _, handler := range server.Handlers {
			if handler.Endpoints().TagKeys != nil {
				for _, newKey := range handler.Endpoints().TagKeys(req.Context()) {
					keys = append(keys, &tagKey{Type: "string", Text: newKey})
				}
			}
		}
		return keys, nil
	})
}

func (server *Server) tagValues(w http.ResponseWriter, req *http.Request) {
	var key valueKey
	handleEndpoint(w, req, &key, func() (response []json.Marshaler, err error) {
		for _, handler := range server.Handlers {
			if handler.Endpoints().TagValues != nil {
				var values []string
				values, err = handler.Endpoints().TagValues(req.Context(), key.Key)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return nil, err
				}

				for _, v := range values {
					response = append(response, &value{Text: v})
				}
			}
		}
		return
	})
}
