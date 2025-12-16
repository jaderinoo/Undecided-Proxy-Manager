package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"upm-backend/internal/models"
	"upm-backend/internal/services"

	"github.com/gin-gonic/gin"
)

var dbService *services.DatabaseService

// SetDatabaseService sets the database service instance
func SetDatabaseService(service *services.DatabaseService) {
	dbService = service
}

// GetDatabaseService gets the database service instance
func GetDatabaseService() *services.DatabaseService {
	return dbService
}

// getNginxService gets the nginx service instance from the nginx handler
func getNginxService() *services.NginxService {
	// This is a simple way to access the nginx service
	// In a more complex setup, you might want to use dependency injection
	return GetNginxService()
}

// GetProxies godoc
// @Summary      Get all proxies
// @Description  Get a list of all proxy configurations
// @Tags         proxies
// @Accept       json
// @Produce      json
// @Success      200  {array}   models.Proxy
// @Failure      500  {object}  map[string]string
// @Router       /proxies [get]
func GetProxies(c *gin.Context) {
	if dbService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database service not initialized"})
		return
	}

	proxies, err := dbService.GetProxies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch proxies: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  proxies,
		"count": len(proxies),
	})
}

// GetProxy godoc
// @Summary      Get proxy by ID
// @Description  Get a specific proxy configuration by ID
// @Tags         proxies
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Proxy ID"
// @Success      200  {object}  models.Proxy
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /proxies/{id} [get]
func GetProxy(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proxy ID"})
		return
	}

	if dbService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database service not initialized"})
		return
	}

	proxy, err := dbService.GetProxy(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Proxy not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": proxy})
}

// CreateProxy godoc
// @Summary      Create a new proxy
// @Description  Create a new proxy configuration
// @Tags         proxies
// @Accept       json
// @Produce      json
// @Param        proxy  body      models.ProxyCreateRequest  true  "Proxy data"
// @Success      201    {object}  models.Proxy
// @Failure      400    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /proxies [post]
func CreateProxy(c *gin.Context) {
	var req models.ProxyCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if dbService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database service not initialized"})
		return
	}

	// Create proxy object
	wsEnabled := false
	if req.WSEnabled != nil {
		wsEnabled = *req.WSEnabled
	}
	proxy := &models.Proxy{
		Name:       req.Name,
		Domain:     req.Domain,
		TargetURL:  req.TargetURL,
		SSLEnabled: req.SSLEnabled,
		WSEnabled:  wsEnabled,
		Status:     "active",
	}

	// If SSL is enabled, check if certificate already exists
	if req.SSLEnabled {
		// First check if a certificate already exists for this domain
		existingCert, err := dbService.GetCertificateByDomain(req.Domain)
		if err == nil && existingCert != nil {
			// Certificate already exists, use it
			proxy.SSLEnabled = true
			proxy.SSLPath = existingCert.CertPath
			fmt.Printf("Found existing certificate for %s, enabling SSL\n", req.Domain)
		} else {
			// No existing certificate, generate Let's Encrypt certificate
			certService := services.NewCertificateService("/etc/ssl/certs")

			// Generate Let's Encrypt certificate
			certificate, err := certService.GenerateLetsEncryptCertificate(req.Domain)
			if err != nil {
				// If certificate generation fails, disable SSL and continue
				proxy.SSLEnabled = false
				proxy.Status = "active" // Still create proxy but without SSL

				// Log the error but don't fail the proxy creation
				fmt.Printf("Warning: Failed to generate Let's Encrypt certificate for %s: %v. Creating proxy without SSL.\n", req.Domain, err)
			} else {
				// Certificate generated successfully, save it to database
				if err := dbService.CreateCertificate(certificate); err != nil {
					fmt.Printf("Warning: Failed to save certificate to database: %v\n", err)
				} else {
					// Set SSL path in proxy using the certificate path from database
					proxy.SSLPath = certificate.CertPath
				}
			}
		}
	}

	// Save to database
	if err := dbService.CreateProxy(proxy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create proxy: " + err.Error()})
		return
	}

	// Generate nginx configuration
	nginxService := getNginxService()
	if nginxService != nil {
		if err := nginxService.GenerateProxyConfig(proxy); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate nginx config: " + err.Error()})
			return
		}

		// Test nginx configuration
		if err := nginxService.TestNginxConfig(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid nginx configuration: " + err.Error()})
			return
		}

		// Reload nginx
		if err := nginxService.ReloadNginx(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload nginx: " + err.Error()})
			return
		}
	}

	// Prepare response with SSL status
	response := gin.H{"data": proxy}

	// Add SSL certificate generation status
	if req.SSLEnabled {
		if proxy.SSLEnabled {
			response["ssl_status"] = "certificate_generated"
			response["ssl_message"] = "Let's Encrypt certificate generated successfully"
		} else {
			response["ssl_status"] = "certificate_failed"
			response["ssl_message"] = "Failed to generate Let's Encrypt certificate. Proxy created without SSL."
		}
	}

	c.JSON(http.StatusCreated, response)
}

