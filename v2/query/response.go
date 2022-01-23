package query

import (
	"encoding/json"
	"errors"
	"time"
)

// TimeSeriesResponse is the response from a timeseries Query.
type TimeSeriesResponse struct {
	Target     string      `json:"target"`     // name of the target
	DataPoints []DataPoint `json:"datapoints"` // values for the target
}

// DataPoint contains one entry returned by a Query.
type DataPoint struct {
	Timestamp time.Time
	Value     int64
}

// MarshalJSON converts a DataPoint to JSON.
func (d *DataPoint) MarshalJSON() ([]byte, error) {
	out := []int64{d.Value, d.Timestamp.UnixNano() / 1000000}
	return json.Marshal(out)
}

// UnmarshalJSON converts a JSON structure to a DataPoint.
func (d *DataPoint) UnmarshalJSON(input []byte) (err error) {
	var in []int64

	if err = json.Unmarshal(input, &in); err == nil {
		*d = DataPoint{
			Value:     in[0],
			Timestamp: time.Unix(0, in[1]*1000000),
		}
	}
	return
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
func (table *TableResponse) MarshalJSON() (output []byte, err error) {
	var columns []tableResponseColumn
	var rows []tableResponseRow
	var colTypes []string
	var rowCount int

	colTypes, rowCount, err = table.getColumnDetails()

	if err == nil {
		columns, err = table.buildColumns(colTypes)
	}
	if err == nil {
		rows, err = table.buildRows(rowCount)
	}
	if err == nil {
		output, err = json.Marshal(tableResponse{
			Type:    "table",
			Columns: columns,
			Rows:    rows,
		})
	}
	return
}

func (table *TableResponse) getColumnDetails() (colTypes []string, rowCount int, err error) {
	for _, entry := range table.Columns {
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

func (table *TableResponse) buildColumns(colTypes []string) (columns []tableResponseColumn, err error) {
	for index, entry := range colTypes {
		columns = append(columns, tableResponseColumn{
			Text: table.Columns[index].Text,
			Type: entry,
		})
	}
	return
}

func (table *TableResponse) buildRows(rowCount int) (rows []tableResponseRow, err error) {
	for row := 0; row < rowCount; row++ {
		newRow := make(tableResponseRow, len(table.Columns))

		for column, entry := range table.Columns {
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
