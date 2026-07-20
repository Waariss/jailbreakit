package version

import (
	"runtime/debug"
	"strings"
)

var (
	Version = "dev"
	Author  = "Waariss"
	Commit  = "unknown"
	Date    = "unknown"
)

// Resolved returns the release linker value, module version, or a local fallback.
func Resolved() string {
	moduleVersion, settings := buildInfo()
	resolved := resolveVersion(Version, moduleVersion)
	if Commit == "unknown" {
		Commit = settings["vcs.revision"]
		if Commit == "" {
			Commit = "unknown"
		}
	}
	if Date == "unknown" {
		Date = settings["vcs.time"]
		if Date == "" {
			Date = "unknown"
		}
	}
	return resolved
}

func buildInfo() (string, map[string]string) {
	info, ok := debug.ReadBuildInfo()
	if !ok || info == nil {
		return "", nil
	}
	settings := make(map[string]string, len(info.Settings))
	for _, setting := range info.Settings {
		settings[setting.Key] = setting.Value
	}
	return info.Main.Version, settings
}

func resolveVersion(explicit, module string) string {
	explicit = strings.TrimSpace(explicit)
	if explicit != "" && explicit != "dev" && explicit != "unknown" {
		return explicit
	}
	module = strings.TrimSpace(module)
	if module != "" && module != "(devel)" {
		return module
	}
	return "dev"
}
