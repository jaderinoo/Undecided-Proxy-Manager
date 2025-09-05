package models

// Settings represents the application settings
type Settings struct {
	// Core settings (from .env - read-only in UI)
	CoreSettings CoreSettings `json:"core_settings"`

	// UI-manageable settings (stored in database)
	UISettings UISettings `json:"ui_settings"`
}

// CoreSettings represents settings that come from .env and are read-only in UI
type CoreSettings struct {
	DatabasePath        string `json:"database_path"`
	Environment         string `json:"environment"`
	BackendPort         string `json:"backend_port"`
	AdminPassword       string `json:"admin_password"` // Masked for security
	JWTSecret           string `json:"jwt_secret"`     // Masked for security
	LetsEncryptEmail    string `json:"letsencrypt_email"`
	LetsEncryptWebroot  string `json:"letsencrypt_webroot"`
	LetsEncryptCertPath string `json:"letsencrypt_cert_path"`
	DNSCheckInterval    string `json:"dns_check_interval"`
	PublicIPService     string `json:"public_ip_service"`
}

// UISettings represents settings that can be managed through the UI
type UISettings struct {
	DisplayName     string `json:"display_name"`
	Theme           string `json:"theme"`
	Language        string `json:"language"`
	EnableDynamicDNS bool  `json:"enable_dynamic_dns"`
}

// SettingsUpdateRequest represents a request to update UI settings
type SettingsUpdateRequest struct {
	DisplayName     *string `json:"display_name,omitempty"`
	Theme           *string `json:"theme,omitempty"`
	Language        *string `json:"language,omitempty"`
	EnableDynamicDNS *bool  `json:"enable_dynamic_dns,omitempty"`
}

// MaskSensitiveData masks sensitive information in core settings
func (cs *CoreSettings) MaskSensitiveData() {
	if cs.AdminPassword != "" {
		cs.AdminPassword = "***masked***"
	}
	if cs.JWTSecret != "" {
		cs.JWTSecret = "***masked***"
	}
}
