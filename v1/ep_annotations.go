package simplejson

import (
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

	var request AnnotationRequest
	handleEndpoint(w, req, &request, func() (response interface{}, err error) {
		args := AnnotationRequestArgs{
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
				return
			}

			annotations = append(annotations, newAnnotations...)
		}

		for index := range annotations {
			annotations[index].Request = request.Annotation
		}

		return annotations, nil
	})
}
