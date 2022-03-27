package simplejson

import (
	"encoding/json"
	"github.com/clambin/simplejson/v3/annotation"
	"net/http"
)

func (server *Server) annotations(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Headers", "accept, content-type")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		return
	}

	var request annotation.Request
	handleEndpoint(w, req, &request, func() (response []json.Marshaler, err error) {
		var annotations []annotation.Annotation
		for _, h := range server.Handlers {
			if h.Endpoints().Annotations == nil {
				continue
			}

			var newAnnotations []annotation.Annotation
			newAnnotations, err = h.Endpoints().Annotations(request)

			if err == nil {
				annotations = append(annotations, newAnnotations...)
			}
		}

		for index := range annotations {
			annotations[index].Request = request.Annotation
			response = append(response, &annotations[index])
		}
		return
	})
}
