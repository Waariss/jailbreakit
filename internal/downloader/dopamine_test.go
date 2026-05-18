package downloader

import "testing"

func TestTagFromVersion(t *testing.T) {
	tests := map[string]string{
		"2.5 Beta 3": "2.5b3",
		"2.5b3":      "2.5b3",
		"2.4.9":      "2.4.9",
	}
	for input, want := range tests {
		if got := tagFromVersion(input); got != want {
			t.Fatalf("tagFromVersion(%q) = %q, want %q", input, got, want)
		}
	}
}
