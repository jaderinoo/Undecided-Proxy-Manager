package services

import (
	"os"
	"os/exec"
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

func TestGenerateProxyConfig_RateLimitEnabled_RendersLimitReqDirectives(t *testing.T) {
	svc := newTestNginxService(t)

	proxy := &models.Proxy{
		ID:               4,
		Domain:           "ratelimited.example.com",
		TargetURL:        "http://backend:8080",
		RateLimitEnabled: true,
		RateLimitRPS:     15,
	}
	if err := svc.GenerateProxyConfig(proxy); err != nil {
		t.Fatalf("GenerateProxyConfig returned error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(svc.SitesEnabledPath, "proxy-4.conf"))
	if err != nil {
		t.Fatalf("expected config file: %v", err)
	}

	if !strings.Contains(string(content), "limit_req_zone $binary_remote_addr zone=proxy_4:10m rate=15r/s;") {
		t.Errorf("expected limit_req_zone directive, got:\n%s", content)
	}
	if !strings.Contains(string(content), "limit_req zone=proxy_4 burst=30 nodelay;") {
		t.Errorf("expected limit_req directive with burst=2x rate, got:\n%s", content)
	}
}

func TestGenerateProxyConfig_RateLimitDisabled_NoLimitReqDirectives(t *testing.T) {
	svc := newTestNginxService(t)

	proxy := &models.Proxy{
		ID:               5,
		Domain:           "unlimited.example.com",
		TargetURL:        "http://backend:8080",
		RateLimitEnabled: false,
	}
	if err := svc.GenerateProxyConfig(proxy); err != nil {
		t.Fatalf("GenerateProxyConfig returned error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(svc.SitesEnabledPath, "proxy-5.conf"))
	if err != nil {
		t.Fatalf("expected config file: %v", err)
	}

	if strings.Contains(string(content), "limit_req") {
		t.Errorf("expected no limit_req directives when rate limiting is disabled, got:\n%s", content)
	}
}

// TestGenerateProxyConfig_RenderedConfigIsValidNginxSyntax renders configs
// with rate limiting on and off and, if the nginx binary is available,
// actually runs `nginx -t` against them to catch template syntax errors that
// a plain string-content test would miss.
func TestGenerateProxyConfig_RenderedConfigIsValidNginxSyntax(t *testing.T) {
	nginxPath, err := exec.LookPath("nginx")
	if err != nil {
		t.Skip("nginx binary not available, skipping syntax validation")
	}

	svc := newTestNginxService(t)

	proxies := []*models.Proxy{
		{ID: 10, Domain: "a.example.com", TargetURL: "http://127.0.0.1:8080", RateLimitEnabled: true, RateLimitRPS: 15},
		{ID: 11, Domain: "b.example.com", TargetURL: "http://127.0.0.1:9090", RateLimitEnabled: false, SSLEnabled: true},
		{ID: 12, Domain: "c.example.com", TargetURL: "http://127.0.0.1:7000", RateLimitEnabled: true, RateLimitRPS: 5, WSEnabled: true},
	}
	for _, p := range proxies {
		if err := svc.GenerateProxyConfig(p); err != nil {
			t.Fatalf("GenerateProxyConfig(%d) returned error: %v", p.ID, err)
		}
	}

	// Build a minimal, self-contained nginx.conf that includes the rendered
	// site configs, mirroring the http-context inclusion used in production.
	nginxRoot := t.TempDir()
	mainConf := `pid ` + filepath.Join(nginxRoot, "nginx.pid") + `;
error_log ` + filepath.Join(nginxRoot, "error.log") + `;
events {}
http {
    access_log ` + filepath.Join(nginxRoot, "access.log") + `;
    map $http_upgrade $connection_upgrade {
        default upgrade;
        '' close;
    }
    include ` + filepath.Join(svc.SitesEnabledPath, "*.conf") + `;
}
`
	mainConfPath := filepath.Join(nginxRoot, "nginx.conf")
	if err := os.WriteFile(mainConfPath, []byte(mainConf), 0644); err != nil {
		t.Fatalf("failed to write test nginx.conf: %v", err)
	}

	// `nginx -t` in this unprivileged test process also tries to load
	// certificates and bind to ports 80/443, which fail here for reasons
	// unrelated to the template (missing cert files, no CAP_NET_BIND_SERVICE).
	// What we actually care about is that the config *parses*, which nginx
	// reports via "syntax is ok" before attempting either of those steps.
	cmd := exec.Command(nginxPath, "-t", "-c", mainConfPath)
	output, err := cmd.CombinedOutput()
	if !strings.Contains(string(output), "syntax is ok") {
		t.Errorf("nginx -t did not report valid syntax: %v\noutput:\n%s", err, output)
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
