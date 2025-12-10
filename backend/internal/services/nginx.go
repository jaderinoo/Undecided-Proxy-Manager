package services

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"upm-backend/internal/models"
)

type NginxService struct {
	ConfigPath      string
	ReloadCommand   string
	TemplatePath    string
	ContainerName   string
	DatabaseService *DatabaseService
}

func NewNginxService(configPath, reloadCommand, containerName string, dbService *DatabaseService) *NginxService {
	return &NginxService{
		ConfigPath:      configPath,
		ReloadCommand:   reloadCommand,
		TemplatePath:    filepath.Join(configPath, "proxy-template.conf"),
		ContainerName:   containerName,
		DatabaseService: dbService,
	}
}

// GenerateProxyConfig generates nginx configuration for a proxy
func (n *NginxService) GenerateProxyConfig(proxy *models.Proxy) error {
	// Read the template
	tmpl, err := template.ParseFiles(n.TemplatePath)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Get allowed IP ranges and backend configuration from DNS record
	var allowedRanges []string
	var includeBackend bool
	var backendURL string
	if n.DatabaseService != nil {
		dnsRecord, err := n.DatabaseService.GetDNSRecordByDomain(proxy.Domain)
		if err != nil {
			return fmt.Errorf("failed to get DNS record for domain %s: %w", proxy.Domain, err)
		}
		if dnsRecord != nil {
			// Parse comma-separated IP ranges
			if dnsRecord.AllowedIPRanges != "" {
				ranges := strings.Split(dnsRecord.AllowedIPRanges, ",")
				for _, r := range ranges {
					r = strings.TrimSpace(r)
					if r != "" {
						allowedRanges = append(allowedRanges, r)
					}
				}
			}
			// Get backend configuration
			includeBackend = dnsRecord.IncludeBackend
			backendURL = dnsRecord.BackendURL
		}
	}

	// Prepare template data
	// Prefer cert paths from DB certificate if present
	certPath := fmt.Sprintf("/etc/ssl/certs/%s.crt", proxy.Domain)
	keyPath := fmt.Sprintf("/etc/ssl/certs/%s.key", proxy.Domain)
	var hasCertInDB bool
	if n.DatabaseService != nil {
		cert, err := n.DatabaseService.GetCertificateByDomain(proxy.Domain)
		if err != nil {
			fmt.Printf("Certificate lookup for %s: %v\n", proxy.Domain, err)
		} else if cert != nil {
			hasCertInDB = true
			fmt.Printf("Found certificate in DB for %s: certPath=%s, keyPath=%s\n", proxy.Domain, cert.CertPath, cert.KeyPath)
			if cert.CertPath != "" {
				certPath = cert.CertPath
			}
			if cert.KeyPath != "" {
				keyPath = cert.KeyPath
			}
		} else {
			fmt.Printf("No certificate found in DB for %s\n", proxy.Domain)
		}
	}

	sslEnabled := proxy.SSLEnabled
	fmt.Printf("Proxy %s: initial SSL state=%v, hasCertInDB=%v\n", proxy.Domain, sslEnabled, hasCertInDB)

	// If proxy not marked SSL but certificate exists in DB, check files and auto-enable if valid
	if !sslEnabled && hasCertInDB {
		certValid := isValidPEMFile(certPath, "CERTIFICATE")
		keyValid := isValidPEMFile(keyPath, "PRIVATE KEY")
		fmt.Printf("Checking cert files for %s: cert=%v (path: %s), key=%v (path: %s)\n", proxy.Domain, certValid, certPath, keyValid, keyPath)
		if certValid && keyValid {
			sslEnabled = true
			if n.DatabaseService != nil {
				proxy.SSLEnabled = true
				proxy.SSLPath = certPath
				if err := n.DatabaseService.UpdateProxy(proxy); err != nil {
					fmt.Printf("Failed to persist SSL enable for %s: %v\n", proxy.Domain, err)
				} else {
					fmt.Printf("Auto-enabled SSL for %s based on valid certificate files\n", proxy.Domain)
				}
			}
		} else {
			fmt.Printf("Certificate exists in DB for %s but files are invalid - SSL not enabled. cert=%v, key=%v\n", proxy.Domain, certValid, keyValid)
		}
	}

	// If SSL is enabled, validate files one more time and disable if invalid (to prevent nginx errors)
	if sslEnabled {
		certValid := isValidPEMFile(certPath, "CERTIFICATE")
		keyValid := isValidPEMFile(keyPath, "PRIVATE KEY")
		if !certValid || !keyValid {
			fmt.Printf("SSL disabled for %s: certificate files invalid (cert=%v, key=%v). Paths: %s, %s\n", proxy.Domain, certValid, keyValid, certPath, keyPath)
			sslEnabled = false
			// Update DB to reflect disabled state
			if n.DatabaseService != nil {
				proxy.SSLEnabled = false
				_ = n.DatabaseService.UpdateProxy(proxy)
			}
		}
	}

	sanitizedRanges := sanitizeAllowedRanges(allowedRanges)

	data := struct {
		Domain         string
		TargetURL      string
		SSLEnabled     bool
		SSLPath        string
		CertPath       string
		KeyPath        string
		AllowedRanges  []string
		IncludeBackend bool
		BackendURL     string
	}{
		Domain:         proxy.Domain,
		TargetURL:      proxy.TargetURL,
		SSLEnabled:     sslEnabled,
		SSLPath:        "/etc/nginx/ssl",
		CertPath:       certPath,
		KeyPath:        keyPath,
		AllowedRanges:  sanitizedRanges,
		IncludeBackend: includeBackend,
		BackendURL:     backendURL,
	}

	// Generate config content
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Write config file
	configFile := filepath.Join(n.ConfigPath, fmt.Sprintf("proxy-%d.conf", proxy.ID))
	if err := os.WriteFile(configFile, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Copy config file to sites-enabled directory (shared volume)
	enabledPath := filepath.Join("/etc/nginx/sites-enabled", fmt.Sprintf("proxy-%d.conf", proxy.ID))
	if err := os.WriteFile(enabledPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to copy config to sites-enabled: %w", err)
	}

	return nil
}

// RemoveProxyConfig removes nginx configuration for a proxy
func (n *NginxService) RemoveProxyConfig(proxyID int) error {
	// Remove config file from sites-enabled directory (shared volume)
	enabledPath := filepath.Join("/etc/nginx/sites-enabled", fmt.Sprintf("proxy-%d.conf", proxyID))
	if err := os.Remove(enabledPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove config from sites-enabled: %w", err)
	}

	// Remove config file from backend container
	configFile := filepath.Join(n.ConfigPath, fmt.Sprintf("proxy-%d.conf", proxyID))
	if err := os.Remove(configFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove config file: %w", err)
	}

	return nil
}

// ReloadNginx reloads nginx configuration
func (n *NginxService) ReloadNginx() error {
	// Check if we're in a Docker environment
	if n.isDockerEnvironment() {
		return n.reloadNginxDocker()
	}
	return n.reloadNginxDirect()
}

// isDockerEnvironment checks if we're running in a Docker environment
func (n *NginxService) isDockerEnvironment() bool {
	// Check if we have a container name and the reload command contains "docker exec"
	return n.ContainerName != "" && strings.Contains(n.ReloadCommand, "docker exec")
}

// reloadNginxDocker reloads nginx using Docker exec
func (n *NginxService) reloadNginxDocker() error {
	cmd := exec.Command("sh", "-c", n.ReloadCommand)
	output, err := cmd.CombinedOutput()
	
	// Log the output for debugging
	fmt.Printf("Nginx reload output: %s\n", string(output))
	
 	if err != nil {
		// If container doesn't exist or nginx is not available, log warning but don't fail
		if strings.Contains(string(output), "not found") ||
			strings.Contains(err.Error(), "not found") ||
			strings.Contains(string(output), "command not found") ||
			strings.Contains(string(output), "No such container") ||
			strings.Contains(string(output), "Container") {
			// nginx container not available, skip reload
			fmt.Printf("Nginx container not available, skipping reload: %s\n", string(output))
			return nil
		}
		return fmt.Errorf("failed to reload nginx via Docker: %s, error: %w", string(output), err)
	}
	
	// Check if there are warnings in the output
	if strings.Contains(string(output), "warn") {
		fmt.Printf("Nginx reload completed with warnings: %s\n", string(output))
	}
	
	return nil
}

// reloadNginxDirect reloads nginx directly (for non-Docker environments)
func (n *NginxService) reloadNginxDirect() error {
	cmd := exec.Command("sh", "-c", n.ReloadCommand)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If nginx is not available, log warning but don't fail
		if strings.Contains(string(output), "not found") ||
			strings.Contains(err.Error(), "not found") ||
			strings.Contains(string(output), "command not found") {
			// nginx not available, skip reload
			return nil
		}
		return fmt.Errorf("failed to reload nginx: %s, error: %w", string(output), err)
	}
	return nil
}

