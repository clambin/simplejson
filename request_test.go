package grafana_json_test

import (
	"encoding/json"
	"github.com/clambin/grafana-json"
	"github.com/stretchr/testify/assert"
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

	var output grafana_json.QueryRequest

	if err := json.Unmarshal([]byte(input), &output); assert.Nil(t, err) {
		assert.Equal(t, uint64(100), output.MaxDataPoints)
		// assert.Equal(t, server.QueryRequestDuration(1*time.Hour), output.Interval)
		// assert.Equal(t, 1*time.Hour, time.Duration(output.Interval))
		assert.Equal(t, time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), output.Range.From)
		assert.Equal(t, time.Date(2020, 12, 31, 0, 0, 0, 0, time.UTC), output.Range.To)
		if assert.Len(t, output.Targets, 2); err == nil {
			assert.Equal(t, "A", output.Targets[0].Target)
			assert.Equal(t, "dataserie", output.Targets[0].Type)
			assert.Equal(t, "B", output.Targets[1].Target)
			assert.Equal(t, "table", output.Targets[1].Type)
		}
	}
}
