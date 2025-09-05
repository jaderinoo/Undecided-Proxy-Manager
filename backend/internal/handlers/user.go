package handlers

import (
	"net/http"
	"strconv"

	"upm-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// GetUsers godoc
// @Summary      Get all users
// @Description  Get a list of all users
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {array}   models.User
// @Failure      500  {object}  map[string]string
// @Router       /users [get]
func GetUsers(c *gin.Context) {
	// TODO: Implement database query
	users := []models.User{
		{
			ID:       1,
			Username: "admin",
			Email:    "admin@example.com",
			IsActive: true,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  users,
		"count": len(users),
	})
}

// GetUser godoc
// @Summary      Get user by ID
// @Description  Get a specific user by ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  models.User
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /users/{id} [get]
func GetUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// TODO: Implement database query
	user := models.User{
		ID:       id,
		Username: "admin",
		Email:    "admin@example.com",
		IsActive: true,
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

// CreateUser godoc
// @Summary      Create a new user
// @Description  Create a new user account
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user  body      models.UserCreateRequest  true  "User data"
// @Success      201   {object}  models.User
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /users [post]
func CreateUser(c *gin.Context) {
	var req models.UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement database insert
	user := models.User{
		ID:       1,
		Username: req.Username,
		Email:    req.Email,
		IsActive: true,
	}

	c.JSON(http.StatusCreated, gin.H{"data": user})
}

// UpdateUser godoc
// @Summary      Update a user
// @Description  Update an existing user account
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id    path      int  true  "User ID"
// @Param        user  body      models.User  true  "User data"
// @Success      200   {object}  models.User
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /users/{id} [put]
func UpdateUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req models.User
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement database update
	// For now, return a mock response with the provided ID
	user := models.User{
		ID:       id,
		Username: req.Username,
		Email:    req.Email,
		IsActive: req.IsActive,
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

// DeleteUser godoc
// @Summary      Delete a user
// @Description  Delete a user account
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      204  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /users/{id} [delete]
func DeleteUser(c *gin.Context) {
	_, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// TODO: Implement database delete
	c.JSON(http.StatusNoContent, gin.H{"message": "User deleted successfully"})
}
