package grafana_json

import (
	"context"
	"fmt"
	"sort"
)

// TargetTable is a convenience struct for handlers that support multiple targets
type TargetTable map[string]struct {
	QueryFunc      QueryFunc
	TableQueryFunc TableQueryFunc
}

func (tt TargetTable) Targets() (targets []string) {
	for target, functions := range tt {
		if functions.TableQueryFunc != nil || functions.QueryFunc != nil {
			targets = append(targets, target)
		}
	}
	sort.Strings(targets)
	return
}

func (tt TargetTable) RunQuery(ctx context.Context, target string, args *TimeSeriesQueryArgs) (response *QueryResponse, err error) {
	builder, ok := tt[target]

	if ok == false || builder.QueryFunc == nil {
		return nil, fmt.Errorf("unknown target '%s' for TimeSeries Query", target)
	}

	return builder.QueryFunc(ctx, target, args)
}

func (tt TargetTable) RunTableQuery(ctx context.Context, target string, args *TableQueryArgs) (response *TableQueryResponse, err error) {
	builder, ok := tt[target]

	if ok == false || builder.TableQueryFunc == nil {
		return nil, fmt.Errorf("unknown target '%s' for Table Query", target)
	}

	return builder.TableQueryFunc(ctx, target, args)
}
