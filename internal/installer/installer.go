package installer

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Waariss/jailbreakit/internal/device"
)

var supportedInstallers = []string{"appinst", "ipainstaller"}

type Options struct {
	IPAPath    string
	Host       string
	User       string
	Port       int
	LocalPort  int
	DevicePort int
	RemoteDir  string
	Installer  string
	DryRun     bool
}

type commandSpec struct {
	Name string
	Args []string
}

func Run(options Options, stdout, stderr io.Writer) error {
	options = withDefaults(options)
	if err := validate(options); err != nil {
		return err
	}

	if isHostInstaller(options.Installer) {
		return runHostInstall(options, stdout, stderr)
	}

	host := options.Host
	port := options.Port
	var tunnel *exec.Cmd
	usesUSBTunnel := host == ""
	if usesUSBTunnel {
		host = "127.0.0.1"
		port = options.LocalPort
		tunnel = exec.Command("iproxy", strconv.Itoa(options.LocalPort), strconv.Itoa(options.DevicePort))
		fmt.Fprintf(stdout, "[*] USB tunnel: iproxy %d %d\n", options.LocalPort, options.DevicePort)
		if !options.DryRun {
			tunnel.Stdout = stdout
			tunnel.Stderr = stderr
			if err := tunnel.Start(); err != nil {
				return fmt.Errorf("start iproxy failed: %w", err)
			}
			defer stopTunnel(tunnel)
			time.Sleep(750 * time.Millisecond)
		}
	}

	remotePath := remoteIPAPath(options.RemoteDir, options.IPAPath)
	copyCmd := scpCommand(options.IPAPath, options.User, host, port, remotePath)
	installCmd := sshCommand(options.User, host, port, installRemoteArgs(options.Installer, remotePath))

	if options.DryRun {
		printCommand(stdout, copyCmd)
		printCommand(stdout, installCmd)
		return nil
	}

	fmt.Fprintf(stdout, "[*] Copying %s to %s@%s:%s\n", options.IPAPath, options.User, host, remotePath)
	if err := runCommand(copyCmd, stdout, stderr); err != nil {
		return fmt.Errorf("copy IPA failed: %w", err)
	}

	fmt.Fprintf(stdout, "[*] Installing with %s on the iPhone\n", installerLabel(options.Installer))
	if err := runCommand(installCmd, stdout, stderr); err != nil {
		if options.Installer == "auto" && isMissingRemoteInstaller(err) && usesUSBTunnel && hostInstallerAvailable() {
			fmt.Fprintln(stdout, "[*] No device-side IPA installer found; falling back to ideviceinstaller on host")
			return runHostInstall(options, stdout, stderr)
		}
		return fmt.Errorf("install IPA failed: %w", err)
	}

	fmt.Fprintln(stdout, "[+] IPA install done")
	return nil
}

func MissingTools() []string {
	var missing []string
	for _, name := range []string{"ssh", "scp", "iproxy"} {
		if _, ok := device.LookPath(name); !ok {
			missing = append(missing, name)
		}
	}
	return missing
}

func MissingHostTools() []string {
	if hostInstallerAvailable() {
		return nil
	}
	return []string{"ideviceinstaller"}
}

func InstallHint() string {
	switch runtime.GOOS {
	case "darwin":
		return "brew install libusbmuxd libimobiledevice ideviceinstaller"
	case "linux":
		return "install openssh-client, libusbmuxd tools, and ideviceinstaller, e.g. sudo apt install -y openssh-client libusbmuxd-tools ideviceinstaller"
	}
	return ""
}

func withDefaults(options Options) Options {
	if options.User == "" {
		options.User = "root"
	}
	if options.Port == 0 {
		options.Port = 22
	}
	if options.LocalPort == 0 {
		options.LocalPort = 2222
	}
	if options.DevicePort == 0 {
		options.DevicePort = 22
	}
	if options.RemoteDir == "" {
		options.RemoteDir = "/tmp"
	}
	if options.Installer == "" {
		options.Installer = "auto"
	}
	return options
}

func validate(options Options) error {
	if strings.TrimSpace(options.IPAPath) == "" {
		return fmt.Errorf("missing IPA path")
	}
	if !strings.HasSuffix(strings.ToLower(options.IPAPath), ".ipa") {
		return fmt.Errorf("IPA path must end with .ipa")
	}
	if _, err := os.Stat(options.IPAPath); err != nil {
		return err
	}
	if options.LocalPort < 1 || options.LocalPort > 65535 {
		return fmt.Errorf("invalid local port %d", options.LocalPort)
	}
	if options.DevicePort < 1 || options.DevicePort > 65535 {
		return fmt.Errorf("invalid device port %d", options.DevicePort)
	}
	if options.Port < 1 || options.Port > 65535 {
		return fmt.Errorf("invalid SSH port %d", options.Port)
	}
	if options.Installer != "auto" && !isHostInstaller(options.Installer) && strings.ContainsAny(options.Installer, " \t\r\n") {
		return fmt.Errorf("installer must be a single command name")
	}
	return nil
}

