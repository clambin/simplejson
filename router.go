package simplejson

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-http-utils/headers"
	"net/http"
)

func NewRouter(handlers map[string]Handler, options ...Option) *Server {
	s := Server{
		Handlers: handlers,
		Router:   chi.NewRouter(),
	}
	for _, o := range options {
		o.apply(&s)
	}

	s.Router.Use(middleware.Heartbeat("/"))
	s.Router.Group(func(r chi.Router) {
		r.Use(middleware.Logger)
		if s.prometheusMetrics != nil {
			r.Use(s.prometheusMetrics.Handle)
		}
		r.Post("/search", s.Search)
		r.Post("/query", s.Query)
		r.Post("/annotations", s.Annotations)
		r.Options("/annotations", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set(headers.AccessControlAllowOrigin, "*")
			w.Header().Set(headers.AccessControlAllowMethods, "POST")
			w.Header().Set(headers.AccessControlAllowHeaders, "accept, content-type")
		})
		r.Post("/tag-keys", s.TagValues)
		r.Post("/tag-values", s.TagValues)
	})

	return &s
}
