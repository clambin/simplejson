package simplejson_test

import (
	"context"
	grafanajson "github.com/clambin/grafana-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestMultiHandler(t *testing.T) {
	s := grafanajson.Server{Handlers: []grafanajson.Handler{
		&Handler1{},
		&Handler2{},
	}}

	go func() {
		_ = s.Run(8080)
	}()

	listener, err := net.Listen("tcp4", ":0")
	if err != nil {
		panic(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port

	httpServer := http.Server{
		Handler: s.GetRouter(),
	}
	go func() {
		err2 := httpServer.Serve(listener)
		if err2 != http.ErrServerClosed {
			panic(err2)
		}
	}()

	require.Eventually(t, func() bool {
		_, err = call(port, "/search", http.MethodPost, "")
		return err == nil
	}, 500*time.Millisecond, 10*time.Millisecond)

	var response string
	response, err = call(port, "/search", http.MethodPost, "")
	require.NoError(t, err)
	assert.Equal(t, `["handler1","handler2"]`, response)

	req := `{
	"maxDataPoints": 100,
	"interval": "1y",
	"range": {
		"from": "2020-01-01T00:00:00.000Z",
		"to": "2020-12-31T00:00:00.000Z"
	},
	"targets": [
		{ "target": "handler1" },
		{ "target": "handler2", "type": "table" }
	]
}`

	response, err = call(port, "/query", http.MethodPost, req)
	require.NoError(t, err)
	assert.Equal(t, "[{\"target\":\"handler1\",\"datapoints\":[]},{\"type\":\"table\",\"columns\":null,\"rows\":null}]", response)
}

type Handler1 struct {
}

func (h1 *Handler1) Endpoints() grafanajson.Endpoints {
	return grafanajson.Endpoints{
		Search: h1.Search,
		Query:  h1.Query,
	}
}

func (h1 *Handler1) Search() []string {
	return []string{"handler1"}
}

func (h1 *Handler1) Query(_ context.Context, _ string, _ *grafanajson.TimeSeriesQueryArgs) (response *grafanajson.TimeSeriesResponse, err error) {
	response = &grafanajson.TimeSeriesResponse{
		Target:     "handler1",
		DataPoints: []grafanajson.DataPoint{},
	}
	return
}

type Handler2 struct {
}

func (h2 *Handler2) Endpoints() grafanajson.Endpoints {
	return grafanajson.Endpoints{
		Search:     h2.Search,
		TableQuery: h2.TableQuery,
	}
}

func (h2 *Handler2) Search() []string {
	return []string{"handler2"}
}

func (h2 *Handler2) TableQuery(_ context.Context, _ string, _ *grafanajson.TableQueryArgs) (response *grafanajson.TableQueryResponse, err error) {
	response = &grafanajson.TableQueryResponse{}
	return
}
