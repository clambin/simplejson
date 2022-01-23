package annotation_test

import (
	"encoding/json"
	"github.com/clambin/simplejson/v2/annotation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestAnnotation_MarshalJSON(t *testing.T) {
	ann := annotation.Annotation{
		Time:  time.Date(2022, time.January, 23, 0, 0, 0, 0, time.UTC),
		Title: "foo",
		Text:  "bar",
		Tags:  []string{"A", "B"},
		Request: annotation.RequestDetails{
			Name:       "snafu",
			Datasource: "datasource",
			Enable:     true,
			Query:      "A == 10",
		},
	}

	body, err := json.Marshal(ann)
	require.NoError(t, err)
	assert.Equal(t, `{"annotation":{"name":"snafu","datasource":"datasource","enable":true,"query":"A == 10"},"time":1642896000000,"title":"foo","text":"bar","tags":["A","B"]}`, string(body))
}
