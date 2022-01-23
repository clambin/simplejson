package simplejson

import (
	"context"
	"github.com/clambin/simplejson/v2/annotation"
	"github.com/clambin/simplejson/v2/query"
)

// Handler implements the different Grafana SimpleJSON endpoints.  The interface only contains a single Endpoints() function,
// so that a handler only has to implement the endpoint functions (query, tablequery, annotation, etc.) that it needs.
type Handler interface {
	Endpoints() Endpoints
}

// Endpoints contains the functions that implement each of the SimpleJson endpoints
type Endpoints struct {
	Query       TimeSeriesQueryFunc // /query endpoint: handles timeSeries queries
	TableQuery  TableQueryFunc      // /query endpoint: handles table queries
	Annotations AnnotationsFunc     // /annotation endpoint: handles requests for annotation
	TagKeys     TagKeysFunc         // /tag-keys endpoint: returns all supported tag names
	TagValues   TagValuesFunc       // /tag-values endpoint: returns all supported values for the specified tag name
}

// TimeSeriesQueryFunc handles timeseries queries
type TimeSeriesQueryFunc func(ctx context.Context, args query.Args) (*query.TimeSeriesResponse, error)

// TableQueryFunc handles for table queries
type TableQueryFunc func(ctx context.Context, args query.Args) (*query.TableResponse, error)

// TagKeysFunc returns supported tag names
type TagKeysFunc func(ctx context.Context) []string

// TagValuesFunc returns supported values for the specified tag name
type TagValuesFunc func(ctx context.Context, key string) ([]string, error)

// AnnotationsFunc handles requests for annotation
type AnnotationsFunc func(name, query string, args annotation.Args) ([]annotation.Annotation, error)
