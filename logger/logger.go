package logger

import (
	"fmt"
	"time"
)

const (
	// Colors
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
	White  = "\033[97m"

	// Text formatting
	Bold = "\033[1m"
)

type Level string

const (
	INFO    Level = "INFO"
	ERROR   Level = "ERROR"
	WARNING Level = "WARN"
	DEBUG   Level = "DEBUG"
)

func colorizeLevel(level Level) string {
	switch level {
	case INFO:
		return fmt.Sprintf("%s%s%s", Green, level, Reset)
	case ERROR:
		return fmt.Sprintf("%s%s%s", Red, level, Reset)
	case WARNING:
		return fmt.Sprintf("%s%s%s", Yellow, level, Reset)
	case DEBUG:
		return fmt.Sprintf("%s%s%s", Cyan, level, Reset)
	default:
		return string(level)
	}
}

func formatTime(t time.Time) string {
	return fmt.Sprintf("%s%s%s", Purple, t.Format("2006-01-02 15:04:05.000"), Reset)
}

func formatRequestID(requestID string) string {
	if requestID == "" {
		return ""
	}
	return fmt.Sprintf("%s[%s]%s", Yellow, requestID, Reset)
}

func formatMethod(method string) string {
	switch method {
	case "GET":
		return fmt.Sprintf("%s%s%s", Blue, method, Reset)
	case "POST":
		return fmt.Sprintf("%s%s%s", Green, method, Reset)
	case "PUT":
		return fmt.Sprintf("%s%s%s", Yellow, method, Reset)
	case "DELETE":
		return fmt.Sprintf("%s%s%s", Red, method, Reset)
	case "OPTIONS":
		return fmt.Sprintf("%s%s%s", Cyan, method, Reset)
	default:
		return method
	}
}

func formatPath(path string) string {
	return fmt.Sprintf("%s%s%s", White, path, Reset)
}

func formatStatus(status int) string {
	switch {
	case status >= 500:
		return fmt.Sprintf("%s%d%s", Red, status, Reset)
	case status >= 400:
		return fmt.Sprintf("%s%d%s", Yellow, status, Reset)
	case status >= 300:
		return fmt.Sprintf("%s%d%s", Cyan, status, Reset)
	case status >= 200:
		return fmt.Sprintf("%s%d%s", Green, status, Reset)
	default:
		return fmt.Sprintf("%d", status)
	}
}

func formatDuration(duration time.Duration) string {
	return fmt.Sprintf("%s%v%s", Cyan, duration.Round(time.Millisecond), Reset)
}

func LogRequest(level Level, requestID, method, path string, status int, duration time.Duration, msg string) {
	timestamp := formatTime(time.Now())
	levelStr := colorizeLevel(level)
	requestIDStr := formatRequestID(requestID)
	methodStr := formatMethod(method)
	pathStr := formatPath(path)
	statusStr := formatStatus(status)
	durationStr := formatDuration(duration)

	fmt.Printf("%s %s %s %s %s [%s] %v %s\n",
		timestamp,
		levelStr,
		requestIDStr,
		methodStr,
		pathStr,
		statusStr,
		durationStr,
		msg,
	)
}

func LogError(requestID string, err error, msg string) {
	timestamp := formatTime(time.Now())
	levelStr := colorizeLevel(ERROR)
	requestIDStr := formatRequestID(requestID)

	fmt.Printf("%s %s %s %s: %v\n",
		timestamp,
		levelStr,
		requestIDStr,
		msg,
		err,
	)
}

func LogInfo(msg string) {
	timestamp := formatTime(time.Now())
	levelStr := colorizeLevel(INFO)

	fmt.Printf("%s %s %s\n",
		timestamp,
		levelStr,
		msg,
	)
} 