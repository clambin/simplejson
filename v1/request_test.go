package simplejson_test

import (
	"encoding/json"
	"github.com/clambin/simplejson/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestRequests(t *testing.T) {
	input := `{
	"maxDataPoints": 100,
	"interval": "1h",
	"range": {
		"from": "2020-01-01T00:00:00.000Z",
		"to": "2020-12-31T00:00:00.000Z"
	},
	"targets": [
		{ "target": "A", "type": "dataserie" },
		{ "target": "B", "type": "table" }
	]
}`

	var output simplejson.QueryRequest

	err := json.Unmarshal([]byte(input), &output)
	require.NoError(t, err)
	assert.Equal(t, uint64(100), output.MaxDataPoints)
	// assert.Equal(t, server.QueryRequestDuration(1*time.Hour), output.Interval)
	// assert.Equal(t, 1*time.Hour, time.Duration(output.Interval))
	assert.Equal(t, time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), output.Range.From)
	assert.Equal(t, time.Date(2020, 12, 31, 0, 0, 0, 0, time.UTC), output.Range.To)
	require.Len(t, output.Targets, 2)
	assert.Equal(t, "A", output.Targets[0].Target)
	assert.Equal(t, "dataserie", output.Targets[0].Type)
	assert.Equal(t, "B", output.Targets[1].Target)
	assert.Equal(t, "table", output.Targets[1].Type)
}
