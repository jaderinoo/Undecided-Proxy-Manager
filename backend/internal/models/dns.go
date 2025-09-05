package models

import (
	"time"
)

// DNSProvider represents different DNS providers
type DNSProvider string

const (
	ProviderNamecheap DNSProvider = "namecheap"
)

// DNSConfig represents the configuration for a DNS provider
type DNSConfig struct {
	ID         int         `json:"id" db:"id"`
	Provider   DNSProvider `json:"provider" db:"provider"`
	Domain     string      `json:"domain" db:"domain"`
	Username   string      `json:"username" db:"username"`
	Password   string      `json:"password" db:"password"`
	IsActive   bool        `json:"is_active" db:"is_active"`
	LastUpdate *time.Time  `json:"last_update,omitempty" db:"last_update"`
	LastIP     string      `json:"last_ip,omitempty" db:"last_ip"`
	CreatedAt  time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at" db:"updated_at"`
}

// DNSRecord represents a DNS record that can be updated dynamically
type DNSRecord struct {
	ID         int        `json:"id" db:"id"`
	ConfigID   int        `json:"config_id" db:"config_id"`
	Host       string     `json:"host" db:"host"` // "@" for root domain, "www" for subdomain
	CurrentIP  string     `json:"current_ip" db:"current_ip"`
	LastUpdate *time.Time `json:"last_update,omitempty" db:"last_update"`
	IsActive   bool       `json:"is_active" db:"is_active"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
}

// DNSConfigCreateRequest represents the request to create a DNS configuration
type DNSConfigCreateRequest struct {
	Provider string `json:"provider" binding:"required"`
	Domain   string `json:"domain" binding:"required"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// DNSConfigUpdateRequest represents the request to update a DNS configuration
type DNSConfigUpdateRequest struct {
	Provider *string `json:"provider,omitempty"`
	Domain   *string `json:"domain,omitempty"`
	Username *string `json:"username,omitempty"`
	Password *string `json:"password,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
}

// DNSRecordCreateRequest represents the request to create a DNS record
type DNSRecordCreateRequest struct {
	ConfigID int    `json:"config_id" binding:"required"`
	Host     string `json:"host" binding:"required"`
}

// DNSRecordUpdateRequest represents the request to update a DNS record
type DNSRecordUpdateRequest struct {
	Host     *string `json:"host,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
}

// DNSUpdateResponse represents the response from a DNS update operation
type DNSUpdateResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	NewIP     string `json:"new_ip,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// DNSStatus represents the current status of DNS configuration
type DNSStatus struct {
	ConfigID    int         `json:"config_id"`
	Domain      string      `json:"domain"`
	Provider    string      `json:"provider"`
	IsActive    bool        `json:"is_active"`
	LastUpdate  *time.Time  `json:"last_update,omitempty"`
	LastIP      string      `json:"last_ip,omitempty"`
	RecordCount int         `json:"record_count"`
	Records     []DNSRecord `json:"records,omitempty"`
}
