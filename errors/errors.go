package errors

import (
	"encoding/json"
	"net/http"
	"time"
)

type APIError struct {
	Message    string `json:"message"`
	Code       int    `json:"code"`
	ErrorType  string `json:"error_type,omitempty"`
	RequestID  string `json:"request_id,omitempty"`
	Timestamp  string `json:"timestamp"`
}

func NewAPIError(message string, code int) *APIError {
	return &APIError{
		Message:   message,
		Code:      code,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

func (e *APIError) WithType(errorType string) *APIError {
	e.ErrorType = errorType
	return e
}

func (e *APIError) WithRequestID(requestID string) *APIError {
	e.RequestID = requestID
	return e
}

func (e *APIError) RespondWithError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.Code)
	json.NewEncoder(w).Encode(e)
}

// Common error types
var (
	ErrBadRequest = func(msg string) *APIError {
		return NewAPIError(msg, http.StatusBadRequest).WithType("bad_request")
	}
	
	ErrUnauthorized = func(msg string) *APIError {
		return NewAPIError(msg, http.StatusUnauthorized).WithType("unauthorized")
	}
	
	ErrInternalServer = func(msg string) *APIError {
		return NewAPIError(msg, http.StatusInternalServerError).WithType("internal_server_error")
	}

	ErrTooManyRequests = func(msg string) *APIError {
		return NewAPIError(msg, http.StatusTooManyRequests).WithType("too_many_requests")
	}
) 