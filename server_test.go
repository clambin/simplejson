package grafana_json_test

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/clambin/grafana-json"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func TestAPIServer_TimeSeries(t *testing.T) {
	s := grafana_json.Create(newAPIHandler(), 8080)

	go func() {
		err := s.Run()

		assert.Nil(t, err)
	}()

	time.Sleep(1 * time.Second)

	body, err := call("http://localhost:8080/", "GET", "")
	if assert.Nil(t, err) {
		assert.Equal(t, "Hello", body)
	}

	body, err = call("http://localhost:8080/metrics", "GET", "")
	if assert.Nil(t, err) {
		assert.Contains(t, body, "grafana_api_duration_seconds")
		assert.Contains(t, body, "grafana_api_duration_seconds_sum")
		assert.Contains(t, body, "grafana_api_duration_seconds_count")
	}

	body, err = call("http://localhost:8080/search", "POST", "")
	if assert.Nil(t, err) {
		assert.Equal(t, `["A","B","C","Crash"]`, body)
	}

	req := `{
	"maxDataPoints": 100,
	"interval": "1y",
	"range": {
		"from": "2020-01-01T00:00:00.000Z",
		"to": "2020-12-31T00:00:00.000Z"
	},
	"targets": [
		{ "target": "A", "type": "timeserie" },
		{ "target": "B" }
	]
}`
	body, err = call("http://localhost:8080/query", "POST", req)

	if assert.Nil(t, err) {
		assert.Equal(t, `[{"target":"A","datapoints":[[100,1577836800000],[101,1577836860000],[103,1577836920000]]},{"target":"B","datapoints":[[100,1577836800000],[99,1577836860000],[98,1577836920000]]}]`, body)
	}
}

func TestAPIServer_Table(t *testing.T) {
	s := grafana_json.Create(newAPIHandler(), 8088)

	go func() {
		err := s.Run()

		assert.Nil(t, err)
	}()

	time.Sleep(1 * time.Second)

	body, err := call("http://localhost:8088/search", "POST", "")
	if assert.Nil(t, err) {
		assert.Equal(t, `["A","B","C","Crash"]`, body)
	}

	req := `{
	"maxDataPoints": 100,
	"interval": "1y",
	"range": {
		"from": "2020-01-01T00:00:00.000Z",
		"to": "2020-12-31T00:00:00.000Z"
	},
	"targets": [
		{ "target": "C", "type": "table" }
	]
}`
	body, err = call("http://localhost:8080/query", "POST", req)

	if assert.Nil(t, err) {
		assert.Equal(t, `[{"type":"table","columns":[{"text":"Time","type":"time"},{"text":"Label","type":"string"},{"text":"Series A","type":"number"},{"text":"Series B","type":"number"}],"rows":[["2020-01-01T00:00:00Z","foo",42,64.5],["2020-01-01T00:01:00Z","bar",43,100]]}]`, body)
	}
}

func BenchmarkAPIServer(b *testing.B) {
	s := grafana_json.Create(newAPIHandler(), 8082)

	go func() {
		err := s.Run()

		assert.Nil(b, err)
	}()

	time.Sleep(1 * time.Second)

	req := `{
	"maxDataPoints": 100,
	"interval": "1y",
	"range": {
		"from": "2020-01-01T00:00:00.000Z",
		"to": "2020-12-31T00:00:00.000Z"
	},
	"targets": [
		{ "target": "A" },
		{ "target": "B" }
	]
}`
	var body string
	var err error

	b.ResetTimer()
	for i := 0; i < 10000; i++ {
		body, err = call("http://localhost:8082/query", "POST", req)
	}

	if assert.Nil(b, err) {
		assert.Equal(b, `[{"target":"A","datapoints":[[100,1577836800000],[101,1577836860000],[103,1577836920000]]},{"target":"B","datapoints":[[100,1577836800000],[99,1577836860000],[98,1577836920000]]}]`, body)
	}

}

func call(url, method, body string) (string, error) {
	client := &http.Client{}
	reqBody := bytes.NewBuffer([]byte(body))
	req, _ := http.NewRequest(method, url, reqBody)
	resp, err := client.Do(req)

	if err == nil {
		defer resp.Body.Close()
		if body, err := ioutil.ReadAll(resp.Body); err == nil {
			return string(body), nil
		}
	}

	return "", err
}

//
//
// Test APIHandler
//

type testAPIHandler struct {
}

func newAPIHandler() *testAPIHandler {
	return &testAPIHandler{}
}

func (handler *testAPIHandler) Search() []string {
	return []string{"A", "B", "C", "Crash"}
}

func (handler *testAPIHandler) Query(target string, _ *grafana_json.QueryRequest) (response *grafana_json.QueryResponse, err error) {
	switch target {
	case "A":
		response = &grafana_json.QueryResponse{
			Target: "A",
			DataPoints: []grafana_json.QueryResponseDataPoint{
				{Timestamp: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), Value: 100},
				{Timestamp: time.Date(2020, 1, 1, 0, 1, 0, 0, time.UTC), Value: 101},
				{Timestamp: time.Date(2020, 1, 1, 0, 2, 0, 0, time.UTC), Value: 103},
			},
		}
	case "B":
		response = &grafana_json.QueryResponse{
			Target: "B",
			DataPoints: []grafana_json.QueryResponseDataPoint{
				{Timestamp: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), Value: 100},
				{Timestamp: time.Date(2020, 1, 1, 0, 1, 0, 0, time.UTC), Value: 99},
				{Timestamp: time.Date(2020, 1, 1, 0, 2, 0, 0, time.UTC), Value: 98},
			},
		}
	case "Crash":
		return response, errors.New("server crash")
	}

	return
}

func (handler *testAPIHandler) QueryTable(target string, _ *grafana_json.QueryRequest) (response *grafana_json.QueryTableResponse, err error) {
	switch target {
	case "C":
		response = &grafana_json.QueryTableResponse{Columns: []grafana_json.QueryTableResponseColumn{
			{Text: "Time", Data: grafana_json.QueryTableResponseTimeColumn{
				time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2020, 1, 1, 0, 1, 0, 0, time.UTC),
			}},
			{Text: "Label", Data: grafana_json.QueryTableResponseStringColumn{"foo", "bar"}},
			{Text: "Series A", Data: grafana_json.QueryTableResponseNumberColumn{42, 43}},
			{Text: "Series B", Data: grafana_json.QueryTableResponseNumberColumn{64.5, 100.0}},
		}}
	default:
		err = fmt.Errorf("not implemented")
	}
	return
}
