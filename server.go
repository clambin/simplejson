package simplejson

import (
	"github.com/clambin/httpserver"
	"net/http"
	"sort"
	"time"
)

// Server receives SimpleJSON requests from Grafana and dispatches them to the handler that serves the specified target.
type Server struct {
	Handlers          map[string]Handler
	queryMetrics      *QueryMetrics
	httpServerOptions []httpserver.Option
	httpServer        *httpserver.Server
}

func New(handlers map[string]Handler, options ...Option) (*Server, error) {
	s := Server{Handlers: handlers}
	for _, o := range options {
		o.apply(&s)
	}

	s.httpServerOptions = append(s.httpServerOptions, httpserver.WithHandlers{Handlers: []httpserver.Handler{
		{Path: "/", Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })},
		{Path: "/search", Handler: http.HandlerFunc(s.Search), Methods: []string{http.MethodPost}},
		{Path: "/query", Handler: http.HandlerFunc(s.Query), Methods: []string{http.MethodPost}},
		{Path: "/annotations", Handler: http.HandlerFunc(s.Annotations), Methods: []string{http.MethodPost, http.MethodOptions}},
		{Path: "/tag-keys", Handler: http.HandlerFunc(s.TagKeys), Methods: []string{http.MethodPost}},
		{Path: "/tag-values", Handler: http.HandlerFunc(s.TagValues), Methods: []string{http.MethodPost}},
	}})

	var err error
	s.httpServer, err = httpserver.New(s.httpServerOptions...)

	return &s, err
}

// Run starts the SimpleJSon Server.
func (s *Server) Run() error {
	return s.httpServer.Run()
}

// Shutdown stops a running Server.
func (s *Server) Shutdown(timeout time.Duration) error {
	return s.httpServer.Shutdown(timeout)
}

// Targets returns a sorted list of supported targets
func (s *Server) Targets() []string {
	var targets []string
	for target := range s.Handlers {
		targets = append(targets, target)
	}
	sort.Strings(targets)
	return targets
}
