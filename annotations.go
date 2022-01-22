package simplejson

import (
	"encoding/json"
	"time"
)

// AnnotationRequest is a request for annotation.
type AnnotationRequest struct {
	AnnotationRequestArgs
	Annotation AnnotationRequestDetails `json:"annotation"`
}

// AnnotationRequestArgs contains arguments for the Annotations endpoint.
type AnnotationRequestArgs struct {
	Args
}

// AnnotationRequestDetails specifies which annotation should be returned.
type AnnotationRequestDetails struct {
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
	Request AnnotationRequestDetails
}

// MarshalJSON converts an Annotation to JSON.
func (annotation *Annotation) MarshalJSON() (output []byte, err error) {
	// must be an easier way than this?
	jsonResponse := struct {
		Request AnnotationRequestDetails `json:"annotation"`
		Time    int64                    `json:"time"`
		Title   string                   `json:"title"`
		Text    string                   `json:"text"`
		Tags    []string                 `json:"tags"`
	}{
		Request: annotation.Request,
		Time:    annotation.Time.UnixNano() / 1000000,
		Title:   annotation.Title,
		Text:    annotation.Text,
		Tags:    annotation.Tags,
	}

	return json.Marshal(jsonResponse)
}
