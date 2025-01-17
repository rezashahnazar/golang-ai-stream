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

## Testing

The project includes comprehensive test coverage for all major components:

### Running Tests

Run all tests:

```bash
go test ./...
```

Run tests with coverage:

```bash
go test -cover ./...
```

Generate coverage report:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test Structure

- **Handler Tests**: Test request validation, error handling, and response formatting
- **Middleware Tests**: Test logging, security headers, CORS, and rate limiting
- **Config Tests**: Test environment variable loading and default values
- **Integration Tests**: Test complete request flow and streaming functionality

### Test Coverage

Tests cover:

- Input validation
- Error handling
- Configuration loading
- Middleware functionality
- Rate limiting behavior
- Security headers
- CORS configuration
- Request/Response formatting

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
- `github.com/stretchr/testify` - Testing assertions and mocks

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Test Coverage

| Package      | Coverage  |
| ------------ | --------- |
| `config`     | 100.0%    |
| `errors`     | 100.0%    |
| `handlers`   | 71.6%     |
| `logger`     | 97.5%     |
| `middleware` | 91.9%     |
| `main`       | 0.0%      |
| **Total**    | **70.6%** |

### Detailed Coverage Report

#### Config Package (100.0%)

- `LoadConfig`: 100.0%
- `getEnvWithDefault`: 100.0%

#### Errors Package (100.0%)

- `NewAPIError`: 100.0%
- `WithType`: 100.0%
- `WithRequestID`: 100.0%
- `RespondWithError`: 100.0%

#### Handlers Package (71.6%)

- `NewChatHandler`: 100.0%
- `validateRequest`: 100.0%
- `HandleChat`: 68.3%
- `writeSSEMessage`: 75.0%

#### Logger Package (97.5%)

- `colorizeLevel`: 100.0%
- `formatTime`: 100.0%
- `formatRequestID`: 66.7%
- `formatMethod`: 100.0%
- `formatPath`: 100.0%
- `formatStatus`: 100.0%
- `formatDuration`: 100.0%
- `LogRequest`: 100.0%
- `LogError`: 100.0%
- `LogInfo`: 100.0%

#### Middleware Package (91.9%)

- `WriteHeader`: 100.0%
- `Write`: 100.0%
- `Logger`: 100.0%
- `SecurityHeaders`: 100.0%
- `CORS`: 100.0%
- `NewRateLimiter`: 100.0%
- `tryConsume`: 100.0%
- `min`: 100.0%
- `RateLimit`: 100.0%

Note: The main package has 0% coverage as it contains the server initialization code which is not covered by unit tests. This is typical for main packages as they are usually tested through integration tests.

### Areas for Improvement

1. **Handlers Package (71.6%)**

   - Increase coverage of `HandleChat` function (currently 68.3%)
   - Improve coverage of `writeSSEMessage` function (currently 75.0%)
   - Add more test cases for streaming functionality and error handling

2. **Logger Package (97.5%)**

   - Improve coverage of `formatRequestID` function (currently 66.7%)

3. **Middleware Package (91.9%)**

   - Add tests for `Hijack` and `Flush` methods
   - Consider if these methods need to be implemented or can be removed

4. **Main Package (0.0%)**
   - Add integration tests for server initialization
   - Test graceful shutdown scenarios
   - Test configuration loading and error handling

## Running Tests

To run tests with coverage:

```bash
go test ./... -coverprofile=coverage.out
```

To view detailed coverage report:

```bash
go tool cover -func=coverage.out
```

To view coverage in HTML format:

```bash
go tool cover -html=coverage.out
```
