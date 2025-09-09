package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"upm-backend/internal/models"
	"upm-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type DNSHandler struct {
	dnsService        *services.DNSService
	schedulerService  *services.SchedulerService
}

var dnsHandler *DNSHandler

func NewDNSHandler(dnsService *services.DNSService) *DNSHandler {
	return &DNSHandler{
		dnsService: dnsService,
	}
}

// SetDNSService sets the DNS service instance
func SetDNSService(service *services.DNSService) {
	dnsHandler = NewDNSHandler(service)
}

// SetSchedulerService sets the scheduler service instance
func SetSchedulerService(service *services.SchedulerService) {
	if dnsHandler != nil {
		dnsHandler.schedulerService = service
	}
}

// DNS Config handlers

// GetDNSConfigs returns all DNS configurations
func GetDNSConfigs(c *gin.Context) {
	configs, err := dnsHandler.dnsService.DbService.GetDNSConfigs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"configs": configs})
}

// GetDNSConfig returns a specific DNS configuration
func GetDNSConfig(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	config, err := dnsHandler.dnsService.DbService.GetDNSConfig(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"config": config})
}

// CreateDNSConfig creates a new DNS configuration
func CreateDNSConfig(c *gin.Context) {
	var req models.DNSConfigCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate that username and password are provided for dynamic DNS providers
	if req.Provider == "namecheap" {
		if req.Username == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required for dynamic DNS providers"})
			return
		}
		if req.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required for dynamic DNS providers"})
			return
		}
	}

	config := &models.DNSConfig{
		Provider: models.DNSProvider(req.Provider),
		Domain:   req.Domain,
		Username: req.Username,
		Password: req.Password,
		IsActive: true,
	}

	if err := dnsHandler.dnsService.DbService.CreateDNSConfig(config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"config": config})
}

// UpdateDNSConfig updates an existing DNS configuration
func UpdateDNSConfig(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req models.DNSConfigUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config, err := dnsHandler.dnsService.DbService.GetDNSConfig(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Update fields if provided
	if req.Provider != nil {
		config.Provider = models.DNSProvider(*req.Provider)
	}
	if req.Domain != nil {
		config.Domain = *req.Domain
	}
	if req.Username != nil {
		config.Username = *req.Username
	}
	if req.Password != nil {
		config.Password = *req.Password
	}
	if req.IsActive != nil {
		config.IsActive = *req.IsActive
	}

	if err := dnsHandler.dnsService.DbService.UpdateDNSConfig(config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"config": config})
}

// DeleteDNSConfig deletes a DNS configuration
func DeleteDNSConfig(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := dnsHandler.dnsService.DbService.DeleteDNSConfig(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "DNS configuration deleted successfully"})
}

// DNS Record handlers

// GetDNSRecords returns all DNS records for a specific configuration
func GetDNSRecords(c *gin.Context) {
	configIDStr := c.Query("config_id")
	if configIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "config_id query parameter is required"})
		return
	}

	configID, err := strconv.Atoi(configIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
		return
	}

	records, err := dnsHandler.dnsService.DbService.GetDNSRecords(configID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"records": records})
}

// GetDNSRecord returns a specific DNS record
func GetDNSRecord(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	record, err := dnsHandler.dnsService.DbService.GetDNSRecord(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"record": record})
}

// CreateDNSRecord creates a new DNS record
func CreateDNSRecord(c *gin.Context) {
	var req models.DNSRecordCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	record := &models.DNSRecord{
		ConfigID:              req.ConfigID,
		Host:                  req.Host,
		AllowedIPRanges:       req.AllowedIPRanges,
		DynamicDNSRefreshRate: req.DynamicDNSRefreshRate,
		IsActive:              true,
	}

	if err := dnsHandler.dnsService.DbService.CreateDNSRecord(record); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Start scheduled job if refresh rate is set
	if dnsHandler.schedulerService != nil && record.DynamicDNSRefreshRate != nil && *record.DynamicDNSRefreshRate > 0 {
		if err := dnsHandler.schedulerService.StartScheduledJob(record.ID, *record.DynamicDNSRefreshRate); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Warning: Failed to start scheduled job for record %d: %v\n", record.ID, err)
		}
	}

	c.JSON(http.StatusCreated, gin.H{"record": record})
}

