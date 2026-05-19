package recommender

import (
	"strconv"
	"strings"

	"github.com/Waariss/jailbreakit/internal/device"
	"github.com/Waariss/jailbreakit/internal/matrix"
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
			case "chimera", "unc0ver", "checkra1n", "odyssey", "taurine":
				result.Options = append(result.Options, Option{
					Name:    entry.Tool,
					Version: entry.Version,
					Mode:    "recommend-only",
					Type:    statusType(entry.Status),
					Reason:  "Matched iOS compatibility matrix from " + source + ". Runner automation is not implemented yet.",
				})
			}
		}
		if len(result.Options) > 0 {
			return result
		}
	}

	major := majorVersion(info.OSVersion)
	if checkm8Supported(info.Chip) && major >= 15 && major <= 16 {
		result.Options = append(result.Options, Option{
			Name:    "palera1n",
			Version: "2.x",
			Mode:    "rootless or rootful fakefs",
			Type:    "semi-tethered",
			Reason:  "A8-A11/checkm8 device with iOS 15+; good when USB/DFU flow is acceptable.",
		})
	}

	if len(result.Options) == 0 {
		if majorVersion(info.OSVersion) >= 17 {
			result.Warnings = append(result.Warnings, "No tool available for this iOS version in jailbreakit.")
		} else {
			result.Warnings = append(result.Warnings, "No local rule matched. Verify the exact model/iOS against The Apple Wiki before proceeding.")
		}
	}
	if info.OSVersion == "" {
		result.Warnings = append(result.Warnings, "iOS version is unknown; recommendation confidence is low.")
	}
	if info.Chip == "" {
		result.Warnings = append(result.Warnings, "Chip is unknown; palera1n/checkm8 confidence is low.")
	}

	return result
}

func statusType(status string) string {
	if strings.EqualFold(status, "yes") || status == "" {
		return "semi-untethered"
	}
	return strings.ToLower(status)
}

func checkm8Supported(chip string) bool {
	switch strings.ToUpper(strings.TrimSpace(chip)) {
	case "A8", "A8X", "A9", "A9X", "A10", "A10X", "A11":
		return true
	default:
		return false
	}
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
