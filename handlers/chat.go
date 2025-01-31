package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang-ai-stream/config"
	"golang-ai-stream/logger"
	"golang-ai-stream/middleware"
	"golang-ai-stream/models"

	"github.com/sashabaranov/go-openai"
)

// ChatCompletionStreamer interface for better testability
type ChatCompletionStreamer interface {
	Recv() (*openai.ChatCompletionStreamResponse, error)
	Close()
}

// OpenAIClient interface for better testability
type OpenAIClient interface {
	CreateChatCompletionStream(ctx context.Context, req openai.ChatCompletionRequest) (ChatCompletionStreamer, error)
}

type ChatHandler struct {
	client OpenAIClient
	config *config.Config
}

func NewChatHandler(client OpenAIClient, cfg *config.Config) *ChatHandler {
	return &ChatHandler{
		client: client,
		config: cfg,
	}
}

func (h *ChatHandler) validateRequest(reqBody *models.ChatRequest) error {
	if strings.TrimSpace(reqBody.Prompt) == "" {
		return fmt.Errorf("prompt cannot be empty")
	}
	if len(reqBody.Prompt) > h.config.MaxPromptLength {
		return fmt.Errorf("prompt exceeds maximum length of %d characters", h.config.MaxPromptLength)
	}
	return nil
}

func (h *ChatHandler) HandleChat(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value(middleware.RequestIDKey).(string)
	
	// Set headers before any potential error responses
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	flusher, ok := w.(http.Flusher)
	if !ok {
		logger.LogError(requestID, fmt.Errorf("streaming not supported"), "Streaming unsupported")
		chunk := models.ChatResponse{
			Content:   "Streaming unsupported by client",
			RequestID: requestID,
			Type:     "error",
		}
		writeSSEMessage(w, flusher, chunk)
		return
	}

	var reqBody models.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		logger.LogError(requestID, err, "Invalid request payload")
		chunk := models.ChatResponse{
			Content:   "Invalid request payload",
			RequestID: requestID,
			Type:     "error",
		}
		writeSSEMessage(w, flusher, chunk)
		return
	}

	if err := h.validateRequest(&reqBody); err != nil {
		logger.LogError(requestID, err, "Request validation failed")
		chunk := models.ChatResponse{
			Content:   err.Error(),
			RequestID: requestID,
			Type:     "error",
		}
		writeSSEMessage(w, flusher, chunk)
		return
	}

	logger.LogRequest(logger.INFO, requestID, r.Method, r.URL.Path, http.StatusOK, 0, 
		fmt.Sprintf("Processing chat request with prompt length: %d", len(reqBody.Prompt)))

	chatReq := openai.ChatCompletionRequest{
		Model:    "anthropic/claude-3.5-sonnet",
		Messages: []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: reqBody.Prompt}},
		Stream:   true,
	}

	stream, err := h.client.CreateChatCompletionStream(r.Context(), chatReq)
	if err != nil {
		logger.LogError(fmt.Sprintf("[%s] Error creating chat completion stream", requestID), err, "error")
		chunk := models.ChatResponse{
			Content:   "Failed to create chat completion stream",
			RequestID: requestID,
			Type:     "error",
		}
		writeSSEMessage(w, flusher, chunk)
		return
	}
	defer stream.Close()

	errCh := make(chan error, 1)
	go func() {
		for {
			response, err := stream.Recv()
			if err != nil {
				errCh <- err
				return
			}

			content := response.Choices[0].Delta.Content
			if content == "" {
				continue
			}

			chunk := models.ChatResponse{
				Content:   content,
				RequestID: requestID,
				Type:     "content",
			}
			if err := writeSSEMessage(w, flusher, chunk); err != nil {
				errCh <- err
				return
			}
		}
	}()

	select {
	case <-r.Context().Done():
		logger.LogInfo(fmt.Sprintf("[%s] Client disconnected", requestID))
		chunk := models.ChatResponse{
			Content:   "Client disconnected",
			RequestID: requestID,
			Type:     "error",
		}
		writeSSEMessage(w, flusher, chunk)
		return
	case err := <-errCh:
		if errors.Is(err, context.Canceled) {
			return // Already handled by context.Done() case
		}
		if err == io.EOF {
			chunk := models.ChatResponse{
				Content:   "",
				RequestID: requestID,
				Type:     "done",
			}
			writeSSEMessage(w, flusher, chunk)
			return
		}
		logger.LogError(fmt.Sprintf("[%s] Failed to receive chat completion", requestID), err, "error")
		chunk := models.ChatResponse{
			Content:   "Failed to create chat completion stream",
			RequestID: requestID,
			Type:     "error",
		}
		writeSSEMessage(w, flusher, chunk)
		return
	}
}

func writeSSEMessage(w http.ResponseWriter, flusher http.Flusher, chunk models.ChatResponse) error {
	chunkJSON, err := json.Marshal(chunk)
	if err != nil {
		return fmt.Errorf("error marshaling chunk: %v", err)
	}
	
	_, err = fmt.Fprintf(w, "data: %s\n\n", chunkJSON)
	if err != nil {
		return fmt.Errorf("error writing to response: %v", err)
	}
	
	flusher.Flush()
	return nil
} 