// UpdateDNSRecord updates an existing DNS record
func UpdateDNSRecord(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req models.DNSRecordUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	record, err := dnsHandler.dnsService.DbService.GetDNSRecord(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Update fields if provided
	if req.Host != nil {
		record.Host = *req.Host
	}
	if req.AllowedIPRanges != nil {
		record.AllowedIPRanges = *req.AllowedIPRanges
	}
	if req.DynamicDNSRefreshRate != nil {
		record.DynamicDNSRefreshRate = req.DynamicDNSRefreshRate
	}
	if req.IsActive != nil {
		record.IsActive = *req.IsActive
	}

	if err := dnsHandler.dnsService.DbService.UpdateDNSRecord(record); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update scheduled job if refresh rate changed
	if dnsHandler.schedulerService != nil {
		if err := dnsHandler.schedulerService.UpdateScheduledJob(record.ID, record.DynamicDNSRefreshRate); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Warning: Failed to update scheduled job for record %d: %v\n", record.ID, err)
		}
	}

	c.JSON(http.StatusOK, gin.H{"record": record})
}

// DeleteDNSRecord deletes a DNS record
func DeleteDNSRecord(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := dnsHandler.dnsService.DbService.DeleteDNSRecord(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Stop scheduled job if it exists
	if dnsHandler.schedulerService != nil {
		dnsHandler.schedulerService.StopScheduledJob(id)
	}

	c.JSON(http.StatusOK, gin.H{"message": "DNS record deleted successfully"})
}

// DNS Update handlers

// UpdateDNSRecord updates a specific DNS record
func UpdateDNSRecordNow(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if dnsHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DNS handler not initialized"})
		return
	}

	response, err := dnsHandler.dnsService.UpdateDNSRecord(id)
	if err != nil {
		// If response is nil, create a default error response
		if response == nil {
			response = &models.DNSUpdateResponse{
				Success: false,
				Message: fmt.Sprintf("DNS update failed: %v", err),
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "response": response})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": response})
}

// UpdateAllDNSRecords updates all active DNS records
func UpdateAllDNSRecords(c *gin.Context) {
	responses, err := dnsHandler.dnsService.UpdateAllDNSRecords()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"responses": responses})
}

// GetDNSStatus returns the current status of all DNS configurations
func GetDNSStatus(c *gin.Context) {
	statuses, err := dnsHandler.dnsService.GetDNSStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"statuses": statuses})
}

// GetPublicIP returns the current public IP address
func GetPublicIP(c *gin.Context) {
	ip, err := dnsHandler.dnsService.GetPublicIP()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ip": ip})
}

// GetScheduledJobs returns information about active scheduled jobs
func GetScheduledJobs(c *gin.Context) {
	if dnsHandler.schedulerService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Scheduler service not available"})
		return
	}

	activeJobs := dnsHandler.schedulerService.GetActiveJobs()
	c.JSON(http.StatusOK, gin.H{"active_jobs": activeJobs})
}

// PauseScheduledJob pauses a scheduled job
func PauseScheduledJob(c *gin.Context) {
	if dnsHandler.schedulerService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Scheduler service not available"})
		return
	}

	recordIDStr := c.Param("recordId")
	recordID, err := strconv.Atoi(recordIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid record ID"})
		return
	}

	err = dnsHandler.schedulerService.PauseScheduledJob(recordID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job paused successfully"})
}

// ResumeScheduledJob resumes a paused scheduled job
func ResumeScheduledJob(c *gin.Context) {
	if dnsHandler.schedulerService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Scheduler service not available"})
		return
	}

	recordIDStr := c.Param("recordId")
	recordID, err := strconv.Atoi(recordIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid record ID"})
		return
	}

	err = dnsHandler.schedulerService.ResumeScheduledJob(recordID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job resumed successfully"})
}
