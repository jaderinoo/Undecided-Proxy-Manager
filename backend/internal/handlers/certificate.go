package handlers

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"upm-backend/internal/models"
	"upm-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// GetCertificates godoc
// @Summary      Get all certificates
// @Description  Get a list of all SSL certificates
// @Tags         certificates
// @Accept       json
// @Produce      json
// @Success      200  {array}   models.Certificate
// @Failure      500  {object}  map[string]string
// @Router       /certificates [get]
func GetCertificates(c *gin.Context) {
	if dbService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database service not initialized"})
		return
	}

	certificates, err := dbService.GetCertificates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch certificates: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  certificates,
		"count": len(certificates),
	})
}

// GetCertificate godoc
// @Summary      Get certificate by ID
// @Description  Get a specific certificate by ID
// @Tags         certificates
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Certificate ID"
// @Success      200  {object}  models.Certificate
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /certificates/{id} [get]
func GetCertificate(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid certificate ID"})
		return
	}

	if dbService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database service not initialized"})
		return
	}

	certificate, err := dbService.GetCertificate(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Certificate not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": certificate})
}

// CreateCertificate godoc
// @Summary      Create a new certificate
// @Description  Create a new SSL certificate (manual upload or Let's Encrypt)
// @Tags         certificates
// @Accept       json
// @Produce      json
// @Param        certificate  body      models.CertificateCreateRequest  true  "Certificate data"
// @Success      201    {object}  models.Certificate
// @Failure      400    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /certificates [post]
func CreateCertificate(c *gin.Context) {
	var req models.CertificateCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if dbService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database service not initialized"})
		return
	}

	// Create certificate object
	certificate := &models.Certificate{
		Domain:    req.Domain,
		CertPath:  req.CertPath,
		KeyPath:   req.KeyPath,
		ExpiresAt: req.ExpiresAt,
		IsValid:   true,
	}

	// Save to database
	if err := dbService.CreateCertificate(certificate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create certificate: " + err.Error()})
		return
	}

	// Enable SSL on matching proxies
	enableSSLForDomain(req.Domain, certificate.CertPath)

	c.JSON(http.StatusCreated, gin.H{"data": certificate})
}

// UpdateCertificate godoc
// @Summary      Update a certificate
// @Description  Update an existing certificate
// @Tags         certificates
// @Accept       json
// @Produce      json
// @Param        id     path      int                        true  "Certificate ID"
// @Param        certificate  body      models.CertificateUpdateRequest  true  "Certificate data"
// @Success      200    {object}  models.Certificate
// @Failure      400    {object}  map[string]string
// @Failure      404    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /certificates/{id} [put]
func UpdateCertificate(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid certificate ID"})
		return
	}

	var req models.CertificateUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if dbService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database service not initialized"})
		return
	}

	// Get existing certificate
	certificate, err := dbService.GetCertificate(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Certificate not found"})
		return
	}

	// Update fields if provided
	if req.Domain != nil {
		certificate.Domain = *req.Domain
	}
	if req.CertPath != nil {
		certificate.CertPath = *req.CertPath
	}
	if req.KeyPath != nil {
		certificate.KeyPath = *req.KeyPath
	}
	if req.ExpiresAt != nil {
		certificate.ExpiresAt = *req.ExpiresAt
	}
	if req.IsValid != nil {
		certificate.IsValid = *req.IsValid
	}

	// Save to database
	if err := dbService.UpdateCertificate(certificate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update certificate: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": certificate})
}

// DeleteCertificate godoc
// @Summary      Delete a certificate
// @Description  Delete a certificate
// @Tags         certificates
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Certificate ID"
// @Success      204  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /certificates/{id} [delete]
func DeleteCertificate(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid certificate ID"})
		return
	}

	if dbService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database service not initialized"})
		return
	}

	// Get certificate info before deletion for logging
	certificate, err := dbService.GetCertificate(id)
	if err == nil {
		log.Printf("Deleting certificate ID %d for domain: %s", id, certificate.Domain)
	}

	// Remove from database
	if err := dbService.DeleteCertificate(id); err != nil {
		log.Printf("Failed to delete certificate ID %d: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to delete certificate: " + err.Error()})
		return
	}

	// Disable SSL on proxies that were using this certificate
	disableSSLForDomain(certificate.Domain)

	log.Printf("Successfully deleted certificate ID %d", id)
	c.JSON(http.StatusNoContent, gin.H{"message": "Certificate deleted successfully"})
}

// GetCertificateProxies godoc
// @Summary      Get proxies using a certificate
// @Description  Get all proxies that use a specific certificate
// @Tags         certificates
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Certificate ID"
// @Success      200  {array}   models.Proxy
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /certificates/{id}/proxies [get]
func GetCertificateProxies(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid certificate ID"})
		return
	}

	if dbService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database service not initialized"})
		return
	}

	// Get certificate first to get domain
	certificate, err := dbService.GetCertificate(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Certificate not found"})
		return
	}

	// Get proxies that use this certificate (by domain matching)
	proxies, err := dbService.GetProxiesByDomain(certificate.Domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch proxies: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  proxies,
		"count": len(proxies),
	})
}

