package handlers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang-ai-stream/config"
	"golang-ai-stream/middleware"
	"golang-ai-stream/models"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/require"
)

type mockStream struct {
	ctx    context.Context
	closed bool
	err    error
}

func (m *mockStream) Recv() (*openai.ChatCompletionStreamResponse, error) {
	// Wait a bit to ensure we can catch the cancellation
	time.Sleep(50 * time.Millisecond)

	if m.closed {
		return nil, context.Canceled
	}

	if m.err != nil {
		return nil, m.err
	}

	select {
	case <-m.ctx.Done():
		m.closed = true
		return nil, m.ctx.Err()
	default:
		return &openai.ChatCompletionStreamResponse{
			Choices: []openai.ChatCompletionStreamChoice{
				{
					Delta: openai.ChatCompletionStreamChoiceDelta{
						Content: "test",
					},
				},
			},
		}, nil
	}
}

func (m *mockStream) Close() {
	m.closed = true
}

type mockClient struct {
	err    error
	stream *mockStream
}

func (m *mockClient) CreateChatCompletionStream(ctx context.Context, req openai.ChatCompletionRequest) (ChatCompletionStreamer, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.stream != nil {
		m.stream.ctx = ctx
		return m.stream, nil
	}
	return &mockStream{ctx: ctx}, nil
}

func collectResponses(t *testing.T, w *httptest.ResponseRecorder) []models.ChatResponse {
	var responses []models.ChatResponse
	scanner := bufio.NewScanner(w.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		var response models.ChatResponse
		err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &response)
		require.NoError(t, err)
		responses = append(responses, response)
	}
	return responses
}

func TestChatHandler_HandleChat(t *testing.T) {
	tests := []struct {
		name           string
		prompt        string
		setupClient   func(*mockClient)
		cancelContext bool
		wantCount     int
		wantTypes     []string
		wantContent   string
	}{
		{
			name:   "client_disconnect",
			prompt: "test",
			cancelContext: true,
			wantCount: 1,
			wantTypes: []string{"error"},
			wantContent: "Client disconnected",
		},
		{
			name:   "stream_error",
			prompt: "test",
			setupClient: func(m *mockClient) {
				m.stream = &mockStream{err: fmt.Errorf("stream error")}
			},
			wantCount: 1,
			wantTypes: []string{"error"},
			wantContent: "Failed to create chat completion stream",
		},
		{
			name:   "client_error",
			prompt: "test",
			setupClient: func(m *mockClient) {
				m.err = fmt.Errorf("client error")
			},
			wantCount: 1,
			wantTypes: []string{"error"},
			wantContent: "Failed to create chat completion stream",
		},
		{
			name:   "empty_prompt",
			prompt: "",
			wantCount: 1,
			wantTypes: []string{"error"},
			wantContent: "prompt cannot be empty",
		},
		{
			name:   "long_prompt",
			prompt: strings.Repeat("a", 101),
			wantCount: 1,
			wantTypes: []string{"error"},
			wantContent: "prompt exceeds maximum length of 100 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mockClient{}
			if tt.setupClient != nil {
				tt.setupClient(client)
			}

			handler := NewChatHandler(client, &config.Config{MaxPromptLength: 100})

			reqBody := models.ChatRequest{Prompt: tt.prompt}
			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewReader(body))
			req = req.WithContext(context.WithValue(context.Background(), middleware.RequestIDKey, uuid.New().String()))

			if tt.cancelContext {
				ctx, cancel := context.WithCancel(req.Context())
				req = req.WithContext(ctx)
				go func() {
					time.Sleep(20 * time.Millisecond)
					cancel()
				}()
			}

			w := httptest.NewRecorder()

			done := make(chan struct{})
			go func() {
				handler.HandleChat(w, req)
				close(done)
			}()

			<-done

			responses := collectResponses(t, w)
			require.Equal(t, tt.wantCount, len(responses), "response count mismatch")
			
			for i, wantType := range tt.wantTypes {
				require.Equal(t, wantType, responses[i].Type, "response type mismatch at index %d", i)
			}

			if tt.wantContent != "" {
				var found bool
				for _, resp := range responses {
					if resp.Content == tt.wantContent {
						found = true
						break
					}
				}
				require.True(t, found, "expected content %q not found in responses", tt.wantContent)
			}
		})
	}
} 