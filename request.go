package grafana_json

import (
	"time"
)

// QueryRequest is a Query request. For each specified QueryRequestTarget, the server will call the Query endpoint
// with the provided TimeSeriesQueryArgs
type QueryRequest struct {
	Targets []QueryRequestTarget `json:"targets"`
	TimeSeriesQueryArgs
}

// CommonQueryArgs contains common arguments used by endpoints
type CommonQueryArgs struct {
	Range        QueryRequestRange `json:"range"`
	AdHocFilters []AdHocFilter
}

// TimeSeriesQueryArgs contains the arguments for a Query
type TimeSeriesQueryArgs struct {
	CommonQueryArgs
	// Interval      QueryRequestDuration `json:"interval"`
	MaxDataPoints uint64 `json:"maxDataPoints"`
}

// TableQueryArgs contains the arguments for a TableQuery
type TableQueryArgs struct {
	CommonQueryArgs
}

// QueryRequestRange is part of the common arguments
type QueryRequestRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// QueryRequestTarget specifies the requested target. Target contains the target's name. Type specifies
// the target's type ("dataserie" or "" for Queries, "table" for TableQueries).
type QueryRequestTarget struct {
	Target string `json:"target"`
	Type   string `json:"type"`
}

// AdHocFilter specifies the ad hoc filters, whose keys & values are returned by the /tag-key and /tag-values endpoints.
type AdHocFilter struct {
	Value     string `json:"value"`
	Operator  string `json:"operator"`
	Condition string `json:"condition"`
	Key       string `json:"key"`
}

// type QueryRequestDuration time.Duration

/* TODO: intervals can go to "1y", which time.ParseDuration doesn't handle
func (d *QueryRequestDuration) MarshalJSON() ([]byte, error) {
	out := time.Duration(*d).String()
	return json.Marshal(out)
}


func (d *QueryRequestDuration) UnmarshalJSON(input []byte) (err error) {
	in := ""
	if err = json.Unmarshal(input, &in); err == nil {
		var value time.Duration
		value, err = time.ParseDuration(in)
		*d = QueryRequestDuration(value)
	}
	return
}
*/

// AnnotationRequest is a request for annotations
type AnnotationRequest struct {
	AnnotationRequestArgs
	Annotation AnnotationRequestDetails `json:"annotation"`
}

// AnnotationRequestArgs contains arguments for the Annotations endpoint
type AnnotationRequestArgs struct {
	CommonQueryArgs
}

// AnnotationRequestDetails specifies which annotations should be returned
type AnnotationRequestDetails struct {
	Name       string `json:"name"`
	Datasource string `json:"datasource"`
	Enable     bool   `json:"enable"`
	Query      string `json:"query"`
}
