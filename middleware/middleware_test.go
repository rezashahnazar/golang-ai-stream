package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Test with existing request ID
	t.Run("with existing request ID", func(t *testing.T) {
		req.Header.Set("X-Request-ID", "test-id")
		Logger(handler).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "test-id", w.Header().Get("X-Request-ID"))
	})

	// Test without request ID
	t.Run("without request ID", func(t *testing.T) {
		req.Header.Del("X-Request-ID")
		Logger(handler).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
	})
}

func TestSecurityHeaders(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	SecurityHeaders(handler).ServeHTTP(w, req)

	expectedHeaders := map[string]string{
		"X-Content-Type-Options":     "nosniff",
		"X-Frame-Options":            "DENY",
		"X-XSS-Protection":           "1; mode=block",
		"Strict-Transport-Security":  "max-age=31536000; includeSubDomains",
		"Referrer-Policy":            "strict-origin-when-cross-origin",
	}

	for header, expected := range expectedHeaders {
		assert.Equal(t, expected, w.Header().Get(header))
	}
}

func TestCORS(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "OPTIONS request",
			method:         "OPTIONS",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GET request",
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/test", nil)
			w := httptest.NewRecorder()

			CORS(handler).ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
			assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
		})
	}
}

func TestRateLimit(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	limiter := NewRateLimiter(2) // 2 requests per second
	rateLimitedHandler := RateLimit(limiter)(handler)

	makeRequest := func() int {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		rateLimitedHandler.ServeHTTP(w, req)
		return w.Code
	}

	// First request should succeed
	assert.Equal(t, http.StatusOK, makeRequest())

	// Second request should succeed
	assert.Equal(t, http.StatusOK, makeRequest())

	// Third request should be rate limited
	assert.Equal(t, http.StatusTooManyRequests, makeRequest())

	// Wait for token bucket to refill
	time.Sleep(time.Second)

	// Request should succeed again
	assert.Equal(t, http.StatusOK, makeRequest())
}

func TestResponseWriter(t *testing.T) {
	w := httptest.NewRecorder()
	rw := &responseWriter{
		ResponseWriter: w,
		status:        http.StatusOK,
	}

	t.Run("write header once", func(t *testing.T) {
		rw.WriteHeader(http.StatusBadRequest)
		rw.WriteHeader(http.StatusOK) // Should not change the status
		assert.Equal(t, http.StatusBadRequest, rw.status)
	})

	t.Run("automatic status on write", func(t *testing.T) {
		rw = &responseWriter{
			ResponseWriter: w,
		}
		rw.Write([]byte("test"))
		assert.Equal(t, http.StatusOK, rw.status)
	})
} 