// TestNginxConfig tests nginx configuration
func (n *NginxService) TestNginxConfig() error {
	// Check if we're in a Docker environment
	if n.isDockerEnvironment() {
		return n.testNginxConfigDocker()
	}
	return n.testNginxConfigDirect()
}

// testNginxConfigDocker tests nginx configuration using Docker exec
func (n *NginxService) testNginxConfigDocker() error {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("docker exec %s nginx -t", n.ContainerName))
	output, err := cmd.CombinedOutput()
	
	// Log the output for debugging
	fmt.Printf("Nginx config test output: %s\n", string(output))
	
	if err != nil {
		// If container doesn't exist or nginx is not available, log warning but don't fail
		if strings.Contains(string(output), "not found") ||
			strings.Contains(err.Error(), "not found") ||
			strings.Contains(string(output), "command not found") ||
			strings.Contains(string(output), "No such container") ||
			strings.Contains(string(output), "Container") {
			// nginx container not available, skip test
			fmt.Printf("Nginx container not available, skipping config test: %s\n", string(output))
			return nil
		}
		return fmt.Errorf("nginx config test failed via Docker: %s, error: %w", string(output), err)
	}
	
	// Check if there are warnings in the output
	if strings.Contains(string(output), "warn") {
		fmt.Printf("Nginx config test completed with warnings: %s\n", string(output))
	}
	
	return nil
}

