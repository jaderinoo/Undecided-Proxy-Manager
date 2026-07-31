package models

import "testing"

func TestValidateDomain(t *testing.T) {
	valid := []string{
		"example.com",
		"sub.example.com",
		"a.b.c.example.co.uk",
		"localhost",
		"my-app.example.com",
		"xn--fsq.com",
	}
	for _, d := range valid {
		if err := ValidateDomain(d); err != nil {
			t.Errorf("ValidateDomain(%q) = %v, want nil", d, err)
		}
	}

	invalid := []string{
		"",
		"-example.com",
		"example-.com",
		"example..com",
		".example.com",
		"example.com.",
		"exa mple.com",
		"example.com;",
		"example.com\n",
		"http://example.com",
		"example.com/path",
		"example.com{",
		"example.com}",
		"exa\"mple.com",
	}
	for _, d := range invalid {
		if err := ValidateDomain(d); err == nil {
			t.Errorf("ValidateDomain(%q) = nil, want error", d)
		}
	}
}

func TestValidateDomainLength(t *testing.T) {
	long := ""
	for i := 0; i < 260; i++ {
		long += "a"
	}
	if err := ValidateDomain(long); err == nil {
		t.Errorf("ValidateDomain(long domain) = nil, want error")
	}
}

func TestValidateBackendURL(t *testing.T) {
	valid := []string{
		"http://backend:8080",
		"https://backend.internal",
		"http://127.0.0.1:9000",
		"https://example.com/api",
		"http://backend:8080/api/v1",
	}
	for _, u := range valid {
		if err := ValidateBackendURL(u); err != nil {
			t.Errorf("ValidateBackendURL(%q) = %v, want nil", u, err)
		}
	}

	invalid := []string{
		"",
		"ftp://backend:8080",
		"backend:8080",
		"//backend:8080",
		"http://",
		"javascript:alert(1)",
		"http://127.0.0.1:1;\n    }\n    location /steal {",
		"http://backend:8080/api;\n}\nserver{",
		"http://backend:8080 extra",
		"http://backend:8080\t",
		"http://backend\"8080",
		"http://backend'8080",
		"http://backend{8080",
		"http://backend}8080",
		"http://backend;8080",
	}
	for _, u := range invalid {
		if err := ValidateBackendURL(u); err == nil {
			t.Errorf("ValidateBackendURL(%q) = nil, want error", u)
		}
	}
}
