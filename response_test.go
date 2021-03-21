package grafana_json_test

import (
	"encoding/json"
	"github.com/clambin/grafana-json"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestReadResponseDataSeries(t *testing.T) {
	input := `[{
	"target": "A",
	"datapoints": [
		[ 100, 1577836800000 ],
		[ 101, 1577836860000 ],
		[ 102, 1577836920000 ]
	]

}]`

	var output []grafana_json.QueryResponse

	if err := json.Unmarshal([]byte(input), &output); assert.Nil(t, err) {
		if assert.Len(t, output, 1) {
			assert.Equal(t, "A", output[0].Target)
			if assert.Len(t, output[0].DataPoints, 3) {
				assert.Equal(t, int64(100), output[0].DataPoints[0].Value)
				assert.True(t, output[0].DataPoints[0].Timestamp.Equal(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)))
				assert.Equal(t, int64(101), output[0].DataPoints[1].Value)
				assert.True(t, output[0].DataPoints[1].Timestamp.Equal(time.Date(2020, 1, 1, 0, 1, 0, 0, time.UTC)))
				assert.Equal(t, int64(102), output[0].DataPoints[2].Value)
				assert.True(t, output[0].DataPoints[2].Timestamp.Equal(time.Date(2020, 1, 1, 0, 2, 0, 0, time.UTC)))
			}
		}
	}
}

func TestWriteResponseDataSeries(t *testing.T) {
	in := []grafana_json.QueryResponse{{
		Target: "A",
		DataPoints: []grafana_json.QueryResponseDataPoint{
			{Value: 100, Timestamp: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)},
			{Value: 101, Timestamp: time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC)},
			{Value: 102, Timestamp: time.Date(2020, 1, 1, 2, 0, 0, 0, time.UTC)},
		},
	}}

	expected := `[{"target":"A","datapoints":[[100,1577836800000],[101,1577840400000],[102,1577844000000]]}]`

	out, err := json.Marshal(in)

	if assert.Nil(t, err) {
		assert.Equal(t, expected, string(out))
	}
}

func TestWriteResponseTable(t *testing.T) {
	testDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	in := []grafana_json.TableQueryResponse{{
		Columns: []grafana_json.TableQueryResponseColumn{
			{Text: "Time", Data: grafana_json.TableQueryResponseTimeColumn{testDate, testDate}},
			{Text: "Label", Data: grafana_json.TableQueryResponseStringColumn{"foo", "bar"}},
			{Text: "Series A", Data: grafana_json.TableQueryResponseNumberColumn{42, 43}},
			{Text: "Series B", Data: grafana_json.TableQueryResponseNumberColumn{64.5, 100.0}},
		},
	}}

	expected := `[{"type":"table","columns":[{"text":"Time","type":"time"},{"text":"Label","type":"string"},{"text":"Series A","type":"number"},{"text":"Series B","type":"number"}],"rows":[["2020-01-01T00:00:00Z","foo",42,64.5],["2020-01-01T00:00:00Z","bar",43,100]]}]`

	out, err := json.Marshal(in)

	if assert.Nil(t, err) {
		assert.Equal(t, expected, string(out))
	}
}

func TestWriteBadResponseTable(t *testing.T) {
	testDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	in := []grafana_json.TableQueryResponse{{
		Columns: []grafana_json.TableQueryResponseColumn{
			{Text: "Time", Data: grafana_json.TableQueryResponseTimeColumn{testDate, testDate}},
			{Text: "Label", Data: grafana_json.TableQueryResponseStringColumn{"foo"}},
			{Text: "Series A", Data: grafana_json.TableQueryResponseNumberColumn{42, 43}},
			{Text: "Series B", Data: grafana_json.TableQueryResponseNumberColumn{64.5, 100.0, 105.0}},
		},
	}}

	_, err := json.Marshal(in)

	assert.NotNil(t, err)
}

func TestWriteCombinedResponse(t *testing.T) {
	testDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	dataseries := []grafana_json.QueryResponse{{
		Target: "A",
		DataPoints: []grafana_json.QueryResponseDataPoint{
			{Value: 100, Timestamp: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)},
			{Value: 101, Timestamp: time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC)},
			{Value: 102, Timestamp: time.Date(2020, 1, 1, 2, 0, 0, 0, time.UTC)},
		},
	}}

	tables := []grafana_json.TableQueryResponse{{
		Columns: []grafana_json.TableQueryResponseColumn{
			{Text: "Time", Data: grafana_json.TableQueryResponseTimeColumn{testDate, testDate}},
			{Text: "Label", Data: grafana_json.TableQueryResponseStringColumn{"foo", "bar"}},
			{Text: "Series A", Data: grafana_json.TableQueryResponseNumberColumn{42, 43}},
			{Text: "Series B", Data: grafana_json.TableQueryResponseNumberColumn{64.5, 100.0}},
		},
	}}

	packaged := make([]interface{}, 0)
	for _, dataserie := range dataseries {
		packaged = append(packaged, dataserie)
	}
	for _, table := range tables {
		packaged = append(packaged, table)
	}

	output, err := json.Marshal(packaged)

	expected := `[{"target":"A","datapoints":[[100,1577836800000],[101,1577840400000],[102,1577844000000]]},{"Columns":[{"Text":"Time","Data":["2020-01-01T00:00:00Z","2020-01-01T00:00:00Z"]},{"Text":"Label","Data":["foo","bar"]},{"Text":"Series A","Data":[42,43]},{"Text":"Series B","Data":[64.5,100]}]}]`
	if assert.Nil(t, err) {
		assert.Equal(t, expected, string(output))
	}
}
