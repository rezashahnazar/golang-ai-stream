# Golang AI Stream Handler

A robust Go server that demonstrates real-time streaming of AI responses using OpenRouter's API with Claude 3.5 Sonnet. Features enterprise-grade logging, error handling, and security features.

## Features

### Core Functionality

- Real-time streaming of AI responses using Server-Sent Events (SSE)
- Support for OpenRouter API integration with Claude 3.5 Sonnet
- Graceful server shutdown and connection handling
- Client disconnection detection and cleanup

### Security

- Comprehensive security headers (CORS, XSS protection, etc.)
- Rate limiting with token bucket algorithm
- Request validation and sanitization
- Secure streaming implementation

### Monitoring & Debugging

- Colored and structured logging system
- Request ID tracking across the entire request lifecycle
- Detailed error reporting and handling
- Performance metrics (response times, status codes)
- Request/Response lifecycle logging

### API Features

- SSE message types for different events:
  - `connected`: Initial connection confirmation
  - `content`: Actual content chunks
  - `error`: Error messages
  - `done`: Stream completion
- Configurable rate limits and timeouts
- Customizable prompt length validation

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

3. Create a `.env` file in the root directory with the following configuration:

```env
# API Configuration
OPENROUTER_API_KEY=your_api_key_here

# Server Configuration
PORT=:8080
RATE_LIMIT=10
MAX_PROMPT_LENGTH=4000

# Timeout Configuration (in seconds)
READ_TIMEOUT_SECS=15
WRITE_TIMEOUT_SECS=15
IDLE_TIMEOUT_SECS=60
```

## Usage

1. Start the server:

```bash
go run main.go
```

2. The server will start on `http://localhost:8080` (or your configured port)

3. Send requests to the chat endpoint:

```bash
curl -N -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Your question here"}'
```

## API Endpoints

### POST /chat

Streams AI responses for a given prompt using Server-Sent Events.

**Request Headers:**

- `Content-Type: application/json`
- `Accept: text/event-stream` (optional)

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
  "content": "string",
  "request_id": "string",
  "type": "string" // "connected" | "content" | "error" | "done"
}
```

### GET /health

Health check endpoint that returns 200 OK when the server is running.

## Error Handling

The server includes comprehensive error handling for:

- Invalid request payloads
- Rate limit exceeded
- Stream creation failures
- Client disconnections
- Network issues
- API errors

All errors are logged with:

- Unique request ID
- Timestamp
- Error type
- Detailed message
- Stack trace (when applicable)

## Project Structure

- `main.go` - Server initialization and configuration
- `config/` - Configuration management
- `handlers/` - Request handlers and business logic
- `middleware/` - HTTP middleware (logging, security, etc.)
- `models/` - Data models and types
- `errors/` - Error handling and types
- `logger/` - Logging system
- `.env` - Environment variables
- `go.mod` - Go module dependencies

## Dependencies

- `github.com/gorilla/mux` - HTTP router and URL matcher
- `github.com/joho/godotenv` - Environment variable management
- `github.com/sashabaranov/go-openai` - OpenAI API client
- `github.com/google/uuid` - UUID generation

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
