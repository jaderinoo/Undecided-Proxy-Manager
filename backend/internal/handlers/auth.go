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
// @Description  Authenticate admin and return JWT token (Pi-hole style)
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

	// Check if admin password is set
	if cfg.AdminPassword == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Admin password not configured"})
		return
	}

	// Authenticate admin (Pi-hole style - single password)
	if !auth.AuthenticateAdmin(req.Password, cfg.AdminPassword) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken("admin", cfg.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Create admin user response
	user := models.User{
		ID:       1,
		Username: "admin",
		Email:    "admin@upm.local",
		IsActive: true,
	}

	response := models.AuthResponse{
		Token: token,
		User:  user,
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
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
