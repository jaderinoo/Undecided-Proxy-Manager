package services

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"upm-backend/internal/config"
	"upm-backend/internal/models"
)

type CertificateService struct {
	CertPath    string
	LetsEncrypt *LetsEncryptService
	config      *config.Config
}

func NewCertificateService(certPath string) *CertificateService {
	cfg := config.Load()
	return &CertificateService{
		CertPath:    certPath,
		config:      cfg,
		LetsEncrypt: NewLetsEncryptService(cfg.LetsEncryptCertPath, cfg.LetsEncryptWebroot),
	}
}

// ValidateCertificate validates a certificate file
func (c *CertificateService) ValidateCertificate(certPath, keyPath string) (*models.Certificate, error) {
	// Read certificate file
	certData, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}

	// Read private key file
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	// Parse certificate
	cert, err := tls.X509KeyPair(certData, keyData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Parse X.509 certificate
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse X.509 certificate: %w", err)
	}

	// Extract domain from certificate
	domain := x509Cert.Subject.CommonName
	if domain == "" && len(x509Cert.DNSNames) > 0 {
		domain = x509Cert.DNSNames[0]
	}

	// Check if certificate is valid
	isValid := time.Now().Before(x509Cert.NotAfter)

	return &models.Certificate{
		Domain:    domain,
		CertPath:  certPath,
		KeyPath:   keyPath,
		ExpiresAt: x509Cert.NotAfter,
		IsValid:   isValid,
	}, nil
}

// GenerateLetsEncryptCertificate generates a certificate using Let's Encrypt
func (c *CertificateService) GenerateLetsEncryptCertificate(domain string) (*models.Certificate, error) {
	// Check if Let's Encrypt is configured
	if c.config.LetsEncryptEmail == "" {
		return nil, fmt.Errorf("Let's Encrypt email not configured. Set LETSENCRYPT_EMAIL environment variable")
	}

	// In development mode, try Let's Encrypt first, but fall back to placeholder if it fails
	if c.config.Environment == "development" {
		// Try Let's Encrypt first
		cert, err := c.LetsEncrypt.GenerateCertificate(domain)
		if err != nil {
			// If Let's Encrypt fails in development, create a placeholder certificate
			fmt.Printf("Let's Encrypt failed in development mode for %s: %v. Creating placeholder certificate.\n", domain, err)
			return c.createDevelopmentCertificate(domain)
		}
		return cert, nil
	}

	// In production, validate domain is accessible
	if err := c.LetsEncrypt.ValidateDomain(domain); err != nil {
		return nil, fmt.Errorf("domain validation failed: %w", err)
	}

	// Generate certificate using Let's Encrypt
	cert, err := c.LetsEncrypt.GenerateCertificate(domain)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Let's Encrypt certificate: %w", err)
	}

	// Copy certificate to nginx ssl directory
	if err := c.copyCertificateToNginx(domain, cert); err != nil {
		return nil, fmt.Errorf("failed to copy certificate to nginx: %w", err)
	}

	return cert, nil
}

// createDevelopmentCertificate creates a development certificate for testing
func (c *CertificateService) createDevelopmentCertificate(domain string) (*models.Certificate, error) {
	// Create certificate directory
	certDir := filepath.Join(c.CertPath, "certs")
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cert directory: %w", err)
	}

	certPath := filepath.Join(certDir, domain+".crt")
	keyPath := filepath.Join(certDir, domain+".key")

	// Create placeholder certificate files
	if err := c.createPlaceholderCertificate(domain, certPath, keyPath); err != nil {
		return nil, fmt.Errorf("failed to create placeholder certificate: %w", err)
	}

	// Return certificate model with development-specific data
	return &models.Certificate{
		Domain:    domain,
		CertPath:  certPath,
		KeyPath:   keyPath,
		ExpiresAt: time.Now().Add(365 * 24 * time.Hour), // 1 year from now
		IsValid:   true,
	}, nil
}

