package simplejson

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

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

	var request Request
	bytes, err := io.ReadAll(req.Body)
	if err == nil {
		err = json.Unmarshal(bytes, &request)
	}

	if err != nil {
		http.Error(w, "failed to parse request: "+err.Error(), http.StatusBadRequest)
		return
	}

	args := RequestArgs{
		Args: Args{
			Range: request.Range,
		},
	}

	var annotations []Annotation
	for _, h := range server.Handlers {
		if h.Endpoints().Annotations == nil {
			continue
		}

		var newAnnotations []Annotation
		newAnnotations, err = h.Endpoints().Annotations(request.Annotation.Name, request.Annotation.Query, &args)

		if err != nil {
			log.WithError(err).Warning("failed to get annotation from handler")
			continue
		}

		for _, a := range newAnnotations {
			a.Request = request.Annotation
			annotations = append(annotations, a)
		}
	}

	var output []byte
	output, err = json.Marshal(annotations)

	if err != nil {
		http.Error(w, "failed to process annotation request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(output)
}
