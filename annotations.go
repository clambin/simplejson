package grafana_json

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
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

	var request AnnotationRequest
	bytes, err := ioutil.ReadAll(req.Body)
	if err == nil {
		err = json.Unmarshal(bytes, &request)
	}

	if err != nil {
		http.Error(w, "failed to parse request: "+err.Error(), http.StatusBadRequest)
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

	var annotations []Annotation
	for _, h := range server.Handlers {
		if h.Endpoints().Annotations == nil {
			continue
		}

		var newAnnotations []Annotation
		newAnnotations, err = h.Endpoints().Annotations(request.Annotation.Name, request.Annotation.Query, &args)

		if err != nil {
			log.WithError(err).Warning("failed to get annotations from handler")
			continue
		}

		for _, annotation := range newAnnotations {
			annotation.request = request.Annotation
			annotations = append(annotations, annotation)
		}
	}

	var output []byte
	output, err = json.Marshal(annotations)

	if err != nil {
		http.Error(w, "failed to process annotations request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(output)
}
