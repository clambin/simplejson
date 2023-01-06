package simplejson_test

import (
	"bytes"
	"context"
	"fmt"
	"github.com/clambin/go-common/httpserver/middleware"
	"github.com/clambin/simplejson/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewRouter(t *testing.T) {
	r := simplejson.NewRouter(nil)

	for _, path := range []string{"/"} {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, path, nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	for _, path := range []string{"/search", "/query", "/annotations", "/tag-keys", "/tag-values"} {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, path, nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	for _, path := range []string{"/annotations"} {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodOptions, path, nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestNewRouter_Extend(t *testing.T) {
	r := simplejson.NewRouter(nil)

	r.Post("/test", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("Hello"))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Hello", w.Body.String())
}

func TestNewRouter_PrometheusMetrics(t *testing.T) {
	r := simplejson.NewRouter(nil, simplejson.WithHTTPMetrics{Option: middleware.PrometheusMetricsOptions{
		Namespace:   "foo",
		Subsystem:   "bar",
		Application: "snafu",
	}})

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/search", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	n, err := testutil.GatherAndCount(reg)
	require.NoError(t, err)
	assert.Equal(t, 2, n)
}

func TestNewRouter_QueryMetrics(t *testing.T) {
	r := simplejson.NewRouter(handlers, simplejson.WithQueryMetrics{Name: "foobar"})

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/query", bytes.NewBufferString(`{ "targets": [ { "target": "A" } ] }`))
	r.ServeHTTP(w, req)
	if !assert.Equal(t, http.StatusOK, w.Code) {
		t.Log(w.Body)
	}

	n, err := testutil.GatherAndCount(reg)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
}

//
//
// Test Handler
//

type testHandler struct {
	noEndpoints bool

	queryResponse simplejson.Response
	annotations   []simplejson.Annotation
	tags          []string
	tagValues     map[string][]string
}

var _ simplejson.Handler = &testHandler{}

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

	tableQueryResponse = map[string]*simplejson.TableResponse{
		"C": {
			Columns: []simplejson.Column{
				{Text: "Time", Data: simplejson.TimeColumn{
					time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2020, 1, 1, 0, 1, 0, 0, time.UTC),
				}},
				{Text: "Label", Data: simplejson.StringColumn{"foo", "bar"}},
				{Text: "Series A", Data: simplejson.NumberColumn{42, 43}},
				{Text: "Series B", Data: simplejson.NumberColumn{64.5, 100.0}},
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

	s, _ = simplejson.New(handlers)
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

func (handler *testHandler) Query(_ context.Context, _ simplejson.QueryRequest) (response simplejson.Response, err error) {
	return handler.queryResponse, nil
}

func (handler *testHandler) Annotations(_ simplejson.AnnotationRequest) (annotations []simplejson.Annotation, err error) {
	return handler.annotations, nil
}

func (handler *testHandler) Tags(_ context.Context) (tags []string) {
	return handler.tags
}

func (handler *testHandler) TagValues(_ context.Context, tag string) (values []string, err error) {
	var ok bool
	if values, ok = handler.tagValues[tag]; !ok {
		err = fmt.Errorf("unsupported tag '%s'", tag)
	}
	return
}
