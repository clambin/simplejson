package grafana_json_test

import (
	"context"
	grafanajson "github.com/clambin/grafana-json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTargetTable_Targets(t *testing.T) {
	tt := grafanajson.TargetTable{
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

	assert.Equal(t, []string{"both", "query", "tablequery"}, tt.Targets())
}

func queryFunc(_ context.Context, _ string, _ *grafanajson.TimeSeriesQueryArgs) (response *grafanajson.QueryResponse, err error) {
	return
}

func tableQueryFunc(_ context.Context, _ string, _ *grafanajson.TableQueryArgs) (response *grafanajson.TableQueryResponse, err error) {
	return
}
