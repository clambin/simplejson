package simplejson

import (
	"encoding/json"
	"time"
)

// AnnotationRequest is a query for annotation.
type AnnotationRequest struct {
	Annotation RequestDetails `json:"annotation"`
	Args
}

// RequestDetails specifies which annotation should be returned.
type RequestDetails struct {
	Name       string `json:"name"`
	Datasource string `json:"datasource"`
	Enable     bool   `json:"enable"`
	Query      string `json:"query"`
}

// UnmarshalJSON unmarshalls a AnnotationRequest from JSON
func (r *AnnotationRequest) UnmarshalJSON(b []byte) (err error) {
	type Request2 AnnotationRequest
	var c Request2
	err = json.Unmarshal(b, &c)
	if err == nil {
		*r = AnnotationRequest(c)
	}
	return err
}

// Annotation response. The annotation endpoint returns a slice of these.
type Annotation struct {
	Time    time.Time
	TimeEnd time.Time
	Title   string
	Text    string
	Tags    []string
	Request RequestDetails
}

// MarshalJSON converts an Annotation to JSON.
func (annotation Annotation) MarshalJSON() (output []byte, err error) {
	var timeEnd int64
	var isRegion bool
	if !annotation.TimeEnd.IsZero() && !annotation.TimeEnd.Equal(annotation.Time) {
		timeEnd = annotation.TimeEnd.UnixMilli()
		isRegion = true
	}

	jsonResponse := struct {
		Request  RequestDetails `json:"annotation"`
		Time     int64          `json:"time"`
		TimeEnd  int64          `json:"timeEnd,omitempty"`
		IsRegion bool           `json:"isRegion,omitempty"`
		Title    string         `json:"title"`
		Text     string         `json:"text"`
		Tags     []string       `json:"tags"`
	}{
		Request:  annotation.Request,
		Time:     annotation.Time.UnixMilli(),
		TimeEnd:  timeEnd,
		IsRegion: isRegion,
		Title:    annotation.Title,
		Text:     annotation.Text,
		Tags:     annotation.Tags,
	}

	return json.Marshal(jsonResponse)
}
