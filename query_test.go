package simplejson

import (
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRequests(t *testing.T) {
	input := `{
	"maxDataPoints": 100,
	"interval": "1h",
	"range": {
		"from": "2020-01-01T00:00:00.000Z",
		"to": "2020-12-31T00:00:00.000Z"
	},
	"targets": [
		{ "target": "A", "type": "dataserie" },
		{ "target": "B", "type": "table" }
	]
}`

	var output QueryRequest

	err := json.Unmarshal([]byte(input), &output)
	require.NoError(t, err)
	assert.Equal(t, uint64(100), output.MaxDataPoints)
	// assert.Equal(t, server.QueryRequestDuration(1*time.Hour), output.Interval)
	// assert.Equal(t, 1*time.Hour, time.duration(output.Interval))
	assert.Equal(t, time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), output.Range.From)
	assert.Equal(t, time.Date(2020, 12, 31, 0, 0, 0, 0, time.UTC), output.Range.To)
	require.Len(t, output.Targets, 2)
	assert.Equal(t, "A", output.Targets[0].Name)
	assert.Equal(t, "dataserie", output.Targets[0].Type)
	assert.Equal(t, "B", output.Targets[1].Name)
	assert.Equal(t, "table", output.Targets[1].Type)
}

func TestResponse(t *testing.T) {
	tests := []struct {
		name     string
		pass     bool
		response Response
	}{
		{
			name: "timeseries",
			pass: true,
			response: TimeSeriesResponse{
				Target: "A",
				DataPoints: []DataPoint{
					{Value: 100, Timestamp: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)},
					{Value: 101, Timestamp: time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC)},
					{Value: 102, Timestamp: time.Date(2020, 1, 1, 2, 0, 0, 0, time.UTC)},
				},
			},
		},
		{
			name: "table",
			pass: true,
			response: TableResponse{
				Columns: []Column{
					{Text: "Time", Data: TimeColumn{time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, 1, 1, 0, 1, 0, 0, time.UTC)}},
					{Text: "Label", Data: StringColumn{"foo", "bar"}},
					{Text: "Series A", Data: NumberColumn{42, 43}},
					{Text: "Series B", Data: NumberColumn{64.5, 100.0}},
				},
			},
		},
		{
			name:     "combined",
			pass:     true,
			response: makeCombinedQueryResponse(),
		},
		{
			name: "invalid",
			pass: false,
			response: TableResponse{
				Columns: []Column{
					{Text: "Time", Data: TimeColumn{time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, 1, 1, 0, 1, 0, 0, time.UTC)}},
					{Text: "Label", Data: StringColumn{"foo"}},
					{Text: "Series A", Data: NumberColumn{42, 43}},
					{Text: "Series B", Data: NumberColumn{64.5, 100.0, 105.0}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b bytes.Buffer
			w := bufio.NewWriter(&b)
			enc := json.NewEncoder(w)
			enc.SetIndent("", "  ")
			err := enc.Encode(tt.response)

			if !tt.pass {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			_ = w.Flush()

			gp := filepath.Join("testdata", strings.ToLower(t.Name())+".golden")
			if *update {
				t.Logf("updating golden file for %s", t.Name())
				err = os.WriteFile(gp, b.Bytes(), 0644)
				require.NoError(t, err, "failed to update golden file")
			}

			var golden []byte
			golden, err = os.ReadFile(gp)
			require.NoError(t, err)

			assert.Equal(t, string(golden), b.String())

		})
	}
}

type combinedResponse struct {
	responses []interface{}
}

func (r combinedResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.responses)
}

func makeCombinedQueryResponse() combinedResponse {
	testDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	dataseries := []TimeSeriesResponse{{
		Target: "A",
		DataPoints: []DataPoint{
			{Value: 100, Timestamp: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)},
			{Value: 101, Timestamp: time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC)},
			{Value: 102, Timestamp: time.Date(2020, 1, 1, 2, 0, 0, 0, time.UTC)},
		},
	}}

	tables := []TableResponse{{
		Columns: []Column{
			{Text: "Time", Data: TimeColumn{testDate, testDate}},
			{Text: "Label", Data: StringColumn{"foo", "bar"}},
			{Text: "Series A", Data: NumberColumn{42, 43}},
			{Text: "Series B", Data: NumberColumn{64.5, 100.0}},
		},
	}}

	var r combinedResponse
	//r.responses = make([]interface{}, 0)
	for _, dataserie := range dataseries {
		r.responses = append(r.responses, dataserie)
	}
	for _, table := range tables {
		r.responses = append(r.responses, table)
	}

	return r
}

func BenchmarkTimeSeriesResponse_MarshalJSON(b *testing.B) {
	response := buildTimeSeriesResponse(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := response.MarshalJSON(); err != nil {
			b.Fatal(err)
		}
	}
}

func buildTimeSeriesResponse(count int) TimeSeriesResponse {
	var datapoints []DataPoint
	timestamp := time.Date(2022, time.November, 27, 0, 0, 0, 0, time.UTC)
	for i := 0; i < count; i++ {
		datapoints = append(datapoints, DataPoint{
			Timestamp: timestamp,
			Value:     float64(i),
		})
	}
	return TimeSeriesResponse{Target: "foo", DataPoints: datapoints}
}

func BenchmarkTableResponse_MarshalJSON(b *testing.B) {
	response := buildTableResponse(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := response.MarshalJSON(); err != nil {
			b.Fatal(err)
		}
	}
}

func buildTableResponse(count int) TableResponse {
	var timestamps []time.Time
	var values []float64

	timestamp := time.Date(2022, time.November, 27, 0, 0, 0, 0, time.UTC)
	for i := 0; i < count; i++ {
		timestamps = append(timestamps, timestamp)
		values = append(values, 1.0)
		timestamp = timestamp.Add(time.Minute)
	}
	return TableResponse{Columns: []Column{
		{Text: "time", Data: TimeColumn(timestamps)},
		{Text: "value", Data: NumberColumn(values)},
	}}
}
