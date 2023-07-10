package simplejson

import (
	"encoding/json"
	"errors"
	"github.com/mailru/easyjson"
	"strconv"
	"time"
)

// QueryRequest is a Query request. For each specified Target, the server will call the appropriate handler's Query or TableQuery
// function with the provided QueryArgs.
//
//easyjson:skip
type QueryRequest struct {
	Targets []Target `json:"targets"`
	QueryArgs
}

// Target specifies the requested target name and type.
//
//easyjson:skip
type Target struct {
	Name string      `json:"target"` // name of the target.
	Type string      `json:"type"`   // "timeserie" or "" for timeseries. "table" for table queries.
	Data interface{} `json:"data,omitempty"`
}

// QueryArgs contains the arguments for a Query.
//
//easyjson:skip
type QueryArgs struct {
	Args
	MaxDataPoints uint64 `json:"maxDataPoints"`
}

// UnmarshalJSON unmarshalls a QueryRequest from JSON
func (r *QueryRequest) UnmarshalJSON(b []byte) (err error) {
	// workaround to avoid infinite loop
	type Request2 QueryRequest
	var c Request2
	err = json.Unmarshal(b, &c)
	if err == nil {
		*r = QueryRequest(c)
	}
	return err
}

// TimeSeriesResponse is the response from a timeseries Query.
type TimeSeriesResponse struct {
	Target     string      `json:"target"`
	DataPoints []DataPoint `json:"datapoints"`
}

// DataPoint contains one entry returned by a Query.
//
//easyjson:skip
type DataPoint struct {
	Timestamp time.Time
	Value     float64
}

// MarshalJSON converts a DataPoint to JSON.
func (d DataPoint) MarshalJSON() ([]byte, error) {
	return []byte(`[` +
			strconv.FormatFloat(d.Value, 'f', -1, 64) + `,` +
			strconv.FormatInt(d.Timestamp.UnixMilli(), 10) +
			`]`),
		nil
}

// TableResponse is returned by a TableQuery, i.e. a slice of Column structures.
//
//easyjson:skip
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

	if colTypes, rowCount, err = t.getColumnDetails(); err == nil {
		output, err = easyjson.Marshal(tableResponse{
			Type:    "table",
			Columns: t.buildColumns(colTypes),
			Rows:    t.buildRows(rowCount),
		})
	}

	return output, err
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
			return colTypes, rowCount, errors.New("error building table query output: all columns must have the same number of rows")
		}
	}
	return
}

func (t TableResponse) buildColumns(colTypes []string) []tableResponseColumn {
	columns := make([]tableResponseColumn, len(colTypes))
	for index, colType := range colTypes {
		columns[index] = tableResponseColumn{
			Text: t.Columns[index].Text,
			Type: colType,
		}
	}
	return columns
}

func (t TableResponse) buildRows(rowCount int) []tableResponseRow {
	rows := make([]tableResponseRow, rowCount)
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
		rows[row] = newRow
	}
	return rows
}
