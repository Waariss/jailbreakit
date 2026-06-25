package troubleshoot

import "strings"

type Finding struct {
	Title       string
	Suggestions []string
}

func ParsePalera1nLog(log string) []Finding {
	lower := strings.ToLower(log)
	var findings []Finding

	if strings.Contains(lower, "timed out waiting for download mode") {
		findings = append(findings, Finding{
			Title: "Device did not enter download mode in time",
			Suggestions: []string{
				"Try a USB-A cable or USB 2.0 hub.",
				"Avoid flaky USB-C adapters and reconnect directly if possible.",
				"If this repeats on macOS, try palen1x/Linux for the DFU step.",
			},
		})
	}
	if strings.Contains(lower, "usbdeviceopen") || strings.Contains(lower, "exclusive access") {
		findings = append(findings, Finding{
			Title: "USB device is busy or exclusive access failed",
			Suggestions: []string{
				"Close Finder/iTunes/Xcode tools that may be holding the device.",
				"Unplug and reconnect the device, then rerun palera1n.",
				"On Linux, verify udev permissions for libusb/libimobiledevice.",
			},
		})
	}
	if strings.Contains(lower, "dfu mode") && strings.Contains(lower, "reconnected") {
		findings = append(findings, Finding{
			Title: "Device reconnected in DFU mode",
			Suggestions: []string{
				"This can be normal during checkm8 flow; continue if palera1n proceeds.",
				"If it stalls, rerun with verbose logs and use a different cable/port.",
			},
		})
	}
	if strings.Contains(lower, "requires passcode to be disabled") || strings.Contains(lower, "passcode must be disabled") {
		findings = append(findings, Finding{
			Title: "Passcode must be disabled for this device",
			Suggestions: []string{
				"Disable the passcode before running palera1n on A11-class devices.",
				"Back up the device before changing security settings.",
			},
		})
	}
	if strings.Contains(lower, "unable to connect to device") || strings.Contains(lower, "no device found") || strings.Contains(lower, "waiting for devices") {
		findings = append(findings, Finding{
			Title: "Device was not detected",
			Suggestions: []string{
				"Reconnect the cable and unlock/trust the device when iOS prompts.",
				"Check that libimobiledevice can see the device with ideviceinfo.",
				"Try another USB port or cable.",
			},
		})
	}
	if strings.Contains(lower, "failed to enter recovery") || strings.Contains(lower, "failed to enter dfu") {
		findings = append(findings, Finding{
			Title: "Failed to enter recovery or DFU mode",
			Suggestions: []string{
				"Follow the button timing exactly and retry.",
				"Use a USB-A cable or a USB 2.0 hub if USB-C timing is unreliable.",
			},
		})
	}
	if strings.Contains(lower, "pongoos") && strings.Contains(lower, "booting kernel") {
		findings = append(findings, Finding{
			Title: "PongoOS booted the kernel",
			Suggestions: []string{
				"This usually means the device boot/jailbreak flow progressed successfully.",
				"If rootful BindFS was just created, wait for recovery mode and run the rootful boot step.",
			},
		})
	}
	if strings.Contains(lower, "run again without the -b") || strings.Contains(lower, "bindfs to be created") {
		findings = append(findings, Finding{
			Title: "Rootful BindFS creation stage completed",
			Suggestions: []string{
				"Wait until the device reboots into recovery mode.",
				"Run jailbreakit run palera1n --rootful-boot.",
			},
		})
	}

	return findings
}