// testNginxConfigDirect tests nginx configuration directly
func (n *NginxService) testNginxConfigDirect() error {
	cmd := exec.Command("nginx", "-t")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If nginx is not available, log warning but don't fail
		if strings.Contains(string(output), "executable file not found") ||
			strings.Contains(string(output), "not found") ||
			strings.Contains(err.Error(), "executable file not found") {
			// nginx not available, skip test
			return nil
		}
		return fmt.Errorf("nginx config test failed: %s, error: %w", string(output), err)
	}
	return nil
}

// sanitizeAllowedRanges normalizes CIDRs/IPs so nginx doesn't warn about host bits.
func sanitizeAllowedRanges(ranges []string) []string {
	var sanitized []string
	for _, r := range ranges {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}

		// If it parses as CIDR, normalize to network address.
		if ip, ipNet, err := net.ParseCIDR(r); err == nil {
			network := ip.Mask(ipNet.Mask)
			sanitized = append(sanitized, fmt.Sprintf("%s/%d", network.String(), ones(ipNet.Mask)))
			continue
		}

		// Fallback: if it's a bare IP, treat as /32.
		if ip := net.ParseIP(r); ip != nil {
			if ip.To4() != nil {
				sanitized = append(sanitized, fmt.Sprintf("%s/32", ip.String()))
			} else {
				sanitized = append(sanitized, fmt.Sprintf("%s/128", ip.String()))
			}
		}
	}
	return sanitized
}

// ones is a small helper to get prefix length from a mask.
func ones(mask net.IPMask) int {
	ones, _ := mask.Size()
	return ones
}

// isValidPEMFile does a light check: file exists, readable, and contains the expected BEGIN marker.
func isValidPEMFile(path string, marker string) bool {
	b, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	content := string(b)
	
	// For PRIVATE KEY, check for common variants: RSA PRIVATE KEY, EC PRIVATE KEY, or just PRIVATE KEY
	if marker == "PRIVATE KEY" {
		return strings.Contains(content, "-----BEGIN RSA PRIVATE KEY") ||
			strings.Contains(content, "-----BEGIN EC PRIVATE KEY") ||
			strings.Contains(content, "-----BEGIN PRIVATE KEY")
	}
	
	// For other markers (like CERTIFICATE), check exact match
	needle := fmt.Sprintf("-----BEGIN %s", marker)
	return strings.Contains(content, needle)
}

