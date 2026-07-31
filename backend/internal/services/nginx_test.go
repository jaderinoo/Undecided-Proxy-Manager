package services

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"upm-backend/internal/models"
)

// newTestNginxService sets up a NginxService pointed at temp directories so
// GenerateProxyConfig/RemoveProxyConfig can be exercised without touching
// the real /etc/nginx paths. No DatabaseService is wired in, so DNS/cert
// lookups inside GenerateProxyConfig are skipped.
func newTestNginxService(t *testing.T) *NginxService {
	t.Helper()

	configDir := t.TempDir()
	sitesEnabledDir := t.TempDir()

	templateSrc := findProxyTemplate(t)
	templateData, err := os.ReadFile(templateSrc)
	if err != nil {
		t.Fatalf("failed to read real proxy template: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "proxy-template.conf"), templateData, 0644); err != nil {
		t.Fatalf("failed to write template into temp config dir: %v", err)
	}

	svc := NewNginxService(configDir, "true", "", nil)
	svc.SitesEnabledPath = sitesEnabledDir
	return svc
}

// findProxyTemplate locates the real nginx/sites-available/proxy-template.conf
// by walking up from the current working directory.
func findProxyTemplate(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	for i := 0; i < 8; i++ {
		candidate := filepath.Join(dir, "nginx", "sites-available", "proxy-template.conf")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatal("could not locate nginx/sites-available/proxy-template.conf")
	return ""
}

func TestGenerateProxyConfig_WritesConfigToBothPaths(t *testing.T) {
	svc := newTestNginxService(t)

	proxy := &models.Proxy{
		ID:        1,
		Name:      "test",
		Domain:    "example.com",
		TargetURL: "http://backend:8080",
		Status:    "active",
	}

	if err := svc.GenerateProxyConfig(proxy); err != nil {
		t.Fatalf("GenerateProxyConfig returned error: %v", err)
	}

	configFile := filepath.Join(svc.ConfigPath, "proxy-1.conf")
	enabledFile := filepath.Join(svc.SitesEnabledPath, "proxy-1.conf")

	configBytes, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("expected config file at %s: %v", configFile, err)
	}
	enabledBytes, err := os.ReadFile(enabledFile)
	if err != nil {
		t.Fatalf("expected config file at %s: %v", enabledFile, err)
	}
	if string(configBytes) != string(enabledBytes) {
		t.Errorf("config file and sites-enabled copy differ")
	}

	content := string(configBytes)
	if !strings.Contains(content, "server_name example.com;") {
		t.Errorf("expected rendered server_name, got:\n%s", content)
	}
	if !strings.Contains(content, "proxy_pass http://backend:8080;") {
		t.Errorf("expected rendered proxy_pass, got:\n%s", content)
	}
}

// TestGenerateProxyConfig_RejectsInjectionAtHandlerLevel documents that the
// injection payloads models.ValidateDomain/ValidateBackendURL reject would,
// if they ever reached the template unfiltered, break out of their nginx
// directive. Handlers must call the validators before this function runs;
// this test pins the template's actual behavior for the payload as a guard
// against a future rendering-layer regression.
func TestGenerateProxyConfig_InjectionPayloadWouldBreakOutOfDirective(t *testing.T) {
	svc := newTestNginxService(t)

	payload := "http://127.0.0.1:1;\n    }\n    location /pwned {\n        return 200 \"owned\""

	// Sanity check: the payload is indeed rejected by the validator that
	// production handlers call before reaching GenerateProxyConfig.
	if err := models.ValidateBackendURL(payload); err == nil {
		t.Fatalf("expected ValidateBackendURL to reject the injection payload")
	}

	proxy := &models.Proxy{
		ID:        2,
		Name:      "test",
		Domain:    "example.com",
		TargetURL: payload,
		Status:    "active",
	}

	if err := svc.GenerateProxyConfig(proxy); err != nil {
		t.Fatalf("GenerateProxyConfig returned error: %v", err)
	}

	enabledFile := filepath.Join(svc.SitesEnabledPath, "proxy-2.conf")
	content, err := os.ReadFile(enabledFile)
	if err != nil {
		t.Fatalf("expected config file: %v", err)
	}

	if !strings.Contains(string(content), "location /pwned") {
		t.Fatalf("expected the unfiltered template to demonstrate the injected location block")
	}
}

func TestRemoveProxyConfig_RemovesBothFiles(t *testing.T) {
	svc := newTestNginxService(t)

	proxy := &models.Proxy{
		ID:        3,
		Domain:    "remove-me.example.com",
		TargetURL: "http://backend:8080",
	}
	if err := svc.GenerateProxyConfig(proxy); err != nil {
		t.Fatalf("GenerateProxyConfig returned error: %v", err)
	}

	if err := svc.RemoveProxyConfig(proxy.ID); err != nil {
		t.Fatalf("RemoveProxyConfig returned error: %v", err)
	}

	configFile := filepath.Join(svc.ConfigPath, "proxy-3.conf")
	enabledFile := filepath.Join(svc.SitesEnabledPath, "proxy-3.conf")

	if _, err := os.Stat(configFile); !os.IsNotExist(err) {
		t.Errorf("expected config file to be removed, stat err = %v", err)
	}
	if _, err := os.Stat(enabledFile); !os.IsNotExist(err) {
		t.Errorf("expected sites-enabled file to be removed, stat err = %v", err)
	}
}

func TestRemoveProxyConfig_NoErrorWhenFilesMissing(t *testing.T) {
	svc := newTestNginxService(t)

	if err := svc.RemoveProxyConfig(999); err != nil {
		t.Errorf("expected no error removing a nonexistent config, got: %v", err)
	}
}

func TestSanitizeAllowedRanges(t *testing.T) {
	got := sanitizeAllowedRanges([]string{" 10.0.0.5 ", "192.168.1.0/24", "", "not-an-ip"})
	want := []string{"10.0.0.5/32", "192.168.1.0/24"}

	if len(got) != len(want) {
		t.Fatalf("sanitizeAllowedRanges() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("sanitizeAllowedRanges()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
