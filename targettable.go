package grafana_json

import (
	"context"
	"fmt"
	"sort"
)

// TargetTable maps a target to a set of timeseries and table queuries. This can be useful if a Handler supports multiple targets,
// and each target requires its own timeseries and/or table query function.
type TargetTable map[string]struct {
	QueryFunc      QueryFunc
	TableQueryFunc TableQueryFunc
}

// Targets returns the targets mapped in this TargetTable.
func (tt TargetTable) Targets() (targets []string) {
	for target, functions := range tt {
		if functions.TableQueryFunc != nil || functions.QueryFunc != nil {
			targets = append(targets, target)
		}
	}
	sort.Strings(targets)
	return
}

// RunQuery runs a timeseries query against a TargetTable.  It looks up the target in the TargetTable and runs that
// timeseries query. If the target doesn't exist, or doesn't have a timeseries query, it returns an error.
func (tt TargetTable) RunQuery(ctx context.Context, target string, args *TimeSeriesQueryArgs) (response *QueryResponse, err error) {
	builder, ok := tt[target]

	if ok == false || builder.QueryFunc == nil {
		return nil, fmt.Errorf("unknown target '%s' for TimeSeries Query", target)
	}

	return builder.QueryFunc(ctx, target, args)
}

// RunTableQuery runs a table query against a TargetTable.  It looks up the target in the TargetTable and runs that
// table query. If the target doesn't exist, or doesn't have a table query, it returns an error.
func (tt TargetTable) RunTableQuery(ctx context.Context, target string, args *TableQueryArgs) (response *TableQueryResponse, err error) {
	builder, ok := tt[target]

	if ok == false || builder.TableQueryFunc == nil {
		return nil, fmt.Errorf("unknown target '%s' for Table Query", target)
	}

	return builder.TableQueryFunc(ctx, target, args)
}
