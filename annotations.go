package simplejson

import (
	"encoding/json"
	"time"
)

// Request is a request for annotation.
type Request struct {
	RequestArgs
	Annotation RequestDetails `json:"annotation"`
}

// RequestArgs contains arguments for the Annotations endpoint.
type RequestArgs struct {
	Args
}

// RequestDetails specifies which annotation should be returned.
type RequestDetails struct {
	Name       string `json:"name"`
	Datasource string `json:"datasource"`
	Enable     bool   `json:"enable"`
	Query      string `json:"query"`
}

// Annotation response. The annotation endpoint returns a slice of these.
type Annotation struct {
	Time    time.Time
	Title   string
	Text    string
	Tags    []string
	Request RequestDetails
}

// MarshalJSON converts an Annotation to JSON.
func (annotation *Annotation) MarshalJSON() (output []byte, err error) {
	// must be an easier way than this?
	jsonResponse := struct {
		Request RequestDetails `json:"annotation"`
		Time    int64          `json:"time"`
		Title   string         `json:"title"`
		Text    string         `json:"text"`
		Tags    []string       `json:"tags"`
	}{
		Request: annotation.Request,
		Time:    annotation.Time.UnixNano() / 1000000,
		Title:   annotation.Title,
		Text:    annotation.Text,
		Tags:    annotation.Tags,
	}

	return json.Marshal(jsonResponse)
}
