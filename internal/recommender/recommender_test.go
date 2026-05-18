package recommender

import (
	"testing"

	"github.com/waaris/jailbreakit/internal/device"
)

func TestRecommendCheckm8AndDopamine(t *testing.T) {
	result := Recommend(device.Info{
		ProductType: "iPhone8,1",
		Chip:        "A9",
		OSVersion:   "15.8.8",
	})

	if len(result.Options) != 2 {
		t.Fatalf("expected 2 options, got %d: %#v", len(result.Options), result.Options)
	}
	if result.Options[0].Name != "palera1n" {
		t.Fatalf("expected palera1n first, got %q", result.Options[0].Name)
	}
	if result.Options[1].Name != "Dopamine" {
		t.Fatalf("expected Dopamine second, got %q", result.Options[1].Name)
	}
}

func TestRecommendUnknownWarns(t *testing.T) {
	result := Recommend(device.Info{ProductType: "iPhone99,9", OSVersion: "18.0"})
	if len(result.Options) != 0 {
		t.Fatalf("expected no options, got %#v", result.Options)
	}
	if len(result.Warnings) == 0 {
		t.Fatal("expected warnings")
	}
}
