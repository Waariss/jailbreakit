package version

import "testing"

func TestResolveVersionPrefersExplicitValue(t *testing.T) {
	if got := resolveVersion("v1.3.2", "v1.2.0"); got != "v1.3.2" {
		t.Fatalf("resolveVersion() = %q, want v1.3.2", got)
	}
}

func TestResolveVersionUsesModuleVersion(t *testing.T) {
	if got := resolveVersion("dev", "v1.3.2"); got != "v1.3.2" {
		t.Fatalf("resolveVersion() = %q, want v1.3.2", got)
	}
}

func TestResolveVersionUsesDevelopmentFallback(t *testing.T) {
	if got := resolveVersion("dev", "(devel)"); got != "dev" {
		t.Fatalf("resolveVersion() = %q, want dev", got)
	}
}
