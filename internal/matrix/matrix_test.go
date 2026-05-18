package matrix

import "testing"

func TestFallbackIOS15(t *testing.T) {
	entries := fallback("15.8.8")
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Tool != "palera1n" || entries[0].Version != "2.2.1" {
		t.Fatalf("unexpected palera1n entry: %#v", entries[0])
	}
	if entries[1].Tool != "Dopamine" || entries[1].Version != "2.5 Beta 3" {
		t.Fatalf("unexpected Dopamine entry: %#v", entries[1])
	}
}

func TestParseText(t *testing.T) {
	text := `## iOS
15.8.8 palera1n 2.2.1 Semi-Tethered
Dopamine 2.5 Beta 3 Yes
## tvOS`
	entries := parseText("15.8.8", text)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d: %#v", len(entries), entries)
	}
}
