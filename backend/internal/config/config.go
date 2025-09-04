package config

import (
	"os"
)

type Config struct {
	DatabasePath string
	Environment  string
	BackendPort  string
}

func Load() *Config {
	return &Config{
		DatabasePath: getEnv("DB_PATH", "/data/upm.db"),
		Environment:  getEnv("GO_ENV", "development"),
		BackendPort:  getEnv("BACKEND_PORT", "6080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
