package grafana_json

import (
	"encoding/json"
	"net/http"
)

func (server *Server) getTargets() (targets []string) {
	for _, h := range server.Handlers {
		if h.Endpoints().Search != nil {
			targets = append(targets, h.Endpoints().Search()...)
		}
	}
	return
}

func (server *Server) search(w http.ResponseWriter, _ *http.Request) {
	targets := server.getTargets()
	output, err := json.Marshal(targets)

	if err != nil {
		http.Error(w, "failed to create search response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(output)
}
