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
	"testing"
)

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
