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

func TestParsePalera1nRootfulBindFS(t *testing.T) {
	findings := ParsePalera1nLog(`Please wait up to 5 minutes for the bindfs to be created.
Once the device reboots into recovery mode, run again without the -B option to jailbreak.`)

	if len(findings) == 0 {
		t.Fatal("expected rootful BindFS finding")
	}
}

func TestParsePalera1nPasscode(t *testing.T) {
	findings := ParsePalera1nLog(`Product iPhone10,5 requires passcode to be disabled`)
	if len(findings) == 0 {
		t.Fatal("expected passcode finding")
	}
}
