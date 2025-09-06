package services

import (
	"encoding/xml"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"upm-backend/internal/config"
	"upm-backend/internal/models"
)

type DNSService struct {
	DbService *DatabaseService
	config    *config.Config
	client    *http.Client
}

// Namecheap API response structures
type NamecheapResponse struct {
	XMLName  xml.Name `xml:"interface-response"`
	Command  string   `xml:"Command"`
	IP       string   `xml:"IP"`
	ErrCount int      `xml:"ErrCount"`
	Errors   []Error  `xml:"errors>Error"`
}

type Error struct {
	Number string `xml:"Number,attr"`
	Text   string `xml:",chardata"`
}

func NewDNSService(dbService *DatabaseService) *DNSService {
	return &DNSService{
		DbService: dbService,
		config:    config.Load(),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ValidateIPRanges validates a comma-separated list of IP ranges
func (d *DNSService) ValidateIPRanges(ipRanges string) error {
	if ipRanges == "" {
		return nil // Empty is valid
	}

	ranges := strings.Split(ipRanges, ",")
	for _, r := range ranges {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}

		// Check if it's a CIDR notation
		if strings.Contains(r, "/") {
			_, _, err := net.ParseCIDR(r)
			if err != nil {
				return fmt.Errorf("invalid CIDR range '%s': %w", r, err)
			}
		} else {
			// Check if it's a single IP
			ip := net.ParseIP(r)
			if ip == nil {
				return fmt.Errorf("invalid IP address '%s'", r)
			}
		}
	}

	return nil
}

// IsIPAllowed checks if an IP is within the allowed ranges
func (d *DNSService) IsIPAllowed(ip string, allowedRanges string) (bool, error) {
	if allowedRanges == "" {
		return true, nil // No restrictions
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false, fmt.Errorf("invalid IP address '%s'", ip)
	}

	ranges := strings.Split(allowedRanges, ",")
	for _, r := range ranges {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}

		// Check if it's a CIDR notation
		if strings.Contains(r, "/") {
			_, network, err := net.ParseCIDR(r)
			if err != nil {
				return false, fmt.Errorf("invalid CIDR range '%s': %w", r, err)
			}
			if network.Contains(parsedIP) {
				return true, nil
			}
		} else {
			// Check if it's a single IP
			allowedIP := net.ParseIP(r)
			if allowedIP == nil {
				return false, fmt.Errorf("invalid IP address '%s'", r)
			}
			if parsedIP.Equal(allowedIP) {
				return true, nil
			}
		}
	}

	return false, nil
}

// GetPublicIP retrieves the current public IP address using multiple fallback services
func (d *DNSService) GetPublicIP() (string, error) {
	// List of public IP services to try in order
	services := []string{
		d.config.PublicIPService, // Use configured service first
		"https://api.ipify.org",
		"https://ipv4.icanhazip.com",
		"https://api.ip.sb/ip",
		"https://checkip.amazonaws.com",
		"https://ifconfig.me/ip",
		"https://ipecho.net/plain",
	}

	var lastErr error
	for i, service := range services {
		if service == "" {
			continue // Skip empty services
		}

		ip, err := d.tryGetIPFromService(service)
		if err != nil {
			lastErr = err
			fmt.Printf("Failed to get IP from service %d (%s): %v\n", i+1, service, err)
			continue
		}

		if ip != "" {
			fmt.Printf("Successfully got public IP from service %d (%s): %s\n", i+1, service, ip)
			return ip, nil
		}
	}

	return "", fmt.Errorf("failed to get public IP from any service. Last error: %w", lastErr)
}

// tryGetIPFromService attempts to get IP from a specific service
func (d *DNSService) tryGetIPFromService(service string) (string, error) {
	resp, err := d.client.Get(service)
	if err != nil {
		return "", fmt.Errorf("failed to get public IP from %s: %w", service, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get public IP from %s: status %d", service, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body from %s: %w", service, err)
	}

	ip := string(body)
	if ip == "" {
		return "", fmt.Errorf("empty IP address received from %s", service)
	}

	// Basic validation - check if it looks like an IP address
	if len(ip) < 7 || len(ip) > 15 {
		return "", fmt.Errorf("invalid IP address format from %s: %s", service, ip)
	}

	return ip, nil
}

