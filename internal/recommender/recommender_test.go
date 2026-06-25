package recommender

import (
	"testing"

	"github.com/Waariss/jailbreakit/internal/device"
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

func TestRecommendA11IOS16OmitsDopamine(t *testing.T) {
	result := Recommend(device.Info{
		ProductType: "iPhone10,5",
		Chip:        "A11",
		OSVersion:   "16.7.10",
	})

	if len(result.Options) != 1 {
		t.Fatalf("expected only palera1n, got %#v", result.Options)
	}
	if result.Options[0].Name != "palera1n" {
		t.Fatalf("expected palera1n, got %#v", result.Options[0])
	}
	for _, option := range result.Options {
		if option.Name == "Dopamine" {
			t.Fatalf("did not expect Dopamine for A11/iOS 16.7.10: %#v", result.Options)
		}
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

func TestRecommendIOS17NoFallback(t *testing.T) {
	result := Recommend(device.Info{ProductType: "iPhone8,1", Chip: "A9", OSVersion: "17.0"})
	if len(result.Options) != 0 {
		t.Fatalf("expected no options for iOS 17 fallback, got %#v", result.Options)
	}
}

func TestRecommendIOS12RecommendOnly(t *testing.T) {
	result := Recommend(device.Info{ProductType: "iPhone8,1", Chip: "A9", OSVersion: "12.1"})
	if len(result.Options) != 2 {
		t.Fatalf("expected 2 options, got %#v", result.Options)
	}
	if result.Options[0].Mode != "recommend-only" {
		t.Fatalf("expected recommend-only, got %#v", result.Options[0])
	}
}
