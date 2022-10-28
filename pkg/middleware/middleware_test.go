package middleware_test

import (
	"github.com/clambin/simplejson/v3/pkg/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHandleForMethods(t *testing.T) {
	handler := middleware.HandleForMethods(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("OK"))
		}),
		http.MethodPost, http.MethodDelete)

	testCases := []struct {
		method string
		code   int
	}{
		{method: http.MethodGet, code: http.StatusMethodNotAllowed},
		{method: http.MethodPost, code: http.StatusOK},
		{method: http.MethodDelete, code: http.StatusOK},
		{method: http.MethodHead, code: http.StatusMethodNotAllowed},
	}

	for _, tt := range testCases {
		t.Run(tt.method, func(t *testing.T) {
			resp := httptest.NewRecorder()
			req, err := http.NewRequest(tt.method, "", nil)
			require.NoError(t, err)
			handler.ServeHTTP(resp, req)

			assert.Equal(t, tt.code, resp.Code)
		})
	}
}

func TestHTTPMetrics(t *testing.T) {
	var path, method string
	var statusCode int
	var duration time.Duration

	handler := middleware.HTTPMetrics(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			time.Sleep(10 * time.Millisecond)
			_, _ = w.Write([]byte("OK"))
			w.WriteHeader(http.StatusOK)
		}),
		func(p, m string, sc int, d time.Duration) {
			path = p
			method = m
			statusCode = sc
			duration = d
		})

	var testCases = []struct {
		path   string
		method string
	}{
		{path: "/", method: http.MethodGet},
		{path: "/foo", method: http.MethodPost},
		{path: "/bar", method: http.MethodDelete},
	}

	for _, tt := range testCases {
		t.Run(tt.path, func(t *testing.T) {
			resp := httptest.NewRecorder()
			req, err := http.NewRequest(tt.method, tt.path, nil)
			require.NoError(t, err)
			handler.ServeHTTP(resp, req)

			assert.Equal(t, http.StatusOK, resp.Code)
			assert.Equal(t, tt.path, path)
			assert.Equal(t, tt.method, method)
			assert.Equal(t, http.StatusOK, statusCode)
			assert.NotZero(t, duration)
		})
	}
}
