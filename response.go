package simplejson

import (
	"encoding/json"
	"errors"
	"time"
)

// TimeSeriesResponse is returned by a Query.
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

// TableQueryResponse is returned by a TableQuery, i.e. a slice of TableQueryResponseColumn structures.
type TableQueryResponse struct {
	Columns []TableQueryResponseColumn
}

// TableQueryResponseColumn is a column returned by a TableQuery.  Text holds the column's header,
// Data holds the slice of values and should be a TableQueryResponseTimeColumn, a TableQueryResponseStringColumn
// or a TableQueryResponseNumberColumn.
type TableQueryResponseColumn struct {
	Text string
	Data interface{}
}

// TableQueryResponseTimeColumn holds a slice of time.Time values (one per row).
type TableQueryResponseTimeColumn []time.Time

// TableQueryResponseStringColumn holds a slice of string values (one per row).
type TableQueryResponseStringColumn []string

// TableQueryResponseNumberColumn holds a slice of number values (one per row).
type TableQueryResponseNumberColumn []float64

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

// MarshalJSON converts a TableQueryResponse to JSON.
func (table *TableQueryResponse) MarshalJSON() (output []byte, err error) {
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

func (table *TableQueryResponse) getColumnDetails() (colTypes []string, rowCount int, err error) {
	for _, entry := range table.Columns {
		var dataCount int
		switch data := entry.Data.(type) {
		case TableQueryResponseTimeColumn:
			colTypes = append(colTypes, "time")
			dataCount = len(data)
		case TableQueryResponseStringColumn:
			colTypes = append(colTypes, "string")
			dataCount = len(data)
		case TableQueryResponseNumberColumn:
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

func (table *TableQueryResponse) buildColumns(colTypes []string) (columns []tableResponseColumn, err error) {
	for index, entry := range colTypes {
		columns = append(columns, tableResponseColumn{
			Text: table.Columns[index].Text,
			Type: entry,
		})
	}
	return
}

func (table *TableQueryResponse) buildRows(rowCount int) (rows []tableResponseRow, err error) {
	for row := 0; row < rowCount; row++ {
		newRow := make(tableResponseRow, len(table.Columns))

		for column, entry := range table.Columns {
			switch data := entry.Data.(type) {
			case TableQueryResponseTimeColumn:
				newRow[column] = data[row]
			case TableQueryResponseStringColumn:
				newRow[column] = data[row]
			case TableQueryResponseNumberColumn:
				newRow[column] = data[row]
			}
		}

		rows = append(rows, newRow)

	}
	return
}
