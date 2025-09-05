package handlers

import (
	"net/http"

	"upm-backend/internal/auth"
	"upm-backend/internal/config"
	"upm-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// Login godoc
// @Summary      Admin login
// @Description  Authenticate admin and return JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials  body      models.UserLoginRequest  true  "Login credentials"
// @Success      200         {object}  models.AuthResponse
// @Failure      400         {object}  map[string]string
// @Failure      401         {object}  map[string]string
// @Failure      500         {object}  map[string]string
// @Router       /auth/login [post]
func Login(c *gin.Context) {
	var req models.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get config
	cfg := config.Load()

	// Check if admin password is set (unless in dev mode with dev test password)
	if cfg.AdminPassword == "" && !(cfg.DevMode && req.Password == cfg.DevTestPassword) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Admin password not configured"})
		return
	}

	// Authenticate admin (single password) with dev bypass
	if !auth.AuthenticateAdmin(req.Password, cfg.AdminPassword, cfg.DevMode, cfg.DevTestPassword) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token (single admin token)
	token, err := auth.GenerateToken(cfg.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": models.AuthResponse{
		Token: token,
		// Minimal user info for single admin auth
		User: models.User{
			ID:       1,
			Username: "admin",
			Email:    "admin@upm.local",
			IsActive: true,
		},
	}})
}

// Register godoc
// @Summary      User registration (disabled)
// @Description  Registration is disabled - UPM uses single admin authentication
// @Tags         auth
// @Accept       json
// @Produce      json
// @Failure      403   {object}  map[string]string
// @Router       /auth/register [post]
func Register(c *gin.Context) {
	c.JSON(http.StatusForbidden, gin.H{
		"error": "Registration is disabled. UPM uses single admin authentication. Set ADMIN_PASSWORD environment variable to configure admin access.",
	})
}
