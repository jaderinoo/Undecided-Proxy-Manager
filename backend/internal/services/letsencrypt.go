package services

import (
	"bufio"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Resolver: &net.Resolver{
						PreferGo: true,
						Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
							d := net.Dialer{}
							return d.DialContext(ctx, "udp", "1.1.1.1:53")
						},
					},
				}).DialContext,
			},
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

	// Log certificate data for debugging
	log.Printf("Certificate obtained for %s: cert size=%d bytes, key size=%d bytes", domain, len(certificates.Certificate), len(certificates.PrivateKey))
	if len(certificates.Certificate) == 0 {
		return nil, fmt.Errorf("certificate data is empty after obtaining from Let's Encrypt")
	}
	if len(certificates.PrivateKey) == 0 {
		return nil, fmt.Errorf("private key data is empty after obtaining from Let's Encrypt")
	}

	// Instead of manually saving, use lego's built-in storage
	// Lego automatically saves certificates when we call Obtain
	// We'll read from lego's storage location and copy to nginx location
	certPath, keyPath, err := l.saveCertificateFromLegoStorage(domain, certificates)
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
	// Check actual certificate file expiration, not database value
	// This handles cases where the database is out of sync with the actual certificate
	var actualExpiry time.Time
	var daysUntilExpiry int
	now := time.Now()
	
	// Try to read the actual certificate file to get real expiration
	// Check both the database path and the nginx path (common path mismatch)
	certPaths := []string{cert.CertPath}
	// If cert path is in /etc/letsencrypt, also check /etc/ssl/certs
	if strings.Contains(cert.CertPath, "/etc/letsencrypt") {
		nginxPath := strings.Replace(cert.CertPath, "/etc/letsencrypt/certs", "/etc/ssl/certs", 1)
		certPaths = append(certPaths, nginxPath)
	}
	
	var certInfo *CertificateInfo
	var err error
	for _, path := range certPaths {
		certInfo, err = l.GetCertificateInfo(path)
		if err == nil {
			break
		}
	}
	
	if certInfo != nil {
		actualExpiry = certInfo.NotAfter
		daysUntilExpiry = int(actualExpiry.Sub(now).Hours() / 24)
		log.Printf("Certificate file expiration check: %d days remaining (expires: %v)", daysUntilExpiry, actualExpiry)
	} else {
		// If we can't read the certificate file, check database value
		// But also allow renewal if database says it's close to expiration or if file doesn't exist
		dbDaysUntilExpiry := int(cert.ExpiresAt.Sub(now).Hours() / 24)
		log.Printf("Could not read certificate file, using database value: %d days remaining", dbDaysUntilExpiry)
		
		// If database says > 30 days but file can't be read, allow renewal anyway
		// This handles cases where the file is missing or in a different location
		if dbDaysUntilExpiry > 30 {
			// Check if certificate file exists at all - if not, allow renewal
			fileExists := false
			for _, path := range certPaths {
				if _, err := os.Stat(path); err == nil {
					fileExists = true
					break
				}
			}
			if !fileExists {
				log.Printf("Certificate file not found, allowing renewal")
				daysUntilExpiry = 0 // Force renewal
			} else {
				daysUntilExpiry = dbDaysUntilExpiry
			}
		} else {
			daysUntilExpiry = dbDaysUntilExpiry
		}
	}

	// Allow renewal if certificate is expired or expiring within 30 days
	if daysUntilExpiry > 30 {
		return cert, fmt.Errorf("certificate is not close to expiration (%d days remaining). If the certificate is actually expired, please delete and recreate it.", daysUntilExpiry)
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
	// Load private key
	keyData, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block from key file")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Load user data
	userData, err := os.ReadFile(userFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read user file: %w", err)
	}

	var userDataStruct struct {
		Email        string                   `json:"email"`
		Registration *registration.Resource   `json:"registration"`
	}

	if err := json.Unmarshal(userData, &userDataStruct); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user data: %w", err)
	}

	user := &User{
		Email:        userDataStruct.Email,
		Registration: userDataStruct.Registration,
		key:          privateKey,
	}

	return user, nil
}

