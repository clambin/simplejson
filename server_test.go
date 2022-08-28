package simplejson_test

import (
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
	"testing"
	"time"
)

var (
	update = flag.Bool("update", false, "update .golden files")
	s      = simplejson.Server{Handlers: handlers}
)

func TestServer_Metrics(t *testing.T) {
	listener, err := net.Listen("tcp4", ":0")
	if err != nil {
		panic(err)
	}
	target := fmt.Sprintf("http://127.0.0.1:%d/metrics", listener.Addr().(*net.TCPAddr).Port)

	httpServer := http.Server{Handler: s.GetRouter()}
	go func() {
		err2 := httpServer.Serve(listener)
		if !errors.Is(err2, http.ErrServerClosed) {
			panic(err2)
		}
	}()

	var resp *http.Response
	require.Eventually(t, func() bool {
		resp, err = http.Post(target, "", nil)
		if err != nil {
			return false
		}
		_ = resp.Body.Close()
		return resp.StatusCode == http.StatusOK

	}, time.Second, 10*time.Millisecond)

	resp, err = http.Get(target)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body []byte
	body, err = io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	require.NoError(t, err)
	assert.Contains(t, string(body), "http_duration_seconds")
	assert.Contains(t, string(body), "http_duration_seconds_sum")
	assert.Contains(t, string(body), "http_duration_seconds_count")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	err = s.Shutdown(ctx, time.Second)
	if err != nil {
		panic(err)
	}
	cancel()
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