// UpdateNamecheapDNS updates a DNS record using Namecheap's dynamic DNS API
func (d *DNSService) UpdateNamecheapDNS(config *models.DNSConfig, record *models.DNSRecord, newIP string) (*models.DNSUpdateResponse, error) {
	// Validate required fields
	if config.Domain == "" {
		return &models.DNSUpdateResponse{
			Success: false,
			Message: "Domain is required for Namecheap DNS update",
		}, fmt.Errorf("domain is required")
	}
	if record.Host == "" {
		return &models.DNSUpdateResponse{
			Success: false,
			Message: "Host is required for Namecheap DNS update",
		}, fmt.Errorf("host is required")
	}
	if config.Password == "" {
		return &models.DNSUpdateResponse{
			Success: false,
			Message: "Password is required for Namecheap DNS update",
		}, fmt.Errorf("password is required")
	}
	if newIP == "" {
		return &models.DNSUpdateResponse{
			Success: false,
			Message: "IP address is required for Namecheap DNS update",
		}, fmt.Errorf("IP address is required")
	}

	// Namecheap dynamic DNS update URL with proper URL encoding
	baseURL := "https://dynamicdns.park-your-domain.com/update"
	params := url.Values{}
	params.Add("host", record.Host)
	params.Add("domain", config.Domain)
	params.Add("password", config.Password)
	params.Add("ip", newIP)

	requestURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := d.client.Get(requestURL)
	if err != nil {
		return &models.DNSUpdateResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to update DNS: %v", err),
		}, err
	}
	defer resp.Body.Close()

	// Log the request URL for debugging
	fmt.Printf("Namecheap API request URL: %s\n", requestURL)

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return &models.DNSUpdateResponse{
			Success: false,
			Message: fmt.Sprintf("Namecheap API returned status %d: %s", resp.StatusCode, string(body)),
		}, fmt.Errorf("Namecheap API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &models.DNSUpdateResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to read response: %v", err),
		}, err
	}

	// Log the response for debugging
	fmt.Printf("Namecheap API response (status %d): %s\n", resp.StatusCode, string(body))

	// Check if response body is empty
	if len(body) == 0 {
		return &models.DNSUpdateResponse{
			Success: false,
			Message: "Empty response from Namecheap API",
		}, fmt.Errorf("empty response from Namecheap API")
	}

	// Parse XML response with proper charset handling
	var namecheapResp NamecheapResponse

	// Parse XML response - Namecheap API declares UTF-16 but sends UTF-8
	responseBody := string(body)

	// Remove the XML declaration to avoid charset confusion
	// The API declares UTF-16 but actually sends UTF-8
	if strings.HasPrefix(responseBody, "<?xml") {
		// Find the end of the XML declaration
		endDecl := strings.Index(responseBody, "?>")
		if endDecl != -1 {
			responseBody = responseBody[endDecl+2:]
		}
	}

	// Parse as UTF-8 since that's what the content actually is
	decoder := xml.NewDecoder(strings.NewReader(responseBody))
	if err := decoder.Decode(&namecheapResp); err != nil {
		return &models.DNSUpdateResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to parse XML response: %v. Response body: %s", err, responseBody),
		}, err
	}

	// Check for errors in response
	if namecheapResp.ErrCount > 0 {
		errorMsg := "Namecheap API errors:"
		for _, err := range namecheapResp.Errors {
			errorMsg += fmt.Sprintf(" [%s] %s", err.Number, err.Text)
		}
		return &models.DNSUpdateResponse{
			Success: false,
			Message: errorMsg,
		}, fmt.Errorf(errorMsg)
	}

	// Update the record in database
	now := time.Now()
	record.CurrentIP = newIP
	record.LastUpdate = &now

	if err := d.DbService.UpdateDNSRecord(record); err != nil {
		return &models.DNSUpdateResponse{
			Success: false,
			Message: fmt.Sprintf("DNS updated but failed to save to database: %v", err),
		}, err
	}

	// Update config last update time
	config.LastUpdate = &now
	config.LastIP = newIP
	if err := d.DbService.UpdateDNSConfig(config); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: Failed to update DNS config timestamp: %v\n", err)
	}

	return &models.DNSUpdateResponse{
		Success:   true,
		Message:   "DNS record updated successfully",
		NewIP:     newIP,
		UpdatedAt: now.Format(time.RFC3339),
	}, nil
}

