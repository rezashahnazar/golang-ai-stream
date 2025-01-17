package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	APIKey           string
	BaseURL          string
	Port             string
	RateLimit        float64
	MaxPromptLength  int
	ReadTimeoutSecs  int
	WriteTimeoutSecs int
	IdleTimeoutSecs  int
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found: %v", err)
	}

	rateLimit, _ := strconv.ParseFloat(getEnvWithDefault("RATE_LIMIT", "10"), 64)
	maxPromptLen, _ := strconv.Atoi(getEnvWithDefault("MAX_PROMPT_LENGTH", "4000"))
	readTimeout, _ := strconv.Atoi(getEnvWithDefault("READ_TIMEOUT_SECS", "15"))
	writeTimeout, _ := strconv.Atoi(getEnvWithDefault("WRITE_TIMEOUT_SECS", "15"))
	idleTimeout, _ := strconv.Atoi(getEnvWithDefault("IDLE_TIMEOUT_SECS", "60"))

	return &Config{
		APIKey:           os.Getenv("OPENROUTER_API_KEY"),
		BaseURL:          "https://openrouter.ai/api/v1",
		Port:             getEnvWithDefault("PORT", ":8080"),
		RateLimit:        rateLimit,
		MaxPromptLength:  maxPromptLen,
		ReadTimeoutSecs:  readTimeout,
		WriteTimeoutSecs: writeTimeout,
		IdleTimeoutSecs:  idleTimeout,
	}, nil
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
} 