package data

import (
	"github.com/clambin/simplejson/v4"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"time"
)

// CreateTableResponse creates a simplejson TableResponse from a Dataset
func (t Table) CreateTableResponse() *simplejson.TableResponse {
	columns := make([]simplejson.Column, len(t.Frame.Fields))

	for i, f := range t.Frame.Fields {
		columns[i] = makeColumn(f)
	}

	return &simplejson.TableResponse{Columns: columns}
}

func makeColumn(f *data.Field) (column simplejson.Column) {
	name := f.Name
	if name == "" {
		name = "(unknown)"
	}

	var values interface{}
	if f.Len() > 0 {
		switch f.At(0).(type) {
		case time.Time:
			values = simplejson.TimeColumn(getFieldValues[time.Time](f))
		case string:
			values = simplejson.StringColumn(getFieldValues[string](f))
		case float64:
			values = simplejson.NumberColumn(getFieldValues[float64](f))
		}
	}
	return simplejson.Column{
		Text: name,
		Data: values,
	}
}
