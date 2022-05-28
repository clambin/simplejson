package simplejson

import (
	"encoding/json"
	"net/http"
)

// handleEndpoint is a wrapper for simplejson endpoint handlers. It parses the incoming http.Request, calls the processor
// and writes the response to the http.ResponseWriter.
func handleEndpoint(w http.ResponseWriter, req *http.Request, request json.Unmarshaler, processor func() ([]json.Marshaler, error)) {
	var err error
	if req.ContentLength > 0 {
		if err = json.NewDecoder(req.Body).Decode(&request); err != nil {
			http.Error(w, "failed to parse request: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	var response []json.Marshaler
	if response, err = processor(); err != nil {
		http.Error(w, "failed to process request: "+err.Error(), http.StatusInternalServerError)
		return

	}

	if err = json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to create response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
