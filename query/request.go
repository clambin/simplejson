package query

import (
	"encoding/json"
	"github.com/clambin/simplejson/v3/common"
)

// Request is a Query request. For each specified Target, the server will call the appropriate handler's Query or TableQuery
// function with the provided Args.
type Request struct {
	Targets []Target `json:"targets"`
	Args
}

// Target specifies the requested target name and type.
type Target struct {
	Name string `json:"target"` // name of the target.
	Type string `json:"type"`   // "timeserie" or "" for timeseries. "table" for table queries.
}

// Args contains the arguments for a Query.
type Args struct {
	common.Args
	// Interval      QueryRequestDuration `json:"interval"`
	MaxDataPoints uint64 `json:"maxDataPoints"`
}

// UnmarshalJSON unmarshalls a Request from JSON
func (r *Request) UnmarshalJSON(b []byte) (err error) {
	// workaround to avoid infinite loop
	type Request2 Request
	var c Request2
	err = json.Unmarshal(b, &c)
	if err == nil {
		*r = Request(c)
	}
	return err
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
