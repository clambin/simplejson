package simplejson_test

import (
	"context"
	"fmt"
	"github.com/clambin/simplejson/v3"
	"github.com/clambin/simplejson/v3/annotation"
	"github.com/clambin/simplejson/v3/query"
	"time"
)

func Example() {
	s := simplejson.Server{
		Handlers: map[string]simplejson.Handler{
			"A": &handler{},
			"B": &handler{table: true},
		},
	}

	_ = s.Run(8088)
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

func (h *handler) Query(ctx context.Context, req query.Request) (query.Response, error) {
	if h.table == false {
		return h.timeSeriesQuery(ctx, req)
	}
	return h.tableQuery(ctx, req)
}

func (h *handler) timeSeriesQuery(_ context.Context, _ query.Request) (*query.TimeSeriesResponse, error) {
	dataPoints := make([]query.DataPoint, 60)
	timestamp := time.Now().Add(-1 * time.Hour)
	for i := 0; i < 60; i++ {
		dataPoints[i] = query.DataPoint{
			Timestamp: timestamp,
			Value:     int64(i),
		}
		timestamp = timestamp.Add(1 * time.Minute)
	}

	return &query.TimeSeriesResponse{
		DataPoints: dataPoints,
	}, nil
}

func (h *handler) tableQuery(_ context.Context, _ query.Request) (*query.TableResponse, error) {
	timestamps := make(query.TimeColumn, 60)
	seriesA := make(query.NumberColumn, 60)
	seriesB := make(query.NumberColumn, 60)

	timestamp := time.Now().Add(-1 * time.Hour)
	for i := 0; i < 60; i++ {
		timestamps[i] = timestamp
		seriesA[i] = float64(i)
		seriesB[i] = float64(-i)
		timestamp = timestamp.Add(1 * time.Minute)
	}

	return &query.TableResponse{
		Columns: []query.Column{
			{Text: "timestamp", Data: timestamps},
			{Text: "series A", Data: seriesA},
			{Text: "series B", Data: seriesB},
		},
	}, nil
}

func (h *handler) Annotations(_ annotation.Request) ([]annotation.Annotation, error) {
	return []annotation.Annotation{{
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