// UpdateProxyConfig updates nginx configuration for a proxy
func (n *NginxService) UpdateProxyConfig(proxy *models.Proxy) error {
	// Remove old config from sites-enabled directory (shared volume)
	enabledPath := filepath.Join("/etc/nginx/sites-enabled", fmt.Sprintf("proxy-%d.conf", proxy.ID))
	if err := os.Remove(enabledPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove old config from sites-enabled: %w", err)
	}

	// Generate new config (this will overwrite the existing config file and copy to sites-enabled)
	if err := n.GenerateProxyConfig(proxy); err != nil {
		return fmt.Errorf("failed to generate new config: %w", err)
	}

	// Reload nginx to apply the new configuration
	if err := n.ReloadNginx(); err != nil {
		return fmt.Errorf("failed to reload nginx: %w", err)
	}

	return nil
}

// UpdateAdminConfig updates the nginx admin configuration with IP restrictions
func (n *NginxService) UpdateAdminConfig(allowedRanges []string) error {
	templatePath := filepath.Join(n.ConfigPath, "upm-admin.conf.template")
	outputPath := filepath.Join(n.ConfigPath, "upm-admin.conf")
	
	// Read the template
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("failed to parse admin template: %w", err)
	}

	// Prepare template data
	data := struct {
		AllowedRanges []string
	}{
		AllowedRanges: allowedRanges,
	}

	// Generate the configuration
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute admin template: %w", err)
	}

	// Write to the output file
	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write admin config: %w", err)
	}

	// Use envsubst to substitute environment variables and copy to sites-enabled
	enabledPath := "/etc/nginx/sites-enabled/upm-admin.conf"
	if err := n.processConfigWithEnvSubst(outputPath, enabledPath); err != nil {
		return fmt.Errorf("failed to process config with envsubst: %w", err)
	}

	return nil
}

// copyConfigToContainer copies a local config file to the nginx container
func (n *NginxService) copyConfigToContainer(localPath, containerPath string) error {
	// Read the local file
	content, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("failed to read local config: %w", err)
	}

	// Create the file in the container using docker exec
	cmd := exec.Command("docker", "exec", "-i", n.ContainerName, "sh", "-c", fmt.Sprintf("cat > %s", containerPath))
	cmd.Stdin = bytes.NewReader(content)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy config to container: %w", err)
	}

	return nil
}

// GetAdminIPRestrictions gets the current nginx admin IP restrictions
func (n *NginxService) GetAdminIPRestrictions() ([]string, error) {
	// Read the current config from the local file
	configPath := filepath.Join(n.ConfigPath, "upm-admin.conf")
	content, err := os.ReadFile(configPath)
	if err != nil {
		// If the processed config doesn't exist, return default ranges
		return []string{"192.168.50.0/24", "10.6.0.1/32"}, nil
	}

	// Parse the config to extract allow directives
	var allowedRanges []string
	lines := strings.Split(string(content), "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "allow ") {
			// Extract the IP range after "allow "
			rangeStr := strings.TrimSpace(strings.TrimPrefix(line, "allow"))
			rangeStr = strings.TrimSuffix(rangeStr, ";")
			if rangeStr != "" {
				allowedRanges = append(allowedRanges, rangeStr)
			}
		}
	}

	return allowedRanges, nil
}

// processConfigWithEnvSubst processes a config file using envsubst to substitute environment variables
func (n *NginxService) processConfigWithEnvSubst(inputPath, outputPath string) error {
	// Read the input file
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Use envsubst to substitute only specific environment variables
	cmd := exec.Command("envsubst", "$PROD_FRONTEND_PORT $PROD_BACKEND_PORT")
	cmd.Stdin = bytes.NewReader(content)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("envsubst failed: %s, error: %w", buf.String(), err)
	}

	// Write the processed content to the output file
	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write processed config: %w", err)
	}

	return nil
}
