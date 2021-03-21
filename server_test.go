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
	handler := testAPIHandler{}
	s := grafana_json.Create(endpoints(&handler), 8080)

	go func() {
		err := s.Run()

		assert.Nil(t, err)
	}()

	assert.Eventually(t, func() bool {
		body, err := call("http://localhost:8080/", "GET", "")
		if assert.Nil(t, err) {
			return assert.Equal(t, "Hello", body)
		}
		return false
	}, 500*time.Millisecond, 10*time.Millisecond)

	body, err := call("http://localhost:8080/metrics", "GET", "")
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
	handler := testAPIHandler{}
	s := grafana_json.Create(endpoints(&handler), 8081)

	go func() {
		err := s.Run()

		assert.Nil(t, err)
	}()

	assert.Eventually(t, func() bool {
		body, err := call("http://localhost:8081/", "GET", "")
		if assert.Nil(t, err) {
			return assert.Equal(t, "Hello", body)
		}
		return false
	}, 500*time.Millisecond, 10*time.Millisecond)

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
	body, err := call("http://localhost:8081/query", "POST", req)

	if assert.Nil(t, err) {
		assert.Equal(t, `[{"type":"table","columns":[{"text":"Time","type":"time"},{"text":"Label","type":"string"},{"text":"Series A","type":"number"},{"text":"Series B","type":"number"}],"rows":[["2020-01-01T00:00:00Z","foo",42,64.5],["2020-01-01T00:01:00Z","bar",43,100]]}]`, body)
	}
}

func TestAPIServer_MissingEndpoint(t *testing.T) {
	handler := testAPIHandler{}
	eps := endpoints(&handler)
	eps.TableQuery = nil
	s := grafana_json.Create(eps, 8082)

	go func() {
		err := s.Run()

		assert.Nil(t, err)
	}()

	assert.Eventually(t, func() bool {
		body, err := call("http://localhost:8082/", "GET", "")
		if assert.Nil(t, err) {
			return assert.Equal(t, "Hello", body)
		}
		return false
	}, 500*time.Millisecond, 10*time.Millisecond)

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
	_, err := call("http://localhost:8082/query", "POST", req)

	assert.NotNil(t, err)
}

func BenchmarkAPIServer(b *testing.B) {
	handler := testAPIHandler{}
	s := grafana_json.Create(endpoints(&handler), 8082)

	go func() {
		err := s.Run()
		assert.Nil(b, err)
	}()

	assert.Eventually(b, func() bool {
		body, err := call("http://localhost:8082/", "GET", "")
		if assert.Nil(b, err) {
			return assert.Equal(b, "Hello", body)
		}
		return false
	}, 500*time.Millisecond, 10*time.Millisecond)

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

func TestAPIServer_Annotations(t *testing.T) {
	handler := testAPIHandler{}
	s := grafana_json.Create(endpoints(&handler), 8084)

	go func() {
		err := s.Run()

		assert.Nil(t, err)
	}()

	assert.Eventually(t, func() bool {
		body, err := call("http://localhost:8084/", "GET", "")
		if assert.Nil(t, err) {
			return assert.Equal(t, "Hello", body)
		}
		return false
	}, 500*time.Millisecond, 10*time.Millisecond)

	body, err := call("http://localhost:8084/annotations", "OPTIONS", "")

	assert.Nil(t, err)
	assert.Equal(t, "", body)

	req := `{
	"range": {
		"from": "2020-01-01T00:00:00.000Z",
		"to": "2020-12-31T00:00:00.000Z"
	},
	"annotation": {
		"name": "snafu",
		"datasource": "fubar",
		"enable": true,
		"query": ""
	}
}`
	body, err = call("http://localhost:8084/annotations", "POST", req)

	if assert.Nil(t, err) {
		assert.Equal(t, `[{"annotation":{"name":"snafu","datasource":"fubar","enable":true,"query":""},"time":1609459200000,"title":"snafu","text":"bar","tags":["snafu"]}]`, body)
	}
}

func call(url, method, body string) (response string, err error) {
	client := &http.Client{}
	reqBody := bytes.NewBuffer([]byte(body))
	req, _ := http.NewRequest(method, url, reqBody)

	var resp *http.Response
	if resp, err = client.Do(req); err == nil {
		defer resp.Body.Close()
		var buff []byte
		if resp.StatusCode == 200 {
			if buff, err = ioutil.ReadAll(resp.Body); err == nil {
				response = string(buff)
			}
		} else {
			err = errors.New(resp.Status)
		}
	}
	return
}

//
//
// Test Handler
//

type testAPIHandler struct {
}

func endpoints(handler *testAPIHandler) grafana_json.Handler {
	return grafana_json.Handler{
		Search:      handler.Search,
		Query:       handler.Query,
		TableQuery:  handler.TableQuery,
		Annotations: handler.Annotations,
	}
}

func (handler *testAPIHandler) Search() []string {
	return []string{"A", "B", "C", "Crash"}
}

func (handler *testAPIHandler) Query(target string, _ *grafana_json.TimeSeriesQueryArgs) (response *grafana_json.QueryResponse, err error) {
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

func (handler *testAPIHandler) TableQuery(target string, _ *grafana_json.TableQueryArgs) (response *grafana_json.TableQueryResponse, err error) {
	switch target {
	case "C":
		response = &grafana_json.TableQueryResponse{Columns: []grafana_json.TableQueryResponseColumn{
			{Text: "Time", Data: grafana_json.TableQueryResponseTimeColumn{
				time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2020, 1, 1, 0, 1, 0, 0, time.UTC),
			}},
			{Text: "Label", Data: grafana_json.TableQueryResponseStringColumn{"foo", "bar"}},
			{Text: "Series A", Data: grafana_json.TableQueryResponseNumberColumn{42, 43}},
			{Text: "Series B", Data: grafana_json.TableQueryResponseNumberColumn{64.5, 100.0}},
		}}
	default:
		err = fmt.Errorf("not implemented")
	}
	return
}

func (handler *testAPIHandler) Annotations(annotation string, _ *grafana_json.AnnotationRequestArgs) (annotations []grafana_json.Annotation, err error) {
	annotations = append(annotations, grafana_json.Annotation{
		Time:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		Title: annotation,
		Text:  "bar",
		Tags:  []string{"snafu"},
	})
	return
}