// createPlaceholderCertificate creates placeholder certificate files
// This is for development/testing purposes only
func (c *CertificateService) createPlaceholderCertificate(domain, certPath, keyPath string) error {
	// Ensure certificate directory exists
	if err := os.MkdirAll(filepath.Dir(certPath), 0755); err != nil {
		return fmt.Errorf("failed to create certificate directory: %w", err)
	}

	// Create placeholder certificate file
	certContent := fmt.Sprintf(`-----BEGIN CERTIFICATE-----
PLACEHOLDER CERTIFICATE FOR %s
This is a placeholder certificate for development purposes.
In production, this would be a real Let's Encrypt certificate.
-----END CERTIFICATE-----`, domain)

	if err := os.WriteFile(certPath, []byte(certContent), 0644); err != nil {
		return fmt.Errorf("failed to write certificate file: %w", err)
	}

	// Create placeholder private key file
	keyContent := fmt.Sprintf(`-----BEGIN PRIVATE KEY-----
PLACEHOLDER PRIVATE KEY FOR %s
This is a placeholder private key for development purposes.
In production, this would be a real private key.
-----END PRIVATE KEY-----`, domain)

	if err := os.WriteFile(keyPath, []byte(keyContent), 0600); err != nil {
		return fmt.Errorf("failed to write private key file: %w", err)
	}

	return nil
}

// RenewCertificate renews a certificate using Let's Encrypt
func (c *CertificateService) RenewCertificate(cert *models.Certificate) (*models.Certificate, error) {
	// Check if Let's Encrypt is configured
	if c.config.LetsEncryptEmail == "" {
		return nil, fmt.Errorf("Let's Encrypt email not configured. Set LETSENCRYPT_EMAIL environment variable")
	}

	// Renew certificate using Let's Encrypt
	renewedCert, err := c.LetsEncrypt.RenewCertificate(cert)
	if err != nil {
		return nil, fmt.Errorf("failed to renew Let's Encrypt certificate: %w", err)
	}

	return renewedCert, nil
}

// CheckCertificateExpiry checks if a certificate is expiring soon
func (c *CertificateService) CheckCertificateExpiry(cert *models.Certificate) (bool, int) {
	now := time.Now()
	daysUntilExpiry := int(cert.ExpiresAt.Sub(now).Hours() / 24)

	// Consider expiring if less than 30 days
	return daysUntilExpiry <= 30, daysUntilExpiry
}

// GetCertificateInfo extracts information from a certificate file
func (c *CertificateService) GetCertificateInfo(certPath string) (*CertificateInfo, error) {
	certData, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}

	// Parse certificate
	block, _ := pem.Decode(certData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return &CertificateInfo{
		Subject:   cert.Subject.CommonName,
		Issuer:    cert.Issuer.CommonName,
		NotBefore: cert.NotBefore,
		NotAfter:  cert.NotAfter,
		DNSNames:  cert.DNSNames,
		IsValid:   time.Now().After(cert.NotBefore) && time.Now().Before(cert.NotAfter),
	}, nil
}

// CertificateInfo contains information about a certificate
type CertificateInfo struct {
	Subject   string
	Issuer    string
	NotBefore time.Time
	NotAfter  time.Time
	DNSNames  []string
	IsValid   bool
}

// copyCertificateToNginx copies the certificate files to the nginx ssl directory
func (c *CertificateService) copyCertificateToNginx(domain string, cert *models.Certificate) error {
	// Create nginx ssl directory (this is mounted to the host nginx/ssl directory)
	nginxCertDir := c.CertPath
	if err := os.MkdirAll(nginxCertDir, 0755); err != nil {
		return fmt.Errorf("failed to create nginx cert directory: %w", err)
	}

	// Copy certificate file
	certDest := filepath.Join(nginxCertDir, domain+".crt")
	if err := copyFile(cert.CertPath, certDest); err != nil {
		return fmt.Errorf("failed to copy certificate file: %w", err)
	}

	// Copy private key file
	keyDest := filepath.Join(nginxCertDir, domain+".key")
	if err := copyFile(cert.KeyPath, keyDest); err != nil {
		return fmt.Errorf("failed to copy private key file: %w", err)
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Copy file permissions
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	return os.Chmod(dst, sourceInfo.Mode())
}
