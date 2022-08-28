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
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "", bytes.NewBufferString(tt.request))
		s.TagValues(w, req)

		if !tt.pass {
			require.NotEqual(t, http.StatusOK, w.Code)
			continue
		}
		require.Equal(t, http.StatusOK, w.Code)

		body, _ := io.ReadAll(w.Body)

		gp := filepath.Join("testdata", strings.ToLower(t.Name())+"_"+tt.name+".golden")
		if *update {
			err := os.WriteFile(gp, body, 0644)
			require.NoError(t, err)
		}
		golden, err := os.ReadFile(gp)
		require.NoError(t, err)
		assert.Equal(t, string(golden), string(body))
	}
}
