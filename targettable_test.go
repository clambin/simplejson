package grafana_json_test

import (
	"context"
	"fmt"
	grafanajson "github.com/clambin/grafana-json"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	tt = grafanajson.TargetTable{
		"empty": {},
		"query": {
			QueryFunc: queryFunc,
		},
		"tablequery": {
			TableQueryFunc: tableQueryFunc,
		},
		"both": {
			QueryFunc:      queryFunc,
			TableQueryFunc: tableQueryFunc,
		},
	}
)

func TestTargetTable_Targets(t *testing.T) {
	assert.Equal(t, []string{"both", "query", "tablequery"}, tt.Targets())
}

func TestTargetTable_RunQuery(t *testing.T) {
	_, err := tt.RunQuery(context.Background(), "query", &grafanajson.TimeSeriesQueryArgs{})
	assert.Equal(t, "not implemented", err.Error())

	_, err = tt.RunQuery(context.Background(), "tablequery", &grafanajson.TimeSeriesQueryArgs{})
	assert.Equal(t, "unknown target 'tablequery' for TimeSeries Query", err.Error())

	_, err = tt.RunQuery(context.Background(), "invalid", &grafanajson.TimeSeriesQueryArgs{})
	assert.Equal(t, "unknown target 'invalid' for TimeSeries Query", err.Error())
}

func TestTargetTable_RunTableQuery(t *testing.T) {
	_, err := tt.RunTableQuery(context.Background(), "tablequery", &grafanajson.TableQueryArgs{})
	assert.Equal(t, "not implemented", err.Error())

	_, err = tt.RunTableQuery(context.Background(), "query", &grafanajson.TableQueryArgs{})
	assert.Equal(t, "unknown target 'query' for Table Query", err.Error())

	_, err = tt.RunTableQuery(context.Background(), "invalid", &grafanajson.TableQueryArgs{})
	assert.Equal(t, "unknown target 'invalid' for Table Query", err.Error())
}

func queryFunc(_ context.Context, _ string, _ *grafanajson.TimeSeriesQueryArgs) (response *grafanajson.QueryResponse, err error) {
	err = fmt.Errorf("not implemented")
	return
}

func tableQueryFunc(_ context.Context, _ string, _ *grafanajson.TableQueryArgs) (response *grafanajson.TableQueryResponse, err error) {
	err = fmt.Errorf("not implemented")
	return
}
