package main

import (
	"errors"
	"github.com/clambin/grafana-json"
	"time"
)

func main() {
	handler := grafana_json.Handler{
		Search:     Search,
		Query:      Query,
		TableQuery: TableQuery,
	}
	s := grafana_json.Create(handler, 8081)

	_ = s.Run()
}

func Search() []string {
	return []string{"series", "table"}
}

func Query(target string, _ *grafana_json.TimeSeriesQueryArgs) (response *grafana_json.QueryResponse, err error) {
	if target != "series" {
		err = errors.New("unsupported series")
		return
	}

	timestamp := time.Now().Add(-1 * time.Hour)

	response = new(grafana_json.QueryResponse)
	response.DataPoints = make([]grafana_json.QueryResponseDataPoint, 0)

	for i := 0; i < 100; i++ {
		response.DataPoints[i] = grafana_json.QueryResponseDataPoint{
			Timestamp: timestamp,
			Value:     int64(i),
		}
		timestamp = timestamp.Add(1 * time.Second)
	}
	return
}

func TableQuery(target string, _ *grafana_json.TableQueryArgs) (response *grafana_json.TableQueryResponse, err error) {
	if target != "table" {
		err = errors.New("unsupported series")
	}

	timestamps := make(grafana_json.TableQueryResponseTimeColumn, 100)
	seriesA := make(grafana_json.TableQueryResponseNumberColumn, 100)
	seriesB := make(grafana_json.TableQueryResponseNumberColumn, 100)

	timestamp := time.Now().Add(-1 * time.Hour)
	for i := 0; i < 100; i++ {
		timestamps[i] = timestamp
		seriesA[i] = float64(i)
		seriesB[i] = float64(-i)
		timestamp = timestamp.Add(1 * time.Second)
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
