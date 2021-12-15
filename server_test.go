package grafana_json_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/clambin/grafana-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net"
	"net/http"
	"testing"
	"time"
)

var Port int

func TestMain(m *testing.M) {
	s := grafana_json.Server{
		Handlers: []grafana_json.Handler{createHandler()},
	}

	listener, err := net.Listen("tcp4", ":0")
	if err != nil {
		panic(err)
	}
	Port = listener.Addr().(*net.TCPAddr).Port

	httpServer := http.Server{
		Handler: s.GetRouter(),
	}
	go func() {
		err2 := httpServer.Serve(listener)
		if !errors.Is(err2, http.ErrServerClosed) {
			panic(err2)
		}
	}()

	m.Run()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	err = s.Shutdown(ctx, time.Second)
	if err != nil {
		panic(err)
	}
	cancel()
}

func BenchmarkAPIServer(b *testing.B) {
	require.Eventually(b, func() bool {
		body, err := call(Port, "/", "GET", "")
		require.NoError(b, err)
		return assert.Equal(b, "Hello", body)
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
	body, err = call(Port, "/query", "POST", req)

	require.Nil(b, err)
	assert.Equal(b, `[{"target":"A","datapoints":[[100,1577836800000],[101,1577836860000],[103,1577836920000]]},{"target":"B","datapoints":[[100,1577836800000],[99,1577836860000],[98,1577836920000]]}]`, body)
}

func serverRunning(t *testing.T) {
	require.Eventually(t, func() bool {
		body, err := call(Port, "/", "GET", "")
		if assert.Nil(t, err) {
			return assert.Empty(t, body)
		}
		return false
	}, 500*time.Millisecond, 10*time.Millisecond)
}

func call(port int, path, method, body string) (response string, err error) {
	url := fmt.Sprintf("http://localhost:%d%s", port, path)
	client := &http.Client{}
	reqBody := bytes.NewBuffer([]byte(body))
	req, _ := http.NewRequest(method, url, reqBody)

	var resp *http.Response
	if resp, err = client.Do(req); err == nil {
		defer func() {
			_ = resp.Body.Close()
		}()
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
	noEndpoints bool

	queryResponses      map[string]*grafana_json.QueryResponse
	tableQueryResponses map[string]*grafana_json.TableQueryResponse
	annotations         []grafana_json.Annotation
	tags                []string
	tagValues           map[string][]string
}

func createHandler() (handler *testAPIHandler) {
	return &testAPIHandler{
		queryResponses: map[string]*grafana_json.QueryResponse{
			"A": {
				Target: "A",
				DataPoints: []grafana_json.QueryResponseDataPoint{
					{Timestamp: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), Value: 100},
					{Timestamp: time.Date(2020, 1, 1, 0, 1, 0, 0, time.UTC), Value: 101},
					{Timestamp: time.Date(2020, 1, 1, 0, 2, 0, 0, time.UTC), Value: 103},
				},
			},
			"B": {
				Target: "B",
				DataPoints: []grafana_json.QueryResponseDataPoint{
					{Timestamp: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), Value: 100},
					{Timestamp: time.Date(2020, 1, 1, 0, 1, 0, 0, time.UTC), Value: 99},
					{Timestamp: time.Date(2020, 1, 1, 0, 2, 0, 0, time.UTC), Value: 98},
				},
			},
		},
		tableQueryResponses: map[string]*grafana_json.TableQueryResponse{
			"C": {
				Columns: []grafana_json.TableQueryResponseColumn{
					{Text: "Time", Data: grafana_json.TableQueryResponseTimeColumn{
						time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
						time.Date(2020, 1, 1, 0, 1, 0, 0, time.UTC),
					}},
					{Text: "Label", Data: grafana_json.TableQueryResponseStringColumn{"foo", "bar"}},
					{Text: "Series A", Data: grafana_json.TableQueryResponseNumberColumn{42, 43}},
					{Text: "Series B", Data: grafana_json.TableQueryResponseNumberColumn{64.5, 100.0}},
				},
			},
		},
		annotations: []grafana_json.Annotation{{
			Time:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			Title: "foo",
			Text:  "bar",
			Tags:  []string{"snafu"},
		}},
		tags: []string{"foo", "bar"},
		tagValues: map[string][]string{
			"foo": {"A", "B"},
			"bar": {"1", "2"},
		},
	}
}

func (handler *testAPIHandler) Endpoints() grafana_json.Endpoints {
	if handler.noEndpoints {
		return grafana_json.Endpoints{}
	}
	return grafana_json.Endpoints{
		Search:      handler.Search,
		Query:       handler.Query,
		TableQuery:  handler.TableQuery,
		Annotations: handler.Annotations,
		TagKeys:     handler.Tags,
		TagValues:   handler.TagValues,
	}
}

func (handler *testAPIHandler) Search() []string {
	return []string{"A", "B", "C", "Crash"}
}

func (handler *testAPIHandler) Query(_ context.Context, target string, _ *grafana_json.TimeSeriesQueryArgs) (response *grafana_json.QueryResponse, err error) {
	if target == "Crash" {
		err = fmt.Errorf("server crash")
	} else {
		var ok bool
		if response, ok = handler.queryResponses[target]; ok == false {
			err = fmt.Errorf("not implemented: %s", target)
		}
	}
	return
}

func (handler *testAPIHandler) TableQuery(_ context.Context, target string, _ *grafana_json.TableQueryArgs) (response *grafana_json.TableQueryResponse, err error) {
	if target == "Crash" {
		err = fmt.Errorf("server crash")
	} else {
		var ok bool
		if response, ok = handler.tableQueryResponses[target]; ok == false {
			err = fmt.Errorf("not implemented: %s", target)
		}
	}
	return
}

func (handler *testAPIHandler) Annotations(_, _ string, _ *grafana_json.AnnotationRequestArgs) (annotations []grafana_json.Annotation, err error) {
	for _, ann := range handler.annotations {
		annotations = append(annotations, ann)
	}
	return
}

func (handler *testAPIHandler) Tags(_ context.Context) (tags []string) {
	return handler.tags
}

func (handler *testAPIHandler) TagValues(_ context.Context, tag string) (values []string, err error) {
	var ok bool
	if values, ok = handler.tagValues[tag]; ok == false {
		err = fmt.Errorf("unsupported tag '%s'", tag)
	}
	return
}
