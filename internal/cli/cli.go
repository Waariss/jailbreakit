package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/waaris/jailbreakit/internal/deps"
	"github.com/waaris/jailbreakit/internal/device"
	"github.com/waaris/jailbreakit/internal/downloader"
	"github.com/waaris/jailbreakit/internal/recommender"
	"github.com/waaris/jailbreakit/internal/runner/palera1n"
	"github.com/waaris/jailbreakit/internal/sideload"
	"github.com/waaris/jailbreakit/internal/troubleshoot"
)

func Run(args []string) error {
	return run(args, os.Stdin, os.Stdout, os.Stderr)
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return wizard(stdin, stdout, stderr)
	}

	switch args[0] {
	case "help", "-h", "--help":
		if len(args) > 1 && args[1] == "advanced" {
			printAdvancedHelp(stdout)
		} else {
			printHelp(stdout)
		}
	case "doctor":
		return doctorWithArgs(args[1:], stdin, stdout)
	case "detect":
		return detect(stdout)
	case "recommend":
		return recommend(args[1:], stdout)
	case "run":
		return runWorkflow(args[1:], stdout, stderr)
	case "signer":
		return signer(args[1:], stdin, stdout)
	case "download":
		return download(args[1:], stdout)
	case "troubleshoot":
		return parseTroubleshoot(args[1:], stdout)
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}

	return nil
}

func printHelp(w io.Writer) {
	fmt.Fprintln(w, `jailbreakit - iOS jailbreak assistant

Usage:
  jailbreakit

Common:
  jailbreakit doctor
  jailbreakit detect
  jailbreakit recommend --ios 15.8.8 --product iPhone8,1

Advanced:
  jailbreakit help advanced`)
}

func printAdvancedHelp(w io.Writer) {
	fmt.Fprintln(w, `jailbreakit advanced commands

Actions:
  jailbreakit run palera1n --rootless
  jailbreakit run palera1n --rootful
  jailbreakit run palera1n --rootful-boot
  jailbreakit run dopamine --version "2.5 Beta 3"
  jailbreakit run dopamine --url <ipa-url>
  jailbreakit run dopamine --sideload-cmd "plumesign sign --package {ipa} --apple-id --register-and-install"

Utility:
  jailbreakit signer install
  jailbreakit signer install --platform macos
  jailbreakit download dopamine --out ./downloads
  jailbreakit doctor --install
  jailbreakit troubleshoot --from-log palera1n.log`)
}

func wizard(stdin io.Reader, stdout, stderr io.Writer) error {
	prompt := newPrompt(stdin, stdout)

	fmt.Fprintln(stdout, "jailbreakit")
	if err := ensureDependencies(prompt, stdout); err != nil {
		return err
	}

	fmt.Fprintln(stdout, "[*] Detecting device...")
	info, err := device.Detect()
	if err != nil {
		fmt.Fprintf(stdout, "[!] Auto-detect failed: %v\n", err)
		if !prompt.confirm("Enter device info manually?", true) {
			return nil
		}
		info = device.Enrich(device.Info{
			ProductType: prompt.ask("ProductType, e.g. iPhone8,1"),
			OSVersion:   prompt.ask("iOS version, e.g. 15.8.8"),
		})
	}

	fmt.Fprintln(stdout)
	printDevice(stdout, info)

	result := recommender.Recommend(info)
	fmt.Fprintln(stdout)
	if len(result.Options) == 0 {
		for _, warning := range result.Warnings {
			fmt.Fprintf(stdout, "[!] %s\n", warning)
		}
		return nil
	}

	fmt.Fprintln(stdout, "Options:")
	for i, option := range result.Options {
		fmt.Fprintf(stdout, "[%d] %s %s - %s, %s\n", i+1, option.Name, option.Version, option.Mode, option.Type)
	}
	for _, warning := range result.Warnings {
		fmt.Fprintf(stdout, "[!] %s\n", warning)
	}

	choice := 1
	if len(result.Options) > 1 {
		choice = prompt.choose("Choose jailbreak route", len(result.Options))
	}
	option := result.Options[choice-1]

	switch strings.ToLower(option.Name) {
	case "palera1n":
		return wizardPalera1n(prompt, stdout, stderr)
	case "dopamine":
		return runDopamineInteractive(prompt, "", "downloads", "", option.Version, stdout)
	default:
		return fmt.Errorf("no wizard handler for %s", option.Name)
	}
}

