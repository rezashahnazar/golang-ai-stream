package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"golang-ai-stream/config"
	"golang-ai-stream/handlers"
	"golang-ai-stream/logger"
	"golang-ai-stream/middleware"

	"github.com/gorilla/mux"
	"github.com/sashabaranov/go-openai"
)

// openAIClientWrapper wraps the OpenAI client to implement our interface
type openAIClientWrapper struct {
	client *openai.Client
}

type streamWrapper struct {
	stream *openai.ChatCompletionStream
}

func (s *streamWrapper) Recv() (*openai.ChatCompletionStreamResponse, error) {
	resp, err := s.stream.Recv()
	if err != nil {
		return nil, err
	}
	return &openai.ChatCompletionStreamResponse{
		ID:      resp.ID,
		Object:  resp.Object,
		Created: resp.Created,
		Model:   resp.Model,
		Choices: resp.Choices,
	}, nil
}

func (s *streamWrapper) Close() {
	s.stream.Close()
}

func (w *openAIClientWrapper) CreateChatCompletionStream(ctx context.Context, req openai.ChatCompletionRequest) (handlers.ChatCompletionStreamer, error) {
	stream, err := w.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, err
	}
	return &streamWrapper{stream: stream}, nil
}

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.LogError("", err, "Failed to load configuration")
		os.Exit(1)
	}

	// Initialize OpenAI client
	config := openai.DefaultConfig(cfg.APIKey)
	config.BaseURL = cfg.BaseURL
	client := openai.NewClientWithConfig(config)

	// Wrap the OpenAI client
	clientWrapper := &openAIClientWrapper{client: client}

	// Initialize handlers
	chatHandler := handlers.NewChatHandler(clientWrapper, cfg)

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimit)

	// Setup router with middleware
	r := mux.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.CORS)
	r.Use(middleware.RateLimit(rateLimiter))
	
	// Routes
	r.HandleFunc("/chat", chatHandler.HandleChat).Methods("POST", "OPTIONS")
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Create server with timeouts
	srv := &http.Server{
		Addr:         cfg.Port,
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.ReadTimeoutSecs) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeoutSecs) * time.Second,
		IdleTimeout:  time.Duration(cfg.IdleTimeoutSecs) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.LogInfo(fmt.Sprintf("Server running on http://localhost%s", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.LogError("", err, "Server failed to start")
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	logger.LogInfo("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ReadTimeoutSecs)*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.LogError("", err, "Server forced to shutdown")
	}
	logger.LogInfo("Server gracefully stopped")
}
