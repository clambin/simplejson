package simplejson

import (
	"bytes"
	"context"
	"github.com/clambin/simplejson/v3/query"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func BenchmarkServer_Query(b *testing.B) {
	s := &Server{
		Name: "benchmark",
		Handlers: map[string]Handler{
			"A": &handler{},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest(http.MethodPost, "", bytes.NewBufferString(`{
	"maxDataPoints": 100,
	"interval": "1y",
	"range": {
		"from": "2020-01-01T00:00:00.000Z",
		"to": "2020-12-31T00:00:00.000Z"
	},
	"targets": [{ "target": "A" }]
}`))
		w := httptest.NewRecorder()
		s.query(w, req)
		if w.Code != http.StatusOK {
			b.Fatalf("unexpected http code: %d", w.Code)
		}
	}
}

type handler struct {
}

var _ Handler = &handler{}

func (h handler) Endpoints() Endpoints {
	return Endpoints{Query: h.query}
}

func (h handler) query(_ context.Context, _ query.Request) (query.Response, error) {
	return &query.TableResponse{
		Columns: []query.Column{
			{Text: "time", Data: query.TimeColumn{time.Now()}},
			{Text: "value", Data: query.NumberColumn{1}},
		},
	}, nil
}
