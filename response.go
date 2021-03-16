package grafana_json

import (
	"encoding/json"
	"errors"
	"time"
)

type QueryResponse struct {
	Target     string                   `json:"target"`
	DataPoints []QueryResponseDataPoint `json:"datapoints"`
}

type QueryResponseDataPoint struct {
	Timestamp time.Time
	Value     int64
}

func (d *QueryResponseDataPoint) MarshalJSON() ([]byte, error) {
	out := []int64{d.Value, d.Timestamp.UnixNano() / 1000000}
	return json.Marshal(out)
}

func (d *QueryResponseDataPoint) UnmarshalJSON(input []byte) (err error) {
	var in []int64

	if err = json.Unmarshal(input, &in); err == nil {
		*d = QueryResponseDataPoint{
			Value:     in[0],
			Timestamp: time.Unix(0, in[1]*1000000),
		}
	}
	return
}

type QueryTableResponse struct {
	Columns []QueryTableResponseColumn
}

type QueryTableResponseColumn struct {
	Text string
	Data interface{}
}

type QueryTableResponseTimeColumn []time.Time
type QueryTableResponseStringColumn []string
type QueryTableResponseNumberColumn []float64

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

func (table *QueryTableResponse) MarshalJSON() (output []byte, err error) {
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

func (table *QueryTableResponse) getColumnDetails() (colTypes []string, rowCount int, err error) {
	for _, entry := range table.Columns {
		var dataCount int
		switch data := entry.Data.(type) {
		case QueryTableResponseTimeColumn:
			colTypes = append(colTypes, "time")
			dataCount = len(data)
		case QueryTableResponseStringColumn:
			colTypes = append(colTypes, "string")
			dataCount = len(data)
		case QueryTableResponseNumberColumn:
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

func (table *QueryTableResponse) buildColumns(colTypes []string) (columns []tableResponseColumn, err error) {
	for index, entry := range colTypes {
		columns = append(columns, tableResponseColumn{
			Text: table.Columns[index].Text,
			Type: entry,
		})
	}
	return
}

func (table *QueryTableResponse) buildRows(rowCount int) (rows []tableResponseRow, err error) {
	for row := 0; row < rowCount; row++ {
		newRow := make(tableResponseRow, len(table.Columns))

		for column, entry := range table.Columns {
			switch data := entry.Data.(type) {
			case QueryTableResponseTimeColumn:
				newRow[column] = data[row]
			case QueryTableResponseStringColumn:
				newRow[column] = data[row]
			case QueryTableResponseNumberColumn:
				newRow[column] = data[row]
			}
		}

		rows = append(rows, newRow)

	}
	return
}
