package readiness

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Waariss/jailbreakit/internal/device"
)

type LookupFunc func(string) (string, error)
type RunFunc func(context.Context, string, ...string) ([]byte, error)
type InteractiveRunFunc func(context.Context, string, ...string) ([]byte, error)
type DetectFunc func() (device.Info, error)

type Checker struct {
	Lookup         LookupFunc
	Run            RunFunc
	RunInteractive InteractiveRunFunc
	Detect         DetectFunc
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
	Host        string
	Port        int
	User        string
	DeviceFrida bool
	Interactive bool
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
	Frida       ToolCheck        `json:"frida"`
	FridaPS     ToolCheck        `json:"frida_ps"`
	Objection   ToolCheck        `json:"objection"`
	Device      DeviceFridaCheck `json:"device"`
	Suggestions []string         `json:"suggestions"`
}

type DeviceFridaCheck struct {
	Checked bool   `json:"checked"`
	OK      bool   `json:"ok"`
	Command string `json:"command,omitempty"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
	Hint    string `json:"hint,omitempty"`
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
		RunInteractive: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			cmd := exec.CommandContext(ctx, name, args...)
			cmd.Stdin = os.Stdin
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
	if c.RunInteractive == nil {
		c.RunInteractive = DefaultChecker().RunInteractive
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
		c.Tool("Host dependency ideviceprofile", "ideviceprofile", true),
		c.Tool("Host dependency pymobiledevice3", "pymobiledevice3", true),
		c.Tool("Host dependency cfgutil", "cfgutil", true),
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

func (c Checker) FridaDeviceReadiness() DeviceFridaCheck {
	c = c.withDefaults()
	result := DeviceFridaCheck{
		Checked: true,
		Command: "frida-ps -U",
	}
	if _, err := c.Lookup("frida-ps"); err != nil {
		result.Error = "frida-ps is not available on the host"
		result.Hint = "Install host tools with: python3 -m pip install frida-tools"
		return result
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	out, err := c.Run(ctx, "frida-ps", "-U")
	result.Output = strings.TrimSpace(string(out))
	if err != nil {
		result.Error = strings.TrimSpace(string(out))
		if result.Error == "" {
			result.Error = err.Error()
		}
		result.Hint = "Ensure a paired USB device is connected and frida-server is installed and running on the authorized jailbroken device."
		return result
	}
	result.OK = true
	return result
}

func (c Checker) SSH(options SSHOptions) SSHCheck {
	if strings.TrimSpace(options.Host) == "" {
		return SSHCheck{
			Checked: false,
			Hint:    "pass --ssh-host 127.0.0.1 --ssh-port 2222 --ssh-user root after starting iproxy; add --ssh-interactive for password auth",
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
	args := []string{"-o", "ConnectTimeout=5"}
	if !options.Interactive {
		args = append(args, "-o", "BatchMode=yes")
	}
	args = append(args,
		"-p", strconv.Itoa(options.Port),
		fmt.Sprintf("%s@%s", options.User, options.Host),
		"echo", "jailbreakit-ssh-ok",
	)
	var run func(context.Context, string, ...string) ([]byte, error) = c.Run
	if options.Interactive {
		run = c.RunInteractive
	}
	out, err := run(ctx, "ssh", args...)
	result.Output = strings.TrimSpace(string(out))
	if err != nil {
		result.Error = strings.TrimSpace(string(out))
		if result.Error == "" {
			result.Error = err.Error()
		}
		return result
	}
	result.OK = strings.Contains(result.Output, "jailbreakit-ssh-ok")
	return result
}

func (c Checker) Lab(options SSHOptions) LabCheck {
	frida := c.FridaReadiness()
	if options.DeviceFrida {
		frida.Device = c.FridaDeviceReadiness()
	}
	return LabCheck{
		HostDependencies: c.HostDependencies(),
		Device:           c.Device(),
		IPAInstall:       c.IPAInstallReadiness(),
		Frida:            frida,
		SSH:              c.SSH(options),
		SuggestedTunnel:  "iproxy 2222 22",
	}
}

func PrintLab(w io.Writer, report LabCheck) {
	fmt.Fprintln(w, "Lab readiness")
	if report.Device.Detected {
		info := report.Device.Info
		fmt.Fprintf(w, "[+] Device: %s (%s), %s, iOS %s\n", valueOrUnknown(info.ModelName), valueOrUnknown(info.ProductType), valueOrUnknown(info.Chip), valueOrUnknown(info.OSVersion))
	} else {
		fmt.Fprintf(w, "[-] Device: not detected (%s)\n", valueOrUnknown(report.Device.Error))
	}
	printSummary(w, "Host tools", report.HostDependencies)
	printSummary(w, "IPA install", report.IPAInstall)
	printSummary(w, "Frida host", []ToolCheck{report.Frida.Frida, report.Frida.FridaPS, report.Frida.Objection})
	if report.Frida.Device.Checked {
		if report.Frida.Device.OK {
			fmt.Fprintln(w, "[+] Frida device: ready")
		} else {
			fmt.Fprintf(w, "[-] Frida device: %s\n", shortError(report.Frida.Device.Error))
		}
	} else {
		fmt.Fprintln(w, "[>] Frida device: skipped (use lab-check --device)")
	}
	printSSH(w, report.SSH)
	if !report.Device.Detected || report.SSH.Checked && !report.SSH.OK {
		fmt.Fprintf(w, "[>] Start tunnel: %s\n", report.SuggestedTunnel)
	}
	if missing := missingBinaries(report.HostDependencies); len(missing) > 0 {
		fmt.Fprintf(w, "[>] Next: jailbreakit doctor (%s missing)\n", strings.Join(missing, ", "))
	}
	if !report.Device.Detected {
		fmt.Fprintln(w, "[>] Next: connect, unlock, and trust the iPhone")
	}
	if !report.Frida.Frida.Available || !report.Frida.FridaPS.Available || !report.Frida.Objection.Available {
		fmt.Fprintln(w, "[>] Next: python3 -m pip install frida-tools objection")
	}
	if report.Frida.Device.Checked && !report.Frida.Device.OK && report.Frida.Device.Hint != "" {
		fmt.Fprintf(w, "[>] Next: %s\n", report.Frida.Device.Hint)
	}
}

func PrintFrida(w io.Writer, report FridaCheck) {
	fmt.Fprintln(w, "Frida readiness")
	printCompactTool(w, "Host Frida", report.Frida)
	printCompactTool(w, "frida-ps", report.FridaPS)
	printCompactTool(w, "Objection", report.Objection)
	if report.Device.Checked {
		if report.Device.OK {
			fmt.Fprintln(w, "[+] Device Frida: ready")
		} else {
			fmt.Fprintf(w, "[-] Device Frida: %s\n", shortError(report.Device.Error))
		}
	} else {
		fmt.Fprintln(w, "[>] Device Frida: skipped (run with --device)")
	}
	if !report.Frida.Available || !report.FridaPS.Available || !report.Objection.Available {
		fmt.Fprintln(w, "[>] Install: python3 -m pip install frida-tools objection")
	}
	if report.Device.Checked && !report.Device.OK && report.Device.Hint != "" {
		fmt.Fprintf(w, "[>] %s\n", report.Device.Hint)
	}
}

func printSummary(w io.Writer, label string, tools []ToolCheck) {
	ready := 0
	for _, tool := range tools {
		if tool.Available {
			ready++
		}
	}
	if ready == len(tools) {
		fmt.Fprintf(w, "[+] %s: ready (%d/%d)\n", label, ready, len(tools))
	} else {
		fmt.Fprintf(w, "[!] %s: incomplete (%d/%d)\n", label, ready, len(tools))
		for _, tool := range tools {
			if !tool.Available {
				fmt.Fprintf(w, "    missing: %s\n", tool.Binary)
			}
		}
	}
}

func printCompactTool(w io.Writer, label string, tool ToolCheck) {
	if !tool.Available {
		fmt.Fprintf(w, "[-] %s: missing\n", label)
		return
	}
	if tool.Version != "" {
		fmt.Fprintf(w, "[+] %s: %s\n", label, tool.Version)
		return
	}
	fmt.Fprintf(w, "[+] %s: ready\n", label)
}

func missingBinaries(tools []ToolCheck) []string {
	var missing []string
	for _, tool := range tools {
		if !tool.Available {
			missing = append(missing, tool.Binary)
		}
	}
	return missing
}

func printTool(w io.Writer, tool ToolCheck) {
	if tool.Available {
		if tool.Version != "" {
			fmt.Fprintf(w, "[+] %s: %s (%s)\n", tool.Name, tool.Path, tool.Version)
			return
		}
		if tool.Error != "" && tool.Error != "missing" {
			fmt.Fprintf(w, "[!] %s: %s (version check: %s)\n", tool.Name, tool.Path, shortError(tool.Error))
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

func shortError(value string) string {
	value = strings.TrimSpace(value)
	for _, line := range strings.Split(value, "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Host key verification failed") ||
			strings.Contains(line, "Permission denied") ||
			strings.Contains(line, "Connection refused") ||
			strings.Contains(line, "Connection timed out") ||
			strings.Contains(line, "Could not resolve hostname") {
			return line
		}
	}
	if index := strings.IndexByte(value, '\n'); index >= 0 {
		value = value[:index]
	}
	if len(value) > 160 {
		value = value[:157] + "..."
	}
	return valueOrUnknown(value)
}

func printSSH(w io.Writer, check SSHCheck) {
	if !check.Checked {
		fmt.Fprintf(w, "[>] SSH: skipped (%s)\n", check.Hint)
		return
	}
	if check.OK {
		fmt.Fprintf(w, "[+] SSH: ready (%s@%s:%d)\n", check.User, check.Host, check.Port)
		return
	}
	fmt.Fprintf(w, "[-] SSH: failed (%s@%s:%d)\n", check.User, check.Host, check.Port)
	if check.Error != "" {
		fmt.Fprintf(w, "    error: %s\n", shortError(check.Error))
	}
	if strings.Contains(check.Error, "Host key verification failed") {
		fmt.Fprintf(w, "[>] Next: verify the current key, then run ssh-keygen -R '[%s]:%d'\n", check.Host, check.Port)
	} else if strings.Contains(check.Error, "Permission denied") {
		fmt.Fprintln(w, "[>] Next: use an SSH key/agent or rerun with --ssh-interactive")
	}
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
