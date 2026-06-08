package services

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"upm-backend/internal/models"
)

// CertificateRenewalService periodically renews expiring Let's Encrypt certificates.
type CertificateRenewalService struct {
	db       *DatabaseService
	nginx    *NginxService
	interval time.Duration
	stopChan chan struct{}
	wg       sync.WaitGroup
	running  bool
	mu       sync.Mutex
}

// NewCertificateRenewalService creates a certificate auto-renewal scheduler.
func NewCertificateRenewalService(db *DatabaseService, nginx *NginxService, interval time.Duration) *CertificateRenewalService {
	return &CertificateRenewalService{
		db:       db,
		nginx:    nginx,
		interval: interval,
		stopChan: make(chan struct{}),
	}
}

// Start begins the background renewal scheduler.
func (s *CertificateRenewalService) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.stopChan = make(chan struct{})
	s.mu.Unlock()

	s.wg.Add(1)
	go s.run()
	log.Printf("Certificate auto-renewal scheduler started (interval: %v)", s.interval)
}

// Stop shuts down the background renewal scheduler.
func (s *CertificateRenewalService) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	close(s.stopChan)
	s.mu.Unlock()
	s.wg.Wait()
	log.Printf("Certificate auto-renewal scheduler stopped")
}

func (s *CertificateRenewalService) run() {
	defer s.wg.Done()

	// Run once at startup so recently expired certs are picked up promptly.
	s.runRenewalCheck()

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.runRenewalCheck()
		case <-s.stopChan:
			return
		}
	}
}

func (s *CertificateRenewalService) runRenewalCheck() {
	log.Printf("Running scheduled certificate renewal check")
	responses := s.RenewEligibleCertificates()

	renewed := 0
	failed := 0
	for _, response := range responses {
		if response.Success {
			renewed++
		} else if !strings.HasPrefix(response.Message, "Skipped:") {
			failed++
		}
	}

	if renewed > 0 || failed > 0 {
		log.Printf("Certificate renewal check complete: %d renewed, %d failed", renewed, failed)
	} else {
		log.Printf("Certificate renewal check complete: no certificates due for renewal")
	}
}

// RenewEligibleCertificates renews all Let's Encrypt certificates expiring within 30 days.
func (s *CertificateRenewalService) RenewEligibleCertificates() []models.CertificateRenewResponse {
	certificates, err := s.db.GetCertificates()
	if err != nil {
		return []models.CertificateRenewResponse{{
			Success: false,
			Message: fmt.Sprintf("Failed to fetch certificates: %v", err),
		}}
	}

	certService := NewCertificateService("/etc/nginx/ssl")
	var responses []models.CertificateRenewResponse
	nginxReloadNeeded := false

	for _, certificate := range certificates {
		response := models.CertificateRenewResponse{
			Domain: certificate.Domain,
		}

		if !isLetsEncryptCertificate(&certificate) {
			response.Success = false
			response.Message = "Skipped: manual certificates must be renewed externally"
			responses = append(responses, response)
			continue
		}

		expiringSoon, daysUntilExpiry := certService.CheckCertificateExpiry(&certificate)
		if !expiringSoon {
			response.Success = false
			response.Message = fmt.Sprintf("Skipped: not due for renewal (%d days remaining)", daysUntilExpiry)
			responses = append(responses, response)
			continue
		}

		log.Printf("Auto-renewing certificate for domain: %s (ID: %d)", certificate.Domain, certificate.ID)
		renewedCert, err := s.renewCertificate(&certificate, certService)
		if err != nil {
			log.Printf("Certificate renewal failed for %s: %v", certificate.Domain, err)
			response.Success = false
			response.Message = err.Error()
			responses = append(responses, response)
			continue
		}

		response.Success = true
		response.Message = "Certificate renewed successfully"
		response.Certificate = renewedCert
		responses = append(responses, response)
		nginxReloadNeeded = true
	}

	if nginxReloadNeeded && s.nginx != nil {
		if err := s.nginx.ReloadNginx(); err != nil {
			log.Printf("Warning: Failed to reload nginx after certificate renewal: %v", err)
		} else {
			log.Printf("Nginx reloaded successfully after certificate renewal")
		}
	}

	return responses
}

func (s *CertificateRenewalService) renewCertificate(certificate *models.Certificate, certService *CertificateService) (*models.Certificate, error) {
	renewedCert, err := certService.RenewCertificate(certificate)
	if err != nil {
		return nil, err
	}

	certificate.CertPath = renewedCert.CertPath
	certificate.KeyPath = renewedCert.KeyPath
	certificate.ExpiresAt = renewedCert.ExpiresAt
	certificate.IsValid = renewedCert.IsValid
	certificate.UpdatedAt = time.Now()

	if err := s.db.UpdateCertificate(certificate); err != nil {
		return nil, fmt.Errorf("failed to update certificate: %w", err)
	}

	return certificate, nil
}

func isLetsEncryptCertificate(cert *models.Certificate) bool {
	return strings.Contains(cert.CertPath, "/etc/letsencrypt") ||
		strings.Contains(cert.CertPath, "/etc/ssl/certs")
}
