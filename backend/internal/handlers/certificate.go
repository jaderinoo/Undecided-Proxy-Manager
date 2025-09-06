package handlers

import (
	"net/http"
	"strconv"
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

	// Remove from database
	if err := dbService.DeleteCertificate(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to delete certificate: " + err.Error()})
		return
	}

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
	renewedCert, err := certService.RenewCertificate(certificate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to renew certificate: " + err.Error()})
		return
	}

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate certificate: " + err.Error()})
		return
	}

	// Save to database
	if err := dbService.CreateCertificate(certificate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save certificate: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": certificate, "message": "Let's Encrypt certificate generated successfully"})
}
