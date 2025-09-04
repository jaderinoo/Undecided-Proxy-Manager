package config

import (
	"os"
)

type Config struct {
	DatabasePath  string
	Environment   string
	BackendPort   string
	AdminPassword string
	JWTSecret     string
}

func Load() *Config {
	return &Config{
		DatabasePath:  getEnv("DB_PATH", "/data/upm.db"),
		Environment:   getEnv("GO_ENV", "development"),
		BackendPort:   getEnv("BACKEND_PORT", "6080"),
		AdminPassword: getEnv("ADMIN_PASSWORD", ""),
		JWTSecret:     getEnv("JWT_SECRET", "upm-default-secret-change-in-production"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
