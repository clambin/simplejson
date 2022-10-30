package middleware

import (
	"net/http"
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
