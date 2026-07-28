package sideload

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPreviewCommandQuotesIPAPath(t *testing.T) {
	got, err := PreviewCommand("/tmp/My App.ipa", "sign --package {ipa}")
	if err != nil {
		t.Fatal(err)
	}
	if got != "sign --package '/tmp/My App.ipa'" {
		t.Fatalf("PreviewCommand() = %q", got)
	}
}

func TestPreviewCommandRequiresPlaceholder(t *testing.T) {
	_, err := PreviewCommand("App.ipa", "sign --package App.ipa")
	if err == nil || !strings.Contains(err.Error(), "{ipa}") {
		t.Fatalf("expected placeholder error, got %v", err)
	}
}

func TestResolveCommandUsesEnvironment(t *testing.T) {
	t.Setenv(envCommand, "custom {ipa}")
	if got := ResolveCommand(""); got != "custom {ipa}" {
		t.Fatalf("ResolveCommand() = %q", got)
	}
}

func TestPlumesignPathUsesPATH(t *testing.T) {
	dir := t.TempDir()
	signer := filepath.Join(dir, "plumesign")
	if err := os.WriteFile(signer, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	got, err := plumesignPath()
	if err != nil {
		t.Fatal(err)
	}
	if got != signer {
		t.Fatalf("plumesignPath() = %q, want %q", got, signer)
	}
}
