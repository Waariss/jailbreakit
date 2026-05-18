package recommender

import (
	"strconv"
	"strings"

	"github.com/waaris/jailbreakit/internal/device"
	"github.com/waaris/jailbreakit/internal/matrix"
)

type Result struct {
	Options  []Option
	Warnings []string
}

type Option struct {
	Name    string
	Version string
	Mode    string
	Type    string
	Reason  string
}

func Recommend(info device.Info) Result {
	var result Result
	if info.OSVersion != "" {
		entries, source := matrix.LookupIOS(info.OSVersion)
		for _, entry := range entries {
			switch strings.ToLower(entry.Tool) {
			case "palera1n":
				if checkm8Supported(info.Chip) {
					result.Options = append(result.Options, Option{
						Name:    "palera1n",
						Version: entry.Version,
						Mode:    "rootless or rootful fakefs",
						Type:    strings.ToLower(entry.Status),
						Reason:  "Matched iOS compatibility matrix from " + source + ".",
					})
				}
			case "dopamine":
				result.Options = append(result.Options, Option{
					Name:    "Dopamine",
					Version: entry.Version,
					Mode:    "rootless",
					Type:    "semi-untethered",
					Reason:  "Matched iOS compatibility matrix from " + source + ".",
				})
			}
		}
		if len(result.Options) > 0 {
			return result
		}
	}

	if checkm8Supported(info.Chip) && majorVersion(info.OSVersion) >= 15 {
		result.Options = append(result.Options, Option{
			Name:    "palera1n",
			Version: "2.x",
			Mode:    "rootless or rootful fakefs",
			Type:    "semi-tethered",
			Reason:  "A8-A11/checkm8 device with iOS 15+; good when USB/DFU flow is acceptable.",
		})
	}

	if dopamineLikelySupported(info.OSVersion) {
		result.Options = append(result.Options, Option{
			Name:    "Dopamine",
			Version: "2.x",
			Mode:    "rootless",
			Type:    "semi-untethered",
			Reason:  "Useful for lab devices where sideloading is easier than DFU/checkm8 flow.",
		})
	}

	if len(result.Options) == 0 {
		result.Warnings = append(result.Warnings, "No local rule matched. Verify the exact model/iOS against The Apple Wiki before proceeding.")
	}
	if info.OSVersion == "" {
		result.Warnings = append(result.Warnings, "iOS version is unknown; recommendation confidence is low.")
	}
	if info.Chip == "" {
		result.Warnings = append(result.Warnings, "Chip is unknown; palera1n/checkm8 confidence is low.")
	}

	return result
}

func checkm8Supported(chip string) bool {
	switch strings.ToUpper(strings.TrimSpace(chip)) {
	case "A8", "A8X", "A9", "A9X", "A10", "A10X", "A11":
		return true
	default:
		return false
	}
}

func dopamineLikelySupported(version string) bool {
	major := majorVersion(version)
	if major == 15 {
		return true
	}
	if major == 16 {
		minor := minorVersion(version)
		return minor >= 0 && minor <= 6
	}
	return false
}

func majorVersion(version string) int {
	parts := strings.Split(version, ".")
	if len(parts) == 0 {
		return -1
	}
	n, err := strconv.Atoi(parts[0])
	if err != nil {
		return -1
	}
	return n
}

func minorVersion(version string) int {
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return 0
	}
	n, err := strconv.Atoi(parts[1])
	if err != nil {
		return -1
	}
	return n
}
