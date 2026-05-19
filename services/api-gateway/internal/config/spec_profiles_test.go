package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSpecProfileFor_BloodDeathKnightUsesManualScreenshotReferences(t *testing.T) {
	profile, ok := SpecProfileFor("Death Knight", "Blood")
	if !ok {
		t.Fatalf("expected Blood Death Knight profile")
	}
	if profile.Label != "Blood Death Knight" {
		t.Fatalf("expected Blood Death Knight label, got %s", profile.Label)
	}
	if profile.Role != "Tank" {
		t.Fatalf("expected Tank role, got %s", profile.Role)
	}
	if profile.SourceURL != "guide reference screenshots/blood-deathbringer.png" {
		t.Fatalf("expected screenshot source, got %s", profile.SourceURL)
	}
	if len(profile.Rotation) != 3 {
		t.Fatalf("expected three Blood Death Knight reference sections, got %d", len(profile.Rotation))
	}
	if len(profile.Opener) != 0 {
		t.Fatalf("expected manual screenshot profiles not to synthesize opener sections")
	}

	mechanics := strings.ToLower(strings.Join(profile.KeyMechanics, " "))
	if !strings.Contains(mechanics, "rune") {
		t.Fatalf("expected Blood DK mechanics to mention runes, got %s", mechanics)
	}
	if !strings.Contains(mechanics, "runic power") {
		t.Fatalf("expected Blood DK mechanics to mention runic power, got %s", mechanics)
	}
}

func TestSpecProfileFor_NormalizesHyphenatedSpecs(t *testing.T) {
	profile, ok := SpecProfileFor("Hunter", "Beast-Mastery")
	if !ok {
		t.Fatalf("expected Beast Mastery Hunter profile")
	}
	if profile.Label != "Beast Mastery Hunter" {
		t.Fatalf("expected Beast Mastery Hunter label, got %s", profile.Label)
	}
}

func TestSpecProfileFor_DisambiguatesSharedSpecNamesByClass(t *testing.T) {
	paladin, ok := SpecProfileFor("Paladin", "Holy")
	if !ok {
		t.Fatalf("expected Holy Paladin profile")
	}
	priest, ok := SpecProfileFor("Priest", "Holy")
	if !ok {
		t.Fatalf("expected Holy Priest profile")
	}
	if paladin.Label != "Holy Paladin" {
		t.Fatalf("expected Holy Paladin label, got %s", paladin.Label)
	}
	if priest.Label != "Holy Priest" {
		t.Fatalf("expected Holy Priest label, got %s", priest.Label)
	}
	if paladin.SourceURL == priest.SourceURL {
		t.Fatalf("expected shared spec names to keep class-specific screenshot sources")
	}
}

func TestManualSpecProfilesReferenceScreenshots(t *testing.T) {
	seen := map[string]string{}
	for _, profile := range manualSpecProfiles {
		if profile.Class == "" || profile.Spec == "" || profile.Role == "" {
			t.Fatalf("expected class, spec, and role for every manual profile: %+v", profile)
		}
		if len(profile.Mechanics) == 0 {
			t.Fatalf("expected mechanics for %s %s", profile.Spec, profile.Class)
		}
		if len(profile.Screenshots) == 0 {
			t.Fatalf("expected screenshot references for %s %s", profile.Spec, profile.Class)
		}
		for _, screenshot := range profile.Screenshots {
			if screenshot.HeroTalent == "" || screenshot.File == "" {
				t.Fatalf("expected hero talent and file for %s %s screenshot: %+v", profile.Spec, profile.Class, screenshot)
			}
			if !strings.HasPrefix(screenshot.File, "guide reference screenshots/") {
				t.Fatalf("expected screenshot path under guide reference screenshots, got %s", screenshot.File)
			}
			if existing, ok := seen[screenshot.File]; ok {
				t.Fatalf("screenshot %s is referenced by both %s and %s %s", screenshot.File, existing, profile.Spec, profile.Class)
			}
			seen[screenshot.File] = profile.Spec + " " + profile.Class
		}
	}
	entries, err := os.ReadDir(filepath.Join("..", "..", "..", "..", "guide reference screenshots"))
	if err != nil {
		t.Fatalf("read guide reference screenshots: %v", err)
	}
	expected := map[string]struct{}{}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".png" {
			continue
		}
		expected["guide reference screenshots/"+entry.Name()] = struct{}{}
	}
	if len(seen) != len(expected) {
		t.Fatalf("expected %d screenshot references, got %d", len(expected), len(seen))
	}
	for file := range expected {
		if _, ok := seen[file]; !ok {
			t.Fatalf("expected manual profile reference for %s", file)
		}
	}
}
