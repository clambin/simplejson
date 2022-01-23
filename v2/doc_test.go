package simplejson_test

import (
	"context"
	"github.com/clambin/simplejson/v2"
	"github.com/clambin/simplejson/v2/query"
	"time"
)

type handler struct{}

func (h handler) Endpoints() simplejson.Endpoints {
	return simplejson.Endpoints{
		Query:      h.Query,
		TableQuery: h.TableQuery,
	}
}

func (h *handler) Query(_ context.Context, _ *query.Args) (response *query.TimeSeriesResponse, err error) {
	response = &query.TimeSeriesResponse{
		DataPoints: make([]query.DataPoint, 60),
	}

	timestamp := time.Now().Add(-1 * time.Hour)
	for i := 0; i < 60; i++ {
		response.DataPoints[i] = query.DataPoint{
			Timestamp: timestamp,
			Value:     int64(i),
		}
		timestamp = timestamp.Add(1 * time.Minute)
	}
	return
}

func (h *handler) TableQuery(_ context.Context, _ *query.Args) (response *query.TableResponse, err error) {
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

	response = &query.TableResponse{
		Columns: []query.Column{
			{Text: "timestamp", Data: timestamps},
			{Text: "series A", Data: seriesA},
			{Text: "series B", Data: seriesB},
		},
	}

	return
}

func Example() {
	s := simplejson.Server{
		Handlers: map[string]simplejson.Handler{
			"A": &handler{},
		},
	}

	_ = s.Run(8088)
}
