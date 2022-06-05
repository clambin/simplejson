package simplejson_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/clambin/simplejson/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"os"
	"path/filepath"
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

	gp := filepath.Join("testdata", t.Name()+".golden")
	if *update {
		t.Logf("updating golden file for %s", t.Name())
		err = os.WriteFile(gp, []byte(body), 0644)
		require.NoError(t, err, "failed to update golden file")
	}

	var golden []byte
	golden, err = os.ReadFile(gp)
	require.NoError(t, err)

	assert.Equal(t, string(golden), body)
}

func TestServer_Query_Fail(t *testing.T) {
	req := `{
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
	serverRunning(t)
	_, err := call(Port, "/query", http.MethodPost, req)
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

	gp := filepath.Join("testdata", t.Name()+".golden")
	if *update {
		t.Logf("updating golden file for %s", t.Name())
		err = os.WriteFile(gp, []byte(body), 0644)
		require.NoError(t, err, "failed to update golden file")
	}

	var g []byte
	g, err = os.ReadFile(gp)
	require.NoError(t, err)

	assert.Equal(t, body, string(g))
}

func TestServer_TableQuery_Fail(t *testing.T) {
	req := `{
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
	serverRunning(t)
	_, err := call(Port, "/query", http.MethodPost, req)
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
