package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"golang-ai-stream/config"
	"golang-ai-stream/handlers"
	"golang-ai-stream/middleware"

	"github.com/gorilla/mux"
	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
)

type mockStream struct {
	closed bool
}

func (s *mockStream) Recv() (*openai.ChatCompletionStreamResponse, error) {
	return &openai.ChatCompletionStreamResponse{}, nil
}

func (s *mockStream) Close() {
	s.closed = true
}

type mockClient struct {
	stream *mockStream
}

func (m *mockClient) CreateChatCompletionStream(ctx context.Context, req openai.ChatCompletionRequest) (handlers.ChatCompletionStreamer, error) {
	if m.stream == nil {
		m.stream = &mockStream{}
	}
	return m.stream, nil
}

func TestCreateServer(t *testing.T) {
	// Set test environment variables
	os.Setenv("PORT", ":8081")
	os.Setenv("API_KEY", "test-key")
	os.Setenv("BASE_URL", "http://test.com")
	os.Setenv("RATE_LIMIT", "10")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("API_KEY")
		os.Unsetenv("BASE_URL")
		os.Unsetenv("RATE_LIMIT")
	}()

	// Load configuration
	cfg, err := config.LoadConfig()
	assert.NoError(t, err)

	// Create router
	r := mux.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.CORS)
	r.Use(middleware.RateLimit(middleware.NewRateLimiter(cfg.RateLimit)))

	// Add health endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Create server
	srv := &http.Server{
		Addr:         cfg.Port,
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.ReadTimeoutSecs) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeoutSecs) * time.Second,
		IdleTimeout:  time.Duration(cfg.IdleTimeoutSecs) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Errorf("Server failed to start: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Test health endpoint
	resp, err := http.Get(fmt.Sprintf("http://localhost%s/health", cfg.Port))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	assert.NoError(t, srv.Shutdown(ctx))
}

func TestRoutes(t *testing.T) {
	// Set test environment variables
	os.Setenv("PORT", ":8082")
	os.Setenv("API_KEY", "test-key")
	os.Setenv("BASE_URL", "http://test.com")
	os.Setenv("RATE_LIMIT", "10")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("API_KEY")
		os.Unsetenv("BASE_URL")
		os.Unsetenv("RATE_LIMIT")
	}()

	// Load configuration
	cfg, err := config.LoadConfig()
	assert.NoError(t, err)

	// Create router
	r := mux.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.CORS)
	r.Use(middleware.RateLimit(middleware.NewRateLimiter(cfg.RateLimit)))

	// Add routes
	chatHandler := handlers.NewChatHandler(&mockClient{}, cfg)
	r.HandleFunc("/chat", chatHandler.HandleChat).Methods("POST", "OPTIONS")
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Test routes
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "health check",
			method:         "GET",
			path:           "/health",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "chat options",
			method:         "OPTIONS",
			path:           "/chat",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "not found",
			method:         "GET",
			path:           "/notfound",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.path, nil)
			assert.NoError(t, err)

			// Add request ID to context
			req = req.WithContext(context.WithValue(req.Context(), middleware.RequestIDKey, "test-id"))

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
} 