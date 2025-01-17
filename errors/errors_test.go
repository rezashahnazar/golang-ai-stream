package errors

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewAPIError(t *testing.T) {
	message := "test error"
	code := http.StatusBadRequest

	err := NewAPIError(message, code)

	assert.Equal(t, message, err.Message)
	assert.Equal(t, code, err.Code)
	assert.NotEmpty(t, err.Timestamp)

	// Verify timestamp format
	_, parseErr := time.Parse(time.RFC3339, err.Timestamp)
	assert.NoError(t, parseErr)
}

func TestAPIError_WithType(t *testing.T) {
	err := NewAPIError("test error", http.StatusBadRequest)
	errorType := "test_error"

	err = err.WithType(errorType)

	assert.Equal(t, errorType, err.ErrorType)
}

func TestAPIError_WithRequestID(t *testing.T) {
	err := NewAPIError("test error", http.StatusBadRequest)
	requestID := "test-request-id"

	err = err.WithRequestID(requestID)

	assert.Equal(t, requestID, err.RequestID)
}

func TestAPIError_RespondWithError(t *testing.T) {
	err := NewAPIError("test error", http.StatusBadRequest).
		WithType("test_error").
		WithRequestID("test-request-id")

	w := httptest.NewRecorder()
	err.RespondWithError(w)

	// Check status code
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Check content type
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	// Verify response body
	var response APIError
	decodeErr := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, decodeErr)
	assert.Equal(t, err.Message, response.Message)
	assert.Equal(t, err.Code, response.Code)
	assert.Equal(t, err.ErrorType, response.ErrorType)
	assert.Equal(t, err.RequestID, response.RequestID)
	assert.Equal(t, err.Timestamp, response.Timestamp)
}

func TestCommonErrors(t *testing.T) {
	tests := []struct {
		name           string
		errorFunc      func(string) *APIError
		expectedCode   int
		expectedType   string
		expectedStatus int
	}{
		{
			name:           "bad request error",
			errorFunc:      ErrBadRequest,
			expectedCode:   http.StatusBadRequest,
			expectedType:   "bad_request",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized error",
			errorFunc:      ErrUnauthorized,
			expectedCode:   http.StatusUnauthorized,
			expectedType:   "unauthorized",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "internal server error",
			errorFunc:      ErrInternalServer,
			expectedCode:   http.StatusInternalServerError,
			expectedType:   "internal_server_error",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "too many requests error",
			errorFunc:      ErrTooManyRequests,
			expectedCode:   http.StatusTooManyRequests,
			expectedType:   "too_many_requests",
			expectedStatus: http.StatusTooManyRequests,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := "test message"
			err := tt.errorFunc(message)

			assert.Equal(t, message, err.Message)
			assert.Equal(t, tt.expectedCode, err.Code)
			assert.Equal(t, tt.expectedType, err.ErrorType)

			// Test response
			w := httptest.NewRecorder()
			err.RespondWithError(w)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
} 