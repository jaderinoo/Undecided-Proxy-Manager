package services

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"upm-backend/internal/config"
	"upm-backend/internal/models"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/http/webroot"
	"github.com/go-acme/lego/v4/registration"
)

// LetsEncryptService handles Let's Encrypt certificate operations
type LetsEncryptService struct {
	config     *config.Config
	certPath   string
	webroot    string
	httpClient *http.Client
}

// User represents a Let's Encrypt user
type User struct {
	Email        string
	Registration *registration.Resource
	key          *rsa.PrivateKey
}

// GetEmail returns the user's email
func (u *User) GetEmail() string {
	return u.Email
}

// GetRegistration returns the user's registration
func (u *User) GetRegistration() *registration.Resource {
	return u.Registration
}

// GetPrivateKey returns the user's private key
func (u *User) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

// NewLetsEncryptService creates a new Let's Encrypt service
func NewLetsEncryptService(certPath, webroot string) *LetsEncryptService {
	return &LetsEncryptService{
		config:   config.Load(),
		certPath: certPath,
		webroot:  webroot,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GenerateCertificate generates a Let's Encrypt certificate for the given domain
func (l *LetsEncryptService) GenerateCertificate(domain string) (*models.Certificate, error) {
	// Create user
	user, err := l.createOrGetUser()
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create ACME client
	client, err := l.createACMEClient(user)
	if err != nil {
		return nil, fmt.Errorf("failed to create ACME client: %w", err)
	}

	// Set up HTTP-01 challenge
	err = l.setupHTTP01Challenge(client)
	if err != nil {
		return nil, fmt.Errorf("failed to setup HTTP-01 challenge: %w", err)
	}

	// Request certificate
	request := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
	}

	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain certificate: %w", err)
	}

	// Save certificate files
	certPath, keyPath, err := l.saveCertificate(domain, certificates)
	if err != nil {
		return nil, fmt.Errorf("failed to save certificate: %w", err)
	}

	// Parse certificate to get expiration date
	expiresAt, err := l.getCertificateExpiration(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate expiration: %w", err)
	}

	return &models.Certificate{
		Domain:    domain,
		CertPath:  certPath,
		KeyPath:   keyPath,
		ExpiresAt: expiresAt,
		IsValid:   true,
	}, nil
}

// RenewCertificate renews an existing Let's Encrypt certificate
func (l *LetsEncryptService) RenewCertificate(cert *models.Certificate) (*models.Certificate, error) {
	// Check if certificate is close to expiration (less than 30 days)
	now := time.Now()
	daysUntilExpiry := int(cert.ExpiresAt.Sub(now).Hours() / 24)

	if daysUntilExpiry > 30 {
		return cert, fmt.Errorf("certificate is not close to expiration (%d days remaining)", daysUntilExpiry)
	}

	// Generate new certificate
	newCert, err := l.GenerateCertificate(cert.Domain)
	if err != nil {
		return nil, fmt.Errorf("failed to renew certificate: %w", err)
	}

	// Update the existing certificate with new paths and expiration
	cert.CertPath = newCert.CertPath
	cert.KeyPath = newCert.KeyPath
	cert.ExpiresAt = newCert.ExpiresAt
	cert.IsValid = true
	cert.UpdatedAt = time.Now()

	return cert, nil
}

// createOrGetUser creates or retrieves a Let's Encrypt user
func (l *LetsEncryptService) createOrGetUser() (*User, error) {
	userDir := filepath.Join(l.certPath, "accounts")
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create user directory: %w", err)
	}

	userFile := filepath.Join(userDir, "user.json")
	keyFile := filepath.Join(userDir, "user.key")

	// Try to load existing user
	if _, err := os.Stat(userFile); err == nil {
		user, err := l.loadUser(userFile, keyFile)
		if err == nil {
			return user, nil
		}
		log.Printf("Failed to load existing user, creating new one: %v", err)
	}

	// Create new user
	user, err := l.createNewUser(userFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create new user: %w", err)
	}

	return user, nil
}

