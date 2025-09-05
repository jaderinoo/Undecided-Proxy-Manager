package handlers

import (
	"net/http"

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
