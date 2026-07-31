package models

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// domainRegex matches a valid DNS hostname: labels of alphanumerics/hyphens
// (no leading/trailing hyphen), separated by dots. This is intentionally
// strict since Domain values are rendered directly into nginx config
// (server_name), and any character outside this set could be used to break
// out of the directive.
var domainRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

// ValidateDomain ensures a domain is a well-formed hostname safe to embed in
// an nginx server_name directive.
func ValidateDomain(domain string) error {
	if domain == "" {
		return fmt.Errorf("domain is required")
	}
	if len(domain) > 253 {
		return fmt.Errorf("domain is too long")
	}
	if !domainRegex.MatchString(domain) {
		return fmt.Errorf("invalid domain: must be a valid hostname (letters, digits, hyphens, dots only)")
	}
	return nil
}

// ValidateBackendURL ensures a URL is well-formed (http/https, valid host,
// no embedded whitespace/control characters) before it is rendered directly
// into an nginx proxy_pass directive.
func ValidateBackendURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("URL is required")
	}
	for _, r := range rawURL {
		if r <= 0x1f || r == 0x7f {
			return fmt.Errorf("invalid URL: control characters are not allowed")
		}
	}
	if strings.ContainsAny(rawURL, " \t\"'{};") {
		return fmt.Errorf("invalid URL: contains disallowed characters")
	}

	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("invalid URL: scheme must be http or https")
	}
	if parsed.Host == "" {
		return fmt.Errorf("invalid URL: host is required")
	}
	if parsed.Hostname() == "" {
		return fmt.Errorf("invalid URL: host is required")
	}

	return nil
}

// ValidateRateLimitRPS ensures a requests-per-second value is a sane bound
// for an nginx limit_req_zone rate directive.
func ValidateRateLimitRPS(rps int) error {
	if rps < 1 || rps > 10000 {
		return fmt.Errorf("rate_limit_rps must be between 1 and 10000")
	}
	return nil
}
