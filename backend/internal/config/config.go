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
	// Development Configuration
	DevMode       bool   // Whether we're in development mode
	DevTestPassword string // Development test password for bypass
	// DNS Configuration
	DNSCheckInterval string // How often to check and update DNS (e.g., "5m", "1h")
	PublicIPService  string // Service to get public IP (e.g., "https://api.ipify.org")
	// Encryption Configuration
	EncryptionKey string // Key for encrypting sensitive data like DNS passwords
	// Let's Encrypt Configuration
	LetsEncryptEmail    string // Email for Let's Encrypt registration
	LetsEncryptWebroot  string // Webroot for HTTP-01 challenges
	LetsEncryptCertPath string // Path to store Let's Encrypt certificates
}

func Load() *Config {
	env := getEnv("GO_ENV", "development")
	devMode := env == "development"

	return &Config{
		DatabasePath:        getEnv("DB_PATH", "/data/upm.db"),
		Environment:         env,
		BackendPort:         getEnv("BACKEND_PORT", "6080"),
		AdminPassword:       getEnv("ADMIN_PASSWORD", ""),
		JWTSecret:           getEnv("JWT_SECRET", "upm-default-secret-change-in-production"),
		DevMode:             devMode,
		DevTestPassword:     getEnv("DEV_TEST_PASSWORD", "devtest"),
		DNSCheckInterval:    getEnv("DNS_CHECK_INTERVAL", "5m"),
		PublicIPService:     getEnv("PUBLIC_IP_SERVICE", "https://api.ipify.org"),
		EncryptionKey:       getEnv("ENCRYPTION_KEY", "upm-default-encryption-key-32byt"),
		LetsEncryptEmail:    getEnv("LETSENCRYPT_EMAIL", ""),
		LetsEncryptWebroot:  getEnv("LETSENCRYPT_WEBROOT", "/var/www/html"),
		LetsEncryptCertPath: getEnv("LETSENCRYPT_CERT_PATH", "/etc/letsencrypt"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
