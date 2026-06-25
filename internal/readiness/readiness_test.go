package readiness

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Waariss/jailbreakit/internal/device"
)

func TestToolUsesLookupAndVersionRunner(t *testing.T) {
	checker := Checker{
		Lookup: func(name string) (string, error) {
			if name != "frida" {
				t.Fatalf("unexpected lookup %q", name)
			}
			return "/usr/local/bin/frida", nil
		},
		Run: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			if name != "frida" || len(args) != 1 || args[0] != "--version" {
				t.Fatalf("unexpected run %s %#v", name, args)
			}
			return []byte("16.1.0\n"), nil
		},
	}

	got := checker.Tool("Frida CLI", "frida", false, "--version")
	if !got.Available || got.Path != "/usr/local/bin/frida" || got.Version != "16.1.0" {
		t.Fatalf("unexpected tool check: %#v", got)
	}
}

func TestToolMissing(t *testing.T) {
	checker := Checker{Lookup: func(name string) (string, error) {
		return "", errors.New("not found")
	}}
	got := checker.Tool("Frida CLI", "frida", false)
	if got.Available || got.Error != "missing" {
		t.Fatalf("expected missing tool, got %#v", got)
	}
}

func TestSSHSkippedWithoutHost(t *testing.T) {
	got := Checker{}.SSH(SSHOptions{})
	if got.Checked || !strings.Contains(got.Hint, "--ssh-host") {
		t.Fatalf("expected skipped SSH check, got %#v", got)
	}
}

func TestSSHUsesSafeBatchModeCommand(t *testing.T) {
	checker := Checker{
		Run: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			joined := strings.Join(args, " ")
			if name != "ssh" || !strings.Contains(joined, "BatchMode=yes") || !strings.Contains(joined, "echo jailbreakit-ssh-ok") {
				t.Fatalf("unexpected SSH command %s %#v", name, args)
			}
			return []byte("jailbreakit-ssh-ok\n"), nil
		},
	}
	got := checker.SSH(SSHOptions{Host: "127.0.0.1", Port: 2222, User: "root"})
	if !got.Checked || !got.OK {
		t.Fatalf("expected successful SSH check, got %#v", got)
	}
}

func TestLabDoesNotRequireRealDevice(t *testing.T) {
	checker := Checker{
		Lookup: func(name string) (string, error) { return "", errors.New("missing") },
		Detect: func() (device.Info, error) {
			return device.Info{}, errors.New("no device")
		},
	}
	got := checker.Lab(SSHOptions{})
	if got.Device.Detected || got.SuggestedTunnel == "" || len(got.HostDependencies) == 0 {
		t.Fatalf("unexpected lab report: %#v", got)
	}
}

func TestPrintLabDoesNotDuplicateBinaryName(t *testing.T) {
	report := LabCheck{
		HostDependencies: []ToolCheck{{Name: "Host dependency ideviceinfo", Binary: "ideviceinfo", Available: true, Path: "/bin/ideviceinfo"}},
		IPAInstall:       []ToolCheck{{Name: "IPA install ssh", Binary: "ssh", Available: true, Path: "/usr/bin/ssh"}},
		Frida: FridaCheck{
			Frida:     ToolCheck{Name: "Frida CLI", Binary: "frida", Available: true, Path: "/bin/frida", Version: "17.0.0"},
			FridaPS:   ToolCheck{Name: "Frida process lister", Binary: "frida-ps", Available: true, Path: "/bin/frida-ps", Version: "17.0.0"},
			Objection: ToolCheck{Name: "Objection", Binary: "objection", Optional: true},
		},
		SSH:             SSHCheck{Hint: "skip"},
		SuggestedTunnel: "iproxy 2222 22",
	}
	var out bytes.Buffer
	PrintLab(&out, report)
	text := out.String()
	if strings.Contains(text, "ideviceinfo ideviceinfo") || strings.Contains(text, "ssh ssh") || strings.Contains(text, "frida frida") {
		t.Fatalf("output duplicated binary name:\n%s", text)
	}
	if !strings.Contains(text, "[+] Host dependency ideviceinfo: /bin/ideviceinfo") {
		t.Fatalf("missing cleaned dependency line:\n%s", text)
	}
}
