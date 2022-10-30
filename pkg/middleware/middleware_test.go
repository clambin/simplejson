package middleware_test

import (
	"github.com/clambin/simplejson/v3/pkg/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
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
