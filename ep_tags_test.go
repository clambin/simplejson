package simplejson_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestTags(t *testing.T) {
	serverRunning(t)

	body, err := call(Port, "/tag-keys", http.MethodPost, "")
	require.NoError(t, err)
	assert.Equal(t, `[{"type":"string","text":"foo"},{"type":"string","text":"bar"}]`, body)
}

func TestTagValues(t *testing.T) {
	serverRunning(t)

	body, err := call(Port, "/tag-values", http.MethodPost, `{"key": "foo"}`)
	require.NoError(t, err)
	assert.Equal(t, `[{"text":"A"},{"text":"B"}]`, body)

	body, err = call(Port, "/tag-values", http.MethodPost, `{"key": "foo"`)
	assert.Error(t, err)

	body, err = call(Port, "/tag-values", http.MethodPost, `{"key": "snafu"}`)
	assert.Error(t, err)
}

func TestAhHocFilter(t *testing.T) {
	serverRunning(t)
	req := `{
	"maxDataPoints": 100,
	"interval": "1y",
	"range": {
		"from": "2020-01-01T00:00:00.000Z",
		"to": "2020-12-31T00:00:00.000Z"
	},
	"targets": [
		{ "target": "B" }
	],
	"adhocFilters": [
    	{
      		"value":"B",
			"operator":"<",
      		"condition":"",
      		"key":"100"
    	}
  	]
}`

	body, err := call(Port, "/query", http.MethodPost, req)
	require.NoError(t, err)
	assert.Equal(t, `[{"target":"B","datapoints":[[100,1577836800000],[99,1577836860000],[98,1577836920000]]}]`, body)
}
