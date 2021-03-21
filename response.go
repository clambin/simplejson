package grafana_json

import (
	"encoding/json"
	"errors"
	"time"
)

// Query (timeserie)

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

// Table Query

type TableQueryResponse struct {
	Columns []TableQueryResponseColumn
}

type TableQueryResponseColumn struct {
	Text string
	Data interface{}
}

type TableQueryResponseTimeColumn []time.Time
type TableQueryResponseStringColumn []string
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

// Annotations

type Annotation struct {
	request AnnotationRequestDetails
	Time    time.Time
	Title   string
	Text    string
	Tags    []string
}

// must be an easier way than this?
func (annotation *Annotation) MarshalJSON() (output []byte, err error) {
	jsonResponse := struct {
		Request AnnotationRequestDetails `json:"annotation"`
		Time    int64                    `json:"time"`
		Title   string                   `json:"title"`
		Text    string                   `json:"text"`
		Tags    []string                 `json:"tags"`
	}{
		Request: annotation.request,
		Time:    annotation.Time.UnixNano() / 1000000,
		Title:   annotation.Title,
		Text:    annotation.Text,
		Tags:    annotation.Tags,
	}

	return json.Marshal(jsonResponse)
}
