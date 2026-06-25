package readiness

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Waariss/jailbreakit/internal/device"
)

type LookupFunc func(string) (string, error)
type RunFunc func(context.Context, string, ...string) ([]byte, error)
type DetectFunc func() (device.Info, error)

type Checker struct {
	Lookup LookupFunc
	Run    RunFunc
	Detect DetectFunc
}

type ToolCheck struct {
	Name      string `json:"name"`
	Binary    string `json:"binary"`
	Available bool   `json:"available"`
	Path      string `json:"path,omitempty"`
	Version   string `json:"version,omitempty"`
	Error     string `json:"error,omitempty"`
	Optional  bool   `json:"optional,omitempty"`
}

type DeviceCheck struct {
	Detected bool        `json:"detected"`
	Info     device.Info `json:"info,omitempty"`
	Error    string      `json:"error,omitempty"`
}

type SSHOptions struct {
	Host string
	Port int
	User string
}

type SSHCheck struct {
	Checked bool   `json:"checked"`
	OK      bool   `json:"ok"`
	Host    string `json:"host,omitempty"`
	Port    int    `json:"port,omitempty"`
	User    string `json:"user,omitempty"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
	Hint    string `json:"hint,omitempty"`
}

type FridaCheck struct {
	Frida       ToolCheck `json:"frida"`
	FridaPS     ToolCheck `json:"frida_ps"`
	Objection   ToolCheck `json:"objection"`
	Suggestions []string  `json:"suggestions"`
}

type LabCheck struct {
	HostDependencies []ToolCheck `json:"host_dependencies"`
	Device           DeviceCheck `json:"device"`
	IPAInstall       []ToolCheck `json:"ipa_install"`
	Frida            FridaCheck  `json:"frida"`
	SSH              SSHCheck    `json:"ssh"`
	SuggestedTunnel  string      `json:"suggested_tunnel"`
}

func DefaultChecker() Checker {
	return Checker{
		Lookup: exec.LookPath,
		Run: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			cmd := exec.CommandContext(ctx, name, args...)
			return cmd.CombinedOutput()
		},
		Detect: device.Detect,
	}
}

func HostOS() string {
	return runtime.GOOS
}

func HostArch() string {
	return runtime.GOARCH
}

func (c Checker) withDefaults() Checker {
	if c.Lookup == nil {
		c.Lookup = exec.LookPath
	}
	if c.Run == nil {
		c.Run = DefaultChecker().Run
	}
	if c.Detect == nil {
		c.Detect = device.Detect
	}
	return c
}

func (c Checker) Tool(name, binary string, optional bool, versionArgs ...string) ToolCheck {
	c = c.withDefaults()
	check := ToolCheck{Name: name, Binary: binary, Optional: optional}
	path, err := c.Lookup(binary)
	if err != nil {
		check.Error = "missing"
		return check
	}
	check.Available = true
	check.Path = path
	if len(versionArgs) > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		out, err := c.Run(ctx, binary, versionArgs...)
		if err == nil {
			check.Version = firstLine(out)
		} else {
			check.Error = strings.TrimSpace(string(out))
			if check.Error == "" {
				check.Error = err.Error()
			}
		}
	}
	return check
}

func (c Checker) HostDependencies() []ToolCheck {
	return []ToolCheck{
		c.Tool("Host dependency ideviceinfo", "ideviceinfo", false),
		c.Tool("Host dependency iproxy", "iproxy", false),
		c.Tool("Host dependency ssh", "ssh", false),
		c.Tool("Host dependency scp", "scp", false),
		c.Tool("Host dependency ideviceinstaller", "ideviceinstaller", false),
		c.Tool("Host dependency curl", "curl", false),
		c.Tool("Host dependency palera1n", "palera1n", true),
	}
}

func (c Checker) Device() DeviceCheck {
	c = c.withDefaults()
	info, err := c.Detect()
	if err != nil {
		return DeviceCheck{Error: err.Error()}
	}
	return DeviceCheck{Detected: true, Info: info}
}

func (c Checker) IPAInstallReadiness() []ToolCheck {
	return []ToolCheck{
		c.Tool("IPA install ssh", "ssh", false),
		c.Tool("IPA install scp", "scp", false),
		c.Tool("IPA install iproxy", "iproxy", false),
		c.Tool("IPA install ideviceinstaller", "ideviceinstaller", false),
		c.Tool("Device-side installer appinst", "appinst", true),
		c.Tool("Device-side installer ipainstaller", "ipainstaller", true),
	}
}

func (c Checker) FridaReadiness() FridaCheck {
	return FridaCheck{
		Frida:     c.Tool("Frida CLI", "frida", false, "--version"),
		FridaPS:   c.Tool("Frida process lister", "frida-ps", false, "--version"),
		Objection: c.Tool("Objection", "objection", false, "version"),
		Suggestions: []string{
			"Install host tools if missing: python3 -m pip install frida-tools objection",
			"List USB-connected processes: frida-ps -U",
			"With a jailbroken device and SSH, ensure frida-server is installed and running on the device.",
		},
	}
}

func (c Checker) SSH(options SSHOptions) SSHCheck {
	if strings.TrimSpace(options.Host) == "" {
		return SSHCheck{
			Checked: false,
			Hint:    "pass --ssh-host 127.0.0.1 --ssh-port 2222 --ssh-user root after starting iproxy",
		}
	}
	c = c.withDefaults()
	if options.Port == 0 {
		options.Port = 22
	}
	if strings.TrimSpace(options.User) == "" {
		options.User = "root"
	}
	result := SSHCheck{
		Checked: true,
		Host:    options.Host,
		Port:    options.Port,
		User:    options.User,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := c.Run(ctx, "ssh",
		"-o", "BatchMode=yes",
		"-o", "ConnectTimeout=5",
		"-p", strconv.Itoa(options.Port),
		fmt.Sprintf("%s@%s", options.User, options.Host),
		"echo", "jailbreakit-ssh-ok",
	)
	result.Output = strings.TrimSpace(string(out))
	if err != nil {
		result.Error = err.Error()
		return result
	}
	result.OK = strings.Contains(result.Output, "jailbreakit-ssh-ok")
	return result
}

func (c Checker) Lab(options SSHOptions) LabCheck {
	return LabCheck{
		HostDependencies: c.HostDependencies(),
		Device:           c.Device(),
		IPAInstall:       c.IPAInstallReadiness(),
		Frida:            c.FridaReadiness(),
		SSH:              c.SSH(options),
		SuggestedTunnel:  "iproxy 2222 22",
	}
}

func PrintLab(w io.Writer, report LabCheck) {
	for _, dep := range report.HostDependencies {
		printTool(w, dep)
	}
	fmt.Fprintln(w)
	if report.Device.Detected {
		info := report.Device.Info
		fmt.Fprintln(w, "[+] Connected iOS device detected")
		fmt.Fprintf(w, "    ProductType: %s\n", valueOrUnknown(info.ProductType))
		fmt.Fprintf(w, "    Model:       %s\n", valueOrUnknown(info.ModelName))
		fmt.Fprintf(w, "    Chip:        %s\n", valueOrUnknown(info.Chip))
		fmt.Fprintf(w, "    iOS:         %s\n", valueOrUnknown(info.OSVersion))
	} else {
		fmt.Fprintf(w, "[-] Connected iOS device: not detected (%s)\n", valueOrUnknown(report.Device.Error))
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "[*] IPA install readiness")
	for _, dep := range report.IPAInstall {
		printTool(w, dep)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "[*] Frida / Objection readiness")
	printTool(w, report.Frida.Frida)
	printTool(w, report.Frida.FridaPS)
	printTool(w, report.Frida.Objection)
	for _, suggestion := range report.Frida.Suggestions {
		fmt.Fprintf(w, "    next: %s\n", suggestion)
	}
	fmt.Fprintln(w)
	printSSH(w, report.SSH)
	fmt.Fprintf(w, "[*] Suggested tunnel: %s\n", report.SuggestedTunnel)
}

func PrintFrida(w io.Writer, report FridaCheck) {
	printTool(w, report.Frida)
	printTool(w, report.FridaPS)
	printTool(w, report.Objection)
	for _, suggestion := range report.Suggestions {
		fmt.Fprintf(w, "[*] %s\n", suggestion)
	}
}

func printTool(w io.Writer, tool ToolCheck) {
	if tool.Available {
		if tool.Version != "" {
			fmt.Fprintf(w, "[+] %s: %s (%s)\n", tool.Name, tool.Path, tool.Version)
			return
		}
		if tool.Error != "" && tool.Error != "missing" {
			fmt.Fprintf(w, "[!] %s: %s (version check: %s)\n", tool.Name, tool.Path, tool.Error)
			return
		}
		fmt.Fprintf(w, "[+] %s: %s\n", tool.Name, tool.Path)
		return
	}
	prefix := "[-]"
	if tool.Optional {
		prefix = "[!]"
	}
	fmt.Fprintf(w, "%s %s: missing\n", prefix, tool.Name)
}

func printSSH(w io.Writer, check SSHCheck) {
	if !check.Checked {
		fmt.Fprintf(w, "[!] SSH check skipped: %s\n", check.Hint)
		return
	}
	if check.OK {
		fmt.Fprintf(w, "[+] SSH check %s@%s:%d: %s\n", check.User, check.Host, check.Port, check.Output)
		return
	}
	fmt.Fprintf(w, "[-] SSH check %s@%s:%d failed: %s\n", check.User, check.Host, check.Port, check.Error)
}

func firstLine(out []byte) string {
	line, err := bytes.NewBuffer(out).ReadString('\n')
	if err != nil && len(line) == 0 {
		line = string(out)
	}
	return strings.TrimSpace(line)
}

func valueOrUnknown(value string) string {
	if strings.TrimSpace(value) == "" {
		return "unknown"
	}
	return strings.TrimSpace(value)
}
