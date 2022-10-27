package set_test

import (
	"github.com/clambin/simplejson/v3/pkg/set"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSet_Has(t *testing.T) {
	input := []string{"A", "B", "C"}
	s := set.Create(input)

	for _, entry := range input {
		assert.True(t, s.Has(entry))
	}

	assert.False(t, s.Has("invalid"))
}