// RenewCertificate godoc
// @Summary      Renew a certificate
// @Description  Renew a certificate using Let's Encrypt
// @Tags         certificates
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Certificate ID"
// @Success      200  {object}  models.Certificate
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /certificates/{id}/renew [post]
func RenewCertificate(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid certificate ID"})
		return
	}

	if dbService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database service not initialized"})
		return
	}

	// Get existing certificate
	certificate, err := dbService.GetCertificate(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Certificate not found"})
		return
	}

	// Create certificate service
	certService := services.NewCertificateService("/etc/nginx/ssl")

	// Renew certificate using Let's Encrypt
	log.Printf("Attempting to renew certificate for domain: %s (ID: %d)", certificate.Domain, id)
	renewedCert, err := certService.RenewCertificate(certificate)
	if err != nil {
		log.Printf("Certificate renewal failed for %s: %v", certificate.Domain, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to renew certificate: " + err.Error()})
		return
	}
	log.Printf("Certificate renewal successful for %s", certificate.Domain)

	// Update certificate in database
	certificate.CertPath = renewedCert.CertPath
	certificate.KeyPath = renewedCert.KeyPath
	certificate.ExpiresAt = renewedCert.ExpiresAt
	certificate.IsValid = renewedCert.IsValid
	certificate.UpdatedAt = time.Now()

	// Save to database
	if err := dbService.UpdateCertificate(certificate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update certificate: " + err.Error()})
		return
	}

	// Certificates are now saved directly to /etc/ssl/certs, so no copy needed
	// Reload nginx to pick up the new certificate
	nginxService := GetNginxService()
	if nginxService != nil {
		if err := nginxService.ReloadNginx(); err != nil {
			// Log error but don't fail the renewal - certificate is already renewed
			log.Printf("Warning: Failed to reload nginx after certificate renewal: %v", err)
		} else {
			log.Printf("Nginx reloaded successfully after certificate renewal for %s", certificate.Domain)
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": certificate, "message": "Certificate renewed successfully"})
}

// GenerateLetsEncryptCertificate godoc
// @Summary      Generate Let's Encrypt certificate
// @Description  Generate a new SSL certificate using Let's Encrypt
// @Tags         certificates
// @Accept       json
// @Produce      json
// @Param        request  body      models.LetsEncryptRequest  true  "Let's Encrypt certificate request"
// @Success      201      {object}  models.Certificate
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /certificates/letsencrypt [post]
func GenerateLetsEncryptCertificate(c *gin.Context) {
	var req models.LetsEncryptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if dbService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database service not initialized"})
		return
	}

	// Create certificate service
	certService := services.NewCertificateService("/etc/nginx/ssl")

	// Generate Let's Encrypt certificate
	certificate, err := certService.GenerateLetsEncryptCertificate(req.Domain)
	if err != nil {
		// Extract and format user-friendly error message
		errorMsg := formatLetsEncryptError(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errorMsg})
		return
	}

	// Save to database
	if err := dbService.CreateCertificate(certificate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save certificate: " + err.Error()})
		return
	}

	// Enable SSL on matching proxies
	enableSSLForDomain(req.Domain, certificate.CertPath)

	// Regenerate nginx config for proxies with this domain and reload nginx
	regenerateNginxConfigForDomain(req.Domain)

	c.JSON(http.StatusCreated, gin.H{"data": certificate, "message": "Let's Encrypt certificate generated successfully"})
}

// enableSSLForDomain marks proxies matching the domain as SSL-enabled and sets SSLPath.
func enableSSLForDomain(domain, certPath string) {
	if dbService == nil {
		return
	}

	proxies, err := dbService.GetProxiesByDomain(domain)
	if err != nil {
		return
	}

	for _, p := range proxies {
		p.SSLEnabled = true
		// Use the cert file path as the stored SSL path reference.
		p.SSLPath = certPath
		_ = dbService.UpdateProxy(&p)
	}
}

// disableSSLForDomain marks proxies matching the domain as SSL-disabled.
func disableSSLForDomain(domain string) {
	if dbService == nil {
		return
	}

	proxies, err := dbService.GetProxiesByDomain(domain)
	if err != nil {
		return
	}

	for _, p := range proxies {
		p.SSLEnabled = false
		_ = dbService.UpdateProxy(&p)
	}
}

