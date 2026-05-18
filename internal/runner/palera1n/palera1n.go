package palera1n

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Mode string

const (
	Rootless          Mode = "rootless"
	RootfulCreateFS   Mode = "rootful-create-bindfs"
	RootfulBootBindFS Mode = "rootful-boot-bindfs"
)

func Run(mode Mode, dryRun bool, stdout, stderr io.Writer) error {
	_, err := RunResult(mode, dryRun, stdout, stderr)
	return err
}

type Result struct {
	Output               string
	NeedsRootfulBootStep bool
	CompletedCheckm8     bool
	USBExclusiveWarning  bool
}

func RunResult(mode Mode, dryRun bool, stdout, stderr io.Writer) (Result, error) {
	args, err := argsForMode(mode)
	if err != nil {
		return Result{}, err
	}

	fmt.Fprintf(stdout, "[*] palera1n %v\n", args)
	if dryRun {
		return Result{}, nil
	}

	var captured bytes.Buffer
	cmd := exec.Command("palera1n", args...)
	cmd.Stdout = io.MultiWriter(stdout, &captured)
	cmd.Stderr = io.MultiWriter(stderr, &captured)
	cmd.Stdin = os.Stdin

	runErr := cmd.Run()
	result := AnalyzeOutput(captured.String())
	printNextSteps(mode, result, stdout)
	return result, runErr
}

func argsForMode(mode Mode) ([]string, error) {
	switch mode {
	case Rootless:
		return []string{"-l", "-Vvv"}, nil
	case RootfulCreateFS:
		return []string{"-f", "-B", "-Vvv"}, nil
	case RootfulBootBindFS:
		return []string{"-f", "-Vvv"}, nil
	default:
		return nil, fmt.Errorf("unsupported palera1n mode %q", mode)
	}
}

func AnalyzeOutput(output string) Result {
	lower := strings.ToLower(output)
	return Result{
		Output:               output,
		NeedsRootfulBootStep: strings.Contains(lower, "run again without the -b") || strings.Contains(lower, "bindfs to be created"),
		CompletedCheckm8:     strings.Contains(lower, "checkmate!") || strings.Contains(lower, "booting kernel"),
		USBExclusiveWarning:  strings.Contains(lower, "usbdeviceopen") || strings.Contains(lower, "exclusive access"),
	}
}

func printNextSteps(mode Mode, result Result, stdout io.Writer) {
	if result.NeedsRootfulBootStep {
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "[+] Rootful BindFS stage started successfully.")
		fmt.Fprintln(stdout, "    Wait until the device reboots into recovery mode, then run:")
		fmt.Fprintln(stdout, "    jailbreakit run palera1n --rootful-boot")
	}
	if result.USBExclusiveWarning && !result.CompletedCheckm8 {
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "[!] USB exclusive access issue detected.")
		fmt.Fprintln(stdout, "    Close Finder/iTunes/Xcode device windows, reconnect the cable, then retry.")
	}
}