func ensureDependencies(prompt *prompt, stdout io.Writer) error {
	missing := deps.MissingRequired()
	if len(missing) == 0 {
		return nil
	}

	fmt.Fprintln(stdout, "[!] Missing dependencies:")
	deps.PrintMissing(missing, stdout)

	runnable := deps.RunnableInstallHints(missing)
	if len(runnable) == 0 {
		return nil
	}
	if !prompt.confirm("Install missing dependencies now?", true) {
		return nil
	}
	return deps.InstallMissing(runnable, stdout)
}

func wizardPalera1n(prompt *prompt, stdout, stderr io.Writer) error {
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "palera1n mode:")
	fmt.Fprintln(stdout, "[1] rootless")
	fmt.Fprintln(stdout, "[2] rootful fakefs")

	mode := palera1n.Rootless
	if prompt.choose("Choose palera1n mode", 2) == 2 {
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "rootful stage:")
		fmt.Fprintln(stdout, "[1] create BindFS (first time)")
		fmt.Fprintln(stdout, "[2] boot existing BindFS")
		mode = palera1n.RootfulCreateFS
		if prompt.choose("Choose rootful stage", 2) == 2 {
			mode = palera1n.RootfulBootBindFS
		}
	}

	result, err := palera1n.RunResult(mode, false, stdout, stderr)
	if err != nil {
		return err
	}
	if mode == palera1n.RootfulCreateFS && result.NeedsRootfulBootStep {
		if prompt.confirm("Run rootful boot step after the device returns to recovery?", true) {
			return palera1n.Run(palera1n.RootfulBootBindFS, false, stdout, stderr)
		}
	}
	return nil
}

func doctor(w io.Writer) error {
	return doctorWithArgs(nil, nil, w)
}

func doctorWithArgs(args []string, stdin io.Reader, w io.Writer) error {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	install := fs.Bool("install", false, "install missing dependencies when possible")
	if err := fs.Parse(args); err != nil {
		return err
	}

	missing := deps.MissingRequired()
	for _, dep := range deps.Required() {
		path, ok := device.LookPath(dep.Binary)
		if ok {
			fmt.Fprintf(w, "[+] %-28s %s\n", dep.Name, path)
		} else {
			fmt.Fprintf(w, "[-] %-28s missing (%s)\n", dep.Name, dep.RequiredFor)
			if dep.InstallHint != "" {
				fmt.Fprintf(w, "    install: %s\n", dep.InstallHint)
			}
		}
	}

	fmt.Fprintln(w)
	if sideload.Configured("") {
		fmt.Fprintf(w, "[+] Dopamine sideload command %s\n", sideload.ResolveCommand(""))
	} else {
		fmt.Fprintln(w, `[-] Dopamine sideload command missing`)
		if hint := sideload.InstallHint(); hint != "" {
			fmt.Fprintf(w, "    install: %s\n", hint)
		}
		fmt.Fprintln(w, `    or use --sideload-cmd "your-signer {ipa}"`)
	}

	if *install {
		runnable := deps.RunnableInstallHints(missing)
		if len(runnable) == 0 {
			return nil
		}
		return deps.InstallMissing(runnable, w)
	}
	return nil
}

func detect(w io.Writer) error {
	info, err := device.Detect()
	if err != nil {
		return err
	}
	printDevice(w, info)
	return nil
}

