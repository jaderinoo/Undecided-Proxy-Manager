package services

import (
	"os"
	"path/filepath"
	"testing"

	"upm-backend/internal/models"
)

func TestCertificateDiskPaths_IncludesLetsEncryptNginxCopies(t *testing.T) {
	cert := &models.Certificate{
		Domain:   "wiki.undecided.dev",
		CertPath: "/etc/letsencrypt/certs/wiki.undecided.dev.crt",
		KeyPath:  "/etc/letsencrypt/certs/wiki.undecided.dev.key",
	}
	paths := certificateDiskPaths(cert)
	want := map[string]bool{
		"/etc/letsencrypt/certs/wiki.undecided.dev.crt": true,
		"/etc/letsencrypt/certs/wiki.undecided.dev.key": true,
		"/etc/ssl/certs/wiki.undecided.dev.crt":         true,
		"/etc/ssl/certs/wiki.undecided.dev.key":         true,
	}
	if len(paths) != len(want) {
		t.Fatalf("certificateDiskPaths() = %v, want %d unique paths", paths, len(want))
	}
	for _, p := range paths {
		if !want[p] {
			t.Errorf("unexpected path %s", p)
		}
	}
}

func TestRemoveCertificateFiles_DeletesStoredPaths(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "example.crt")
	keyPath := filepath.Join(dir, "example.key")
	if err := os.WriteFile(certPath, []byte("cert"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyPath, []byte("key"), 0600); err != nil {
		t.Fatal(err)
	}

	cert := &models.Certificate{
		Domain:   "example.test",
		CertPath: certPath,
		KeyPath:  keyPath,
	}
	if err := RemoveCertificateFiles(cert); err != nil {
		t.Fatalf("RemoveCertificateFiles: %v", err)
	}
	if _, err := os.Stat(certPath); !os.IsNotExist(err) {
		t.Errorf("cert file still exists")
	}
	if _, err := os.Stat(keyPath); !os.IsNotExist(err) {
		t.Errorf("key file still exists")
	}
}

func TestRemoveCertificateFiles_MissingFilesOK(t *testing.T) {
	cert := &models.Certificate{
		Domain:   "missing.test",
		CertPath: filepath.Join(t.TempDir(), "nope.crt"),
		KeyPath:  filepath.Join(t.TempDir(), "nope.key"),
	}
	if err := RemoveCertificateFiles(cert); err != nil {
		t.Fatalf("RemoveCertificateFiles on missing files: %v", err)
	}
}
