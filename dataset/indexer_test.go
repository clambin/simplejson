package dataset_test

import (
	"fmt"
	"github.com/clambin/simplejson/v3/dataset"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestIndexer_Time(t *testing.T) {
	idx := dataset.MakeIndexer[time.Time]()
	indices := make(map[time.Time]int)

	const iterations = 100

	for i := 0; i < iterations; i++ {
		value := time.Date(2022, time.March, 1+rand.Intn(31), 0, 0, 0, 0, time.UTC)

		if index, added := idx.Add(value); added {
			indices[value] = index
		}
	}

	assert.Len(t, indices, idx.Count())

	for value, index := range indices {
		i, found := idx.GetIndex(value)
		assert.True(t, found)
		assert.Equal(t, i, index)
	}
}

func TestIndexer_String(t *testing.T) {
	idx := dataset.MakeIndexer[string]()
	indices := make(map[string]int)

	const iterations = 100

	for i := 0; i < iterations; i++ {
		value := fmt.Sprintf("%02d", 1+rand.Intn(31))

		if index, added := idx.Add(value); added {
			indices[value] = index
		}
	}

	assert.Len(t, indices, idx.Count())

	for value, index := range indices {
		i, found := idx.GetIndex(value)
		assert.True(t, found)
		assert.Equal(t, i, index)
	}
}

func TestIndexer_Reorder(t *testing.T) {
	input := []string{"C", "B", "A"}
	idx := dataset.MakeIndexer[string]()

	for index, value := range input {
		i, added := idx.Add(value)
		assert.True(t, added)
		assert.Equal(t, index, i)
	}

	result := idx.List()
	assert.Equal(t, []string{"A", "B", "C"}, result)

	for index, value := range input {
		i, found := idx.GetIndex(value)
		assert.True(t, found)
		assert.Equal(t, index, i, value)
	}
}
