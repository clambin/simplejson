package grafana_json_test

import (
	grafanaJSON "github.com/clambin/grafana-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestAPIServer_Query(t *testing.T) {
	serverRunning(t)

	body, err := call(Port, "/metrics", "GET", "")
	require.Nil(t, err)
	assert.Contains(t, body, "http_duration_seconds")
	assert.Contains(t, body, "http_duration_seconds_sum")
	assert.Contains(t, body, "http_duration_seconds_count")

	body, err = call(Port, "/search", "POST", "")
	require.NoError(t, err)
	assert.Equal(t, `["A","B","C","Crash"]`, body)

	req := `{
	"maxDataPoints": 100,
	"interval": "1y",
	"range": {
		"from": "2020-01-01T00:00:00.000Z",
		"to": "2020-12-31T00:00:00.000Z"
	},
	"targets": [
		{ "target": "A", "type": "timeserie" },
		{ "target": "B" }
	]
}`

	body, err = call(Port, "/query", "POST", req)

	require.Nil(t, err)
	assert.Equal(t, `[{"target":"A","datapoints":[[100,1577836800000],[101,1577836860000],[103,1577836920000]]},{"target":"B","datapoints":[[100,1577836800000],[99,1577836860000],[98,1577836920000]]}]`, body)
}

func TestAPIServer_TableQuery(t *testing.T) {
	serverRunning(t)

	req := `{
	"maxDataPoints": 100,
	"interval": "1y",
	"range": {
		"from": "2020-01-01T00:00:00.000Z",
		"to": "2020-12-31T00:00:00.000Z"
	},
	"targets": [
		{ "target": "C", "type": "table" }
	]
}`
	body, err := call(Port, "/query", "POST", req)

	require.Nil(t, err)
	assert.Equal(t, `[{"type":"table","columns":[{"text":"Time","type":"time"},{"text":"Label","type":"string"},{"text":"Series A","type":"number"},{"text":"Series B","type":"number"}],"rows":[["2020-01-01T00:00:00Z","foo",42,64.5],["2020-01-01T00:01:00Z","bar",43,100]]}]`, body)
}

func TestAPIServer_MissingEndpoint(t *testing.T) {
	s := grafanaJSON.Server{Handlers: []grafanaJSON.Handler{&testAPIHandler{noEndpoints: true}}}

	go func() {
		err := s.Run(8082)
		require.NoError(t, err)
	}()

	require.Eventually(t, func() bool {
		body, err := call(8082, "/", "GET", "")
		require.NoError(t, err)
		return assert.Empty(t, body)
	}, 500*time.Millisecond, 10*time.Millisecond)

	req := `{
	"maxDataPoints": 100,
	"interval": "1y",
	"range": {
		"from": "2020-01-01T00:00:00.000Z",
		"to": "2020-12-31T00:00:00.000Z"
	},
	"targets": [
		{ "target": "C", "type": "table" }
	]
}`
	_, err := call(Port, "/query", "POST", req)
	assert.NoError(t, err)
}
