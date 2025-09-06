package config

import (
	"log"
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

	// In production, require critical secrets to be set
	if !devMode {
		// ADMIN_PASSWORD is now optional - admin user will be created automatically if not set
		requireEnvVar("JWT_SECRET", "JWT secret is required in production")
		requireEnvVar("ENCRYPTION_KEY", "Encryption key is required in production")
	}

	return &Config{
		DatabasePath:        getEnv("DB_PATH", "/data/upm.db"),
		Environment:         env,
		BackendPort:         getEnv("BACKEND_PORT", "6080"),
		AdminPassword:       getEnv("ADMIN_PASSWORD", ""),
		JWTSecret:           getEnvWithDevDefault("JWT_SECRET", "upm-default-secret-change-in-production", devMode),
		DevMode:             devMode,
		DevTestPassword:     getEnv("DEV_TEST_PASSWORD", "devtest"),
		DNSCheckInterval:    getEnv("DNS_CHECK_INTERVAL", "5m"),
		PublicIPService:     getEnv("PUBLIC_IP_SERVICE", "https://api.ipify.org"),
		EncryptionKey:       getEnvWithDevDefault("ENCRYPTION_KEY", "upm-default-encryption-key-32byt", devMode),
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

// getEnvWithDevDefault returns the environment variable value, or a dev default if in dev mode
func getEnvWithDevDefault(key, devDefault string, devMode bool) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	if devMode {
		return devDefault
	}
	// This should never be reached due to requireEnvVar check above
	return ""
}

// requireEnvVar checks if an environment variable is set and exits if not
func requireEnvVar(key, message string) {
	if os.Getenv(key) == "" {
		log.Fatalf("FATAL: %s. Please set the %s environment variable.", message, key)
	}
}
