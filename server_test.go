package simplejson

import (
	"context"
	"flag"
	"fmt"
	"github.com/clambin/simplejson/v3/annotation"
	"github.com/clambin/simplejson/v3/query"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

var update = flag.Bool("update", false, "update .golden files")

func TestServer_Run_Shutdown(t *testing.T) {
	r := prometheus.NewRegistry()
	srv, err := NewWithRegisterer("test", handlers, r)
	require.NoError(t, err)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err2 := srv.Run()
		assert.NoError(t, err2)
		wg.Done()
	}()

	assert.Eventually(t, func() bool {
		var resp *http.Response
		resp, err = http.Get(fmt.Sprintf("http://localhost:%d/", srv.httpServer.GetPort()))
		if err != nil {
			return false
		}
		_ = resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, time.Second, time.Millisecond)

	var resp *http.Response
	resp, err = http.Post(fmt.Sprintf("http://localhost:%d/search", srv.httpServer.GetPort()), "", nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()

	err = srv.Shutdown(5 * time.Second)
	require.NoError(t, err)
	wg.Wait()
}

func TestServer_Metrics(t *testing.T) {
	r := prometheus.NewRegistry()
	srv, err := NewWithRegisterer("foobar", nil, r)
	require.NoError(t, err)

	req, _ := http.NewRequest(http.MethodPost, "/search", nil)
	w := httptest.NewRecorder()
	srv.httpServer.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	m, err := r.Gather()
	require.NoError(t, err)
	require.Len(t, m, 2)
	for _, metric := range m {
		require.Len(t, metric.GetMetric(), 1)
		switch metric.GetName() {
		case "http_requests_duration_seconds":
			assert.Equal(t, uint64(1), metric.GetMetric()[0].Histogram.GetSampleCount())
		case "http_requests_total":
			assert.Equal(t, 1.0, metric.GetMetric()[0].Counter.GetValue())
		default:
			t.Fatalf("unexpected metric: %s", metric.GetName())
		}
	}
}

/*
func TestServer_Metrics(t *testing.T) {
	r := s.GetRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	m, err := prometheus.DefaultGatherer.Gather()
	require.NoError(t, err)
	var found bool
	for _, entry := range m {
		if *entry.Name == "simplejson_request_duration_seconds" {
			require.Equal(t, pcg.MetricType_SUMMARY, *entry.Type)
			require.NotZero(t, entry.Metric)
			assert.NotZero(t, entry.Metric[0].Summary.GetSampleCount())
			found = true
			break
		}
	}
	assert.True(t, found)
}
*/

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

var _ Handler = &testHandler{}

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

	handlers = map[string]Handler{
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

	s, _ = New("foo", handlers)
)

func (handler *testHandler) Endpoints() (endpoints Endpoints) {
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
