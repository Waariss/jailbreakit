package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallAcceptsFlagsAfterIPA(t *testing.T) {
	dir := t.TempDir()
	ipa := filepath.Join(dir, "App.ipa")
	if err := os.WriteFile(ipa, []byte("ipa"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	err := installIPA([]string{ipa, "--host", "192.168.1.23", "--port", "44", "--installer", "ipainstaller", "--dry-run"}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	if !strings.Contains(output, "scp") || !strings.Contains(output, "-P 44") || !strings.Contains(output, "ipainstaller") {
		t.Fatalf("unexpected dry-run output:\n%s", output)
	}
}
