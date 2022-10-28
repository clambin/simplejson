package simplejson_test

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestServer_Query(t *testing.T) {
	testCases := []struct {
		name    string
		request string
		code    int
		pass    bool
	}{
		{
			name: "timeseries",
			request: `{
				"maxDataPoints": 100,
				"interval": "1y",
				"range": {"from": "2020-01-01T00:00:00.000Z", "to": "2020-12-31T00:00:00.000Z"},
				"targets": [{ "target": "A", "type": "timeserie" },{ "target": "B" }]
}`,
			code: http.StatusOK,
			pass: true,
		},
		{
			name: "table",
			request: `{
				"maxDataPoints": 100,
				"interval": "1y",
				"range": {"from": "2020-01-01T00:00:00.000Z","to": "2020-12-31T00:00:00.000Z"},
				"targets": [{ "target": "C", "type": "table" }]
}`,
			code: http.StatusOK,
			pass: true,
		},
		{
			name: "adhoc",
			request: `{
				"maxDataPoints": 100,
				"interval": "1y",
				"range": {"from": "2020-01-01T00:00:00.000Z","to": "2020-12-31T00:00:00.000Z"},
				"targets": [{ "target": "B" }],
				"adhocFilters": [{"value":"B","operator":"<","condition":"","key":"100"}]
}`,
			code: http.StatusOK,
			pass: true,
		},
		{
			name: "bad_target",
			request: `{
				"maxDataPoints": 100,
				"interval": "1y",
				"range": {"from": "2020-01-01T00:00:00.000Z","to": "2020-12-31T00:00:00.000Z"},
				"targets": [{ "target": "D", "type": "timeserie"}]
}`,
			code: http.StatusInternalServerError,
			pass: false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "", bytes.NewBufferString(tt.request))

			s.Query(w, req)
			require.Equal(t, tt.code, w.Code)

			if !tt.pass {
				return
			}

			body, err := io.ReadAll(w.Body)
			require.NoError(t, err)

			gp := strings.ToLower(filepath.Join("testdata", t.Name()+".golden"))
			if *update {
				t.Logf("updating golden file for %s", t.Name())
				err = os.WriteFile(gp, body, 0644)
				require.NoError(t, err, "failed to update golden file")
			}

			var golden []byte
			golden, err = os.ReadFile(gp)
			require.NoError(t, err)

			assert.Equal(t, string(golden), string(body))
		})
	}
}

func BenchmarkServer_Query(b *testing.B) {
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest(http.MethodPost, "", bytes.NewBufferString(`{
			"maxDataPoints": 100,
			"interval": "1y",
			"range": {"from": "2020-01-01T00:00:00.000Z","to": "2020-12-31T00:00:00.000Z"},
			"targets": [{ "target": "A" }]
}`))
		w := httptest.NewRecorder()
		s.Query(w, req)
		if w.Code != http.StatusOK {
			b.Fatalf("unexpected http code: %d", w.Code)
		}
	}
}

func BenchmarkServer_TableQuery(b *testing.B) {
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest(http.MethodPost, "", bytes.NewBufferString(`{
			"maxDataPoints": 100,
			"interval": "1y",
			"range": {"from": "2020-01-01T00:00:00.000Z","to": "2020-12-31T00:00:00.000Z"},
			"targets": [{ "target": "C" }]
}`))
		w := httptest.NewRecorder()
		s.Query(w, req)
		if w.Code != http.StatusOK {
			b.Fatalf("unexpected http code: %d", w.Code)
		}
	}
}
