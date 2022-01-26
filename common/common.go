package common

import "time"

// Args contains common arguments used by endpoints.
type Args struct {
	Range        Range `json:"range"`
	AdHocFilters []AdHocFilter
}

// Range specified a start and end time for the data to be returned.
type Range struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// AdHocFilter specifies the ad hoc filters, whose keys & values are returned by the /tag-key and /tag-values endpoints.
type AdHocFilter struct {
	Value     string `json:"value"`
	Operator  string `json:"operator"`
	Condition string `json:"condition"`
	Key       string `json:"key"`
}
