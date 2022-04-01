package dataset

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestLessThan(t *testing.T) {
	assert.True(t, isLessThan("A", "B"))

	ts := time.Now()
	assert.True(t, isLessThan(ts, ts.Add(time.Hour)))
	assert.False(t, isLessThan(ts, ts.Add(-time.Hour)))
}
