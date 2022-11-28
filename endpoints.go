package simplejson

import (
	"encoding/json"
	"net/http"
)

func (s *Server) Search(w http.ResponseWriter, _ *http.Request) {
	output, _ := json.Marshal(s.Targets())
	_, _ = w.Write(output)
}

func (s *Server) Query(w http.ResponseWriter, req *http.Request) {
	var request QueryRequest
	handleEndpoint(w, req, &request, func() ([]json.Marshaler, error) {
		return s.handleQuery(req.Context(), request)
	})
}

func (s *Server) Annotations(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Headers", "accept, content-type")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		return
	}

	var request AnnotationRequest
	handleEndpoint(w, req, &request, func() ([]json.Marshaler, error) {
		var annotations []Annotation
		for _, h := range s.Handlers {
			if h.Endpoints().Annotations != nil {
				if newAnnotations, err := h.Endpoints().Annotations(request); err == nil {
					annotations = append(annotations, newAnnotations...)
				}
			}
		}

		var response []json.Marshaler
		for index := range annotations {
			annotations[index].Request = request.Annotation
			response = append(response, &annotations[index])
		}
		return response, nil
	})
}

func (s *Server) TagKeys(w http.ResponseWriter, req *http.Request) {
	handleEndpoint(w, req, nil, func() (keys []json.Marshaler, _ error) {
		for _, handler := range s.Handlers {
			if handler.Endpoints().TagKeys != nil {
				for _, newKey := range handler.Endpoints().TagKeys(req.Context()) {
					keys = append(keys, &tagKey{Type: "string", Text: newKey})
				}
			}
		}
		return keys, nil
	})
}

func (s *Server) TagValues(w http.ResponseWriter, req *http.Request) {
	var key valueKey
	handleEndpoint(w, req, &key, func() ([]json.Marshaler, error) {
		var response []json.Marshaler
		for _, handler := range s.Handlers {
			if handler.Endpoints().TagValues == nil {
				continue
			}
			values, err := handler.Endpoints().TagValues(req.Context(), key.Key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return nil, err
			}

			for _, v := range values {
				response = append(response, &value{Text: v})
			}
		}
		return response, nil
	})
}

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
	if err == nil {
		*r = valueKey(c)
	}
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

// handleEndpoint is a wrapper for simplejson endpoint handlers. It parses the incoming http.Request, calls the processor
// and writes the response to the http.ResponseWriter.
func handleEndpoint(w http.ResponseWriter, req *http.Request, request json.Unmarshaler, processor func() ([]json.Marshaler, error)) {
	if req.ContentLength > 0 {
		if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
			http.Error(w, "failed to parse request: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	response, err := processor()
	if err != nil {
		http.Error(w, "failed to process request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to create response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
