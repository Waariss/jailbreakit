package palera1n

import "testing"

func TestAnalyzeOutputDetectsRootfulNextStep(t *testing.T) {
	result := AnalyzeOutput(`Please wait up to 5 minutes for the bindfs to be created.
Once the device reboots into recovery mode, run again without the -B (Create BindFS) option to jailbreak.`)

	if !result.NeedsRootfulBootStep {
		t.Fatal("expected rootful boot step detection")
	}
}

func TestAnalyzeOutputDoesNotTreatRecoveredUSBWarningAsFatal(t *testing.T) {
	result := AnalyzeOutput(`USBDeviceOpen: another process has device opened for exclusive access
Checkmate!
Booting Kernel...`)

	if !result.USBExclusiveWarning {
		t.Fatal("expected USB warning detection")
	}
	if !result.CompletedCheckm8 {
		t.Fatal("expected completed checkm8 detection")
	}
}
