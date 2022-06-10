package data

import (
	"github.com/clambin/simplejson/v3/query"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"time"
)

// CreateTableResponse creates a simplejson TableResponse from a Dataset
func (t Table) CreateTableResponse() *query.TableResponse {
	columns := make([]query.Column, len(t.Frame.Fields))

	for i, f := range t.Frame.Fields {
		columns[i] = makeColumn(f)
	}

	return &query.TableResponse{Columns: columns}
}

func makeColumn(f *data.Field) (column query.Column) {
	name := f.Name
	if name == "" {
		name = "(unknown)"
	}

	var values interface{}
	if f.Len() > 0 {
		switch f.At(0).(type) {
		case time.Time:
			values = query.TimeColumn(getFieldValues[time.Time](f))
		case string:
			values = query.StringColumn(getFieldValues[string](f))
		case float64:
			values = query.NumberColumn(getFieldValues[float64](f))
		}
	}
	return query.Column{
		Text: name,
		Data: values,
	}
}
