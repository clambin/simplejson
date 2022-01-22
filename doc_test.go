package simplejson_test

import (
	"context"
	"github.com/clambin/simplejson"
	"time"
)

type handler struct{}

func (h handler) Endpoints() simplejson.Endpoints {
	return simplejson.Endpoints{
		Query:      h.Query,
		TableQuery: h.TableQuery,
	}
}

func (h *handler) Query(_ context.Context, _ *simplejson.TimeSeriesQueryArgs) (response *simplejson.TimeSeriesResponse, err error) {
	response = &simplejson.TimeSeriesResponse{
		DataPoints: make([]simplejson.DataPoint, 60),
	}

	timestamp := time.Now().Add(-1 * time.Hour)
	for i := 0; i < 60; i++ {
		response.DataPoints[i] = simplejson.DataPoint{
			Timestamp: timestamp,
			Value:     int64(i),
		}
		timestamp = timestamp.Add(1 * time.Minute)
	}
	return
}

func (h *handler) TableQuery(_ context.Context, _ *simplejson.TableQueryArgs) (response *simplejson.TableQueryResponse, err error) {
	timestamps := make(simplejson.TableQueryResponseTimeColumn, 60)
	seriesA := make(simplejson.TableQueryResponseNumberColumn, 60)
	seriesB := make(simplejson.TableQueryResponseNumberColumn, 60)

	timestamp := time.Now().Add(-1 * time.Hour)
	for i := 0; i < 60; i++ {
		timestamps[i] = timestamp
		seriesA[i] = float64(i)
		seriesB[i] = float64(-i)
		timestamp = timestamp.Add(1 * time.Minute)
	}

	response = &simplejson.TableQueryResponse{
		Columns: []simplejson.TableQueryResponseColumn{
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
