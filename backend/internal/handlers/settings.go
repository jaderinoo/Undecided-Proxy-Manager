package handlers

import (
	"net/http"

	"upm-backend/internal/config"
	"upm-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// GetSettings returns the current application settings
// @Summary Get application settings
// @Description Get current application settings including core (.env) and UI-manageable settings
// @Tags settings
// @Accept json
// @Produce json
// @Success 200 {object} models.Settings
// @Failure 500 {object} map[string]string
// @Router /settings [get]
// @Security BearerAuth
func GetSettings(c *gin.Context) {
	// Load current configuration
	cfg := config.Load()

	// Create core settings and mask sensitive data
	coreSettings := models.CoreSettings{
		DatabasePath:        cfg.DatabasePath,
		Environment:         cfg.Environment,
		BackendPort:         cfg.BackendPort,
		AdminPassword:       cfg.AdminPassword,
		JWTSecret:           cfg.JWTSecret,
		LetsEncryptEmail:    cfg.LetsEncryptEmail,
		LetsEncryptWebroot:  cfg.LetsEncryptWebroot,
		LetsEncryptCertPath: cfg.LetsEncryptCertPath,
		DNSCheckInterval:    cfg.DNSCheckInterval,
		PublicIPService:     cfg.PublicIPService,
	}
	coreSettings.MaskSensitiveData()

	// Get UI settings from database
	uiSettings, err := dbService.GetUISettings()
	if err != nil {
		// If no UI settings exist, use defaults
		uiSettings = models.UISettings{
			DisplayName: "UPM Admin",
			Theme:       "auto",
			Language:    "en",
		}
	}

	settings := models.Settings{
		CoreSettings: coreSettings,
		UISettings:   uiSettings,
	}

	c.JSON(http.StatusOK, settings)
}

// UpdateSettings updates the UI-manageable settings
// @Summary Update UI settings
// @Description Update settings that can be managed through the UI
// @Tags settings
// @Accept json
// @Produce json
// @Param settings body models.SettingsUpdateRequest true "Settings to update"
// @Success 200 {object} models.Settings
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /settings [put]
// @Security BearerAuth
func UpdateSettings(c *gin.Context) {
	var updateReq models.SettingsUpdateRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current UI settings
	currentSettings, err := dbService.GetUISettings()
	if err != nil {
		// If no settings exist, create defaults
		currentSettings = models.UISettings{
			DisplayName: "UPM Admin",
			Theme:       "auto",
			Language:    "en",
		}
	}

	// Update only provided fields
	if updateReq.DisplayName != nil {
		currentSettings.DisplayName = *updateReq.DisplayName
	}
	if updateReq.Theme != nil {
		currentSettings.Theme = *updateReq.Theme
	}
	if updateReq.Language != nil {
		currentSettings.Language = *updateReq.Language
	}

	// Save updated settings
	if err := dbService.SaveUISettings(currentSettings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save settings"})
		return
	}

	// Return updated settings
	cfg := config.Load()
	coreSettings := models.CoreSettings{
		DatabasePath:        cfg.DatabasePath,
		Environment:         cfg.Environment,
		BackendPort:         cfg.BackendPort,
		AdminPassword:       cfg.AdminPassword,
		JWTSecret:           cfg.JWTSecret,
		LetsEncryptEmail:    cfg.LetsEncryptEmail,
		LetsEncryptWebroot:  cfg.LetsEncryptWebroot,
		LetsEncryptCertPath: cfg.LetsEncryptCertPath,
	}
	coreSettings.MaskSensitiveData()

	settings := models.Settings{
		CoreSettings: coreSettings,
		UISettings:   currentSettings,
	}

	c.JSON(http.StatusOK, settings)
}
