package sideload

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/waaris/jailbreakit/internal/device"
)

const envCommand = "JAILBREAKIT_SIDELOAD_CMD"

type Platform struct {
	ID   string
	Name string
	URL  string
}

const defaultSignerPath = "bin/plumesign"

var platforms = []Platform{
	{
		ID:   "macos",
		Name: "macOS universal",
		URL:  "https://github.com/claration/Impactor/releases/download/v2.4.0/plumesign-macos-universal",
	},
	{
		ID:   "linux-aarch64",
		Name: "Linux aarch64",
		URL:  "https://github.com/claration/Impactor/releases/download/v2.4.0/plumesign-linux-aarch64",
	},
	{
		ID:   "linux-x86_64",
		Name: "Linux x86_64",
		URL:  "https://github.com/claration/Impactor/releases/download/v2.4.0/plumesign-linux-x86_64",
	},
}

func Configured(command string) bool {
	return ResolveCommand(command) != ""
}

func AutoCommand() string {
	if _, err := os.Stat(defaultSignerPath); err == nil {
		return "./" + defaultSignerPath + " sign --package {ipa} --apple-id --register-and-install"
	}
	candidates := []string{
		"plumesign sign --package {ipa} --apple-id --register-and-install",
		"plumeimpactor install {ipa}",
		"impactor install {ipa}",
	}
	for _, candidate := range candidates {
		binary := strings.Fields(candidate)[0]
		if _, ok := device.LookPath(binary); ok {
			return candidate
		}
	}
	return ""
}

func InstallHint() string {
	switch runtime.GOOS {
	case "darwin":
		return "jailbreakit signer install --platform macos"
	case "linux":
		return "jailbreakit signer install --platform linux-x86_64"
	}
	return ""
}

func Platforms() []Platform {
	return append([]Platform(nil), platforms...)
}

func DetectPlatform() string {
	switch runtime.GOOS {
	case "darwin":
		return "macos"
	case "linux":
		if runtime.GOARCH == "arm64" {
			return "linux-aarch64"
		}
		return "linux-x86_64"
	default:
		return ""
	}
}

func Install(platformID, outPath string, stdout io.Writer) (string, error) {
	platform, ok := findPlatform(platformID)
	if !ok {
		return "", fmt.Errorf("unsupported signer platform %q", platformID)
	}
	if outPath == "" {
		outPath = defaultSignerPath
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return "", err
	}

	fmt.Fprintf(stdout, "[*] downloading %s\n", platform.URL)
	client := &http.Client{Timeout: 2 * time.Minute}
	resp, err := client.Get(platform.URL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("signer download failed: %s", resp.Status)
	}

	file, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(file, resp.Body); err != nil {
		file.Close()
		return "", err
	}
	if err := file.Close(); err != nil {
		return "", err
	}
	if err := os.Chmod(outPath, 0o755); err != nil {
		return "", err
	}
	fmt.Fprintf(stdout, "[+] installed %s\n", outPath)
	return outPath, nil
}

func Run(ipaPath, command string, stdout io.Writer) error {
	return runOnce(ipaPath, command, stdout, true)
}

func RunWithLoginRetry(ipaPath, command, username string, stdout io.Writer) error {
	err := runOnce(ipaPath, command, stdout, true)
	if err == nil {
		return nil
	}
	if !IsNoAccountError(err) {
		return err
	}
	fmt.Fprintln(stdout, "[!] plumesign account is not logged in")
	if err := Login(username, stdout); err != nil {
		return err
	}
	if err := RunTerminal(ipaPath, command, stdout); err != nil {
		return err
	}
	fmt.Fprintln(stdout, "[+] Dopamine install done")
	return nil
}

func IsNoAccountError(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "no account selected")
}

func IsNotTerminalError(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "not a terminal")
}

func Login(username string, stdout io.Writer) error {
	return LoginWithPassword(username, "", stdout)
}

func LoginWithPassword(username, password string, stdout io.Writer) error {
	args := []string{"account", "login"}
	if strings.TrimSpace(username) != "" {
		args = append(args, "--username", strings.TrimSpace(username))
	}
	if password != "" {
		args = append(args, "--password", password)
	}
	fmt.Fprintln(stdout, "[*] Logging in to Apple Developer account")
	fmt.Fprintln(stdout, "Apple ID 2FA Code: enter the code if Apple asks below")
	cmd := exec.Command("./"+defaultSignerPath, args...)
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = plumesignEnv()
	return cmd.Run()
}

func RunTerminal(ipaPath, command string, stdout io.Writer) error {
	return runOnce(ipaPath, command, stdout, false)
}

func runOnce(ipaPath, command string, stdout io.Writer, captureStderr bool) error {
	template := ResolveCommand(command)
	if template == "" {
		return fmt.Errorf("%s is not configured", envCommand)
	}
	if !strings.Contains(template, "{ipa}") {
		return fmt.Errorf("%s must include {ipa}", envCommand)
	}

	resolved := strings.ReplaceAll(template, "{ipa}", shellQuote(ipaPath))
	fmt.Fprintln(stdout, "[*] Signing and installing Dopamine")

	var stderr bytes.Buffer
	cmd := exec.Command("/bin/sh", "-c", resolved)
	cmd.Stdout = stdout
	if captureStderr {
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	} else {
		cmd.Stderr = os.Stderr
	}
	cmd.Stdin = os.Stdin
	cmd.Env = plumesignEnv()
	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message != "" {
			return fmt.Errorf("%w: %s", err, message)
		}
		return err
	}
	return nil
}

func plumesignEnv() []string {
	env := os.Environ()
	for i, value := range env {
		if strings.HasPrefix(value, "RUST_LOG=") {
			env[i] = "RUST_LOG=plumesign=error,reqwest=warn,omnisette=warn,tungstenite=warn,plume_core=warn"
			return env
		}
	}
	return append(env, "RUST_LOG=plumesign=error,reqwest=warn,omnisette=warn,tungstenite=warn,plume_core=warn")
}

func ResolveCommand(command string) string {
	if strings.TrimSpace(command) != "" {
		return strings.TrimSpace(command)
	}
	if env := strings.TrimSpace(os.Getenv(envCommand)); env != "" {
		return env
	}
	return AutoCommand()
}

func findPlatform(id string) (Platform, bool) {
	for _, platform := range platforms {
		if platform.ID == id {
			return platform, true
		}
	}
	return Platform{}, false
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}
