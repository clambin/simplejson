package dataset_test

import (
	"github.com/clambin/simplejson/v3/dataset"
	"github.com/clambin/simplejson/v3/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestDataset_Basic(t *testing.T) {
	d := dataset.New()
	assert.NotNil(t, d)

	for day := 1; day < 5; day++ {
		d.Add(time.Date(2022, time.January, 5-day, 0, 0, 0, 0, time.UTC), "A", float64(5-day))
	}

	d.AddColumn("B", func(values map[string]float64) float64 {
		return values["A"] * 2
	})

	assert.Equal(t, 4, d.Size())
	assert.Equal(t, []string{"A", "B"}, d.GetColumns())
	assert.Equal(t, []time.Time{
		time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2022, time.January, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2022, time.January, 3, 0, 0, 0, 0, time.UTC),
		time.Date(2022, time.January, 4, 0, 0, 0, 0, time.UTC),
	}, d.GetTimestamps())

	values, ok := d.GetValues("B")
	require.True(t, ok)
	assert.Equal(t, []float64{2, 4, 6, 8}, values)
}

func TestDataset_FilterByRange(t *testing.T) {
	d := dataset.New()
	assert.NotNil(t, d)

	for day := 1; day < 32; day++ {
		d.Add(time.Date(2022, time.January, day, 0, 0, 0, 0, time.UTC), "A", float64(day))
	}

	assert.Equal(t, 31, d.Size())

	d.FilterByRange(time.Time{}, time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC))
	assert.Equal(t, 31, d.Size())

	d.FilterByRange(time.Time{}, time.Date(2022, time.January, 30, 0, 0, 0, 0, time.UTC))
	assert.Equal(t, 30, d.Size())

	d.FilterByRange(time.Date(2022, time.January, 2, 0, 0, 0, 0, time.UTC), time.Time{})
	assert.Equal(t, 29, d.Size())

	d.FilterByRange(time.Date(2022, time.January, 8, 0, 0, 0, 0, time.UTC), time.Date(2022, time.January, 14, 0, 0, 0, 0, time.UTC))
	assert.Equal(t, 7, d.Size())

	values, ok := d.GetValues("A")
	require.True(t, ok)
	assert.Equal(t, []float64{8, 9, 10, 11, 12, 13, 14}, values)

}

func TestDataset_Accumulate(t *testing.T) {
	d := dataset.New()
	assert.NotNil(t, d)

	for day := 1; day < 32; day++ {
		d.Add(time.Date(2022, time.January, day, 0, 0, 0, 0, time.UTC), "A", 1.0)
	}

	d.Accumulate()

	values, ok := d.GetValues("A")
	require.True(t, ok)
	expected := 1.0
	for index, value := range values {
		require.Equal(t, expected, value, index)
		expected++
	}
}

func TestDataset_Copy(t *testing.T) {
	d := dataset.New()
	assert.NotNil(t, d)

	for day := 1; day < 5; day++ {
		ts := time.Date(2022, time.January, day, 0, 0, 0, 0, time.UTC)
		d.Add(ts, "A", 1.0)
	}

	clone := d.Copy()

	d.Accumulate()

	values, ok := d.GetValues("A")
	require.True(t, ok)
	assert.Equal(t, []float64{1, 2, 3, 4}, values)

	values, ok = clone.GetValues("A")
	require.True(t, ok)
	assert.Equal(t, []float64{1, 1, 1, 1}, values)
}

func TestDataset_GenerateTableResponse(t *testing.T) {
	d := dataset.New()
	assert.NotNil(t, d)

	for day := 1; day < 5; day++ {
		ts := time.Date(2022, time.January, 5-day, 0, 0, 0, 0, time.UTC)
		d.Add(ts, "", float64(5-day))
	}

	response := d.GenerateTableResponse()
	assert.Equal(t, &query.TableResponse{
		Columns: []query.Column{
			{
				Text: "timestamp",
				Data: query.TimeColumn{
					time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2022, time.January, 2, 0, 0, 0, 0, time.UTC),
					time.Date(2022, time.January, 3, 0, 0, 0, 0, time.UTC),
					time.Date(2022, time.January, 4, 0, 0, 0, 0, time.UTC),
				},
			},
			{
				Text: "(unknown)",
				Data: query.NumberColumn{1, 2, 3, 4},
			},
		},
	}, response)
}

func BenchmarkDataset_Add(b *testing.B) {
	for i := 0; i < b.N; i++ {
		d := dataset.New()
		for y := 0; y < 5; y++ {
			timestamp := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
			for day := 0; day < 365; day++ {
				d.Add(timestamp, "A", float64(day))
				timestamp = timestamp.Add(-24 * time.Hour)
			}
		}
	}
}

func BenchmarkDataset_GetColumns(b *testing.B) {
	d := dataset.New()
	timestamp := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	for day := 0; day < 5*365; day++ {
		d.Add(timestamp, "A", float64(day))
		timestamp = timestamp.Add(-24 * time.Hour)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = d.GetColumns()
	}
}

func BenchmarkDataset_GetTimestamps(b *testing.B) {
	d := dataset.New()
	timestamp := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	for day := 0; day < 5*365; day++ {
		d.Add(timestamp, "A", float64(day))
		timestamp = timestamp.Add(-24 * time.Hour)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.GetTimestamps()
	}
}

func BenchmarkDataset_GetValues(b *testing.B) {
	d := dataset.New()
	timestamp := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	for day := 0; day < 5*365; day++ {
		d.Add(timestamp, "A", float64(day))
		timestamp = timestamp.Add(-24 * time.Hour)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.GetValues("A")
	}
}

func BenchmarkDataset_FilterByRange(b *testing.B) {
	d := dataset.New()
	timestamp := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	for day := 0; day < 5*365; day++ {
		d.Add(timestamp, "A", float64(day))
		timestamp = timestamp.Add(24 * time.Hour)
	}

	b.ResetTimer()

	start := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	stop := timestamp
	for i := 0; i < b.N; i++ {
		d.FilterByRange(start, stop)
		start = start.Add(12 * time.Hour)
		stop = stop.Add(-12 * time.Hour)
	}
}

func BenchmarkDataset_AddColumn(b *testing.B) {
	d := dataset.New()
	timestamp := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	for day := 0; day < 5*365; day++ {
		d.Add(timestamp, "A", float64(day))
		timestamp = timestamp.Add(-24 * time.Hour)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.AddColumn("B", func(values map[string]float64) float64 {
			return 1
		})
	}
}

func BenchmarkDataset_Copy(b *testing.B) {
	d := dataset.New()
	timestamp := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	for day := 0; day < 5*365; day++ {
		d.Add(timestamp, "A", float64(day))
		timestamp = timestamp.Add(-24 * time.Hour)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = d.Copy()
	}
}
