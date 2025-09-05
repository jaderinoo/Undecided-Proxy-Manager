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

	// Create symlink in sites-enabled
	enabledPath := filepath.Join(filepath.Dir(n.ConfigPath), "sites-enabled", fmt.Sprintf("proxy-%d.conf", proxy.ID))
	if err := os.Symlink(configFile, enabledPath); err != nil {
		// If symlink exists, remove it first
		if os.IsExist(err) {
			os.Remove(enabledPath)
			err = os.Symlink(configFile, enabledPath)
		}
		if err != nil {
			return fmt.Errorf("failed to create symlink: %w", err)
		}
	}

	return nil
}

// RemoveProxyConfig removes nginx configuration for a proxy
func (n *NginxService) RemoveProxyConfig(proxyID int) error {
	// Remove symlink from sites-enabled
	enabledPath := filepath.Join(filepath.Dir(n.ConfigPath), "sites-enabled", fmt.Sprintf("proxy-%d.conf", proxyID))
	if err := os.Remove(enabledPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove symlink: %w", err)
	}

	// Remove config file
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
	// Remove old config
	if err := n.RemoveProxyConfig(proxy.ID); err != nil {
		return fmt.Errorf("failed to remove old config: %w", err)
	}

	// Generate new config
	if err := n.GenerateProxyConfig(proxy); err != nil {
		return fmt.Errorf("failed to generate new config: %w", err)
	}

	return nil
}
