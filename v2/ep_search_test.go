package simplejson_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestSearch(t *testing.T) {
	serverRunning(t)

	body, err := call(Port, "/search", http.MethodPost, "")
	require.NoError(t, err)
	assert.Equal(t, `["A","B","C"]`, body)
}
