package dataset

import (
	"github.com/clambin/simplejson/v3/query"
	"time"
)

// Dataset is a convenience data structure to construct a SimpleJSON table response. Use this when you're adding
// data for a range of (possibly out of order) timestamps.
type Dataset struct {
	data       [][]float64
	timestamps *Indexer[time.Time]
	columns    *Indexer[string]
}

// New creates a new Dataset
func New() *Dataset {
	return &Dataset{
		timestamps: MakeIndexer[time.Time](),
		columns:    MakeIndexer[string](),
	}
}

// Add adds a value for a specified timestamp and column to the dataset.  If there is already a value for that
// timestamp and column, the specified value is added to the existing value.
func (d *Dataset) Add(timestamp time.Time, column string, value float64) {
	d.ensureColumnExists(column)

	row, added := d.timestamps.Add(timestamp)
	if added {
		d.data = append(d.data, make([]float64, d.columns.Count()))
	}
	col, _ := d.columns.GetIndex(column)
	d.data[row][col] += value
}

func (d *Dataset) ensureColumnExists(column string) {
	_, added := d.columns.Add(column)
	if !added {
		return
	}

	// new column. add data for the new column to each row
	for key, entry := range d.data {
		entry = append(entry, 0)
		d.data[key] = entry
	}
}

// Size returns the number of rows in the dataset.
func (d Dataset) Size() int {
	return d.timestamps.Count()
}

// AddColumn adds a new column to the dataset. For each timestamp, processor is called with the values for the
// existing columns. Processor's return value is then added for the new column.
func (d *Dataset) AddColumn(column string, processor func(values map[string]float64) float64) {
	columns := d.columns.List()
	for index, row := range d.data {
		d.data[index] = append(row, processor(d.rowValues(row, columns)))
	}
	d.columns.Add(column)
}

func (d Dataset) rowValues(row []float64, columns []string) (values map[string]float64) {
	values = make(map[string]float64)
	for _, column := range columns {
		idx, _ := d.columns.GetIndex(column)
		values[column] = row[idx]
	}
	return
}

// GetTimestamps returns the (sorted) list of timestamps in the dataset.
func (d Dataset) GetTimestamps() (timestamps []time.Time) {
	return d.timestamps.List()
}

// GetColumns returns the (sorted) list of column names.
func (d Dataset) GetColumns() (columns []string) {
	return d.columns.List()
}

// GetValues returns the value for the specified column for each timestamp in the dataset. The values are sorted by timestamp.
func (d Dataset) GetValues(column string) (values []float64, ok bool) {
	var index int
	index, ok = d.columns.GetIndex(column)

	if !ok {
		return
	}

	values = make([]float64, len(d.data))
	for i, timestamp := range d.timestamps.List() {
		rowIndex, _ := d.timestamps.GetIndex(timestamp)
		values[i] = d.data[rowIndex][index]
	}
	return
}

// FilterByRange removes any rows in the dataset that are outside the specified from/to time range. If from/to is zero,
// it is ignored.
func (d *Dataset) FilterByRange(from, to time.Time) {
	// make a list of all records to be removed, and the remaining timestamps
	timestamps := make([]time.Time, 0, d.timestamps.Count())
	var remove bool
	for _, timestamp := range d.timestamps.List() {
		if !from.IsZero() && timestamp.Before(from) {
			remove = true
			continue
		} else if !to.IsZero() && timestamp.After(to) {
			remove = true
			continue
		}
		timestamps = append(timestamps, timestamp)
	}

	// nothing to do here?
	if !remove {
		return
	}

	// create a new data list from the timestamps we want to keep
	data := make([][]float64, len(timestamps))
	ts := MakeIndexer[time.Time]()
	for index, timestamp := range timestamps {
		i, _ := d.timestamps.GetIndex(timestamp)
		data[index] = d.data[i]
		ts.Add(timestamp)
	}
	d.data = data
	d.timestamps = ts
}

// Accumulate accumulates the values for each column by time. E.g. if the values were 1, 1, 1, 1, the result would be
// 1, 2, 3, 4.
func (d *Dataset) Accumulate() {
	accumulated := make([]float64, d.columns.Count())

	for _, timestamp := range d.timestamps.List() {
		row, _ := d.timestamps.GetIndex(timestamp)
		for index, value := range d.data[row] {
			accumulated[index] += value
		}
		copy(d.data[row], accumulated)
	}
}

// Copy returns a copy of the dataset
func (d Dataset) Copy() (clone *Dataset) {
	clone = &Dataset{
		data:       make([][]float64, len(d.data)),
		timestamps: d.timestamps.Copy(),
		columns:    d.columns.Copy(),
	}
	for index, row := range d.data {
		clone.data[index] = make([]float64, len(row))
		copy(clone.data[index], row)
	}
	return
}

// GenerateTableResponse creates a TableResponse for the dataset
func (d Dataset) GenerateTableResponse() (response *query.TableResponse) {
	response = &query.TableResponse{
		Columns: []query.Column{{
			Text: "timestamp",
			Data: query.TimeColumn(d.GetTimestamps()),
		}},
	}

	for _, column := range d.GetColumns() {
		values, _ := d.GetValues(column)
		if column == "" {
			column = "(unknown)"
		}
		response.Columns = append(response.Columns, query.Column{
			Text: column,
			Data: query.NumberColumn(values),
		})
	}
	return
}
