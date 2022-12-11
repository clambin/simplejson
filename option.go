package simplejson

import (
	"github.com/clambin/go-common/httpserver"
)

// Option specified configuration options for Server
type Option interface {
	apply(server *Server)
}

// WithQueryMetrics will collect the specified metrics to instrument the Server's Handlers.
type WithQueryMetrics struct {
	Name string
}

func (o WithQueryMetrics) apply(s *Server) {
	if o.Name == "" {
		o.Name = "simplejson"
	}
	s.queryMetrics = newQueryMetrics(o.Name)
}

// WithHTTPServerOption will pass the provided option to the underlying HTTP Server
type WithHTTPServerOption struct {
	Option httpserver.Option
}

func (o WithHTTPServerOption) apply(s *Server) {
	s.httpServerOptions = append(s.httpServerOptions, o.Option)
}