// createNewUser creates a new Let's Encrypt user
func (l *LetsEncryptService) createNewUser(userFile, keyFile string) (*User, error) {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Save private key
	keyData := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	if err := os.WriteFile(keyFile, keyData, 0600); err != nil {
		return nil, fmt.Errorf("failed to save private key: %w", err)
	}

	// Create user
	user := &User{
		Email: l.config.LetsEncryptEmail,
		key:   privateKey,
	}

	// Register with Let's Encrypt
	client, err := l.createACMEClient(user)
	if err != nil {
		return nil, fmt.Errorf("failed to create ACME client: %w", err)
	}

	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return nil, fmt.Errorf("failed to register with Let's Encrypt: %w", err)
	}

	user.Registration = reg

	// Save user data
	if err := l.saveUser(user, userFile); err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	return user, nil
}

// loadUser loads an existing user from files
func (l *LetsEncryptService) loadUser(userFile, keyFile string) (*User, error) {
	// This is a simplified implementation
	// In production, you'd want to properly serialize/deserialize the user data
	return nil, fmt.Errorf("user loading not implemented yet")
}

// saveUser saves user data to file
func (l *LetsEncryptService) saveUser(user *User, userFile string) error {
	// This is a simplified implementation
	// In production, you'd want to properly serialize the user data
	return nil
}

// createACMEClient creates an ACME client
func (l *LetsEncryptService) createACMEClient(user *User) (*lego.Client, error) {
	config := lego.NewConfig(user)

	// Use Let's Encrypt staging for development, production for production
	if l.config.Environment == "production" {
		config.CADirURL = lego.LEDirectoryProduction
	} else {
		config.CADirURL = lego.LEDirectoryStaging
	}

	client, err := lego.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create ACME client: %w", err)
	}

	return client, nil
}

// setupHTTP01Challenge sets up HTTP-01 challenge
func (l *LetsEncryptService) setupHTTP01Challenge(client *lego.Client) error {
	// Use webroot provider for HTTP-01 challenge
	provider, err := webroot.NewHTTPProvider(l.webroot)
	if err != nil {
		return fmt.Errorf("failed to create webroot provider: %w", err)
	}

	err = client.Challenge.SetHTTP01Provider(provider)
	if err != nil {
		return fmt.Errorf("failed to set HTTP-01 provider: %w", err)
	}

	return nil
}

// saveCertificate saves certificate and key to files
func (l *LetsEncryptService) saveCertificate(domain string, certs *certificate.Resource) (string, string, error) {
	// Ensure certificate directory exists
	certDir := filepath.Join(l.certPath, "certs")
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create cert directory: %w", err)
	}

	certPath := filepath.Join(certDir, domain+".crt")
	keyPath := filepath.Join(certDir, domain+".key")

	// Save certificate
	if err := os.WriteFile(certPath, certs.Certificate, 0644); err != nil {
		return "", "", fmt.Errorf("failed to save certificate: %w", err)
	}

	// Save private key
	if err := os.WriteFile(keyPath, certs.PrivateKey, 0600); err != nil {
		return "", "", fmt.Errorf("failed to save private key: %w", err)
	}

	return certPath, keyPath, nil
}

// getCertificateExpiration extracts expiration date from certificate
func (l *LetsEncryptService) getCertificateExpiration(certPath string) (time.Time, error) {
	certData, err := os.ReadFile(certPath)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to read certificate: %w", err)
	}

	block, _ := pem.Decode(certData)
	if block == nil {
		return time.Time{}, fmt.Errorf("failed to decode PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert.NotAfter, nil
}

// ValidateDomain validates that a domain is accessible for ACME challenges
func (l *LetsEncryptService) ValidateDomain(domain string) error {
	// Check if domain resolves to this server
	// This is a simplified check - in production you'd want more robust validation
	resp, err := l.httpClient.Get("http://" + domain + "/.well-known/acme-challenge/test")
	if err != nil {
		return fmt.Errorf("domain %s is not accessible: %w", domain, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 404 {
		return fmt.Errorf("domain %s is not properly configured for ACME challenges", domain)
	}

	return nil
}

// GetCertificateInfo returns information about a certificate
func (l *LetsEncryptService) GetCertificateInfo(certPath string) (*CertificateInfo, error) {
	certData, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate: %w", err)
	}

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
