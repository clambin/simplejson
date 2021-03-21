package grafana_json

import (
	"time"
)

type QueryRequest struct {
	Targets []QueryRequestTarget `json:"targets"`
	TimeSeriesQueryArgs
}

type CommonQueryArgs struct {
	Range QueryRequestRange `json:"range"`
}

type TimeSeriesQueryArgs struct {
	CommonQueryArgs
	// Interval      QueryRequestDuration `json:"interval"`
	MaxDataPoints uint64 `json:"maxDataPoints"`
}

type TableQueryArgs struct {
	CommonQueryArgs
}

type QueryRequestRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

type QueryRequestTarget struct {
	Target string `json:"target"`
	Type   string `json:"type"`
}

type QueryRequestDuration time.Duration

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

type AnnotationRequest struct {
	AnnotationRequestArgs
	Annotation AnnotationRequestDetails `json:"annotation"`
}

type AnnotationRequestArgs struct {
	CommonQueryArgs
}

type AnnotationRequestDetails struct {
	Name       string `json:"name"`
	Datasource string `json:"datasource"`
	Enable     bool   `json:"enable"`
	Query      string `json:"query"`
}