// UpdateProxy godoc
// @Summary      Update a proxy
// @Description  Update an existing proxy configuration
// @Tags         proxies
// @Accept       json
// @Produce      json
// @Param        id     path      int                        true  "Proxy ID"
// @Param        proxy  body      models.ProxyUpdateRequest  true  "Proxy data"
// @Success      200    {object}  models.Proxy
// @Failure      400    {object}  map[string]string
// @Failure      404    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /proxies/{id} [put]
func UpdateProxy(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proxy ID"})
		return
	}

	var req models.ProxyUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if dbService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database service not initialized"})
		return
	}

	// Get existing proxy
	proxy, err := dbService.GetProxy(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Proxy not found"})
		return
	}

	// Update fields if provided
	if req.Name != nil {
		proxy.Name = *req.Name
	}
	if req.Domain != nil {
		proxy.Domain = *req.Domain
	}
	if req.TargetURL != nil {
		proxy.TargetURL = *req.TargetURL
	}
	if req.WSEnabled != nil {
		proxy.WSEnabled = *req.WSEnabled
	}
	if req.SSLEnabled != nil {
		// If SSL is being enabled, check if certificate already exists
		if *req.SSLEnabled && !proxy.SSLEnabled {
			// First check if a certificate already exists for this domain
			existingCert, err := dbService.GetCertificateByDomain(proxy.Domain)
			if err == nil && existingCert != nil {
				// Certificate already exists, enable SSL
				proxy.SSLEnabled = true
				proxy.SSLPath = existingCert.CertPath
				fmt.Printf("Found existing certificate for %s, enabling SSL\n", proxy.Domain)
			} else {
				// No existing certificate, generate Let's Encrypt certificate
				certService := services.NewCertificateService("/etc/ssl/certs")

				// Generate Let's Encrypt certificate
				certificate, err := certService.GenerateLetsEncryptCertificate(proxy.Domain)
				if err != nil {
					// If certificate generation fails, keep SSL disabled
					fmt.Printf("Warning: Failed to generate Let's Encrypt certificate for %s: %v. Keeping SSL disabled.\n", proxy.Domain, err)
					proxy.SSLEnabled = false
				} else {
					// Certificate generated successfully, save it to database
					if err := dbService.CreateCertificate(certificate); err != nil {
						fmt.Printf("Warning: Failed to save certificate to database: %v\n", err)
						proxy.SSLEnabled = false
					} else {
						proxy.SSLEnabled = true
						proxy.SSLPath = certificate.CertPath
					}
				}
			}
		} else {
			proxy.SSLEnabled = *req.SSLEnabled
		}
	}

	// Save to database
	if err := dbService.UpdateProxy(proxy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update proxy: " + err.Error()})
		return
	}

	// Update nginx configuration
	nginxService := getNginxService()
	if nginxService != nil {
		if err := nginxService.UpdateProxyConfig(proxy); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update nginx config: " + err.Error()})
			return
		}

		// Test nginx configuration
		if err := nginxService.TestNginxConfig(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid nginx configuration: " + err.Error()})
			return
		}

		// Reload nginx
		if err := nginxService.ReloadNginx(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload nginx: " + err.Error()})
			return
		}
	}

	// Prepare response with SSL status
	response := gin.H{"data": proxy}

	// Add SSL certificate generation status if SSL was enabled
	if req.SSLEnabled != nil && *req.SSLEnabled {
		if proxy.SSLEnabled {
			response["ssl_status"] = "certificate_generated"
			response["ssl_message"] = "Let's Encrypt certificate generated successfully"
		} else {
			response["ssl_status"] = "certificate_failed"
			response["ssl_message"] = "Failed to generate Let's Encrypt certificate. SSL remains disabled."
		}
	}

	c.JSON(http.StatusOK, response)
}

// DeleteProxy godoc
// @Summary      Delete a proxy
// @Description  Delete a proxy configuration
// @Tags         proxies
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Proxy ID"
// @Success      204  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /proxies/{id} [delete]
func DeleteProxy(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proxy ID"})
		return
	}

	if dbService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database service not initialized"})
		return
	}

	// Remove from database
	if err := dbService.DeleteProxy(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to delete proxy: " + err.Error()})
		return
	}

	// Remove nginx configuration
	nginxService := getNginxService()
	if nginxService != nil {
		if err := nginxService.RemoveProxyConfig(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove nginx config: " + err.Error()})
			return
		}

		// Test nginx configuration
		if err := nginxService.TestNginxConfig(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid nginx configuration: " + err.Error()})
			return
		}

		// Reload nginx
		if err := nginxService.ReloadNginx(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload nginx: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusNoContent, gin.H{"message": "Proxy deleted successfully"})
}

// GetProxyCertificate godoc
// @Summary      Get certificate information for a proxy
// @Description  Get certificate details for a specific proxy by domain
// @Tags         proxies
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Proxy ID"
// @Success      200  {object}  models.Certificate
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /proxies/{id}/certificate [get]
func GetProxyCertificate(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proxy ID"})
		return
	}

	if dbService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database service not initialized"})
		return
	}

	// Get proxy first to get domain
	proxy, err := dbService.GetProxy(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Proxy not found"})
		return
	}

	// Get certificate by domain (return if present regardless of flag)
	certificate, err := dbService.GetCertificateByDomain(proxy.Domain)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Certificate not found for this domain"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": certificate})
}
