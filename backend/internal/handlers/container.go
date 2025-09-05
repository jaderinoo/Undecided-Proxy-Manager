package handlers

import (
	"net/http"

	"upm-backend/internal/models"
	"upm-backend/internal/services"

	"github.com/gin-gonic/gin"
)

var dockerService *services.DockerService

// SetDockerService sets the docker service instance
func SetDockerService(service *services.DockerService) {
	dockerService = service
}

// GetContainers godoc
// @Summary      Get all containers
// @Description  Get a list of all containers (running and stopped)
// @Tags         containers
// @Accept       json
// @Produce      json
// @Success      200  {object}  models.ContainerListResponse
// @Failure      500  {object}  map[string]string
// @Router       /containers [get]
func GetContainers(c *gin.Context) {
	if dockerService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Docker service not initialized"})
		return
	}

	containers, err := dockerService.GetRunningContainers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch containers: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.ContainerListResponse{
		Containers: containers,
		Count:      len(containers),
	})
}

// GetContainer godoc
// @Summary      Get container by ID
// @Description  Get a specific container by ID with detailed information
// @Tags         containers
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Container ID"
// @Success      200  {object}  models.Container
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /containers/{id} [get]
func GetContainer(c *gin.Context) {
	containerID := c.Param("id")
	if containerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Container ID is required"})
		return
	}

	if dockerService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Docker service not initialized"})
		return
	}

	container, err := dockerService.GetContainerByID(containerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Container not found: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": container})
}

// GetContainerStats godoc
// @Summary      Get container stats
// @Description  Get real-time statistics for a specific container
// @Tags         containers
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Container ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /containers/{id}/stats [get]
func GetContainerStats(c *gin.Context) {
	containerID := c.Param("id")
	if containerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Container ID is required"})
		return
	}

	if dockerService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Docker service not initialized"})
		return
	}

	stats, err := dockerService.GetContainerStats(containerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to get container stats: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}
