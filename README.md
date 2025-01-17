# Golang AI Stream Handler

A lightweight Go server that demonstrates real-time streaming of AI responses using OpenRouter's API with Claude 3.5 Sonnet.

## Features

- Real-time streaming of AI responses
- Simple REST API endpoint
- Support for OpenRouter API integration
- Server-Sent Events (SSE) for efficient streaming

## Prerequisites

- Go 1.23.5 or higher
- OpenRouter API key
- pnpm (for frontend development)

## Installation

1. Clone the repository:

```bash
git clone https://github.com/rezashahnazar/golang-ai-stream.git
cd golang-ai-stream
```

2. Install dependencies:

```bash
go mod download
```

3. Create a `.env` file in the root directory:

```env
OPENROUTER_API_KEY=your_api_key_here
```

## Usage

1. Start the server:

```bash
go run main.go
```

2. The server will start on `http://localhost:8080`

3. Send requests to the chat endpoint:

```bash
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Your question here"}'
```

## API Endpoints

### POST /chat

Streams AI responses for a given prompt.

**Request Body:**

```json
{
  "prompt": "string"
}
```

**Response:**
Server-sent events stream with JSON chunks:

```json
{
  "content": "string"
}
```

## Project Structure

- `main.go` - Main server implementation and chat handler
- `.env` - Environment variables configuration
- `go.mod` - Go module dependencies

## Dependencies

- `github.com/gorilla/mux` - HTTP router and URL matcher
- `github.com/joho/godotenv` - Environment variable management
- `github.com/sashabaranov/go-openai` - OpenAI API client

## Error Handling

The server includes error handling for:

- Missing API keys
- Invalid request payloads
- Stream creation failures
- Unsupported streaming capabilities
