package simplejson

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAnnotation_MarshalJSON(t *testing.T) {
	ann := Annotation{
		Time:  time.Date(2022, time.January, 23, 0, 0, 0, 0, time.UTC),
		Title: "foo",
		Text:  "bar",
		Tags:  []string{"A", "B"},
		Request: RequestDetails{
			Name:       "snafu",
			Datasource: "datasource",
			Enable:     true,
			Query:      "A == 10",
		},
	}

	body, err := json.Marshal(ann)
	require.NoError(t, err)

	gp := filepath.Join("testdata", strings.ToLower(t.Name()), "1.golden")
	if *update {
		t.Logf("updating golden file for %s", t.Name())
		err = os.WriteFile(gp, body, 0644)
		require.NoError(t, err, "failed to update golden file")
	}

	var golden []byte
	golden, err = os.ReadFile(gp)
	require.NoError(t, err)
	assert.Equal(t, golden, body)

	ann.TimeEnd = time.Date(2022, time.January, 23, 0, 0, 0, 0, time.UTC)

	body, err = json.Marshal(ann)
	require.NoError(t, err)
	assert.Equal(t, golden, body)

	ann.TimeEnd = time.Date(2022, time.January, 23, 1, 0, 0, 0, time.UTC)

	body, err = json.Marshal(ann)
	require.NoError(t, err)

	gp = filepath.Join("testdata", strings.ToLower(t.Name()), "2.golden")
	if *update {
		t.Logf("updating golden file for %s", t.Name())
		err = os.WriteFile(gp, body, 0644)
		require.NoError(t, err, "failed to update golden file")
	}

	golden, err = os.ReadFile(gp)
	require.NoError(t, err)

	assert.Equal(t, golden, body)
}