func recommend(args []string, w io.Writer) error {
	fs := flag.NewFlagSet("recommend", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	product := fs.String("product", "", "Apple product type, e.g. iPhone8,1")
	ios := fs.String("ios", "", "iOS version, e.g. 15.8.8")
	if err := fs.Parse(args); err != nil {
		return err
	}

	info := device.Enrich(device.Info{ProductType: *product, OSVersion: *ios})
	if info.ProductType == "" || info.OSVersion == "" {
		detected, err := device.Detect()
		if err != nil {
			return fmt.Errorf("pass --product and --ios, or connect a device: %w", err)
		}
		info = detected
	}

	result := recommender.Recommend(info)
	printDevice(w, info)
	fmt.Fprintln(w)
	for i, option := range result.Options {
		fmt.Fprintf(w, "[%d] %s %s\n", i+1, option.Name, option.Version)
		fmt.Fprintf(w, "    Mode: %s\n", option.Mode)
		fmt.Fprintf(w, "    Type: %s\n", option.Type)
		fmt.Fprintf(w, "    Fit:  %s\n", option.Reason)
	}
	if len(result.Warnings) > 0 {
		fmt.Fprintln(w)
		for _, warning := range result.Warnings {
			fmt.Fprintf(w, "[!] %s\n", warning)
		}
	}
	return nil
}

func runWorkflow(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("missing workflow: palera1n")
	}

	switch args[0] {
	case "palera1n":
		fs := flag.NewFlagSet("run palera1n", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		rootless := fs.Bool("rootless", false, "run rootless palera1n")
		rootful := fs.Bool("rootful", false, "run first-time rootful palera1n BindFS creation")
		rootfulBoot := fs.Bool("rootful-boot", false, "boot an existing rootful BindFS")
		dryRun := fs.Bool("dry-run", false, "print command without executing")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		selected := 0
		for _, value := range []bool{*rootless, *rootful, *rootfulBoot} {
			if value {
				selected++
			}
		}
		if selected != 1 {
			return fmt.Errorf("choose exactly one of --rootless, --rootful, or --rootful-boot")
		}
		mode := palera1n.Rootless
		if *rootful {
			mode = palera1n.RootfulCreateFS
		}
		if *rootfulBoot {
			mode = palera1n.RootfulBootBindFS
		}
		return palera1n.Run(mode, *dryRun, stdout, stderr)
	case "dopamine":
		fs := flag.NewFlagSet("run dopamine", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		out := fs.String("out", "downloads", "output directory")
		url := fs.String("url", "", "explicit Dopamine IPA URL")
		version := fs.String("version", "", "Dopamine version, e.g. '2.5 Beta 3'")
		sideloadCmd := fs.String("sideload-cmd", "", "sideload command template, e.g. 'plumesign sign --package {ipa} --apple-id --register-and-install'")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		return runDopamine(*url, *out, *sideloadCmd, *version, stdout)
	default:
		return fmt.Errorf("unknown workflow %q", args[0])
	}
}

func runDopamine(url, out, sideloadCmd, version string, stdout io.Writer) error {
	path, err := downloader.DownloadDopamineVersion(url, out, version)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "[+] saved %s\n", path)

	if !sideload.Configured(sideloadCmd) {
		fmt.Fprintln(stdout, "[!] sideload command is not configured")
		fmt.Fprintln(stdout, `    Use: jailbreakit run dopamine --sideload-cmd "plumesign sign --package {ipa} --apple-id --register-and-install"`)
		return nil
	}
	if err := sideload.RunTerminal(path, sideloadCmd, stdout); err != nil {
		return err
	}
	fmt.Fprintln(stdout, "[+] Dopamine install done")
	printTrustInstructions(stdout, "")
	return nil
}

func runDopamineInteractive(prompt *prompt, url, out, sideloadCmd, version string, stdout io.Writer) error {
	path, err := downloader.DownloadDopamineVersion(url, out, version)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "[+] saved %s\n", path)

	if sideload.Configured(sideloadCmd) {
		return runSideloadInteractive(prompt, path, sideloadCmd, stdout)
	}

	fmt.Fprintln(stdout, "[!] CLI sideload signer is not installed")
	if !prompt.confirm("Install plumesign CLI and continue?", true) {
		if hint := sideload.InstallHint(); hint != "" {
			fmt.Fprintf(stdout, "    install later: %s\n", hint)
		}
		return nil
	}

	platform := chooseSignerPlatform(prompt, stdout)
	if _, err := sideload.Install(platform, "", stdout); err != nil {
		return err
	}
	return runSideloadInteractive(prompt, path, sideloadCmd, stdout)
}

