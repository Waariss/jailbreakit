package deps

import (
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"

	"github.com/waaris/jailbreakit/internal/device"
)

type Dependency struct {
	Name        string
	Binary      string
	RequiredFor string
	InstallHint string
}

type Missing struct {
	Dependency
}

func Required() []Dependency {
	return []Dependency{
		{
			Name:        "palera1n",
			Binary:      "palera1n",
			RequiredFor: "checkm8 jailbreak flow",
			InstallHint: installHint("palera1n"),
		},
		{
			Name:        "libimobiledevice",
			Binary:      "ideviceinfo",
			RequiredFor: "device detection",
			InstallHint: installHint("libimobiledevice"),
		},
		{
			Name:        "curl",
			Binary:      "curl",
			RequiredFor: "downloads",
			InstallHint: installHint("curl"),
		},
	}
}

func MissingRequired() []Missing {
	var missing []Missing
	for _, dep := range Required() {
		if _, ok := device.LookPath(dep.Binary); !ok {
			missing = append(missing, Missing{Dependency: dep})
		}
	}
	return missing
}

func PrintMissing(missing []Missing, w io.Writer) {
	for _, item := range missing {
		fmt.Fprintf(w, "[-] %s missing (%s)\n", item.Name, item.RequiredFor)
		if item.InstallHint != "" {
			fmt.Fprintf(w, "    install: %s\n", item.InstallHint)
		}
	}
}

func InstallMissing(missing []Missing, w io.Writer) error {
	for _, item := range missing {
		if item.InstallHint == "" {
			fmt.Fprintf(w, "[!] no installer known for %s on %s\n", item.Name, runtime.GOOS)
			continue
		}
		fmt.Fprintf(w, "[*] %s\n", item.InstallHint)
		cmd := exec.Command("/bin/sh", "-c", item.InstallHint)
		cmd.Stdout = w
		cmd.Stderr = w
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("install %s failed: %w", item.Name, err)
		}
	}
	return nil
}

func installHint(name string) string {
	switch runtime.GOOS {
	case "darwin":
		switch name {
		case "palera1n":
			return "see https://palera.in"
		case "libimobiledevice":
			return "brew install libimobiledevice"
		case "curl":
			return "brew install curl"
		}
	case "linux":
		switch name {
		case "libimobiledevice":
			return linuxPackageInstall("libimobiledevice-utils")
		case "curl":
			return linuxPackageInstall("curl")
		case "palera1n":
			return "see https://palera.in for Linux install instructions"
		}
	}
	return ""
}

func linuxPackageInstall(pkg string) string {
	if _, ok := device.LookPath("apt"); ok {
		return "sudo apt install -y " + pkg
	}
	if _, ok := device.LookPath("dnf"); ok {
		return "sudo dnf install -y " + pkg
	}
	if _, ok := device.LookPath("pacman"); ok {
		return "sudo pacman -S --needed " + pkg
	}
	return ""
}

func RunnableInstallHints(missing []Missing) []Missing {
	var runnable []Missing
	for _, item := range missing {
		hint := strings.TrimSpace(item.InstallHint)
		if hint == "" || strings.HasPrefix(hint, "see ") {
			continue
		}
		runnable = append(runnable, item)
	}
	return runnable
}
