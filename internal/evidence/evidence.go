package evidence

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Waariss/jailbreakit/internal/readiness"
	"github.com/Waariss/jailbreakit/internal/recommender"
	"github.com/Waariss/jailbreakit/internal/version"
)

const SafetyNote = "Generated for authorized iOS security testing only."

type Report struct {
	Tool            string                `json:"tool"`
	Version         string                `json:"version"`
	Timestamp       string                `json:"timestamp"`
	Host            Host                  `json:"host"`
	Dependencies    []readiness.ToolCheck `json:"dependencies"`
	Device          map[string]string     `json:"device"`
	Recommendations []Recommendation      `json:"recommendations"`
	Readiness       Readiness             `json:"readiness"`
	SafetyNote      string                `json:"safety_note"`
}

type Host struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
}

type Recommendation struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	Mode    string `json:"mode,omitempty"`
	Type    string `json:"type,omitempty"`
	Reason  string `json:"reason,omitempty"`
}

type Readiness struct {
	IPAInstall map[string]readiness.ToolCheck `json:"ipa_install"`
	Frida      readiness.FridaCheck           `json:"frida"`
	SSH        readiness.SSHCheck             `json:"ssh"`
}

func Build(checker readiness.Checker, sshOptions readiness.SSHOptions, now time.Time) Report {
	lab := checker.Lab(sshOptions)
	report := Report{
		Tool:         "jailbreakit",
		Version:      version.Version,
		Timestamp:    now.UTC().Format(time.RFC3339),
		Host:         Host{OS: readiness.HostOS(), Arch: readiness.HostArch()},
		Dependencies: lab.HostDependencies,
		Device:       deviceMap(lab.Device),
		Readiness: Readiness{
			IPAInstall: toolMap(lab.IPAInstall),
			Frida:      lab.Frida,
			SSH:        lab.SSH,
		},
		SafetyNote: SafetyNote,
	}
	if lab.Device.Detected {
		for _, option := range recommender.Recommend(lab.Device.Info).Options {
			report.Recommendations = append(report.Recommendations, Recommendation{
				Name:    option.Name,
				Version: option.Version,
				Mode:    option.Mode,
				Type:    option.Type,
				Reason:  option.Reason,
			})
		}
	}
	return report
}

func Markdown(report Report) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "# jailbreakit Lab Readiness Evidence\n\n")
	fmt.Fprintf(&b, "- Tool: %s\n", report.Tool)
	fmt.Fprintf(&b, "- Version: %s\n", report.Version)
	fmt.Fprintf(&b, "- Timestamp: %s\n", report.Timestamp)
	fmt.Fprintf(&b, "- Host: %s/%s\n", report.Host.OS, report.Host.Arch)
	fmt.Fprintf(&b, "- Safety note: %s\n\n", report.SafetyNote)

	fmt.Fprintln(&b, "## Available Host Dependencies")
	for _, dep := range report.Dependencies {
		fmt.Fprintf(&b, "- %s\n", toolLine(dep))
	}
	fmt.Fprintln(&b)

	fmt.Fprintln(&b, "## Connected Device")
	if len(report.Device) == 0 {
		fmt.Fprintln(&b, "- Not detected")
	} else {
		for _, key := range []string{"product_type", "model", "chip", "ios"} {
			if value := report.Device[key]; strings.TrimSpace(value) != "" {
				fmt.Fprintf(&b, "- %s: %s\n", key, value)
			}
		}
	}
	fmt.Fprintln(&b)

	fmt.Fprintln(&b, "## Recommended Jailbreak / Testing Route")
	if len(report.Recommendations) == 0 {
		fmt.Fprintln(&b, "- No route recommendation available from current device data.")
	} else {
		for _, rec := range report.Recommendations {
			fmt.Fprintf(&b, "- %s %s: %s, %s. %s\n", rec.Name, rec.Version, rec.Mode, rec.Type, rec.Reason)
		}
	}
	fmt.Fprintln(&b)

	fmt.Fprintln(&b, "## IPA Install Readiness")
	for _, name := range []string{"ssh", "scp", "iproxy", "ideviceinstaller", "appinst", "ipainstaller"} {
		if tool, ok := report.Readiness.IPAInstall[name]; ok {
			fmt.Fprintf(&b, "- %s\n", toolLine(tool))
		}
	}
	fmt.Fprintln(&b)

	fmt.Fprintln(&b, "## Frida / Objection Readiness")
	fmt.Fprintf(&b, "- %s\n", toolLine(report.Readiness.Frida.Frida))
	fmt.Fprintf(&b, "- %s\n", toolLine(report.Readiness.Frida.FridaPS))
	fmt.Fprintf(&b, "- %s\n", toolLine(report.Readiness.Frida.Objection))
	for _, suggestion := range report.Readiness.Frida.Suggestions {
		fmt.Fprintf(&b, "- Next: %s\n", suggestion)
	}
	fmt.Fprintln(&b)

	fmt.Fprintln(&b, "## SSH Readiness")
	ssh := report.Readiness.SSH
	if !ssh.Checked {
		fmt.Fprintf(&b, "- Skipped: %s\n", ssh.Hint)
	} else if ssh.OK {
		fmt.Fprintf(&b, "- OK: %s@%s:%d (%s)\n", ssh.User, ssh.Host, ssh.Port, ssh.Output)
	} else {
		fmt.Fprintf(&b, "- Failed: %s@%s:%d (%s)\n", ssh.User, ssh.Host, ssh.Port, ssh.Error)
	}
	return b.Bytes()
}

func JSON(report Report) ([]byte, error) {
	return json.MarshalIndent(report, "", "  ")
}

func deviceMap(check readiness.DeviceCheck) map[string]string {
	if !check.Detected {
		return map[string]string{}
	}
	info := check.Info
	return map[string]string{
		"product_type": info.ProductType,
		"model":        info.ModelName,
		"chip":         info.Chip,
		"ios":          info.OSVersion,
	}
}

func toolMap(tools []readiness.ToolCheck) map[string]readiness.ToolCheck {
	result := make(map[string]readiness.ToolCheck, len(tools))
	for _, tool := range tools {
		result[tool.Binary] = tool
	}
	return result
}

func toolLine(tool readiness.ToolCheck) string {
	if tool.Available {
		if tool.Version != "" {
			return fmt.Sprintf("%s: available at %s (%s)", tool.Binary, tool.Path, tool.Version)
		}
		return fmt.Sprintf("%s: available at %s", tool.Binary, tool.Path)
	}
	if tool.Optional {
		return fmt.Sprintf("%s: optional, missing", tool.Binary)
	}
	return fmt.Sprintf("%s: missing", tool.Binary)
}