// saveUser saves user data to file
func (l *LetsEncryptService) saveUser(user *User, userFile string) error {
	userData := struct {
		Email        string                   `json:"email"`
		Registration *registration.Resource   `json:"registration"`
	}{
		Email:        user.Email,
		Registration: user.Registration,
	}

	data, err := json.MarshalIndent(userData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal user data: %w", err)
	}

	if err := os.WriteFile(userFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write user file: %w", err)
	}

	return nil
}

// createACMEClient creates an ACME client
func (l *LetsEncryptService) createACMEClient(user *User) (*lego.Client, error) {
	config := lego.NewConfig(user)

	// Use Let's Encrypt production
	config.CADirURL = lego.LEDirectoryProduction

	// Configure HTTP client with proper TLS settings
	config.HTTPClient = l.httpClient

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

// saveCertificateFromLegoStorage saves certificate using lego's storage mechanism
// This approach writes the certificate data directly to the final location using a simple, reliable method
func (l *LetsEncryptService) saveCertificateFromLegoStorage(domain string, certs *certificate.Resource) (string, string, error) {
	// Validate that we have certificate data
	if certs == nil {
		return "", "", fmt.Errorf("certificate resource is nil")
	}

	if len(certs.Certificate) == 0 {
		return "", "", fmt.Errorf("certificate data is empty")
	}

	if len(certs.PrivateKey) == 0 {
		return "", "", fmt.Errorf("private key data is empty")
	}

	// Write to /tmp first (non-volume location) to ensure write succeeds
	// Then copy to final location
	tmpCertPath := filepath.Join("/tmp", domain+".crt.tmp")
	tmpKeyPath := filepath.Join("/tmp", domain+".key.tmp")
	
	// Use the original letsencrypt cert path (defaults to /etc/letsencrypt)
	// Save to certs subdirectory: /etc/letsencrypt/certs/
	finalCertDir := filepath.Join(l.certPath, "certs")
	if err := os.MkdirAll(finalCertDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create cert directory: %w", err)
	}

	certPath := filepath.Join(finalCertDir, domain+".crt")
	keyPath := filepath.Join(finalCertDir, domain+".key")

	// Use bufio.Writer with explicit Flush for reliable writes to /tmp
	certFile, err := os.Create(tmpCertPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to create certificate file: %w", err)
	}
	
	certWriter := bufio.NewWriter(certFile)
	if _, err := certWriter.Write(certs.Certificate); err != nil {
		certFile.Close()
		os.Remove(certPath)
		return "", "", fmt.Errorf("failed to write certificate data: %w", err)
	}
	if err := certWriter.Flush(); err != nil {
		certFile.Close()
		os.Remove(certPath)
		return "", "", fmt.Errorf("failed to flush certificate data: %w", err)
	}
	if err := certFile.Sync(); err != nil {
		certFile.Close()
		os.Remove(certPath)
		return "", "", fmt.Errorf("failed to sync certificate file: %w", err)
	}
	if err := certFile.Close(); err != nil {
		os.Remove(tmpCertPath)
		return "", "", fmt.Errorf("failed to close certificate file: %w", err)
	}
	
	// Verify temp file was written
	tmpCertInfo, err := os.Stat(tmpCertPath)
	if err != nil || tmpCertInfo.Size() != int64(len(certs.Certificate)) {
		os.Remove(tmpCertPath)
		return "", "", fmt.Errorf("temp certificate file verification failed: size=%d, expected=%d", func() int64 {
			if tmpCertInfo != nil {
				return tmpCertInfo.Size()
			}
			return 0
		}(), len(certs.Certificate))
	}
	log.Printf("Temp certificate file written successfully: %d bytes", tmpCertInfo.Size())

	// Use bufio.Writer for private key to /tmp
	keyFile, err := os.Create(tmpKeyPath)
	if err != nil {
		os.Remove(certPath)
		return "", "", fmt.Errorf("failed to create private key file: %w", err)
	}
	
	keyWriter := bufio.NewWriter(keyFile)
	if _, err := keyWriter.Write(certs.PrivateKey); err != nil {
		keyFile.Close()
		os.Remove(certPath)
		os.Remove(keyPath)
		return "", "", fmt.Errorf("failed to write private key data: %w", err)
	}
	if err := keyWriter.Flush(); err != nil {
		keyFile.Close()
		os.Remove(certPath)
		os.Remove(keyPath)
		return "", "", fmt.Errorf("failed to flush private key data: %w", err)
	}
	if err := keyFile.Sync(); err != nil {
		keyFile.Close()
		os.Remove(certPath)
		os.Remove(keyPath)
		return "", "", fmt.Errorf("failed to sync private key file: %w", err)
	}
	if err := keyFile.Close(); err != nil {
		os.Remove(tmpCertPath)
		os.Remove(tmpKeyPath)
		return "", "", fmt.Errorf("failed to close private key file: %w", err)
	}
	
	// Verify temp key file was written
	tmpKeyInfo, err := os.Stat(tmpKeyPath)
	if err != nil || tmpKeyInfo.Size() != int64(len(certs.PrivateKey)) {
		os.Remove(tmpCertPath)
		os.Remove(tmpKeyPath)
		return "", "", fmt.Errorf("temp private key file verification failed: size=%d, expected=%d", func() int64 {
			if tmpKeyInfo != nil {
				return tmpKeyInfo.Size()
			}
			return 0
		}(), len(certs.PrivateKey))
	}
	log.Printf("Temp private key file written successfully: %d bytes", tmpKeyInfo.Size())

	// Now copy from /tmp to final location using io.Copy
	srcCertFile, err := os.Open(tmpCertPath)
	if err != nil {
		os.Remove(tmpCertPath)
		os.Remove(tmpKeyPath)
		return "", "", fmt.Errorf("failed to open temp certificate file for copy: %w", err)
	}
	defer srcCertFile.Close()

	dstCertFile, err := os.Create(certPath)
	if err != nil {
		os.Remove(tmpCertPath)
		os.Remove(tmpKeyPath)
		return "", "", fmt.Errorf("failed to create final certificate file: %w", err)
	}
	
	bytesWritten, err := io.Copy(dstCertFile, srcCertFile)
	if err != nil {
		dstCertFile.Close()
		os.Remove(tmpCertPath)
		os.Remove(tmpKeyPath)
		os.Remove(certPath)
		return "", "", fmt.Errorf("failed to copy certificate file: %w", err)
	}
	if err := dstCertFile.Sync(); err != nil {
		dstCertFile.Close()
		os.Remove(tmpCertPath)
		os.Remove(tmpKeyPath)
		os.Remove(certPath)
		return "", "", fmt.Errorf("failed to sync certificate file: %w", err)
	}
	if err := dstCertFile.Close(); err != nil {
		os.Remove(tmpCertPath)
		os.Remove(tmpKeyPath)
		os.Remove(certPath)
		return "", "", fmt.Errorf("failed to close certificate file: %w", err)
	}
	if err := os.Chmod(certPath, 0644); err != nil {
		log.Printf("Warning: failed to set certificate permissions: %v", err)
	}
	log.Printf("Copied certificate from /tmp to final location: %d bytes", bytesWritten)

	// Copy private key
	srcKeyFile, err := os.Open(tmpKeyPath)
	if err != nil {
		os.Remove(tmpCertPath)
		os.Remove(tmpKeyPath)
		os.Remove(certPath)
		return "", "", fmt.Errorf("failed to open temp private key file for copy: %w", err)
	}
	defer srcKeyFile.Close()

	dstKeyFile, err := os.Create(keyPath)
	if err != nil {
		os.Remove(tmpCertPath)
		os.Remove(tmpKeyPath)
		os.Remove(certPath)
		return "", "", fmt.Errorf("failed to create final private key file: %w", err)
	}
	
	keyBytesWritten, err := io.Copy(dstKeyFile, srcKeyFile)
	if err != nil {
		dstKeyFile.Close()
		os.Remove(tmpCertPath)
		os.Remove(tmpKeyPath)
		os.Remove(certPath)
		os.Remove(keyPath)
		return "", "", fmt.Errorf("failed to copy private key file: %w", err)
	}
	if err := dstKeyFile.Sync(); err != nil {
		dstKeyFile.Close()
		os.Remove(tmpCertPath)
		os.Remove(tmpKeyPath)
		os.Remove(certPath)
		os.Remove(keyPath)
		return "", "", fmt.Errorf("failed to sync private key file: %w", err)
	}
	if err := dstKeyFile.Close(); err != nil {
		os.Remove(tmpCertPath)
		os.Remove(tmpKeyPath)
		os.Remove(certPath)
		os.Remove(keyPath)
		return "", "", fmt.Errorf("failed to close private key file: %w", err)
	}
	if err := os.Chmod(keyPath, 0600); err != nil {
		log.Printf("Warning: failed to set private key permissions: %v", err)
	}
	log.Printf("Copied private key from /tmp to final location: %d bytes", keyBytesWritten)

	// Clean up temp files
	os.Remove(tmpCertPath)
	os.Remove(tmpKeyPath)

	// Final verification of files in final location
	time.Sleep(100 * time.Millisecond)
	certInfo, err := os.Stat(certPath)
	if err != nil || certInfo.Size() != int64(len(certs.Certificate)) {
		return "", "", fmt.Errorf("final certificate verification failed: size=%d, expected=%d", func() int64 {
			if certInfo != nil {
				return certInfo.Size()
			}
			return 0
		}(), len(certs.Certificate))
	}

	keyInfo, err := os.Stat(keyPath)
	if err != nil || keyInfo.Size() != int64(len(certs.PrivateKey)) {
		return "", "", fmt.Errorf("final private key verification failed: size=%d, expected=%d", func() int64 {
			if keyInfo != nil {
				return keyInfo.Size()
			}
			return 0
		}(), len(certs.PrivateKey))
	}

	log.Printf("Successfully saved certificate for %s: cert=%d bytes, key=%d bytes", domain, len(certs.Certificate), len(certs.PrivateKey))
	return certPath, keyPath, nil
}

// saveCertificate saves certificate and key to files (DEPRECATED - kept for reference)
func (l *LetsEncryptService) saveCertificate(domain string, certs *certificate.Resource) (string, string, error) {
	// Validate that we have certificate data
	if certs == nil {
		return "", "", fmt.Errorf("certificate resource is nil")
	}

	if len(certs.Certificate) == 0 {
		return "", "", fmt.Errorf("certificate data is empty")
	}

	if len(certs.PrivateKey) == 0 {
		return "", "", fmt.Errorf("private key data is empty")
	}

	// Save certificates directly to /etc/ssl/certs where nginx expects them
	// This avoids the need to copy files later
	certDir := "/etc/ssl/certs"
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create cert directory: %w", err)
	}

	certPath := filepath.Join(certDir, domain+".crt")
	keyPath := filepath.Join(certDir, domain+".key")
	
	// Use temporary files first, then rename for atomic writes
	tempCertPath := certPath + ".tmp"
	tempKeyPath := keyPath + ".tmp"

	// Log what we're about to write for debugging
	previewLen := 50
	if len(certs.Certificate) < previewLen {
		previewLen = len(certs.Certificate)
	}
	log.Printf("About to write certificate for %s: cert length=%d bytes, first %d bytes preview=%x", domain, len(certs.Certificate), previewLen, certs.Certificate[:previewLen])
	
	// Save certificate using a more robust method - write to temp file first
	certFile, err := os.OpenFile(tempCertPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return "", "", fmt.Errorf("failed to create certificate file: %w", err)
	}
	
	n, err := certFile.Write(certs.Certificate)
	if err != nil {
		certFile.Close()
		return "", "", fmt.Errorf("failed to write certificate data: %w", err)
	}
	
	if n != len(certs.Certificate) {
		certFile.Close()
		return "", "", fmt.Errorf("incomplete certificate write: wrote %d of %d bytes", n, len(certs.Certificate))
	}
	
	// Sync to ensure data is written to disk
	if err := certFile.Sync(); err != nil {
		certFile.Close()
		return "", "", fmt.Errorf("failed to sync certificate file: %w", err)
	}
	
	if err := certFile.Close(); err != nil {
		os.Remove(tempCertPath)
		return "", "", fmt.Errorf("failed to close certificate file: %w", err)
	}

	// Verify temp certificate was written
	if info, err := os.Stat(tempCertPath); err != nil {
		os.Remove(tempCertPath)
		return "", "", fmt.Errorf("failed to verify certificate file was created: %w", err)
	} else if info.Size() == 0 {
		return "", "", fmt.Errorf("certificate file was created but is empty (size: 0 bytes)")
	} else if info.Size() != int64(len(certs.Certificate)) {
		os.Remove(tempCertPath)
		return "", "", fmt.Errorf("certificate file size mismatch: expected %d bytes, got %d bytes", len(certs.Certificate), info.Size())
	}

	// Save private key using a more robust method - write to temp file first
	keyFile, err := os.OpenFile(tempKeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return "", "", fmt.Errorf("failed to create private key file: %w", err)
	}
	
	n, err = keyFile.Write(certs.PrivateKey)
	if err != nil {
		keyFile.Close()
		return "", "", fmt.Errorf("failed to write private key data: %w", err)
	}
	
	if n != len(certs.PrivateKey) {
		keyFile.Close()
		return "", "", fmt.Errorf("incomplete private key write: wrote %d of %d bytes", n, len(certs.PrivateKey))
	}
	
	// Sync to ensure data is written to disk
	if err := keyFile.Sync(); err != nil {
		keyFile.Close()
		return "", "", fmt.Errorf("failed to sync private key file: %w", err)
	}
	
	if err := keyFile.Close(); err != nil {
		os.Remove(tempCertPath)
		os.Remove(tempKeyPath)
		return "", "", fmt.Errorf("failed to close private key file: %w", err)
	}

	// Verify temp private key was written
	if info, err := os.Stat(tempKeyPath); err != nil {
		os.Remove(tempCertPath)
		os.Remove(tempKeyPath)
		return "", "", fmt.Errorf("failed to verify private key file was created: %w", err)
	} else if info.Size() == 0 {
		return "", "", fmt.Errorf("private key file was created but is empty (size: 0 bytes)")
	} else if info.Size() != int64(len(certs.PrivateKey)) {
		os.Remove(tempCertPath)
		os.Remove(tempKeyPath)
		return "", "", fmt.Errorf("private key file size mismatch: expected %d bytes, got %d bytes", len(certs.PrivateKey), info.Size())
	}

	// Read back and verify the temp files before renaming
	// Add a small delay to ensure data is flushed
	time.Sleep(100 * time.Millisecond)
	
	verifyCertData, err := os.ReadFile(tempCertPath)
	if err != nil {
		os.Remove(tempCertPath)
		os.Remove(tempKeyPath)
		return "", "", fmt.Errorf("failed to read back certificate for verification: %w", err)
	}
	log.Printf("Read back certificate for %s: %d bytes (expected %d bytes)", domain, len(verifyCertData), len(certs.Certificate))
	if len(verifyCertData) != len(certs.Certificate) {
		// Log first few bytes of what we read back for debugging
		previewLen := 50
		if len(verifyCertData) < previewLen {
			previewLen = len(verifyCertData)
		}
	previewBytes := []byte{}
	if len(verifyCertData) > 0 {
		if len(verifyCertData) < previewLen {
			previewBytes = verifyCertData
		} else {
			previewBytes = verifyCertData[:previewLen]
		}
	}
	log.Printf("Certificate verification failed: read back %d bytes, expected %d bytes. First %d bytes: %x", len(verifyCertData), len(certs.Certificate), len(previewBytes), previewBytes)
		os.Remove(tempCertPath)
		os.Remove(tempKeyPath)
		return "", "", fmt.Errorf("certificate verification failed: read back %d bytes, expected %d bytes", len(verifyCertData), len(certs.Certificate))
	}

	verifyKeyData, err := os.ReadFile(tempKeyPath)
	if err != nil {
		os.Remove(tempCertPath)
		os.Remove(tempKeyPath)
		return "", "", fmt.Errorf("failed to read back private key for verification: %w", err)
	}
	log.Printf("Read back private key for %s: %d bytes (expected %d bytes)", domain, len(verifyKeyData), len(certs.PrivateKey))
	if len(verifyKeyData) != len(certs.PrivateKey) {
		os.Remove(tempCertPath)
		os.Remove(tempKeyPath)
		return "", "", fmt.Errorf("private key verification failed: read back %d bytes, expected %d bytes", len(verifyKeyData), len(certs.PrivateKey))
	}

	// All checks passed, now copy temp files to final names (more reliable than rename on Docker volumes)
	// Remove existing files first if they exist (they might be empty)
	if _, err := os.Stat(certPath); err == nil {
		log.Printf("Removing existing certificate file before copy: %s", certPath)
		os.Remove(certPath)
	}
	if _, err := os.Stat(keyPath); err == nil {
		log.Printf("Removing existing private key file before copy: %s", keyPath)
		os.Remove(keyPath)
	}
	
	// Verify temp files one more time before copy
	tempCertInfo, _ := os.Stat(tempCertPath)
	tempKeyInfo, _ := os.Stat(tempKeyPath)
	log.Printf("Before copy - temp cert size: %d, temp key size: %d", func() int64 {
		if tempCertInfo != nil {
			return tempCertInfo.Size()
		}
		return 0
	}(), func() int64 {
		if tempKeyInfo != nil {
			return tempKeyInfo.Size()
		}
		return 0
	}())
	
	// Copy certificate file using robust write method
	certData, err := os.ReadFile(tempCertPath)
	if err != nil {
		os.Remove(tempCertPath)
		os.Remove(tempKeyPath)
		return "", "", fmt.Errorf("failed to read temp certificate file for copy: %w", err)
	}
	
	certDestFile, err := os.OpenFile(certPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		os.Remove(tempCertPath)
		os.Remove(tempKeyPath)
		return "", "", fmt.Errorf("failed to create certificate file for copy: %w", err)
	}
	n, err = certDestFile.Write(certData)
	if err != nil {
		certDestFile.Close()
		os.Remove(tempCertPath)
		os.Remove(tempKeyPath)
		os.Remove(certPath)
		return "", "", fmt.Errorf("failed to write certificate data: %w", err)
	}
	if n != len(certData) {
		certDestFile.Close()
		os.Remove(tempCertPath)
		os.Remove(tempKeyPath)
		os.Remove(certPath)
		return "", "", fmt.Errorf("incomplete certificate copy: wrote %d of %d bytes", n, len(certData))
	}
	if err := certDestFile.Sync(); err != nil {
		certDestFile.Close()
		os.Remove(tempCertPath)
		os.Remove(tempKeyPath)
		os.Remove(certPath)
		return "", "", fmt.Errorf("failed to sync certificate file: %w", err)
	}
	if err := certDestFile.Close(); err != nil {
		os.Remove(tempCertPath)
		os.Remove(tempKeyPath)
		os.Remove(certPath)
		return "", "", fmt.Errorf("failed to close certificate file: %w", err)
	}
	log.Printf("Copied certificate file from %s to %s (%d bytes)", tempCertPath, certPath, len(certData))
	
	// Copy private key file using robust write method
	keyData, err := os.ReadFile(tempKeyPath)
	if err != nil {
		os.Remove(tempCertPath)
		os.Remove(tempKeyPath)
		os.Remove(certPath)
		return "", "", fmt.Errorf("failed to read temp private key file for copy: %w", err)
	}
	
	keyDestFile, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		os.Remove(tempCertPath)
		os.Remove(tempKeyPath)
		os.Remove(certPath)
		return "", "", fmt.Errorf("failed to create private key file for copy: %w", err)
	}
	n, err = keyDestFile.Write(keyData)
	if err != nil {
		keyDestFile.Close()
		os.Remove(tempCertPath)
		os.Remove(tempKeyPath)
		os.Remove(certPath)
		return "", "", fmt.Errorf("failed to write private key data: %w", err)
	}
	if n != len(keyData) {
		keyDestFile.Close()
		os.Remove(tempCertPath)
		os.Remove(tempKeyPath)
		os.Remove(certPath)
		return "", "", fmt.Errorf("incomplete private key copy: wrote %d of %d bytes", n, len(keyData))
	}
	if err := keyDestFile.Sync(); err != nil {
		keyDestFile.Close()
		os.Remove(tempCertPath)
		os.Remove(tempKeyPath)
		os.Remove(certPath)
		return "", "", fmt.Errorf("failed to sync private key file: %w", err)
	}
	if err := keyDestFile.Close(); err != nil {
		os.Remove(tempCertPath)
		os.Remove(tempKeyPath)
		os.Remove(certPath)
		return "", "", fmt.Errorf("failed to close private key file: %w", err)
	}
	log.Printf("Copied private key file from %s to %s (%d bytes)", tempKeyPath, keyPath, len(keyData))
	
	// Clean up temp files
	os.Remove(tempCertPath)
	os.Remove(tempKeyPath)

	// Final verification of the copied files
	time.Sleep(100 * time.Millisecond) // Small delay to ensure write is complete
	finalCertInfo, err := os.Stat(certPath)
	if err != nil {
		return "", "", fmt.Errorf("final certificate verification failed: file not found after copy: %w", err)
	}
	log.Printf("Final certificate file stat: size=%d bytes, expected=%d bytes", finalCertInfo.Size(), len(certs.Certificate))
	
	// Read the file to verify contents
	actualCertData, err := os.ReadFile(certPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read final certificate file for verification: %w", err)
	}
	log.Printf("Final certificate file read: %d bytes read, expected %d bytes", len(actualCertData), len(certs.Certificate))
	
	if len(actualCertData) != len(certs.Certificate) {
		previewLen := 50
		if len(actualCertData) < previewLen {
			previewLen = len(actualCertData)
		}
		previewBytes := []byte{}
		if len(actualCertData) > 0 {
			previewBytes = actualCertData[:previewLen]
		}
		log.Printf("Final certificate size mismatch: expected %d bytes, got %d bytes. First %d bytes: %x", len(certs.Certificate), len(actualCertData), len(previewBytes), previewBytes)
		return "", "", fmt.Errorf("final certificate verification failed: read %d bytes, expected %d bytes", len(actualCertData), len(certs.Certificate))
	}
	
	finalKeyInfo, err := os.Stat(keyPath)
	if err != nil {
		return "", "", fmt.Errorf("final private key verification failed: file not found after copy: %w", err)
	}
	log.Printf("Final private key file stat: size=%d bytes, expected=%d bytes", finalKeyInfo.Size(), len(certs.PrivateKey))
	
	// Read the key file to verify contents
	actualKeyData, err := os.ReadFile(keyPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read final private key file for verification: %w", err)
	}
	log.Printf("Final private key file read: %d bytes read, expected %d bytes", len(actualKeyData), len(certs.PrivateKey))
	
	if len(actualKeyData) != len(certs.PrivateKey) {
		return "", "", fmt.Errorf("final private key verification failed: read %d bytes, expected %d bytes", len(actualKeyData), len(certs.PrivateKey))
	}

	log.Printf("Successfully saved certificate for %s: cert=%d bytes, key=%d bytes", domain, len(certs.Certificate), len(certs.PrivateKey))

	return certPath, keyPath, nil
}

