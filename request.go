package simplejson

import (
	"context"
)

// TimeSeriesRequest is a Query request. For each specified RequestTarget, the server will call the Query endpoint
// with the provided TimeSeriesQueryArgs.
type TimeSeriesRequest struct {
	Targets []RequestTarget `json:"targets"`
	TimeSeriesQueryArgs
}

// TimeSeriesQueryArgs contains the arguments for a Query.
type TimeSeriesQueryArgs struct {
	Args
	// Interval      QueryRequestDuration `json:"interval"`
	MaxDataPoints uint64 `json:"maxDataPoints"`
}

// TableQueryArgs contains the arguments for a TableQuery.
type TableQueryArgs struct {
	Args
}

// RequestTarget specifies the requested target name and type.
type RequestTarget struct {
	Target string `json:"target"` // name of the target.
	Type   string `json:"type"`   // "timeserie" or "" for timeseries. "table" for table queries.
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

// TimeSeriesQueryFunc handles timeseries queries
type TimeSeriesQueryFunc func(ctx context.Context, target string, args *TimeSeriesQueryArgs) (*TimeSeriesResponse, error)

// TableQueryFunc handles for table queries
type TableQueryFunc func(ctx context.Context, target string, args *TableQueryArgs) (*TableQueryResponse, error)
