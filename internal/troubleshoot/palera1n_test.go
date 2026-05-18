package troubleshoot

import "testing"

func TestParsePalera1nLog(t *testing.T) {
	findings := ParsePalera1nLog(`Timed out waiting for download mode
USBDeviceOpen exclusive access failed`)

	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}
	if findings[0].Title == "" || len(findings[0].Suggestions) == 0 {
		t.Fatalf("expected populated finding: %#v", findings[0])
	}
}
