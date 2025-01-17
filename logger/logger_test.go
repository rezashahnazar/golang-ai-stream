package logger

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestLogRequest(t *testing.T) {
	tests := []struct {
		name      string
		level     Level
		requestID string
		method    string
		path      string
		status    int
		duration  time.Duration
		msg       string
	}{
		{
			name:      "info request with all fields",
			level:     INFO,
			requestID: "test-id",
			method:    "GET",
			path:      "/test",
			status:    200,
			duration:  100 * time.Millisecond,
			msg:      "test message",
		},
		{
			name:      "error request",
			level:     ERROR,
			requestID: "test-id",
			method:    "POST",
			path:      "/test",
			status:    500,
			duration:  200 * time.Millisecond,
			msg:      "error message",
		},
		{
			name:      "warning request",
			level:     WARNING,
			requestID: "test-id",
			method:    "PUT",
			path:      "/test",
			status:    400,
			duration:  150 * time.Millisecond,
			msg:      "warning message",
		},
		{
			name:      "debug request",
			level:     DEBUG,
			requestID: "test-id",
			method:    "DELETE",
			path:      "/test",
			status:    300,
			duration:  50 * time.Millisecond,
			msg:      "debug message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				LogRequest(tt.level, tt.requestID, tt.method, tt.path, tt.status, tt.duration, tt.msg)
			})

			// Verify all components are present
			assert.Contains(t, output, string(tt.level))
			assert.Contains(t, output, tt.requestID)
			assert.Contains(t, output, tt.method)
			assert.Contains(t, output, tt.path)
			assert.Contains(t, output, tt.msg)
		})
	}
}

func TestLogError(t *testing.T) {
	requestID := "test-id"
	err := errors.New("test error")
	msg := "error occurred"

	output := captureOutput(func() {
		LogError(requestID, err, msg)
	})

	assert.Contains(t, output, string(ERROR))
	assert.Contains(t, output, requestID)
	assert.Contains(t, output, msg)
	assert.Contains(t, output, err.Error())
}

func TestLogInfo(t *testing.T) {
	msg := "info message"

	output := captureOutput(func() {
		LogInfo(msg)
	})

	assert.Contains(t, output, string(INFO))
	assert.Contains(t, output, msg)
}

func TestColorizeLevel(t *testing.T) {
	tests := []struct {
		name  string
		level Level
		want  string
	}{
		{
			name:  "info level",
			level: INFO,
			want:  Green + string(INFO) + Reset,
		},
		{
			name:  "error level",
			level: ERROR,
			want:  Red + string(ERROR) + Reset,
		},
		{
			name:  "warning level",
			level: WARNING,
			want:  Yellow + string(WARNING) + Reset,
		},
		{
			name:  "debug level",
			level: DEBUG,
			want:  Cyan + string(DEBUG) + Reset,
		},
		{
			name:  "unknown level",
			level: "UNKNOWN",
			want:  "UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := colorizeLevel(tt.level)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatMethod(t *testing.T) {
	tests := []struct {
		name   string
		method string
		want   string
	}{
		{
			name:   "GET method",
			method: "GET",
			want:   Blue + "GET" + Reset,
		},
		{
			name:   "POST method",
			method: "POST",
			want:   Green + "POST" + Reset,
		},
		{
			name:   "PUT method",
			method: "PUT",
			want:   Yellow + "PUT" + Reset,
		},
		{
			name:   "DELETE method",
			method: "DELETE",
			want:   Red + "DELETE" + Reset,
		},
		{
			name:   "OPTIONS method",
			method: "OPTIONS",
			want:   Cyan + "OPTIONS" + Reset,
		},
		{
			name:   "unknown method",
			method: "UNKNOWN",
			want:   "UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatMethod(tt.method)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatStatus(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   string
	}{
		{
			name:   "2xx status",
			status: 200,
			want:   Green + "200" + Reset,
		},
		{
			name:   "3xx status",
			status: 301,
			want:   Cyan + "301" + Reset,
		},
		{
			name:   "4xx status",
			status: 404,
			want:   Yellow + "404" + Reset,
		},
		{
			name:   "5xx status",
			status: 500,
			want:   Red + "500" + Reset,
		},
		{
			name:   "unknown status",
			status: 100,
			want:   "100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatStatus(tt.status)
			assert.Equal(t, tt.want, got)
		})
	}
} 