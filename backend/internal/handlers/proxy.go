package handlers

import (
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
	proxy := &models.Proxy{
		Name:       req.Name,
		Domain:     req.Domain,
		TargetURL:  req.TargetURL,
		SSLEnabled: req.SSLEnabled,
		Status:     "active",
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

	c.JSON(http.StatusCreated, gin.H{"data": proxy})
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
	if req.SSLEnabled != nil {
		proxy.SSLEnabled = *req.SSLEnabled
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

	c.JSON(http.StatusOK, gin.H{"data": proxy})
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
