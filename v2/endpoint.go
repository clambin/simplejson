package simplejson

import (
	"encoding/json"
	"io"
	"net/http"
)

// handleEndpoint is a wrapper for simplejson endpoint handlers. It parses the incoming http.Request, calls the processor
// with that request and writes the response to the http.ResponseWriter.
func handleEndpoint(w http.ResponseWriter, req *http.Request, request interface{}, processor func() (interface{}, error)) {
	var err error
	if request != nil {
		var body []byte
		body, err = io.ReadAll(req.Body)
		if err == nil {
			err = json.Unmarshal(body, &request)
		}
	}

	if err != nil {
		http.Error(w, "failed to parse request: "+err.Error(), http.StatusBadRequest)
		return
	}

	var response interface{}
	response, err = processor()

	if err != nil {
		http.Error(w, "failed to process request: "+err.Error(), http.StatusInternalServerError)
		return

	}

	var output []byte
	output, err = json.Marshal(response)

	if err != nil {
		http.Error(w, "failed to create response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(output)
}
