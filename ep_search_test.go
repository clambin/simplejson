package simplejson_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
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
