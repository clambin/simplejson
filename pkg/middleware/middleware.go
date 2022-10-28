package middleware

import (
	"net/http"
	"time"
)

// HandleForMethods wil only call the next handler if the request is for the specified method(s)
func HandleForMethods(next http.Handler, methods ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isValidMethod(r.Method, methods) {
			next.ServeHTTP(w, r)
			return
		}
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	})
}

func isValidMethod(method string, methods []string) bool {
	for _, m := range methods {
		if m == method {
			return true
		}
	}
	return false
}

type MetricsRecorder func(path, method string, statusCode int, duration time.Duration)

// HTTPMetrics will call the next handler and pass statistics to a MetricsRecorder to store the required metrics
func HTTPMetrics(next http.Handler, record MetricsRecorder) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lrw := &loggingResponseWriter{
			responseWriter: w,
			statusCode:     http.StatusOK, // if the handler doesn't call WriteHeader(), default to HTTP 200
		}
		start := time.Now()
		next.ServeHTTP(lrw, r)
		record(r.URL.Path, r.Method, lrw.statusCode, time.Since(start))
	})
}

// loggingResponseWriter records the HTTP status code of a ResponseWriter, so we can use it to log response times for
// individual status codes.
type loggingResponseWriter struct {
	responseWriter http.ResponseWriter
	statusCode     int
}

// WriteHeader implements the http.ResponseWriter interface.
func (w *loggingResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.responseWriter.WriteHeader(code)
}

// Write implements the http.ResponseWriter interface.
func (w *loggingResponseWriter) Write(body []byte) (int, error) {
	return w.responseWriter.Write(body)
}

// Header implements the http.ResponseWriter interface
func (w *loggingResponseWriter) Header() http.Header {
	return w.responseWriter.Header()
}
