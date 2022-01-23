package simplejson_test

import (
	"encoding/json"
	"github.com/clambin/simplejson/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	var output []simplejson.TimeSeriesResponse

	err := json.Unmarshal([]byte(input), &output)
	require.NoError(t, err)
	require.Len(t, output, 1)
	assert.Equal(t, "A", output[0].Target)
	require.Len(t, output[0].DataPoints, 3)
	assert.Equal(t, int64(100), output[0].DataPoints[0].Value)
	assert.True(t, output[0].DataPoints[0].Timestamp.Equal(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)))
	assert.Equal(t, int64(101), output[0].DataPoints[1].Value)
	assert.True(t, output[0].DataPoints[1].Timestamp.Equal(time.Date(2020, 1, 1, 0, 1, 0, 0, time.UTC)))
	assert.Equal(t, int64(102), output[0].DataPoints[2].Value)
	assert.True(t, output[0].DataPoints[2].Timestamp.Equal(time.Date(2020, 1, 1, 0, 2, 0, 0, time.UTC)))
}

func TestWriteResponseDataSeries(t *testing.T) {
	in := []simplejson.TimeSeriesResponse{{
		Target: "A",
		DataPoints: []simplejson.DataPoint{
			{Value: 100, Timestamp: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)},
			{Value: 101, Timestamp: time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC)},
			{Value: 102, Timestamp: time.Date(2020, 1, 1, 2, 0, 0, 0, time.UTC)},
		},
	}}

	expected := `[{"target":"A","datapoints":[[100,1577836800000],[101,1577840400000],[102,1577844000000]]}]`

	out, err := json.Marshal(in)
	require.NoError(t, err)
	assert.Equal(t, expected, string(out))
}

func TestWriteResponseTable(t *testing.T) {
	testDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	in := []simplejson.TableQueryResponse{{
		Columns: []simplejson.TableQueryResponseColumn{
			{Text: "Time", Data: simplejson.TableQueryResponseTimeColumn{testDate, testDate}},
			{Text: "Label", Data: simplejson.TableQueryResponseStringColumn{"foo", "bar"}},
			{Text: "Series A", Data: simplejson.TableQueryResponseNumberColumn{42, 43}},
			{Text: "Series B", Data: simplejson.TableQueryResponseNumberColumn{64.5, 100.0}},
		},
	}}

	expected := `[{"type":"table","columns":[{"text":"Time","type":"time"},{"text":"Label","type":"string"},{"text":"Series A","type":"number"},{"text":"Series B","type":"number"}],"rows":[["2020-01-01T00:00:00Z","foo",42,64.5],["2020-01-01T00:00:00Z","bar",43,100]]}]`

	out, err := json.Marshal(in)

	require.NoError(t, err)
	assert.Equal(t, expected, string(out))
}

func TestWriteBadResponseTable(t *testing.T) {
	testDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	in := []simplejson.TableQueryResponse{{
		Columns: []simplejson.TableQueryResponseColumn{
			{Text: "Time", Data: simplejson.TableQueryResponseTimeColumn{testDate, testDate}},
			{Text: "Label", Data: simplejson.TableQueryResponseStringColumn{"foo"}},
			{Text: "Series A", Data: simplejson.TableQueryResponseNumberColumn{42, 43}},
			{Text: "Series B", Data: simplejson.TableQueryResponseNumberColumn{64.5, 100.0, 105.0}},
		},
	}}

	_, err := json.Marshal(in)
	assert.Error(t, err)
}

func TestWriteCombinedResponse(t *testing.T) {
	testDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	dataseries := []simplejson.TimeSeriesResponse{{
		Target: "A",
		DataPoints: []simplejson.DataPoint{
			{Value: 100, Timestamp: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)},
			{Value: 101, Timestamp: time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC)},
			{Value: 102, Timestamp: time.Date(2020, 1, 1, 2, 0, 0, 0, time.UTC)},
		},
	}}

	tables := []simplejson.TableQueryResponse{{
		Columns: []simplejson.TableQueryResponseColumn{
			{Text: "Time", Data: simplejson.TableQueryResponseTimeColumn{testDate, testDate}},
			{Text: "Label", Data: simplejson.TableQueryResponseStringColumn{"foo", "bar"}},
			{Text: "Series A", Data: simplejson.TableQueryResponseNumberColumn{42, 43}},
			{Text: "Series B", Data: simplejson.TableQueryResponseNumberColumn{64.5, 100.0}},
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
	require.NoError(t, err)
	assert.Equal(t, expected, string(output))
}
