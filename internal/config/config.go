package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// API Keys
	GeminiAPIKey    string
	DiscordWebhook string
	DiscordWebhookGlobal string

	// Server Configuration
	Port    string
	GinMode string

	// Timezone
	Timezone string

	// Logging
	LogLevel string
}

func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// It's okay if .env doesn't exist, we'll use environment variables
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("error loading .env file: %w", err)
		}
	}

	cfg := &Config{
		GeminiAPIKey:         getEnv("GEMINI_API_KEY", ""),
		DiscordWebhook:      getEnv("DISCORD_WEBHOOK", ""),
		DiscordWebhookGlobal: getEnv("DISCORD_WEBHOOK_GLOBAL", getEnv("DISCORD_WEBHOOK", "")), // Fallback to main webhook
		Port:                getEnv("PORT", "6005"),
		GinMode:             getEnv("GIN_MODE", "release"),
		Timezone:            getEnv("TZ", "Asia/Jakarta"),
		LogLevel:            getEnv("LOG_LEVEL", "info"),
	}

	// Validate required configuration
	if cfg.GeminiAPIKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is required")
	}
	if cfg.DiscordWebhook == "" {
		return nil, fmt.Errorf("DISCORD_WEBHOOK is required")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}