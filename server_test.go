package simplejson_test

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/clambin/simplejson/v3"
	"github.com/clambin/simplejson/v3/annotation"
	"github.com/clambin/simplejson/v3/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net"
	"net/http"
	"os"
	"testing"
	"time"
)

var (
	Port   int
	update = flag.Bool("update", false, "update .golden files")
)

func TestMain(m *testing.M) {
	s := simplejson.Server{
		Handlers: handlers}

	listener, err := net.Listen("tcp4", ":0")
	if err != nil {
		panic(err)
	}
	Port = listener.Addr().(*net.TCPAddr).Port

	httpServer := http.Server{
		Handler: s.GetRouter(),
	}
	go func() {
		err2 := httpServer.Serve(listener)
		if !errors.Is(err2, http.ErrServerClosed) {
			panic(err2)
		}
	}()

	rc := m.Run()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	err = s.Shutdown(ctx, time.Second)
	if err != nil {
		panic(err)
	}
	cancel()

	os.Exit(rc)
}

func TestServer_Metrics(t *testing.T) {
	serverRunning(t)

	body, err := call(Port, "/metrics", http.MethodGet, "")
	require.NoError(t, err)
	assert.Contains(t, body, "http_duration_seconds")
	assert.Contains(t, body, "http_duration_seconds_sum")
	assert.Contains(t, body, "http_duration_seconds_count")
}

func BenchmarkServer(b *testing.B) {
	require.Eventually(b, func() bool {
		body, err := call(Port, "/", http.MethodPost, "")
		return err == nil && body == ""
	}, 500*time.Millisecond, 10*time.Millisecond)

	req := `{
	"maxDataPoints": 100,
	"interval": "1y",
	"range": {
		"from": "2020-01-01T00:00:00.000Z",
		"to": "2020-12-31T00:00:00.000Z"
	},
	"targets": [
		{ "target": "A" },
		{ "target": "B" }
	]
}`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := call(Port, "/query", http.MethodPost, req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func serverRunning(t *testing.T) {
	require.Eventually(t, func() bool {
		body, err := call(Port, "/", http.MethodGet, "")
		return err == nil && body == ""
	}, 500*time.Millisecond, 10*time.Millisecond)
}

func call(port int, path, method, body string) (response string, err error) {
	url := fmt.Sprintf("http://127.0.0.1:%d%s", port, path)
	client := &http.Client{}
	reqBody := bytes.NewBuffer([]byte(body))
	req, _ := http.NewRequest(method, url, reqBody)

	var resp *http.Response
	resp, err = client.Do(req)

	if err != nil {
		return
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("call failed: %s", resp.Status)
	}

	var buff []byte
	if buff, err = io.ReadAll(resp.Body); err == nil {
		response = string(buff)
	}

	return
}

//
//
// Test Handler
//

type testHandler struct {
	noEndpoints bool

	queryResponse query.Response
	annotations   []annotation.Annotation
	tags          []string
	tagValues     map[string][]string
}

var _ simplejson.Handler = &testHandler{}

var (
	queryResponses = map[string]*query.TimeSeriesResponse{
		"A": {
			Target: "A",
			DataPoints: []query.DataPoint{
				{Timestamp: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), Value: 100},
				{Timestamp: time.Date(2020, 1, 1, 0, 1, 0, 0, time.UTC), Value: 101},
				{Timestamp: time.Date(2020, 1, 1, 0, 2, 0, 0, time.UTC), Value: 103},
			},
		},
		"B": {
			Target: "B",
			DataPoints: []query.DataPoint{
				{Timestamp: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), Value: 100},
				{Timestamp: time.Date(2020, 1, 1, 0, 1, 0, 0, time.UTC), Value: 99},
				{Timestamp: time.Date(2020, 1, 1, 0, 2, 0, 0, time.UTC), Value: 98},
			},
		},
	}

	tableQueryResponse = map[string]*query.TableResponse{
		"C": {
			Columns: []query.Column{
				{Text: "Time", Data: query.TimeColumn{
					time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2020, 1, 1, 0, 1, 0, 0, time.UTC),
				}},
				{Text: "Label", Data: query.StringColumn{"foo", "bar"}},
				{Text: "Series A", Data: query.NumberColumn{42, 43}},
				{Text: "Series B", Data: query.NumberColumn{64.5, 100.0}},
			},
		},
	}

	annotations = []annotation.Annotation{{
		Time:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		Title: "foo",
		Text:  "bar",
		Tags:  []string{"snafu"},
	}}

	tags = []string{"foo", "bar"}

	tagValues = map[string][]string{
		"foo": {"A", "B"},
		"bar": {"1", "2"},
	}

	handlers = map[string]simplejson.Handler{
		"A": &testHandler{
			queryResponse: queryResponses["A"],
			annotations:   annotations,
			tags:          tags,
			tagValues:     tagValues,
		},
		"B": &testHandler{
			queryResponse: queryResponses["B"],
		},
		"C": &testHandler{
			queryResponse: tableQueryResponse["C"],
		},
	}
)

func (handler *testHandler) Endpoints() (endpoints simplejson.Endpoints) {
	if handler.noEndpoints {
		return
	}
	if handler.queryResponse != nil {
		endpoints.Query = handler.Query
	}
	if len(handler.annotations) > 0 {
		endpoints.Annotations = handler.Annotations
	}
	if len(handler.tags) > 0 {
		endpoints.TagKeys = handler.Tags
	}
	if len(handler.tagValues) > 0 {
		endpoints.TagValues = handler.TagValues
	}
	return
}

func (handler *testHandler) Query(_ context.Context, _ query.Request) (response query.Response, err error) {
	return handler.queryResponse, nil
}

func (handler *testHandler) Annotations(_ annotation.Request) (annotations []annotation.Annotation, err error) {
	return handler.annotations, nil
}

func (handler *testHandler) Tags(_ context.Context) (tags []string) {
	return handler.tags
}

func (handler *testHandler) TagValues(_ context.Context, tag string) (values []string, err error) {
	var ok bool
	if values, ok = handler.tagValues[tag]; ok == false {
		err = fmt.Errorf("unsupported tag '%s'", tag)
	}
	return
}
