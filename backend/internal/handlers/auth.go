package handlers

import (
	"net/http"

	"upm-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// Login godoc
// @Summary      User login
// @Description  Authenticate user and return JWT token
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

	// TODO: Implement authentication logic
	// For now, return a mock response
	user := models.User{
		ID:       1,
		Username: req.Username,
		Email:    "user@example.com",
		IsActive: true,
	}

	response := models.AuthResponse{
		Token: "mock-jwt-token",
		User:  user,
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// Register godoc
// @Summary      User registration
// @Description  Register a new user account
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body      models.UserCreateRequest  true  "User registration data"
// @Success      201   {object}  models.AuthResponse
// @Failure      400   {object}  map[string]string
// @Failure      409   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /auth/register [post]
func Register(c *gin.Context) {
	var req models.UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement registration logic
	// For now, return a mock response
	user := models.User{
		ID:       1,
		Username: req.Username,
		Email:    req.Email,
		IsActive: true,
	}

	response := models.AuthResponse{
		Token: "mock-jwt-token",
		User:  user,
	}

	c.JSON(http.StatusCreated, gin.H{"data": response})
}
