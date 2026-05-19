package config

import (
	"strings"
	"testing"
)

func TestSpecProfileFor_BloodDeathKnightHasRotationSteps(t *testing.T) {
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
	if len(profile.Rotation) != 2 {
		t.Fatalf("expected two Blood Death Knight rotation sections (one per hero talent), got %d", len(profile.Rotation))
	}
	for _, section := range profile.Rotation {
		if section.HeroTalent == "" {
			t.Fatalf("expected hero talent on each rotation section")
		}
		if len(section.Steps) == 0 {
			t.Fatalf("expected rotation steps for %s", section.HeroTalent)
		}
	}
	if len(profile.Opener) != 0 {
		t.Fatalf("expected no opener sections, got %d", len(profile.Opener))
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
	if len(paladin.Rotation) == 0 || len(priest.Rotation) == 0 {
		t.Fatalf("expected both Holy Paladin and Holy Priest to have rotation sections")
	}
	if paladin.Rotation[0].Steps[0].Text == priest.Rotation[0].Steps[0].Text {
		t.Fatalf("expected Holy Paladin and Holy Priest to have different rotation steps")
	}
}

func TestManualSpecProfiles_AllHaveRequiredFields(t *testing.T) {
	for _, profile := range manualSpecProfiles {
		if profile.Class == "" || profile.Spec == "" || profile.Role == "" {
			t.Fatalf("expected class, spec, and role for every manual profile: %+v", profile)
		}
		if len(profile.Mechanics) == 0 {
			t.Fatalf("expected mechanics for %s %s", profile.Spec, profile.Class)
		}
		if len(profile.HeroTalents) == 0 {
			t.Fatalf("expected hero talent rotations for %s %s", profile.Spec, profile.Class)
		}
		for _, ht := range profile.HeroTalents {
			if ht.HeroTalent == "" {
				t.Fatalf("expected hero talent name for %s %s", profile.Spec, profile.Class)
			}
			if len(ht.Steps) == 0 {
				t.Fatalf("expected rotation steps for %s %s / %s", profile.Spec, profile.Class, ht.HeroTalent)
			}
		}
	}
}

func TestManualSpecProfiles_BuildsCorrectRotationSections(t *testing.T) {
	profile, ok := SpecProfileFor("Warrior", "Fury")
	if !ok {
		t.Fatalf("expected Fury Warrior profile")
	}
	if len(profile.Rotation) != 2 {
		t.Fatalf("expected 2 rotation sections for Fury Warrior, got %d", len(profile.Rotation))
	}
	for _, section := range profile.Rotation {
		if section.Context != "Rotation" {
			t.Fatalf("expected context 'Rotation', got %s", section.Context)
		}
		if section.HeroTalent == "" {
			t.Fatalf("expected hero talent on rotation section")
		}
		if len(section.Steps) == 0 {
			t.Fatalf("expected steps in rotation section for %s", section.HeroTalent)
		}
		for _, step := range section.Steps {
			if strings.TrimSpace(step.Text) == "" {
				t.Fatalf("expected non-empty step text in %s rotation", section.HeroTalent)
			}
		}
	}
}
