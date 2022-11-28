package data_test

import (
	"github.com/clambin/simplejson/v4"
	"github.com/clambin/simplejson/v4/pkg/data"
	grafanaData "github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTable_FilterByTime(t *testing.T) {
	d := createTable(10)

	output := d.Filter(simplejson.Args{
		Range: simplejson.Range{
			From: time.Date(2022, 6, 5, 0, 0, 0, 0, time.UTC),
			To:   time.Date(2022, 6, 7, 0, 0, 0, 0, time.UTC),
		},
	})
	assert.Equal(t, []time.Time{
		time.Date(2022, time.June, 5, 0, 0, 0, 0, time.UTC),
		time.Date(2022, time.June, 6, 0, 0, 0, 0, time.UTC),
		time.Date(2022, time.June, 7, 0, 0, 0, 0, time.UTC),
	}, output.GetTimestamps())
}

func TestTable_FilterByTime_Empty(t *testing.T) {
	table := data.Table{Frame: grafanaData.NewFrame("bad")}

	f := table.Filter(simplejson.Args{
		Range: simplejson.Range{
			From: time.Date(2022, 6, 5, 0, 0, 0, 0, time.UTC),
			To:   time.Date(2022, 6, 7, 0, 0, 0, 0, time.UTC),
		},
	})
	assert.NotNil(t, f.Frame)
}
