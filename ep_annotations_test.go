package simplejson_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestServer_Annotations(t *testing.T) {
	serverRunning(t)

	body, err := call(Port, "/annotations", http.MethodOptions, "")

	require.NoError(t, err)
	assert.Equal(t, "", body)

	req := `{
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
}`
	body, err = call(Port, "/annotations", http.MethodPost, req)

	require.NoError(t, err)
	assert.Equal(t, `[{"annotation":{"name":"snafu","datasource":"fubar","enable":true,"query":""},"time":1609459200000,"title":"foo","text":"bar","tags":["snafu"]}]
`, body)
}