func runSideloadInteractive(prompt *prompt, path, sideloadCmd string, stdout io.Writer) error {
	fmt.Fprintln(stdout, "[*] Apple ID login")
	username := prompt.ask("Apple ID email")
	password := prompt.askSecret("Apple ID password")
	if err := sideload.LoginWithPassword(username, password, stdout); err != nil {
		return err
	}
	if err := sideload.RunTerminal(path, sideloadCmd, stdout); err != nil {
		return err
	}
	fmt.Fprintln(stdout, "[+] Dopamine install done")
	printTrustInstructions(stdout, username)
	return nil
}

func printTrustInstructions(stdout io.Writer, username string) {
	fmt.Fprintln(stdout, "[*] On iPhone: Settings > General > VPN & Device Management")
	if username != "" {
		fmt.Fprintf(stdout, "[*] Trust developer profile: %s\n", username)
	} else {
		fmt.Fprintln(stdout, "[*] Trust the Apple Developer profile used for signing")
	}
	fmt.Fprintln(stdout, "[*] Then open Dopamine and tap Jailbreak")
}

func signer(args []string, stdin io.Reader, stdout io.Writer) error {
	if len(args) == 0 || args[0] != "install" {
		return fmt.Errorf("usage: jailbreakit signer install --platform macos")
	}
	prompt := newPrompt(stdin, stdout)
	fs := flag.NewFlagSet("signer install", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	platform := fs.String("platform", "", "macos, linux-aarch64, or linux-x86_64")
	out := fs.String("out", "", "output path")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	selected := *platform
	if selected == "" {
		selected = chooseSignerPlatform(prompt, stdout)
	}
	_, err := sideload.Install(selected, *out, stdout)
	return err
}

func chooseSignerPlatform(prompt *prompt, stdout io.Writer) string {
	platforms := sideload.Platforms()
	detected := sideload.DetectPlatform()
	fmt.Fprintln(stdout, "plumesign platform:")
	for i, platform := range platforms {
		suffix := ""
		if platform.ID == detected {
			suffix = " (detected)"
		}
		fmt.Fprintf(stdout, "[%d] %s%s\n", i+1, platform.Name, suffix)
	}
	choice := prompt.choose("Choose platform", len(platforms))
	return platforms[choice-1].ID
}

func download(args []string, w io.Writer) error {
	if len(args) == 0 || args[0] != "dopamine" {
		return fmt.Errorf("usage: jailbreakit download dopamine --out ./downloads")
	}
	fs := flag.NewFlagSet("download dopamine", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	out := fs.String("out", "downloads", "output directory")
	url := fs.String("url", "", "explicit Dopamine IPA URL")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	path, err := downloader.DownloadDopamine(*url, *out)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "[+] saved %s\n", path)
	return nil
}

func parseTroubleshoot(args []string, w io.Writer) error {
	fs := flag.NewFlagSet("troubleshoot", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	logPath := fs.String("from-log", "", "path to palera1n log")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *logPath == "" {
		return fmt.Errorf("usage: jailbreakit troubleshoot --from-log palera1n.log")
	}
	content, err := os.ReadFile(*logPath)
	if err != nil {
		return err
	}
	findings := troubleshoot.ParsePalera1nLog(string(content))
	if len(findings) == 0 {
		fmt.Fprintln(w, "No known palera1n issues matched this log yet.")
		return nil
	}
	for _, finding := range findings {
		fmt.Fprintf(w, "[!] %s\n    %s\n", finding.Title, strings.Join(finding.Suggestions, "\n    "))
	}
	return nil
}

func printDevice(w io.Writer, info device.Info) {
	fmt.Fprintf(w, "ProductType:  %s\n", valueOrUnknown(info.ProductType))
	fmt.Fprintf(w, "Model:        %s\n", valueOrUnknown(info.ModelName))
	fmt.Fprintf(w, "Chip:         %s\n", valueOrUnknown(info.Chip))
	fmt.Fprintf(w, "iOS:          %s\n", valueOrUnknown(info.OSVersion))
}

func valueOrUnknown(value string) string {
	if value == "" {
		return "unknown"
	}
	return value
}
