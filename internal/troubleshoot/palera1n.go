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

	return findings
}