// UpdateDNSRecord updates a specific DNS record
func (d *DNSService) UpdateDNSRecord(recordID int) (*models.DNSUpdateResponse, error) {
	// Get the record
	record, err := d.DbService.GetDNSRecord(recordID)
	if err != nil {
		return &models.DNSUpdateResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to get DNS record: %v", err),
		}, err
	}

	// Get the config
	config, err := d.DbService.GetDNSConfig(record.ConfigID)
	if err != nil {
		return &models.DNSUpdateResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to get DNS config: %v", err),
		}, err
	}

	// Check if config is active
	if !config.IsActive {
		return &models.DNSUpdateResponse{
			Success: false,
			Message: "DNS configuration is not active",
		}, fmt.Errorf("DNS configuration is not active")
	}

	// Check if record is active
	if !record.IsActive {
		return &models.DNSUpdateResponse{
			Success: false,
			Message: "DNS record is not active",
		}, fmt.Errorf("DNS record is not active")
	}

	// Get current public IP
	currentIP, err := d.GetPublicIP()
	if err != nil {
		return &models.DNSUpdateResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to get public IP: %v", err),
		}, err
	}

	// Check if IP is allowed based on allowed ranges
	if record.AllowedIPRanges != "" {
		allowed, err := d.IsIPAllowed(currentIP, record.AllowedIPRanges)
		if err != nil {
			return &models.DNSUpdateResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to validate IP against allowed ranges: %v", err),
			}, err
		}
		if !allowed {
			return &models.DNSUpdateResponse{
				Success: false,
				Message: fmt.Sprintf("Current IP %s is not in the allowed ranges: %s", currentIP, record.AllowedIPRanges),
			}, fmt.Errorf("IP %s not allowed", currentIP)
		}
	}

	// Check if IP has changed
	if record.CurrentIP == currentIP {
		return &models.DNSUpdateResponse{
			Success: true,
			Message: "IP address has not changed, no update needed",
			NewIP:   currentIP,
		}, nil
	}

	// Update based on provider
	switch config.Provider {
	case models.ProviderNamecheap:
		return d.UpdateNamecheapDNS(config, record, currentIP)
	case models.ProviderStatic:
		// Static DNS - just update the record in database without API call
		now := time.Now()
		record.CurrentIP = currentIP
		record.LastUpdate = &now

		if err := d.DbService.UpdateDNSRecord(record); err != nil {
			return &models.DNSUpdateResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to save to database: %v", err),
			}, err
		}

		// Update config last update time
		config.LastUpdate = &now
		config.LastIP = currentIP
		if err := d.DbService.UpdateDNSConfig(config); err != nil {
			fmt.Printf("Warning: Failed to update DNS config timestamp: %v\n", err)
		}

		return &models.DNSUpdateResponse{
			Success:   true,
			Message:   "Static DNS record updated successfully",
			NewIP:     currentIP,
			UpdatedAt: now.Format(time.RFC3339),
		}, nil
	default:
		return &models.DNSUpdateResponse{
			Success: false,
			Message: fmt.Sprintf("Unsupported DNS provider: %s", config.Provider),
		}, fmt.Errorf("unsupported DNS provider: %s", config.Provider)
	}
}

// UpdateAllDNSRecords updates all active DNS records
func (d *DNSService) UpdateAllDNSRecords() ([]models.DNSUpdateResponse, error) {
	configs, err := d.DbService.GetDNSConfigs()
	if err != nil {
		return nil, fmt.Errorf("failed to get DNS configs: %w", err)
	}

	var responses []models.DNSUpdateResponse
	for _, config := range configs {
		if !config.IsActive {
			continue
		}

		records, err := d.DbService.GetDNSRecords(config.ID)
		if err != nil {
			responses = append(responses, models.DNSUpdateResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to get records for config %d: %v", config.ID, err),
			})
			continue
		}

		for _, record := range records {
			if !record.IsActive {
				continue
			}

			response, err := d.UpdateDNSRecord(record.ID)
			if err != nil {
				response = &models.DNSUpdateResponse{
					Success: false,
					Message: fmt.Sprintf("Failed to update record %d: %v", record.ID, err),
				}
			}
			responses = append(responses, *response)
		}
	}

	return responses, nil
}

// GetDNSStatus returns the current status of all DNS configurations
func (d *DNSService) GetDNSStatus() ([]models.DNSStatus, error) {
	configs, err := d.DbService.GetDNSConfigs()
	if err != nil {
		return nil, fmt.Errorf("failed to get DNS configs: %w", err)
	}

	var statuses []models.DNSStatus
	for _, config := range configs {
		records, err := d.DbService.GetDNSRecords(config.ID)
		if err != nil {
			// Log error but continue with other configs
			fmt.Printf("Warning: Failed to get records for config %d: %v\n", config.ID, err)
			records = []models.DNSRecord{}
		}

		status := models.DNSStatus{
			ConfigID:    config.ID,
			Domain:      config.Domain,
			Provider:    string(config.Provider),
			IsActive:    config.IsActive,
			LastUpdate:  config.LastUpdate,
			LastIP:      config.LastIP,
			RecordCount: len(records),
			Records:     records,
		}

		statuses = append(statuses, status)
	}

	return statuses, nil
}

// StartPeriodicUpdates starts a goroutine that periodically updates DNS records
func (d *DNSService) StartPeriodicUpdates() {
	interval, err := time.ParseDuration(d.config.DNSCheckInterval)
	if err != nil {
		fmt.Printf("Warning: Invalid DNS check interval '%s', using default 5m: %v\n", d.config.DNSCheckInterval, err)
		interval = 5 * time.Minute
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				fmt.Printf("Starting periodic DNS update check...\n")
				responses, err := d.UpdateAllDNSRecords()
				if err != nil {
					fmt.Printf("Error during periodic DNS update: %v\n", err)
					continue
				}

				// Log results
				for _, response := range responses {
					if response.Success {
						fmt.Printf("DNS update successful: %s (IP: %s)\n", response.Message, response.NewIP)
					} else {
						fmt.Printf("DNS update failed: %s\n", response.Message)
					}
				}
			}
		}
	}()

	fmt.Printf("Started periodic DNS updates every %v\n", interval)
}