// regenerateNginxConfigForDomain regenerates nginx config for all proxies with the given domain and reloads nginx
func regenerateNginxConfigForDomain(domain string) {
	if dbService == nil {
		return
	}

	nginxService := GetNginxService()
	if nginxService == nil {
		log.Printf("Nginx service not available, skipping config regeneration for domain: %s", domain)
		return
	}

	// Find all proxies with this domain
	proxies, err := dbService.GetProxiesByDomain(domain)
	if err != nil {
		log.Printf("Failed to get proxies for domain %s: %v", domain, err)
		return
	}

	if len(proxies) == 0 {
		log.Printf("No proxies found for domain %s, skipping nginx config regeneration", domain)
		return
	}

	// Regenerate config for each proxy
	for _, proxy := range proxies {
		// Check if a certificate exists and auto-enable SSL if needed
		if !proxy.SSLEnabled {
			existingCert, err := dbService.GetCertificateByDomain(proxy.Domain)
			if err == nil && existingCert != nil {
				proxy.SSLEnabled = true
				proxy.SSLPath = existingCert.CertPath
				if err := dbService.UpdateProxy(&proxy); err != nil {
					log.Printf("Warning: Failed to update proxy SSL status for %s: %v", proxy.Domain, err)
				} else {
					log.Printf("Auto-enabled SSL for %s based on existing certificate", proxy.Domain)
					// Reload proxy from database to ensure we have the latest state
					updatedProxy, err := dbService.GetProxy(proxy.ID)
					if err == nil {
						proxy = *updatedProxy
					}
				}
			}
		}

		// Regenerate nginx configuration
		if err := nginxService.GenerateProxyConfig(&proxy); err != nil {
			log.Printf("Failed to regenerate nginx config for proxy %d (%s): %v", proxy.ID, proxy.Domain, err)
			continue
		}
		log.Printf("Regenerated nginx config for proxy %d (%s)", proxy.ID, proxy.Domain)
	}

	// Test nginx configuration
	if err := nginxService.TestNginxConfig(); err != nil {
		log.Printf("Invalid nginx configuration after certificate creation: %v", err)
		return
	}

	// Reload nginx once for all updated proxies
	if err := nginxService.ReloadNginx(); err != nil {
		log.Printf("Failed to reload nginx after certificate creation: %v", err)
		return
	}

	log.Printf("Successfully regenerated nginx config and reloaded nginx for domain: %s", domain)
}

// formatLetsEncryptError extracts and formats Let's Encrypt errors into user-friendly messages
func formatLetsEncryptError(err error) string {
	if err == nil {
		return "Unknown error occurred"
	}

	errorStr := err.Error()

	// Check for rate limit errors
	if strings.Contains(errorStr, "rateLimited") || strings.Contains(errorStr, "rate limit") {
		// Extract retry time if available
		retryTime := ""
		if strings.Contains(errorStr, "retry after") {
			parts := strings.Split(errorStr, "retry after")
			if len(parts) > 1 {
				retryTime = strings.TrimSpace(strings.Split(parts[1], ":")[0])
			}
		}

		// Extract certificate count if available
		certCount := ""
		if strings.Contains(errorStr, "already issued") {
			// Look for number before "already issued"
			re := regexp.MustCompile(`(\d+)\s+already issued`)
			matches := re.FindStringSubmatch(errorStr)
			if len(matches) > 1 {
				certCount = matches[1]
			}
		}

		msg := "Let's Encrypt rate limit reached"
		if certCount != "" {
			msg += fmt.Sprintf(": %s certificates already issued for this domain in the last 168 hours", certCount)
		} else {
			msg += ": Too many certificates already issued for this domain in the last 168 hours"
		}
		if retryTime != "" {
			msg += fmt.Sprintf(". You can retry after %s", retryTime)
		} else {
			msg += ". Please wait 168 hours before requesting another certificate for this domain."
		}
		msg += " See https://letsencrypt.org/docs/rate-limits/ for more information."
		return msg
	}

	// Check for other common Let's Encrypt errors
	if strings.Contains(errorStr, "urn:ietf:params:acme:error") {
		// Extract the error type
		if strings.Contains(errorStr, "invalidEmail") {
			return "Invalid email address for Let's Encrypt registration. Please check your LETSENCRYPT_EMAIL configuration."
		}
		if strings.Contains(errorStr, "connection") || strings.Contains(errorStr, "timeout") {
			return "Connection error while communicating with Let's Encrypt. Please check your internet connection and try again."
		}
		if strings.Contains(errorStr, "unauthorized") {
			return "Authorization failed. The domain may not be pointing to this server, or the HTTP-01 challenge cannot be completed."
		}
		if strings.Contains(errorStr, "dns") {
			return "DNS validation failed. Please ensure the domain DNS records are correctly configured."
		}
	}

	// For other errors, return a cleaned-up version
	// Remove technical prefixes but keep the useful information
	cleaned := errorStr
	if strings.Contains(cleaned, "failed to generate Let's Encrypt certificate:") {
		cleaned = strings.TrimPrefix(cleaned, "failed to generate Let's Encrypt certificate: ")
	}
	if strings.Contains(cleaned, "failed to obtain certificate:") {
		cleaned = strings.TrimPrefix(cleaned, "failed to obtain certificate: ")
	}

	return cleaned
}
