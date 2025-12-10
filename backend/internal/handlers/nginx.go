package handlers

import (
	"fmt"
	"net/http"

	"upm-backend/internal/models"
	"upm-backend/internal/services"

	"github.com/gin-gonic/gin"
)

var nginxService *services.NginxService

// SetNginxService sets the nginx service instance
func SetNginxService(service *services.NginxService) {
	nginxService = service
}

// GetNginxService returns the nginx service instance
func GetNginxService() *services.NginxService {
	return nginxService
}

// ReloadNginx godoc
// @Summary      Reload nginx configuration
// @Description  Manually reload nginx configuration
// @Tags         nginx
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /nginx/reload [post]
func ReloadNginx(c *gin.Context) {
	if nginxService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Nginx service not initialized"})
		return
	}

	// Test nginx configuration first
	if err := nginxService.TestNginxConfig(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid nginx configuration: " + err.Error()})
		return
	}

	// Reload nginx
	if err := nginxService.ReloadNginx(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload nginx: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Nginx configuration reloaded successfully"})
}

// TestNginxConfig godoc
// @Summary      Test nginx configuration
// @Description  Test nginx configuration without reloading
// @Tags         nginx
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /nginx/test [post]
func TestNginxConfig(c *gin.Context) {
	if nginxService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Nginx service not initialized"})
		return
	}

	// Test nginx configuration
	if err := nginxService.TestNginxConfig(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Nginx configuration test failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Nginx configuration is valid"})
}

// UpdateAdminIPRestrictions godoc
// @Summary      Update nginx admin IP restrictions
// @Description  Update the IP ranges allowed to access the admin interface
// @Tags         nginx
// @Accept       json
// @Produce      json
// @Param        request body models.AdminIPRestrictionsRequest true "IP restrictions"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /api/v1/nginx/admin-ip-restrictions [PUT]
func UpdateAdminIPRestrictions(c *gin.Context) {
	var req models.AdminIPRestrictionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get nginx service
	nginxService := GetNginxService()
	if nginxService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Nginx service not available"})
		return
	}

	// Update admin configuration
	if err := nginxService.UpdateAdminConfig(req.AllowedRanges); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update admin config: " + err.Error()})
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

	c.JSON(http.StatusOK, gin.H{"message": "Admin IP restrictions updated successfully"})
}

// GetAdminIPRestrictions godoc
// @Summary      Get nginx admin IP restrictions
// @Description  Get the current IP ranges allowed to access the admin interface
// @Tags         nginx
// @Produce      json
// @Success      200 {object} map[string][]string
// @Failure      500 {object} map[string]string
// @Router       /api/v1/nginx/admin-ip-restrictions [GET]
func GetAdminIPRestrictions(c *gin.Context) {
	// Get nginx service
	nginxService := GetNginxService()
	if nginxService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Nginx service not available"})
		return
	}

	// Get current IP restrictions
	allowedRanges, err := nginxService.GetAdminIPRestrictions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get admin IP restrictions: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"allowed_ranges": allowedRanges})
}

// RegenerateProxyConfig godoc
// @Summary      Regenerate nginx configuration for a specific domain
// @Description  Regenerate nginx configuration for a proxy based on its domain name
// @Tags         nginx
// @Accept       json
// @Produce      json
// @Param        domain query string true "Domain name to regenerate config for"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /nginx/regenerate-config [post]
func RegenerateProxyConfig(c *gin.Context) {
	domain := c.Query("domain")
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Domain parameter is required"})
		return
	}

	// Get nginx service
	nginxService := GetNginxService()
	if nginxService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Nginx service not available"})
		return
	}

	// Get database service
	if dbService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database service not available"})
		return
	}

	// Find proxy by domain
	proxies, err := dbService.GetProxies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get proxies: " + err.Error()})
		return
	}

	var targetProxy *models.Proxy
	for _, proxy := range proxies {
		if proxy.Domain == domain {
			targetProxy = &proxy
			break
		}
	}

	if targetProxy == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Proxy not found for domain: " + domain})
		return
	}

	// Check if a certificate exists for this domain and auto-enable SSL if needed
	if !targetProxy.SSLEnabled {
		existingCert, err := dbService.GetCertificateByDomain(targetProxy.Domain)
		if err == nil && existingCert != nil {
			// Certificate exists, enable SSL on the proxy
			targetProxy.SSLEnabled = true
			targetProxy.SSLPath = existingCert.CertPath
			// Update the proxy in the database
			if err := dbService.UpdateProxy(targetProxy); err != nil {
				fmt.Printf("Warning: Failed to update proxy SSL status for %s: %v\n", targetProxy.Domain, err)
			} else {
				fmt.Printf("Auto-enabled SSL for %s based on existing certificate\n", targetProxy.Domain)
				// Reload the proxy from database to ensure we have the latest state
				updatedProxy, err := dbService.GetProxy(targetProxy.ID)
				if err == nil {
					targetProxy = updatedProxy
				}
			}
		} else {
			if err != nil {
				fmt.Printf("Certificate lookup for %s: %v\n", targetProxy.Domain, err)
			} else {
				fmt.Printf("No certificate found in database for %s\n", targetProxy.Domain)
			}
		}
	}

	// Regenerate nginx configuration
	if err := nginxService.GenerateProxyConfig(targetProxy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to regenerate nginx config: " + err.Error()})
		return
	}

	// Reload proxy from database to get final SSL status after GenerateProxyConfig
	finalProxy, err := dbService.GetProxy(targetProxy.ID)
	if err == nil {
		targetProxy = finalProxy
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

	message := fmt.Sprintf("Nginx configuration regenerated successfully for domain: %s", domain)
	if targetProxy.SSLEnabled {
		message += " (SSL enabled)"
	} else {
		// Check if certificate exists to provide helpful message
		existingCert, err := dbService.GetCertificateByDomain(targetProxy.Domain)
		if err == nil && existingCert != nil {
			message += " (SSL disabled - certificate files may be missing or invalid, check backend logs)"
		} else {
			message += " (SSL disabled - no certificate found)"
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": message, "ssl_enabled": targetProxy.SSLEnabled})
}
