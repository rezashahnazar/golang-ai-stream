package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

type requestBody struct {
	Prompt string `json:"prompt"`
}

type responseChunk struct {
	Content string `json:"content"`
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found: %v", err)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/chat", chatHandler).Methods("POST")
	fmt.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	var reqBody requestBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		http.Error(w, "OpenRouter API key not set", http.StatusInternalServerError)
		return
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://openrouter.ai/api/v1"
	client := openai.NewClientWithConfig(config)

	chatReq := openai.ChatCompletionRequest{
		Model:    "anthropic/claude-3.5-sonnet",
		Messages: []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: reqBody.Prompt}},
		Stream:   true,
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	stream, err := client.CreateChatCompletionStream(context.Background(), chatReq)
	if err != nil {
		http.Error(w, "Error creating chat completion stream: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer stream.Close()

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	for {
		response, err := stream.Recv()
		if err != nil {
			break
		}

		content := response.Choices[0].Delta.Content
		if content == "" {
			continue
		}

		chunk := responseChunk{Content: content}
		chunkJSON, _ := json.Marshal(chunk)
		fmt.Fprintf(w, "data: %s\n\n", chunkJSON)
		flusher.Flush()
	}
}
