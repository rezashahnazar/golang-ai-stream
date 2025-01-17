package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	// Save current env vars
	oldEnv := map[string]string{
		"OPENROUTER_API_KEY":    os.Getenv("OPENROUTER_API_KEY"),
		"PORT":                  os.Getenv("PORT"),
		"RATE_LIMIT":           os.Getenv("RATE_LIMIT"),
		"MAX_PROMPT_LENGTH":    os.Getenv("MAX_PROMPT_LENGTH"),
		"READ_TIMEOUT_SECS":    os.Getenv("READ_TIMEOUT_SECS"),
		"WRITE_TIMEOUT_SECS":   os.Getenv("WRITE_TIMEOUT_SECS"),
		"IDLE_TIMEOUT_SECS":    os.Getenv("IDLE_TIMEOUT_SECS"),
	}

	// Restore env vars after test
	defer func() {
		for k, v := range oldEnv {
			if v != "" {
				os.Setenv(k, v)
			} else {
				os.Unsetenv(k)
			}
		}
	}()

	tests := []struct {
		name     string
		envVars  map[string]string
		expected *Config
	}{
		{
			name: "default values",
			envVars: map[string]string{
				"OPENROUTER_API_KEY": "test-key",
			},
			expected: &Config{
				APIKey:           "test-key",
				BaseURL:          "https://openrouter.ai/api/v1",
				Port:             ":8080",
				RateLimit:        10,
				MaxPromptLength:  4000,
				ReadTimeoutSecs:  15,
				WriteTimeoutSecs: 15,
				IdleTimeoutSecs:  60,
			},
		},
		{
			name: "custom values",
			envVars: map[string]string{
				"OPENROUTER_API_KEY":    "custom-key",
				"PORT":                  ":3000",
				"RATE_LIMIT":           "20",
				"MAX_PROMPT_LENGTH":    "5000",
				"READ_TIMEOUT_SECS":    "30",
				"WRITE_TIMEOUT_SECS":   "30",
				"IDLE_TIMEOUT_SECS":    "120",
			},
			expected: &Config{
				APIKey:           "custom-key",
				BaseURL:          "https://openrouter.ai/api/v1",
				Port:             ":3000",
				RateLimit:        20,
				MaxPromptLength:  5000,
				ReadTimeoutSecs:  30,
				WriteTimeoutSecs: 30,
				IdleTimeoutSecs:  120,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			for k := range oldEnv {
				os.Unsetenv(k)
			}

			// Set test environment
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			cfg, err := LoadConfig()
			assert.NoError(t, err)
			assert.Equal(t, tt.expected.APIKey, cfg.APIKey)
			assert.Equal(t, tt.expected.BaseURL, cfg.BaseURL)
			assert.Equal(t, tt.expected.Port, cfg.Port)
			assert.Equal(t, tt.expected.RateLimit, cfg.RateLimit)
			assert.Equal(t, tt.expected.MaxPromptLength, cfg.MaxPromptLength)
			assert.Equal(t, tt.expected.ReadTimeoutSecs, cfg.ReadTimeoutSecs)
			assert.Equal(t, tt.expected.WriteTimeoutSecs, cfg.WriteTimeoutSecs)
			assert.Equal(t, tt.expected.IdleTimeoutSecs, cfg.IdleTimeoutSecs)
		})
	}
}

func TestGetEnvWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "existing env var",
			key:          "TEST_KEY",
			defaultValue: "default",
			envValue:     "custom",
			expected:     "custom",
		},
		{
			name:         "missing env var",
			key:          "MISSING_KEY",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := getEnvWithDefault(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
} 