func runHostInstall(options Options, stdout, stderr io.Writer) error {
	cmd := hostInstallCommand(options.IPAPath)
	if options.DryRun {
		printCommand(stdout, cmd)
		return nil
	}
	fmt.Fprintln(stdout, "[*] Installing with ideviceinstaller on host")
	var commandStderr bytes.Buffer
	if err := runCommand(cmd, stdout, io.MultiWriter(stderr, &commandStderr)); err != nil {
		if isSignatureVerificationFailure(commandStderr.String()) {
			fmt.Fprintln(stdout, "[>] IPA signature rejected: use jailbreakit sideload <app.ipa>")
			fmt.Fprintln(stdout, "[>] Or install AppSync with appinst/ipainstaller on an authorized jailbroken device")
		}
		return fmt.Errorf("host IPA install failed: %w", err)
	}
	fmt.Fprintln(stdout, "[+] IPA install done")
	return nil
}

func isSignatureVerificationFailure(output string) bool {
	output = strings.ToLower(output)
	return strings.Contains(output, "applicationverificationfailed") ||
		strings.Contains(output, "no code signature found") ||
		strings.Contains(output, "failed to verify code signature")
}

func hostInstallCommand(ipaPath string) commandSpec {
	return commandSpec{Name: "ideviceinstaller", Args: []string{"install", ipaPath}}
}

func isHostInstaller(installer string) bool {
	return installer == "host" || installer == "ideviceinstaller"
}

func hostInstallerAvailable() bool {
	_, ok := device.LookPath("ideviceinstaller")
	return ok
}

func isMissingRemoteInstaller(err error) bool {
	return strings.Contains(err.Error(), "exit status 127")
}

func installRemoteArgs(installer, remotePath string) []string {
	if installer != "auto" {
		return []string{installer + " " + shellQuote(remotePath)}
	}
	return []string{autoInstallerCommand(remotePath)}
}

func autoInstallerCommand(remotePath string) string {
	return "ipa_path=" + shellQuote(remotePath) + "; " +
		`for tool in appinst ipainstaller; do ` +
		`if command -v "$tool" >/dev/null 2>&1; then ` +
		`echo "[*] using $tool"; exec "$tool" "$ipa_path"; ` +
		`fi; ` +
		`done; ` +
		`echo "[!] no supported IPA installer found on iPhone: appinst, ipainstaller" >&2; ` +
		`echo "[!] install AppSync Unified/appinst or ipainstaller, then retry; or pass --installer <command>" >&2; ` +
		`exit 127`
}

func installerLabel(installer string) string {
	if installer == "auto" {
		return "auto (" + strings.Join(supportedInstallers, ", ") + ", ideviceinstaller fallback)"
	}
	if isHostInstaller(installer) {
		return "ideviceinstaller"
	}
	return installer
}

func scpCommand(localPath, user, host string, port int, remotePath string) commandSpec {
	return commandSpec{
		Name: "scp",
		Args: append(scpCommonArgs(port), localPath, fmt.Sprintf("%s@%s:%s", user, host, remotePath)),
	}
}

func sshCommand(user, host string, port int, remoteArgs []string) commandSpec {
	args := append(sshCommonArgs(port), fmt.Sprintf("%s@%s", user, host))
	args = append(args, remoteArgs...)
	return commandSpec{Name: "ssh", Args: args}
}

func sshCommonArgs(port int) []string {
	return []string{
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-p", strconv.Itoa(port),
	}
}

func scpCommonArgs(port int) []string {
	return []string{
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-P", strconv.Itoa(port),
	}
}

func remoteIPAPath(remoteDir, ipaPath string) string {
	return strings.TrimRight(remoteDir, "/") + "/" + sanitizeRemoteName(filepath.Base(ipaPath))
}

func sanitizeRemoteName(name string) string {
	var builder strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_' {
			builder.WriteRune(r)
			continue
		}
		builder.WriteByte('_')
	}
	if builder.Len() == 0 {
		return "App.ipa"
	}
	return builder.String()
}

func runCommand(spec commandSpec, stdout, stderr io.Writer) error {
	cmd := exec.Command(spec.Name, spec.Args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func printCommand(w io.Writer, spec commandSpec) {
	fmt.Fprintf(w, "[dry-run] %s %s\n", spec.Name, strings.Join(spec.Args, " "))
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}

func stopTunnel(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	_ = cmd.Process.Kill()
	_, _ = cmd.Process.Wait()
}
