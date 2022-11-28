package simplejson_test

import (
	"context"
	"fmt"
	"github.com/clambin/simplejson/v4"
	"time"
)

func Example() {
	s, err := simplejson.New("test", map[string]simplejson.Handler{
		"A": &handler{},
		"B": &handler{table: true},
	})

	if err == nil {
		_ = s.Run()
	}
}

type handler struct{ table bool }

func (h *handler) Endpoints() simplejson.Endpoints {
	return simplejson.Endpoints{
		Query:       h.Query,
		Annotations: h.Annotations,
		TagKeys:     h.TagKeys,
		TagValues:   h.TagValues,
	}
}

func (h *handler) Query(ctx context.Context, req simplejson.QueryRequest) (simplejson.Response, error) {
	if h.table == false {
		return h.timeSeriesQuery(ctx, req)
	}
	return h.tableQuery(ctx, req)
}

func (h *handler) timeSeriesQuery(_ context.Context, _ simplejson.QueryRequest) (simplejson.TimeSeriesResponse, error) {
	dataPoints := make([]simplejson.DataPoint, 60)
	timestamp := time.Now().Add(-1 * time.Hour)
	for i := 0; i < 60; i++ {
		dataPoints[i] = simplejson.DataPoint{
			Timestamp: timestamp,
			Value:     float64(i),
		}
		timestamp = timestamp.Add(1 * time.Minute)
	}

	return simplejson.TimeSeriesResponse{
		DataPoints: dataPoints,
	}, nil
}

func (h *handler) tableQuery(_ context.Context, _ simplejson.QueryRequest) (simplejson.TableResponse, error) {
	timestamps := make(simplejson.TimeColumn, 60)
	seriesA := make(simplejson.NumberColumn, 60)
	seriesB := make(simplejson.NumberColumn, 60)

	timestamp := time.Now().Add(-1 * time.Hour)
	for i := 0; i < 60; i++ {
		timestamps[i] = timestamp
		seriesA[i] = float64(i)
		seriesB[i] = float64(-i)
		timestamp = timestamp.Add(1 * time.Minute)
	}

	return simplejson.TableResponse{Columns: []simplejson.Column{
		{Text: "timestamp", Data: timestamps},
		{Text: "series A", Data: seriesA},
		{Text: "series B", Data: seriesB},
	}}, nil
}

func (h *handler) Annotations(_ simplejson.AnnotationRequest) ([]simplejson.Annotation, error) {
	return []simplejson.Annotation{{
		Time:  time.Now().Add(-5 * time.Minute),
		Title: "foo",
		Text:  "bar",
	}}, nil
}

func (h *handler) TagKeys(_ context.Context) []string {
	return []string{"some-key"}
}

func (h *handler) TagValues(_ context.Context, key string) ([]string, error) {
	if key != "some-key" {
		return nil, fmt.Errorf("invalid key: %s", key)
	}
	return []string{"A", "B", "C"}, nil
}
