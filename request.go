package grafana_json

import (
	"time"
)

type QueryRequest struct {
	Range QueryRequestRange `json:"range"`
	// Interval      QueryRequestDuration `json:"interval"`
	MaxDataPoints uint64               `json:"maxDataPoints"`
	Targets       []QueryRequestTarget `json:"targets"`
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

/*
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
