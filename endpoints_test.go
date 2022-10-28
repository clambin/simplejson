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

func TestSearch(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "", nil)

	s.Search(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	body, _ := io.ReadAll(w.Body)
	assert.Equal(t, `["A","B","C"]`, string(body))
}

func TestServer_Annotations(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "", bytes.NewBufferString(`{
	"range": {
		"from": "2020-01-01T00:00:00.000Z",
		"to": "2020-12-31T00:00:00.000Z"
	},
	"annotation": {
		"name": "snafu",
		"datasource": "fubar",
		"enable": true,
		"query": ""
	}
}`))

	s.Annotations(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	body, err := io.ReadAll(w.Body)
	require.NoError(t, err)

	gp := filepath.Join("testdata", t.Name()+".golden")
	if *update {
		t.Logf("updating golden file for %s", t.Name())
		err = os.WriteFile(gp, body, 0644)
		require.NoError(t, err, "failed to update golden file")
	}

	var g []byte
	g, err = os.ReadFile(gp)
	require.NoError(t, err)

	assert.Equal(t, string(body), string(g))
}

func TestServer_Annotations_Options(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodOptions, "", nil)

	s.Annotations(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "accept, content-type", w.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, http.MethodPost, w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestTags(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "", nil)
	s.TagKeys(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	body, err := io.ReadAll(w.Body)
	require.NoError(t, err)
	assert.Equal(t, `[{"type":"string","text":"foo"},{"type":"string","text":"bar"}]
`, string(body))
}

func TestTagValues(t *testing.T) {
	testCases := []struct {
		name    string
		request string
		pass    bool
	}{
		{
			name:    "foo",
			request: `{"key": "foo"}`,
			pass:    true,
		},
		{
			name:    "invalid",
			request: `{"key": "foo"`,
			pass:    false,
		},
		{
			name:    "bad_target",
			request: `{"key": "foo"`,
			pass:    false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "", bytes.NewBufferString(tt.request))
			s.TagValues(w, req)

			if !tt.pass {
				require.NotEqual(t, http.StatusOK, w.Code)
				return
			}
			require.Equal(t, http.StatusOK, w.Code)

			body, _ := io.ReadAll(w.Body)

			gp := strings.ToLower(filepath.Join("testdata", t.Name()+".golden"))
			if *update {
				err := os.WriteFile(gp, body, 0644)
				require.NoError(t, err)
			}
			golden, err := os.ReadFile(gp)
			require.NoError(t, err)
			assert.Equal(t, string(golden), string(body))
		})
	}
}
