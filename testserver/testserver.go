package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/clambin/simplejson/v3"
	"github.com/clambin/simplejson/v3/annotation"
	"github.com/clambin/simplejson/v3/query"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

func main() {
	s := simplejson.Server{Handlers: map[string]simplejson.Handler{
		"A": &handler{},
		"B": &handler{table: true},
		"C": &annoHandler{},
	}}

	log.SetLevel(log.DebugLevel)
	err := s.Run(8080)
	if errors.Is(err, http.ErrServerClosed) == false {
		panic(err)
	}
}

type handler struct{ table bool }

func (h handler) Endpoints() simplejson.Endpoints {
	return simplejson.Endpoints{
		Query:       h.Query,
		Annotations: h.Annotations,
		TagKeys:     h.TagKeys,
		TagValues:   h.TagValues,
	}
}

func (h *handler) Query(ctx context.Context, req query.Request) (response query.Response, err error) {
	if h.table == false {
		return h.timeSeriesQuery(ctx, req)
	}
	return h.tableQuery(ctx, req)
}

func (h *handler) timeSeriesQuery(_ context.Context, req query.Request) (response *query.TimeSeriesResponse, err error) {
	for _, filter := range req.Args.AdHocFilters {
		log.WithFields(log.Fields{
			"key":       filter.Key,
			"operator":  filter.Operator,
			"condition": filter.Condition,
			"value":     filter.Value,
		}).Info("table request received")
	}

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

func (h *handler) tableQuery(_ context.Context, req query.Request) (response *query.TableResponse, err error) {
	for _, filter := range req.Args.AdHocFilters {
		log.WithFields(log.Fields{
			"key":       filter.Key,
			"operator":  filter.Operator,
			"condition": filter.Condition,
			"value":     filter.Value,
		}).Info("table request received")
	}

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

func (h *handler) Annotations(_ annotation.Request) (annotations []annotation.Annotation, err error) {
	annotations = []annotation.Annotation{
		{
			Time:  time.Now().Add(-5 * time.Minute),
			Title: "foo",
			Text:  "bar",
			Tags:  []string{"A", "B"},
		},
	}
	return
}

func (h *handler) TagKeys(_ context.Context) (keys []string) {
	return []string{"some-key"}
}

func (h *handler) TagValues(_ context.Context, key string) (values []string, err error) {
	if key != "some-key" {
		return nil, fmt.Errorf("invalid key: %s", key)
	}
	return []string{"A", "B", "C"}, nil
}

type annoHandler struct {
}

func (a annoHandler) Endpoints() simplejson.Endpoints {
	return simplejson.Endpoints{Query: a.annotations}
}

func (annoHandler) annotations(_ context.Context, _ query.Request) (response query.Response, err error) {
	response = &query.TableResponse{Columns: []query.Column{
		{Text: "start", Data: query.TimeColumn([]time.Time{time.Now().Add(-5 * time.Minute)})},
		{Text: "stop", Data: query.TimeColumn([]time.Time{time.Now().Add(-4 * time.Minute)})},
		{Text: "title", Data: query.StringColumn([]string{"bar"})},
		{Text: "name", Data: query.StringColumn([]string{"foo"})},
		{Text: "id", Data: query.NumberColumn([]float64{1.0})},
		{Text: "tags", Data: query.StringColumn([]string{"A"})},
	}}

	return
}
