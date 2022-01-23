package simplejson_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/clambin/simplejson/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

var Port int

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

	m.Run()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	err = s.Shutdown(ctx, time.Second)
	if err != nil {
		panic(err)
	}
	cancel()
}

func TestServer_Metrics(t *testing.T) {
	serverRunning(t)

	body, err := call(Port, "/metrics", http.MethodGet, "")
	require.NoError(t, err)
	assert.Contains(t, body, "http_duration_seconds")
	assert.Contains(t, body, "http_duration_seconds_sum")
	assert.Contains(t, body, "http_duration_seconds_count")
}

func BenchmarkAPIServer(b *testing.B) {
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
		return "", errors.New(resp.Status)
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

type testAPIHandler struct {
	noEndpoints bool

	queryResponse      *simplejson.TimeSeriesResponse
	tableQueryResponse *simplejson.TableQueryResponse
	annotations        []simplejson.Annotation
	tags               []string
	tagValues          map[string][]string
}

var _ simplejson.Handler = &testAPIHandler{}

var (
	queryResponses = map[string]*simplejson.TimeSeriesResponse{
		"A": {
			Target: "A",
			DataPoints: []simplejson.DataPoint{
				{Timestamp: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), Value: 100},
				{Timestamp: time.Date(2020, 1, 1, 0, 1, 0, 0, time.UTC), Value: 101},
				{Timestamp: time.Date(2020, 1, 1, 0, 2, 0, 0, time.UTC), Value: 103},
			},
		},
		"B": {
			Target: "B",
			DataPoints: []simplejson.DataPoint{
				{Timestamp: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), Value: 100},
				{Timestamp: time.Date(2020, 1, 1, 0, 1, 0, 0, time.UTC), Value: 99},
				{Timestamp: time.Date(2020, 1, 1, 0, 2, 0, 0, time.UTC), Value: 98},
			},
		},
	}

	tableQueryResponse = map[string]*simplejson.TableQueryResponse{
		"C": {
			Columns: []simplejson.TableQueryResponseColumn{
				{Text: "Time", Data: simplejson.TableQueryResponseTimeColumn{
					time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2020, 1, 1, 0, 1, 0, 0, time.UTC),
				}},
				{Text: "Label", Data: simplejson.TableQueryResponseStringColumn{"foo", "bar"}},
				{Text: "Series A", Data: simplejson.TableQueryResponseNumberColumn{42, 43}},
				{Text: "Series B", Data: simplejson.TableQueryResponseNumberColumn{64.5, 100.0}},
			},
		},
	}

	annotations = []simplejson.Annotation{{
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
		"A": &testAPIHandler{
			queryResponse: queryResponses["A"],
			annotations:   annotations,
			tags:          tags,
			tagValues:     tagValues,
		},
		"B": &testAPIHandler{
			queryResponse: queryResponses["B"],
		},
		"C": &testAPIHandler{
			tableQueryResponse: tableQueryResponse["C"],
		},
	}
)

func (handler *testAPIHandler) Endpoints() (endpoints simplejson.Endpoints) {
	if handler.noEndpoints {
		return
	}
	if handler.queryResponse != nil {
		endpoints.Query = handler.Query
	}
	if handler.tableQueryResponse != nil {
		endpoints.TableQuery = handler.TableQuery
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

func (handler *testAPIHandler) Query(_ context.Context, _ *simplejson.TimeSeriesQueryArgs) (response *simplejson.TimeSeriesResponse, err error) {
	return handler.queryResponse, nil
}

func (handler *testAPIHandler) TableQuery(_ context.Context, _ *simplejson.TableQueryArgs) (response *simplejson.TableQueryResponse, err error) {
	return handler.tableQueryResponse, nil
}

func (handler *testAPIHandler) Annotations(_, _ string, _ *simplejson.AnnotationRequestArgs) (annotations []simplejson.Annotation, err error) {
	return handler.annotations, nil
}

func (handler *testAPIHandler) Tags(_ context.Context) (tags []string) {
	return handler.tags
}

func (handler *testAPIHandler) TagValues(_ context.Context, tag string) (values []string, err error) {
	var ok bool
	if values, ok = handler.tagValues[tag]; ok == false {
		err = fmt.Errorf("unsupported tag '%s'", tag)
	}
	return
}
