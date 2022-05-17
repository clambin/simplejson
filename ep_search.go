package simplejson

import (
	"encoding/json"
	"net/http"
)

func (s *Server) search(w http.ResponseWriter, _ *http.Request) {
	output, _ := json.Marshal(s.Targets())
	_, _ = w.Write(output)
}
