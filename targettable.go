package grafana_json

import "sort"

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