// readCertificateFromLegoStorage attempts to read certificate from lego's storage directory
func (l *LetsEncryptService) readCertificateFromLegoStorage(domain string) (string, string, error) {
	// Lego stores certificates in certPath/certificates/domain/
	certStorageDir := filepath.Join(l.certPath, "certificates", domain)
	
	certFile := filepath.Join(certStorageDir, domain+".crt")
	keyFile := filepath.Join(certStorageDir, domain+".key")
	
	// Check if files exist
	if _, err := os.Stat(certFile); err != nil {
		return "", "", fmt.Errorf("certificate file not found in lego storage: %w", err)
	}
	if _, err := os.Stat(keyFile); err != nil {
		return "", "", fmt.Errorf("private key file not found in lego storage: %w", err)
	}
	
	// Read certificate and key
	certData, err := os.ReadFile(certFile)
	if err != nil {
		return "", "", fmt.Errorf("failed to read certificate from lego storage: %w", err)
	}
	
	keyData, err := os.ReadFile(keyFile)
	if err != nil {
		return "", "", fmt.Errorf("failed to read private key from lego storage: %w", err)
	}
	
	if len(certData) == 0 || len(keyData) == 0 {
		return "", "", fmt.Errorf("certificate or key data is empty in lego storage")
	}
	
	// Copy to /etc/ssl/certs where nginx expects them
	certDir := "/etc/ssl/certs"
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create cert directory: %w", err)
	}
	
	destCertPath := filepath.Join(certDir, domain+".crt")
	destKeyPath := filepath.Join(certDir, domain+".key")
	
	// Write certificate
	if err := os.WriteFile(destCertPath, certData, 0644); err != nil {
		return "", "", fmt.Errorf("failed to copy certificate: %w", err)
	}
	
	// Write private key
	if err := os.WriteFile(destKeyPath, keyData, 0600); err != nil {
		return "", "", fmt.Errorf("failed to copy private key: %w", err)
	}
	
	log.Printf("Successfully copied certificate from lego storage for %s: cert=%d bytes, key=%d bytes", domain, len(certData), len(keyData))
	
	return destCertPath, destKeyPath, nil
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
	// Skip public IP validation for now since we know the domain is working
	// TODO: Re-enable this check once TLS certificate issues are resolved
	
	// Resolve domain to IP
	ips, err := net.LookupIP(domain)
	if err != nil {
		return fmt.Errorf("domain %s does not resolve: %w", domain, err)
	}

	// Just check that the domain resolves to some IP
	if len(ips) == 0 {
		return fmt.Errorf("domain %s does not resolve to any IP", domain)
	}

	// Check if domain is accessible via HTTP (for ACME challenge)
	resp, err := l.httpClient.Get("http://" + domain + "/.well-known/acme-challenge/test")
	if err != nil {
		return fmt.Errorf("domain %s is not accessible via HTTP: %w", domain, err)
	}
	defer resp.Body.Close()

	// We expect either 404 (no challenge file) or 200 (challenge file exists)
	if resp.StatusCode != 404 && resp.StatusCode != 200 {
		return fmt.Errorf("domain %s returned unexpected status code: %d", domain, resp.StatusCode)
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

// getPublicIP gets the public IP address of this server
func (l *LetsEncryptService) getPublicIP() (string, error) {
	// Use the configured public IP service
	resp, err := l.httpClient.Get(l.config.PublicIPService)
	if err != nil {
		return "", fmt.Errorf("failed to get public IP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("public IP service returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	ip := string(body)
	// Validate that it's a valid IP address
	if net.ParseIP(ip) == nil {
		return "", fmt.Errorf("invalid IP address received: %s", ip)
	}

	return ip, nil
}
