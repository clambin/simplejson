package simplejson

import (
	"encoding/json"
	"net/http"
)

func (server *Server) search(w http.ResponseWriter, _ *http.Request) {
	output, _ := json.Marshal(server.Targets())
	_, _ = w.Write(output)
}
