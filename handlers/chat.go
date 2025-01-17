package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang-ai-stream/config"
	"golang-ai-stream/errors"
	"golang-ai-stream/logger"
	"golang-ai-stream/middleware"
	"golang-ai-stream/models"

	"github.com/sashabaranov/go-openai"
)

type ChatHandler struct {
	client *openai.Client
	config *config.Config
}

func NewChatHandler(client *openai.Client, cfg *config.Config) *ChatHandler {
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
	
	var reqBody models.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		logger.LogError(requestID, err, "Invalid request payload")
		errors.ErrBadRequest("Invalid request payload").
			WithRequestID(requestID).
			RespondWithError(w)
		return
	}

	if err := h.validateRequest(&reqBody); err != nil {
		logger.LogError(requestID, err, "Request validation failed")
		errors.ErrBadRequest(err.Error()).
			WithRequestID(requestID).
			RespondWithError(w)
		return
	}

	// Set headers before any potential error responses
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	flusher, ok := w.(http.Flusher)
	if !ok {
		logger.LogError(requestID, fmt.Errorf("streaming not supported"), "Streaming unsupported")
		errors.ErrInternalServer("Streaming unsupported by client").
			WithRequestID(requestID).
			RespondWithError(w)
		return
	}

	// Send initial SSE message to confirm connection
	chunk := models.ChatResponse{
		Content:   "",
		RequestID: requestID,
		Type:     "connected",
	}
	if err := writeSSEMessage(w, flusher, chunk); err != nil {
		logger.LogError(requestID, err, "Error writing initial SSE message")
		return
	}

	logger.LogRequest(logger.INFO, requestID, r.Method, r.URL.Path, http.StatusOK, 0, 
		fmt.Sprintf("Processing chat request with prompt length: %d", len(reqBody.Prompt)))

	chatReq := openai.ChatCompletionRequest{
		Model:    "anthropic/claude-3.5-sonnet",
		Messages: []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: reqBody.Prompt}},
		Stream:   true,
	}

	ctx := r.Context()
	stream, err := h.client.CreateChatCompletionStream(ctx, chatReq)
	if err != nil {
		logger.LogError(requestID, err, "Error creating chat completion stream")
		chunk := models.ChatResponse{
			Content:   "Failed to create chat completion stream",
			RequestID: requestID,
			Type:     "error",
		}
		writeSSEMessage(w, flusher, chunk)
		return
	}
	defer stream.Close()

	// Monitor for client disconnection
	go func() {
		<-ctx.Done()
		stream.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			logger.LogInfo(fmt.Sprintf("[%s] Client disconnected", requestID))
			return
		default:
			response, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					// Send completion message
					chunk := models.ChatResponse{
						Content:   "",
						RequestID: requestID,
						Type:     "done",
					}
					writeSSEMessage(w, flusher, chunk)
					return
				}
				
				logger.LogError(requestID, err, "Error receiving stream response")
				chunk := models.ChatResponse{
					Content:   "An error occurred while processing your request",
					RequestID: requestID,
					Type:     "error",
				}
				writeSSEMessage(w, flusher, chunk)
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
				logger.LogError(requestID, err, "Error writing SSE message")
				return
			}
		}
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