package models

type ChatRequest struct {
    Prompt string `json:"prompt"`
}

type ChatResponse struct {
    Content   string `json:"content"`
    RequestID string `json:"request_id"`
    Type      string `json:"type"`
} 