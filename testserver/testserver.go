package main

import (
	"context"
	"errors"
	"github.com/clambin/grafana-json"
	log "github.com/sirupsen/logrus"
	"time"
)

func main() {
	h := &handler{}
	s := grafana_json.Create(h, 8088)

	log.SetLevel(log.DebugLevel)
	_ = s.Run()
}

type handler struct{}

func (h *handler) Endpoints() grafana_json.Endpoints {
	return grafana_json.Endpoints{
		Search:      h.Search,
		Query:       h.Query,
		TableQuery:  h.TableQuery,
		Annotations: h.Annotations,
	}
}

func (h *handler) Search() []string {
	return []string{"series", "table"}
}

func (h *handler) Query(_ context.Context, target string, _ *grafana_json.TimeSeriesQueryArgs) (response *grafana_json.QueryResponse, err error) {
	if target != "series" {
		err = errors.New("unsupported series")
		return
	}

	response = new(grafana_json.QueryResponse)
	response.DataPoints = make([]grafana_json.QueryResponseDataPoint, 60)

	timestamp := time.Now().Add(-1 * time.Hour)
	for i := 0; i < 60; i++ {
		response.DataPoints[i] = grafana_json.QueryResponseDataPoint{
			Timestamp: timestamp,
			Value:     int64(i),
		}
		timestamp = timestamp.Add(1 * time.Minute)
	}
	return
}

func (h *handler) TableQuery(_ context.Context, target string, _ *grafana_json.TableQueryArgs) (response *grafana_json.TableQueryResponse, err error) {
	if target != "table" {
		err = errors.New("unsupported series")
		return
	}

	timestamps := make(grafana_json.TableQueryResponseTimeColumn, 60)
	seriesA := make(grafana_json.TableQueryResponseNumberColumn, 60)
	seriesB := make(grafana_json.TableQueryResponseNumberColumn, 60)

	timestamp := time.Now().Add(-1 * time.Hour)
	for i := 0; i < 60; i++ {
		timestamps[i] = timestamp
		seriesA[i] = float64(i)
		seriesB[i] = float64(-i)
		timestamp = timestamp.Add(1 * time.Minute)
	}

	response = new(grafana_json.TableQueryResponse)
	response.Columns = []grafana_json.TableQueryResponseColumn{
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

func (h *handler) Annotations(_, _ string, _ *grafana_json.AnnotationRequestArgs) (annotations []grafana_json.Annotation, err error) {
	annotations = append(annotations, grafana_json.Annotation{
		Time:  time.Now().Add(-5 * time.Minute),
		Title: "foo",
		Text:  "bar",
	})

	return
}
