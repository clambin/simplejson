package query_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"github.com/clambin/simplejson/v3/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var update = flag.Bool("update", false, "update .golden files")

func TestWriteResponseDataSeries(t *testing.T) {
	r := query.TimeSeriesResponse{
		Target: "A",
		DataPoints: []query.DataPoint{
			{Value: 100, Timestamp: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)},
			{Value: 101, Timestamp: time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC)},
			{Value: 102, Timestamp: time.Date(2020, 1, 1, 2, 0, 0, 0, time.UTC)},
		},
	}

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	err := json.NewEncoder(w).Encode(r)
	require.NoError(t, err)
	_ = w.Flush()

	gp := filepath.Join("testdata", t.Name()+".golden")
	if *update {
		t.Logf("updating golden file for %s", t.Name())
		err = os.WriteFile(gp, b.Bytes(), 0644)
		require.NoError(t, err, "failed to update golden file")
	}

	var golden []byte
	golden, err = os.ReadFile(gp)
	require.NoError(t, err)

	assert.Equal(t, string(golden), b.String())
}

func TestWriteResponseTable(t *testing.T) {
	testDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	r := query.TableResponse{
		Columns: []query.Column{
			{Text: "Time", Data: query.TimeColumn{testDate, testDate}},
			{Text: "Label", Data: query.StringColumn{"foo", "bar"}},
			{Text: "Series A", Data: query.NumberColumn{42, 43}},
			{Text: "Series B", Data: query.NumberColumn{64.5, 100.0}},
		},
	}

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	err := json.NewEncoder(w).Encode(r)
	require.NoError(t, err)
	_ = w.Flush()

	gp := filepath.Join("testdata", t.Name()+".golden")
	if *update {
		t.Logf("updating golden file for %s", t.Name())
		err = os.WriteFile(gp, b.Bytes(), 0644)
		require.NoError(t, err, "failed to update golden file")
	}

	var golden []byte
	golden, err = os.ReadFile(gp)
	require.NoError(t, err)

	assert.Equal(t, string(golden), b.String())

}

func TestWriteBadResponseTable(t *testing.T) {
	testDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	in := query.TableResponse{
		Columns: []query.Column{
			{Text: "Time", Data: query.TimeColumn{testDate, testDate}},
			{Text: "Label", Data: query.StringColumn{"foo"}},
			{Text: "Series A", Data: query.NumberColumn{42, 43}},
			{Text: "Series B", Data: query.NumberColumn{64.5, 100.0, 105.0}},
		},
	}

	_, err := json.Marshal(in)
	assert.Error(t, err)
}

func TestWriteCombinedResponse(t *testing.T) {
	testDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	dataseries := []query.TimeSeriesResponse{{
		Target: "A",
		DataPoints: []query.DataPoint{
			{Value: 100, Timestamp: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)},
			{Value: 101, Timestamp: time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC)},
			{Value: 102, Timestamp: time.Date(2020, 1, 1, 2, 0, 0, 0, time.UTC)},
		},
	}}

	tables := []query.TableResponse{{
		Columns: []query.Column{
			{Text: "Time", Data: query.TimeColumn{testDate, testDate}},
			{Text: "Label", Data: query.StringColumn{"foo", "bar"}},
			{Text: "Series A", Data: query.NumberColumn{42, 43}},
			{Text: "Series B", Data: query.NumberColumn{64.5, 100.0}},
		},
	}}

	packaged := make([]interface{}, 0)
	for _, dataserie := range dataseries {
		packaged = append(packaged, dataserie)
	}
	for _, table := range tables {
		packaged = append(packaged, table)
	}

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	err := json.NewEncoder(w).Encode(packaged)
	require.NoError(t, err)
	_ = w.Flush()

	gp := filepath.Join("testdata", t.Name()+".golden")
	if *update {
		t.Logf("updating golden file for %s", t.Name())
		err = os.WriteFile(gp, b.Bytes(), 0644)
		require.NoError(t, err, "failed to update golden file")
	}

	var golden []byte
	golden, err = os.ReadFile(gp)
	require.NoError(t, err)

	assert.Equal(t, string(golden), b.String())
}
