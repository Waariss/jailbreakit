package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Waariss/jailbreakit/internal/device"
	"github.com/Waariss/jailbreakit/internal/readiness"
)

func TestInstallAcceptsFlagsAfterIPA(t *testing.T) {
	dir := t.TempDir()
	ipa := filepath.Join(dir, "App.ipa")
	if err := os.WriteFile(ipa, []byte("ipa"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	err := installIPA([]string{ipa, "--host", "192.168.1.23", "--port", "44", "--installer", "ipainstaller", "--dry-run"}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	if !strings.Contains(output, "scp") || !strings.Contains(output, "-P 44") || !strings.Contains(output, "ipainstaller") {
		t.Fatalf("unexpected dry-run output:\n%s", output)
	}
}

func TestPrintDeviceDoesNotPrefixBackslash(t *testing.T) {
	var stdout bytes.Buffer
	printDevice(&stdout, device.Info{ProductType: "iPhone10,5", OSVersion: "16.7.10"})
	output := stdout.String()
	if !strings.HasPrefix(output, "ProductType:") {
		t.Fatalf("expected ProductType prefix, got %q", output)
	}
	if strings.Contains(output, `\ProductType`) {
		t.Fatalf("unexpected backslash in output: %q", output)
	}
}

func TestParseLabCheckFlags(t *testing.T) {
	got, err := parseLabCheckArgs([]string{"--ssh-host", "127.0.0.1", "--ssh-port", "2222", "--ssh-user", "root"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Host != "127.0.0.1" || got.Port != 2222 || got.User != "root" {
		t.Fatalf("unexpected SSH options: %#v", got)
	}
}

func TestParseLabCheckRejectsInvalidPort(t *testing.T) {
	_, err := parseLabCheckArgs([]string{"--ssh-host", "127.0.0.1", "--ssh-port", "70000"})
	if err == nil {
		t.Fatal("expected invalid port error")
	}
}

func TestDispatchFridaCheck(t *testing.T) {
	var stdout bytes.Buffer
	if err := run([]string{"frida-check"}, strings.NewReader(""), &stdout, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "Frida readiness") || !strings.Contains(stdout.String(), "Device Frida: skipped") {
		t.Fatalf("unexpected frida-check output:\n%s", stdout.String())
	}
}

func TestDispatchFridaCheckDevice(t *testing.T) {
	checker := readiness.Checker{
		Lookup: func(name string) (string, error) { return "/bin/" + name, nil },
		Run: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			return []byte("SpringBoard\n"), nil
		},
	}
	var stdout bytes.Buffer
	if err := fridaCheckWithChecker([]string{"--device"}, &stdout, checker); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "Device Frida: ready") {
		t.Fatalf("unexpected output:\n%s", stdout.String())
	}
}

func TestEvidenceRejectsUnknownFormat(t *testing.T) {
	err := evidenceReport([]string{"--format", "xml"}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "unsupported evidence format") {
		t.Fatalf("expected unsupported format error, got %v", err)
	}
}

func TestBurpCAGeneratesProfile(t *testing.T) {
	dir := t.TempDir()
	cert := filepath.Join(dir, "cacert.der")
	out := filepath.Join(dir, "burp.mobileconfig")
	if err := os.WriteFile(cert, []byte{1, 2, 3}, 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	if err := burpCA([]string{"--cert", cert, "--out", out}, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "Certificate Trust Settings") {
		t.Fatalf("expected trust instructions, got:\n%s", stdout.String())
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatal(err)
	}
}

func TestBurpCAInstallWithoutInstallerPrintsHints(t *testing.T) {
	t.Setenv("PATH", "")
	dir := t.TempDir()
	cert := filepath.Join(dir, "cacert.der")
	out := filepath.Join(dir, "burp.mobileconfig")
	if err := os.WriteFile(cert, []byte{1, 2, 3}, 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	if err := burpCA([]string{"--cert", cert, "--out", out, "--install"}, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	for _, want := range []string{
		"no supported iOS profile installer",
		"python3 -m pip install pymobiledevice3",
		"Certificate Trust Settings",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected %q in output:\n%s", want, output)
		}
	}
}
