package evidence

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Waariss/jailbreakit/internal/device"
	"github.com/Waariss/jailbreakit/internal/readiness"
)

func testChecker() readiness.Checker {
	return readiness.Checker{
		Lookup: func(name string) (string, error) {
			switch name {
			case "ideviceinfo", "iproxy", "ssh", "scp", "ideviceinstaller", "curl", "frida", "frida-ps", "objection":
				return "/bin/" + name, nil
			default:
				return "", errors.New("missing")
			}
		},
		Run: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			return []byte(name + " 1.0\n"), nil
		},
		Detect: func() (device.Info, error) {
			return device.Enrich(device.Info{ProductType: "iPhone8,1", OSVersion: "15.8.8"}), nil
		},
	}
}

func TestMarkdownRendering(t *testing.T) {
	report := Build(testChecker(), readiness.SSHOptions{}, time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC))
	out := string(Markdown(report))
	for _, want := range []string{
		"# jailbreakit Lab Readiness Evidence",
		"Generated for authorized iOS security testing only.",
		"## Frida / Objection Readiness",
		"iPhone 6s",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("markdown missing %q:\n%s", want, out)
		}
	}
}

func TestJSONRendering(t *testing.T) {
	report := Build(testChecker(), readiness.SSHOptions{}, time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC))
	out, err := JSON(report)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Report
	if err := json.Unmarshal(out, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Tool != "jailbreakit" || decoded.SafetyNote != SafetyNote || decoded.Host.OS == "" {
		t.Fatalf("unexpected decoded report: %#v", decoded)
	}
	if decoded.Device["product_type"] != "iPhone8,1" {
		t.Fatalf("expected device info in JSON, got %#v", decoded.Device)
	}
}
