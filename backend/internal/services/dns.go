package services

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf16"
	"unicode/utf8"

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
	Language string   `xml:"Language"`
	Done     bool     `xml:"Done"`
}

type Error struct {
	Number string `xml:"Number,attr"`
	Text   string `xml:",chardata"`
}

// utf16CharsetReader is a charset reader for UTF-16 encoded XML
func utf16CharsetReader(charset string, input io.Reader) (io.Reader, error) {
	if strings.ToLower(charset) == "utf-16" {
		// Read the entire input
		data, err := io.ReadAll(input)
		if err != nil {
			return nil, err
		}

		// Check if the content is actually UTF-8 despite the UTF-16 declaration
		// This is a common issue with some APIs that declare UTF-16 but send UTF-8
		if isUTF8Content(data) {
			// Content is actually UTF-8, return as-is
			return strings.NewReader(string(data)), nil
		}

		// Convert UTF-16 to UTF-8
		utf8Data := convertUTF16ToUTF8(data)
		return strings.NewReader(string(utf8Data)), nil
	}

	return input, nil
}

// isUTF8Content checks if the data is actually UTF-8 encoded
func isUTF8Content(data []byte) bool {
	// Check for UTF-8 BOM
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return true
	}

	// Check if the content is valid UTF-8 by trying to decode it
	// If it's valid UTF-8, it should decode without errors
	return utf8.Valid(data)
}

// convertUTF16ToUTF8 converts UTF-16 encoded bytes to UTF-8
func convertUTF16ToUTF8(data []byte) []byte {
	if len(data) < 2 {
		return data
	}

	// Check for BOM (Byte Order Mark)
	var isLittleEndian bool
	if data[0] == 0xFF && data[1] == 0xFE {
		// Little endian BOM
		isLittleEndian = true
		data = data[2:]
	} else if data[0] == 0xFE && data[1] == 0xFF {
		// Big endian BOM
		isLittleEndian = false
		data = data[2:]
	} else {
		// No BOM, assume little endian (common for Windows)
		isLittleEndian = true
	}

	// Convert to UTF-16 code units
	var utf16CodeUnits []uint16
	if isLittleEndian {
		for i := 0; i < len(data); i += 2 {
			if i+1 < len(data) {
				utf16CodeUnits = append(utf16CodeUnits, uint16(data[i])|uint16(data[i+1])<<8)
			}
		}
	} else {
		for i := 0; i < len(data); i += 2 {
			if i+1 < len(data) {
				utf16CodeUnits = append(utf16CodeUnits, uint16(data[i])<<8|uint16(data[i+1]))
			}
		}
	}

	// Convert UTF-16 to UTF-8
	utf8Data := make([]byte, 0, len(utf16CodeUnits)*3)
	for _, r := range utf16.Decode(utf16CodeUnits) {
		utf8Data = utf8.AppendRune(utf8Data, r)
	}

	return utf8Data
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

// GetPublicIP retrieves the current public IP address
func (d *DNSService) GetPublicIP() (string, error) {
	resp, err := d.client.Get(d.config.PublicIPService)
	if err != nil {
		return "", fmt.Errorf("failed to get public IP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get public IP: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	ip := string(body)
	if ip == "" {
		return "", fmt.Errorf("empty IP address received")
	}

	return ip, nil
}

// UpdateNamecheapDNS updates a DNS record using Namecheap's dynamic DNS API
func (d *DNSService) UpdateNamecheapDNS(config *models.DNSConfig, record *models.DNSRecord, newIP string) (*models.DNSUpdateResponse, error) {
	// Namecheap dynamic DNS update URL
	url := fmt.Sprintf("https://dynamicdns.park-your-domain.com/update?host=%s&domain=%s&password=%s&ip=%s",
		record.Host,
		config.Domain,
		config.Password,
		newIP,
	)

	resp, err := d.client.Get(url)
	if err != nil {
		return &models.DNSUpdateResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to update DNS: %v", err),
		}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &models.DNSUpdateResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to read response: %v", err),
		}, err
	}

	// Parse XML response with custom charset reader for UTF-16 support
	var namecheapResp NamecheapResponse
	decoder := xml.NewDecoder(strings.NewReader(string(body)))
	decoder.CharsetReader = utf16CharsetReader

	if err := decoder.Decode(&namecheapResp); err != nil {
		return &models.DNSUpdateResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to parse response: %v", err),
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
	return d.UpdateDNSRecordWithClientIP(recordID, "")
}

// UpdateDNSRecordWithClientIP updates a specific DNS record with client IP for validation
func (d *DNSService) UpdateDNSRecordWithClientIP(recordID int, clientIP string) (*models.DNSUpdateResponse, error) {
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
