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
	// DNS Configuration
	DNSCheckInterval string // How often to check and update DNS (e.g., "5m", "1h")
	PublicIPService  string // Service to get public IP (e.g., "https://api.ipify.org")
}

func Load() *Config {
	return &Config{
		DatabasePath:     getEnv("DB_PATH", "/data/upm.db"),
		Environment:      getEnv("GO_ENV", "development"),
		BackendPort:      getEnv("BACKEND_PORT", "6080"),
		AdminPassword:    getEnv("ADMIN_PASSWORD", ""),
		JWTSecret:        getEnv("JWT_SECRET", "upm-default-secret-change-in-production"),
		DNSCheckInterval: getEnv("DNS_CHECK_INTERVAL", "5m"),
		PublicIPService:  getEnv("PUBLIC_IP_SERVICE", "https://api.ipify.org"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
