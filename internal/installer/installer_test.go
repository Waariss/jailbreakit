package installer

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestRemoteIPAPath(t *testing.T) {
	got := remoteIPAPath("/tmp/", "/Users/test/My App.ipa")
	if got != "/tmp/My_App.ipa" {
		t.Fatalf("remoteIPAPath() = %q", got)
	}
}

func TestValidateRejectsInstallerWithArguments(t *testing.T) {
	ipa := filepath.Join(t.TempDir(), "App.ipa")
	if err := os.WriteFile(ipa, []byte("ipa"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := validate(withDefaults(Options{IPAPath: ipa, Installer: "appinst --flag"}))
	if err == nil || !strings.Contains(err.Error(), "single command") {
		t.Fatalf("expected installer validation error, got %v", err)
	}
}

func TestAutoInstallerRemoteArgs(t *testing.T) {
	got := installRemoteArgs("auto", "/tmp/App.ipa")
	if len(got) != 1 {
		t.Fatalf("installRemoteArgs() = %#v, want one remote command", got)
	}
	if !strings.Contains(got[0], "appinst") || !strings.Contains(got[0], "ipainstaller") || !strings.Contains(got[0], "ipa_path='/tmp/App.ipa'") {
		t.Fatalf("auto command does not include expected content:\n%s", got[0])
	}
}

func TestHostInstallerAliases(t *testing.T) {
	if !isHostInstaller("host") || !isHostInstaller("ideviceinstaller") {
		t.Fatal("expected host installer aliases")
	}
	if isHostInstaller("ipainstaller") {
		t.Fatal("did not expect device-side installer to be host alias")
	}
}

func TestHostInstallCommandUsesCurrentIdeviceinstallerSyntax(t *testing.T) {
	got := hostInstallCommand("App.ipa")
	want := commandSpec{Name: "ideviceinstaller", Args: []string{"install", "App.ipa"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("hostInstallCommand() = %#v, want %#v", got, want)
	}
}

func TestMissingRemoteInstallerDetection(t *testing.T) {
	if !isMissingRemoteInstaller(execExitError("exit status 127")) {
		t.Fatal("expected exit 127 to be treated as missing remote installer")
	}
	if isMissingRemoteInstaller(execExitError("exit status 1")) {
		t.Fatal("did not expect exit 1 to be treated as missing remote installer")
	}
}

func TestScpCommandUsesUppercasePortFlag(t *testing.T) {
	got := scpCommand("App.ipa", "root", "127.0.0.1", 2222, "/tmp/App.ipa")
	want := commandSpec{
		Name: "scp",
		Args: []string{
			"-o", "StrictHostKeyChecking=no",
			"-o", "UserKnownHostsFile=/dev/null",
			"-P", "2222",
			"App.ipa",
			"root@127.0.0.1:/tmp/App.ipa",
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("scpCommand() = %#v, want %#v", got, want)
	}
}

type execExitError string

func (e execExitError) Error() string {
	return string(e)
}

func TestSSHCommandUsesLowercasePortFlag(t *testing.T) {
	got := sshCommand("mobile", "192.168.1.23", 44, installRemoteArgs("ipainstaller", "/tmp/App.ipa"))
	want := commandSpec{
		Name: "ssh",
		Args: []string{
			"-o", "StrictHostKeyChecking=no",
			"-o", "UserKnownHostsFile=/dev/null",
			"-p", "44",
			"mobile@192.168.1.23",
			"ipainstaller '/tmp/App.ipa'",
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("sshCommand() = %#v, want %#v", got, want)
	}
}

func TestWithDefaults(t *testing.T) {
	got := withDefaults(Options{IPAPath: "App.ipa"})
	if got.User != "root" || got.Port != 22 || got.LocalPort != 2222 || got.DevicePort != 22 || got.RemoteDir != "/tmp" || got.Installer != "auto" {
		t.Fatalf("unexpected defaults: %#v", got)
	}
}
