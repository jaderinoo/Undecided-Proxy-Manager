package models

import (
	"time"
)

// DefaultRateLimitRPS is the requests-per-second applied to a new proxy
// when rate limiting is enabled and no explicit rate is provided.
const DefaultRateLimitRPS = 15

type Proxy struct {
	ID               int       `json:"id" db:"id"`
	Name             string    `json:"name" db:"name"`
	Domain           string    `json:"domain" db:"domain"`
	TargetURL        string    `json:"target_url" db:"target_url"`
	SSLEnabled       bool      `json:"ssl_enabled" db:"ssl_enabled"`
	WSEnabled        bool      `json:"ws_enabled" db:"ws_enabled"`
	SSLPath          string    `json:"ssl_path,omitempty" db:"ssl_path"`
	RateLimitEnabled bool      `json:"rate_limit_enabled" db:"rate_limit_enabled"`
	RateLimitRPS     int       `json:"rate_limit_rps" db:"rate_limit_rps"`
	Status           string    `json:"status" db:"status"` // active, inactive, error
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

type ProxyCreateRequest struct {
	Name             string `json:"name" binding:"required"`
	Domain           string `json:"domain" binding:"required"`
	TargetURL        string `json:"target_url" binding:"required"`
	SSLEnabled       bool   `json:"ssl_enabled"`
	WSEnabled        *bool  `json:"ws_enabled,omitempty"`
	RateLimitEnabled *bool  `json:"rate_limit_enabled,omitempty"`
	RateLimitRPS     *int   `json:"rate_limit_rps,omitempty"`
}

type ProxyUpdateRequest struct {
	Name             *string `json:"name,omitempty"`
	Domain           *string `json:"domain,omitempty"`
	TargetURL        *string `json:"target_url,omitempty"`
	SSLEnabled       *bool   `json:"ssl_enabled,omitempty"`
	WSEnabled        *bool   `json:"ws_enabled,omitempty"`
	RateLimitEnabled *bool   `json:"rate_limit_enabled,omitempty"`
	RateLimitRPS     *int    `json:"rate_limit_rps,omitempty"`
}

type Certificate struct {
	ID        int       `json:"id" db:"id"`
	Domain    string    `json:"domain" db:"domain"`
	CertPath  string    `json:"cert_path" db:"cert_path"`
	KeyPath   string    `json:"key_path" db:"key_path"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	IsValid   bool      `json:"is_valid" db:"is_valid"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type CertificateCreateRequest struct {
	Domain    string    `json:"domain" binding:"required"`
	CertPath  string    `json:"cert_path" binding:"required"`
	KeyPath   string    `json:"key_path" binding:"required"`
	ExpiresAt time.Time `json:"expires_at" binding:"required"`
}

type CertificateUpdateRequest struct {
	Domain    *string    `json:"domain,omitempty"`
	CertPath  *string    `json:"cert_path,omitempty"`
	KeyPath   *string    `json:"key_path,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	IsValid   *bool      `json:"is_valid,omitempty"`
}

type LetsEncryptRequest struct {
	Domain string `json:"domain" binding:"required"`
}

type CertificateRenewResponse struct {
	Domain      string       `json:"domain"`
	Success     bool         `json:"success"`
	Message     string       `json:"message"`
	Certificate *Certificate `json:"certificate,omitempty"`
}
