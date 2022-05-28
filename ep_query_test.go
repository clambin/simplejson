package simplejson_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/clambin/simplejson/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestServer_Query(t *testing.T) {
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

	serverRunning(t)
	body, err := call(Port, "/query", http.MethodPost, req)
	require.NoError(t, err)
	assert.Equal(t, `[{"target":"A","datapoints":[[100,1577836800000],[101,1577836860000],[103,1577836920000]]},{"target":"B","datapoints":[[100,1577836800000],[99,1577836860000],[98,1577836920000]]}]
`, body)

	req = `{
	"maxDataPoints": 100,
	"interval": "1y",
	"range": {
		"from": "2020-01-01T00:00:00.000Z",
		"to": "2020-12-31T00:00:00.000Z"
	},
	"targets": [
		{ "target": "D", "type": "timeserie" },
	]
}`

	_, err = call(Port, "/query", http.MethodPost, req)
	require.Error(t, err)

}

func TestServer_TableQuery(t *testing.T) {
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
	serverRunning(t)
	body, err := call(Port, "/query", http.MethodPost, req)
	require.NoError(t, err)
	assert.Equal(t, `[{"type":"table","columns":[{"text":"Time","type":"time"},{"text":"Label","type":"string"},{"text":"Series A","type":"number"},{"text":"Series B","type":"number"}],"rows":[["2020-01-01T00:00:00Z","foo",42,64.5],["2020-01-01T00:01:00Z","bar",43,100]]}]
`, body)

	req = `{
	"maxDataPoints": 100,
	"interval": "1y",
	"range": {
		"from": "2020-01-01T00:00:00.000Z",
		"to": "2020-12-31T00:00:00.000Z"
	},
	"targets": [
		{ "target": "D", "type": "" }
	]
}`
	_, err = call(Port, "/query", http.MethodPost, req)
	require.Error(t, err)

}

func TestServer_MissingEndpoint(t *testing.T) {
	s := simplejson.Server{Handlers: map[string]simplejson.Handler{"C": &testHandler{noEndpoints: true}}}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		// TODO: this may fail if multiple test runs are going in parallel. should be a dynamic port
		err := s.Run(8082)
		require.True(t, errors.Is(err, http.ErrServerClosed))
		wg.Done()
	}()

	require.Eventually(t, func() bool {
		body, err := call(8082, "/", http.MethodGet, "")
		return err == nil && body == ""
	}, 500*time.Millisecond, 10*time.Millisecond)

	const reqTemplate = `{
	"maxDataPoints": 100,
	"interval": "1y",
	"range": {
		"from": "2020-01-01T00:00:00.000Z",
		"to": "2020-12-31T00:00:00.000Z"
	},
	"targets": [ { %s } ]
}`

	for _, target := range []string{
		`"target": "C"`,
		`"target": "D"`,
		`"target": "C", "type": "table"`,
		`"target": "D", "type": "table"`,
	} {
		req := fmt.Sprintf(reqTemplate, target)
		_, err := call(8082, "/query", http.MethodPost, req)
		assert.Error(t, err, target)
	}

	err := s.Shutdown(context.Background(), 15*time.Second)
	require.NoError(t, err)
	wg.Wait()
}
