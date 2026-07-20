package certprofile

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBuildMobileconfigFromPEM(t *testing.T) {
	cert := filepath.Join(t.TempDir(), "cacert.pem")
	content := []byte("-----BEGIN CERTIFICATE-----\nAQIDBAU=\n-----END CERTIFICATE-----\n")
	if err := os.WriteFile(cert, content, 0o644); err != nil {
		t.Fatal(err)
	}

	profile, err := Build(Options{CertPath: cert, Name: "Burp Suite CA"})
	if err != nil {
		t.Fatal(err)
	}
	text := string(profile)
	for _, want := range []string{
		"com.apple.security.root",
		"Burp Suite CA",
		"AQIDBAU=",
		"authorized iOS security testing only",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("profile missing %q:\n%s", want, text)
		}
	}
}

func TestWriteMobileconfig(t *testing.T) {
	dir := t.TempDir()
	cert := filepath.Join(dir, "cacert.der")
	out := filepath.Join(dir, "burp.mobileconfig")
	if err := os.WriteFile(cert, []byte{1, 2, 3}, 0o644); err != nil {
		t.Fatal(err)
	}

	path, err := Write(Options{CertPath: cert, OutPath: out})
	if err != nil {
		t.Fatal(err)
	}
	if path != out {
		t.Fatalf("Write() path = %q, want %q", path, out)
	}
	written, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(written, []byte("burp-ca")) {
		t.Fatalf("unexpected profile content:\n%s", string(written))
	}
}

func TestVerifyMatchesCertificateInProfile(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.CreateCertificate(rand.Reader, &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Burp Suite CA"},
		NotBefore:             time.Now().Add(-time.Minute),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
	}, &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Burp Suite CA"},
		NotBefore:             time.Now().Add(-time.Minute),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
	}, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	certPath := filepath.Join(dir, "burp.der")
	profilePath := filepath.Join(dir, "burp.mobileconfig")
	if err := os.WriteFile(certPath, der, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Write(Options{CertPath: certPath, OutPath: profilePath}); err != nil {
		t.Fatal(err)
	}
	report, err := Verify(certPath, profilePath)
	if err != nil {
		t.Fatal(err)
	}
	if !report.ProfileMatches || report.Fingerprint == "" || report.Subject == "" {
		t.Fatalf("unexpected verification report: %#v", report)
	}
}

func TestTrustInstructions(t *testing.T) {
	var out bytes.Buffer
	TrustInstructions(&out)
	if !strings.Contains(out.String(), "Did you enable full trust") ||
		!strings.Contains(out.String(), "Certificate Trust Settings") ||
		!strings.Contains(out.String(), "enable the toggle for: Burp Suite CA") ||
		!strings.Contains(out.String(), "does not bypass") {
		t.Fatalf("unexpected trust instructions:\n%s", out.String())
	}
}

func TestInstallReportsMissingInstaller(t *testing.T) {
	t.Setenv("PATH", "")
	oldGOOS := runtimeGOOS
	runtimeGOOS = func() string { return "linux" }
	t.Cleanup(func() { runtimeGOOS = oldGOOS })

	var stdout, stderr bytes.Buffer
	err := Install("burp-ca.mobileconfig", &stdout, &stderr)
	if err == nil || !IsMissingInstaller(err) {
		t.Fatalf("Install() err = %v, want missing installer", err)
	}

	InstallerHints(&stdout, "burp-ca.mobileconfig")
	text := stdout.String()
	for _, want := range []string{
		"no supported iOS profile installer",
		"python3 -m pip install pymobiledevice3",
		"Apple Configurator",
		"burp-ca.mobileconfig",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("installer hints missing %q:\n%s", want, text)
		}
	}
}

func TestFirstUsefulLineSkipsTracebackHeader(t *testing.T) {
	got := firstUsefulLine("Traceback (most recent call last):\nPermissionError: [Errno 1] Operation not permitted\n")
	if got != "PermissionError: [Errno 1] Operation not permitted" {
		t.Fatalf("firstUsefulLine() = %q", got)
	}
}
