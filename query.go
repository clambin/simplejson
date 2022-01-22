package simplejson

// QueryRequest is a Query request. For each specified RequestTarget, the server will call the Query endpoint
// with the provided TimeSeriesQueryArgs.
type QueryRequest struct {
	Targets []RequestTarget `json:"targets"`
	TimeSeriesQueryArgs
}

// RequestTarget specifies the requested target name and type.
type RequestTarget struct {
	Target string `json:"target"` // name of the target.
	Type   string `json:"type"`   // "timeserie" or "" for timeseries. "table" for table queries.
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
