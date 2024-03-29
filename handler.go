package simplejson

import "context"

// Handler implements the different Grafana SimpleJSON endpoints.  The interface only contains a single Endpoints() function,
// so that a handler only has to implement the endpoint functions (query, annotation, etc.) that it needs.
type Handler interface {
	Endpoints() Endpoints
}

// Endpoints contains the functions that implement each of the SimpleJson endpoints
type Endpoints struct {
	Query       QueryFunc       // /query endpoint: handles queries
	Annotations AnnotationsFunc // /annotation endpoint: handles requests for annotation
	TagKeys     TagKeysFunc     // /tag-keys endpoint: returns all supported tag names
	TagValues   TagValuesFunc   // /tag-values endpoint: returns all supported values for the specified tag name
}

// QueryFunc handles queries
type QueryFunc func(ctx context.Context, req QueryRequest) (Response, error)

// AnnotationsFunc handles requests for annotation
type AnnotationsFunc func(req AnnotationRequest) ([]Annotation, error)

// TagKeysFunc returns supported tag names
type TagKeysFunc func(ctx context.Context) []string

// TagValuesFunc returns supported values for the specified tag name
type TagValuesFunc func(ctx context.Context, key string) ([]string, error)
