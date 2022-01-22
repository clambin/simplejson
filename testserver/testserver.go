package main

import (
	"context"
	"github.com/clambin/simplejson"
	log "github.com/sirupsen/logrus"
	"time"
)

func main() {
	s := simplejson.Server{Handlers: map[string]simplejson.Handler{
		"A": &handler{},
	}}

	log.SetLevel(log.DebugLevel)
	_ = s.Run(8088)
}

type handler struct{}

func (h *handler) Endpoints() simplejson.Endpoints {
	return simplejson.Endpoints{
		Query:       h.Query,
		TableQuery:  h.TableQuery,
		Annotations: h.Annotations,
		TagKeys:     h.TagKeys,
		TagValues:   h.TagValues,
	}
}

func (h *handler) Query(_ context.Context, req *simplejson.TimeSeriesQueryArgs) (response *simplejson.TimeSeriesResponse, err error) {
	for _, filter := range req.AdHocFilters {
		log.WithFields(log.Fields{
			"key":       filter.Key,
			"operator":  filter.Operator,
			"condition": filter.Condition,
			"value":     filter.Value,
		}).Info("table request received")
	}

	response = &simplejson.TimeSeriesResponse{}
	response.DataPoints = make([]simplejson.DataPoint, 60)

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

func (h *handler) TableQuery(_ context.Context, req *simplejson.TableQueryArgs) (response *simplejson.TableQueryResponse, err error) {
	for _, filter := range req.AdHocFilters {
		log.WithFields(log.Fields{
			"key":       filter.Key,
			"operator":  filter.Operator,
			"condition": filter.Condition,
			"value":     filter.Value,
		}).Info("table request received")
	}

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

	response = new(simplejson.TableQueryResponse)
	response.Columns = []simplejson.TableQueryResponseColumn{
		{
			Text: "timestamp",
			Data: timestamps,
		},
		{
			Text: "series A",
			Data: seriesA,
		},
		{
			Text: "series B",
			Data: seriesB,
		},
	}
	return
}

func (h *handler) Annotations(_, _ string, _ *simplejson.AnnotationRequestArgs) (annotations []simplejson.Annotation, err error) {
	annotations = append(annotations, simplejson.Annotation{
		Time:  time.Now().Add(-5 * time.Minute),
		Title: "foo",
		Text:  "bar",
	})

	return
}

func (h *handler) TagKeys(_ context.Context) (keys []string) {
	return []string{"foo"}
}

func (h *handler) TagValues(_ context.Context, _ string) (values []string, err error) {
	return []string{"A", "B", "C"}, nil
}
