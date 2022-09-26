package simplejson_test

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/clambin/simplejson/v3"
	"github.com/clambin/simplejson/v3/annotation"
	"github.com/clambin/simplejson/v3/query"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

var (
	update = flag.Bool("update", false, "update .golden files")
	s      = simplejson.Server{Handlers: handlers}
)

func TestServer_Run_Shutdown(t *testing.T) {
	wg := sync.WaitGroup{}
	srv := simplejson.Server{Handlers: handlers}
	wg.Add(1)
	go func() {
		err := srv.Run(0)
		require.True(t, errors.Is(err, http.ErrServerClosed))
		wg.Done()
	}()

	// FIXME: race condition between Run setting HTTPServer and code below using it
	time.Sleep(time.Second)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/", nil)
	srv.HTTPServer.Handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	if err := srv.Shutdown(context.Background(), 5*time.Second); err != nil {
		panic(err)
	}
	wg.Wait()
}

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
		if *entry.Name == "http_duration_seconds" {
			require.Equal(t, io_prometheus_client.MetricType_SUMMARY, *entry.Type)
			require.Len(t, entry.Metric, 1)
			assert.NotZero(t, entry.Metric[0].Summary.GetSampleCount())
			found = true
			break
		}
	}
	assert.True(t, found)
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
