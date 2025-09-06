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

	// Get database service
	dbService := GetDatabaseService()
	if dbService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database service not available"})
		return
	}

	// Get admin user from database
	adminUser, err := dbService.GetAdminUser()
	if err != nil {
		// If admin user doesn't exist, check for dev mode bypass
		if cfg.DevMode && req.Password == cfg.DevTestPassword {
			// Generate JWT token for dev mode
			token, err := auth.GenerateToken(cfg.JWTSecret)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"data": models.AuthResponse{
				Token: token,
				User: models.User{
					ID:       1,
					Username: "admin",
					Email:    "admin@upm.local",
					IsActive: true,
				},
			}})
			return
		}
		
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin access is disabled. Set ADMIN_PASSWORD environment variable to enable admin access."})
		return
	}

	// Check if admin user is active
	if !adminUser.IsActive {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin account is disabled"})
		return
	}

	// Authenticate using database password
	if !auth.CheckPasswordHash(req.Password, adminUser.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(cfg.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": models.AuthResponse{
		Token: token,
		User:  *adminUser,
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
