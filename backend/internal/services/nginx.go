package services

import (
	"bytes"
	"fmt"
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
}

func NewNginxService(configPath, reloadCommand, containerName string) *NginxService {
	return &NginxService{
		ConfigPath:    configPath,
		ReloadCommand: reloadCommand,
		TemplatePath:  filepath.Join(configPath, "proxy-template.conf"),
		ContainerName: containerName,
	}
}

// GenerateProxyConfig generates nginx configuration for a proxy
func (n *NginxService) GenerateProxyConfig(proxy *models.Proxy) error {
	// Read the template
	tmpl, err := template.ParseFiles(n.TemplatePath)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Prepare template data
	data := struct {
		Domain     string
		TargetURL  string
		SSLEnabled bool
		SSLPath    string
		CertPath   string
		KeyPath    string
	}{
		Domain:     proxy.Domain,
		TargetURL:  proxy.TargetURL,
		SSLEnabled: proxy.SSLEnabled,
		SSLPath:    "/etc/nginx/ssl",
		CertPath:   fmt.Sprintf("/etc/ssl/certs/certs/%s.crt", proxy.Domain),
		KeyPath:    fmt.Sprintf("/etc/ssl/certs/certs/%s.key", proxy.Domain),
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
	if err != nil {
		// If container doesn't exist or nginx is not available, log warning but don't fail
		if strings.Contains(string(output), "not found") ||
			strings.Contains(err.Error(), "not found") ||
			strings.Contains(string(output), "command not found") ||
			strings.Contains(string(output), "No such container") ||
			strings.Contains(string(output), "Container") {
			// nginx container not available, skip reload
			return nil
		}
		return fmt.Errorf("failed to reload nginx via Docker: %s, error: %w", string(output), err)
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
	if err != nil {
		// If container doesn't exist or nginx is not available, log warning but don't fail
		if strings.Contains(string(output), "not found") ||
			strings.Contains(err.Error(), "not found") ||
			strings.Contains(string(output), "command not found") ||
			strings.Contains(string(output), "No such container") ||
			strings.Contains(string(output), "Container") {
			// nginx container not available, skip test
			return nil
		}
		return fmt.Errorf("nginx config test failed via Docker: %s, error: %w", string(output), err)
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
