package config

import (
	"os"
	"time"
)

type Config struct {
	Port            string
	Environment     string
	LogLevel        string
	ShutdownTimeout time.Duration
}

func Load() Config {
	return Config{
		Port:            getEnv("PORT", "8080"),
		Environment:     getEnv("APP_ENV", "local"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		ShutdownTimeout: getDuration("SHUTDOWN_TIMEOUT", 10*time.Second),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return duration
}
