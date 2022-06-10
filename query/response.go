package query

import (
	"encoding/json"
	"errors"
	"time"
)

// Response interface for timeseries and table responses
type Response interface {
	MarshalJSON() ([]byte, error)
}

// TimeSeriesResponse is the response from a timeseries Query.
type TimeSeriesResponse struct {
	Target     string
	DataPoints []DataPoint
}

// MarshalJSON converts a TimeSeriesResponse to JSON.
func (t TimeSeriesResponse) MarshalJSON() (output []byte, err error) {
	return json.Marshal(struct {
		Target     string      `json:"target"`     // name of the target
		DataPoints []DataPoint `json:"datapoints"` // values for the target
	}{Target: t.Target, DataPoints: t.DataPoints})
}

// DataPoint contains one entry returned by a Query.
type DataPoint struct {
	Timestamp time.Time
	Value     int64
}

// MarshalJSON converts a DataPoint to JSON.
func (d DataPoint) MarshalJSON() ([]byte, error) {
	out := []int64{d.Value, d.Timestamp.UnixMilli()}
	return json.Marshal(out)
}

// TableResponse is returned by a TableQuery, i.e. a slice of Column structures.
type TableResponse struct {
	Columns []Column
}

// Column is a column returned by a TableQuery.  Text holds the column's header,
// Data holds the slice of values and should be a TimeColumn, a StringColumn
// or a NumberColumn.
type Column struct {
	Text string
	Data interface{}
}

// TimeColumn holds a slice of time.Time values (one per row).
type TimeColumn []time.Time

// StringColumn holds a slice of string values (one per row).
type StringColumn []string

// NumberColumn holds a slice of number values (one per row).
type NumberColumn []float64

type tableResponse struct {
	Type    string                `json:"type"`
	Columns []tableResponseColumn `json:"columns"`
	Rows    []tableResponseRow    `json:"rows"`
}

type tableResponseColumn struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

type tableResponseRow []interface{}

// MarshalJSON converts a TableResponse to JSON.
func (t TableResponse) MarshalJSON() (output []byte, err error) {
	var colTypes []string
	var rowCount int

	if colTypes, rowCount, err = t.getColumnDetails(); err != nil {
		return
	}

	return json.Marshal(tableResponse{
		Type:    "table",
		Columns: t.buildColumns(colTypes),
		Rows:    t.buildRows(rowCount),
	})
}

func (t TableResponse) getColumnDetails() (colTypes []string, rowCount int, err error) {
	for _, entry := range t.Columns {
		var dataCount int
		switch data := entry.Data.(type) {
		case TimeColumn:
			colTypes = append(colTypes, "time")
			dataCount = len(data)
		case StringColumn:
			colTypes = append(colTypes, "string")
			dataCount = len(data)
		case NumberColumn:
			colTypes = append(colTypes, "number")
			dataCount = len(data)
		}

		if rowCount == 0 {
			rowCount = dataCount
		}

		if dataCount != rowCount {
			err = errors.New("error building table query output: all columns must have the same number of rows")
			break
		}
	}
	return
}

func (t TableResponse) buildColumns(colTypes []string) (columns []tableResponseColumn) {
	for index, entry := range colTypes {
		columns = append(columns, tableResponseColumn{
			Text: t.Columns[index].Text,
			Type: entry,
		})
	}
	return
}

func (t TableResponse) buildRows(rowCount int) (rows []tableResponseRow) {
	for row := 0; row < rowCount; row++ {
		newRow := make(tableResponseRow, len(t.Columns))

		for column, entry := range t.Columns {
			switch data := entry.Data.(type) {
			case TimeColumn:
				newRow[column] = data[row]
			case StringColumn:
				newRow[column] = data[row]
			case NumberColumn:
				newRow[column] = data[row]
			}
		}

		rows = append(rows, newRow)

	}
	return
}
