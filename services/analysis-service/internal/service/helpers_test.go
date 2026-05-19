package service

import (
	"testing"
	"time"

	"wow-log-analyzer/services/analysis-service/internal/types"
)

func TestDeathKnightRuneCost(t *testing.T) {
	tests := []struct {
		ability string
		want    int
	}{
		{"Heart Strike", 1},
		{"heart strike", 1},
		{"  Heart Strike  ", 1},
		{"Death's Caress", 1},
		{"Death and Decay", 1},
		{"Soul Reaper", 1},
		{"Scourge Strike", 1},
		{"Clawing Shadows", 1},
		{"Festering Strike", 1},
		{"Howling Blast", 1},
		{"Frostscythe", 1},
		{"Marrowrend", 2},
		{"Obliterate", 2},
		{"Unknown Ability", 0},
		{"", 0},
	}

	for _, tt := range tests {
		got := deathKnightRuneCost(tt.ability)
		if got != tt.want {
			t.Errorf("deathKnightRuneCost(%q) = %d, want %d", tt.ability, got, tt.want)
		}
	}
}

func TestDefaultResourceMax(t *testing.T) {
	tests := []struct {
		id   int
		name string
		want float64
	}{
		{1, "", 100},  // Rage
		{2, "", 100},  // Focus
		{3, "", 100},  // Energy
		{4, "", 5},    // Combo Points
		{5, "", 6},    // Runes
		{6, "", 100},  // Runic Power
		{7, "", 5},    // Soul Shards
		{11, "", 10},  // Maelstrom
		{12, "", 6},   // Chi
		{16, "", 4},   // Arcane Charges
		{17, "", 120}, // Fury
		{19, "", 6},   // Essence
		// By name (unknown ID)
		{0, "rage", 100},
		{0, "energy", 100},
		{0, "fury", 120},
		{0, "combo points", 5},
		{0, "chi", 6},
		{0, "arcane charges", 4},
		{0, "maelstrom", 10},
		{0, "insanity", 100},
		{0, "Unknown Resource", 0},
		{999, "Unknown Resource", 0},
	}

	for _, tt := range tests {
		got := defaultResourceMax(tt.id, tt.name)
		if got != tt.want {
			t.Errorf("defaultResourceMax(%d, %q) = %f, want %f", tt.id, tt.name, got, tt.want)
		}
	}
}

func TestResolveResourceStrategy(t *testing.T) {
	tests := []struct {
		class    string
		spec     string
		wantIDs  map[int]bool
		wantPrimary int
	}{
		{"Death Knight", "Blood", map[int]bool{5: true, 6: true}, 6},
		{"Death Knight", "Frost", map[int]bool{5: true, 6: true}, 6},
		{"Death Knight", "Unholy", map[int]bool{5: true, 6: true}, 6},
		{"Rogue", "Assassination", map[int]bool{3: true, 4: true}, 3},
		{"Rogue", "Outlaw", map[int]bool{3: true, 4: true}, 3},
		{"Rogue", "Subtlety", map[int]bool{3: true, 4: true}, 3},
		{"Druid", "Feral", map[int]bool{3: true, 4: true}, 3},
		{"Druid", "Guardian", map[int]bool{1: true}, 1},
		{"Druid", "Balance", map[int]bool{8: true}, 8},
		{"Warrior", "Arms", map[int]bool{1: true}, 1},
		{"Warrior", "Fury", map[int]bool{1: true}, 1},
		{"Demon Hunter", "Havoc", map[int]bool{17: true}, 17},
		{"Demon Hunter", "Vengeance", map[int]bool{18: true}, 18},
		{"Monk", "Windwalker", map[int]bool{3: true, 12: true}, 12},
		{"Hunter", "Beast Mastery", map[int]bool{2: true}, 2},
		{"Shaman", "Enhancement", map[int]bool{11: true}, 11},
		{"Priest", "Shadow", map[int]bool{13: true}, 13},
		{"Mage", "Arcane", map[int]bool{16: true}, 16},
		{"Warlock", "Affliction", map[int]bool{7: true}, 7},
		{"Evoker", "Devastation", map[int]bool{19: true}, 19},
		{"Unknown", "Unknown", nil, 0},
	}

	for _, tt := range tests {
		strategy := resolveResourceStrategy(tt.class, tt.spec)
		if tt.wantIDs == nil {
			if len(strategy.resourceIDs) != 0 {
				t.Errorf("resolveResourceStrategy(%q, %q): expected empty resourceIDs, got %v", tt.class, tt.spec, strategy.resourceIDs)
			}
		} else {
			for id, expected := range tt.wantIDs {
				if strategy.resourceIDs[id] != expected {
					t.Errorf("resolveResourceStrategy(%q, %q): resourceIDs[%d] = %v, want %v", tt.class, tt.spec, id, strategy.resourceIDs[id], expected)
				}
			}
		}
		if strategy.primaryID != tt.wantPrimary {
			t.Errorf("resolveResourceStrategy(%q, %q): primaryID = %d, want %d", tt.class, tt.spec, strategy.primaryID, tt.wantPrimary)
		}
	}
}

func TestBuildResourceCaution(t *testing.T) {
	strategy := resourceStrategy{primaryID: 3, resourceIDs: map[int]bool{3: true}}

	t.Run("high waste pct on primary resource", func(t *testing.T) {
		player := resourceUsageSummary{ResourceTypeID: 3, WastePct: 15, GeneratedPerMinute: 100}
		caution := buildResourceCaution(player, 100, 9, strategy)
		if caution == "" {
			t.Error("expected caution for high waste on primary resource")
		}
	})

	t.Run("high waste pct on non-primary resource", func(t *testing.T) {
		player := resourceUsageSummary{ResourceTypeID: 99, WastePct: 15, GeneratedPerMinute: 100}
		caution := buildResourceCaution(player, 100, 9, strategy)
		if caution == "" {
			t.Error("expected caution for high waste")
		}
	})

	t.Run("low generation behind cohort", func(t *testing.T) {
		player := resourceUsageSummary{ResourceTypeID: 3, WastePct: 2, GeneratedPerMinute: 50}
		caution := buildResourceCaution(player, 100, 2, strategy)
		if caution == "" {
			t.Error("expected caution for low generation on primary resource")
		}
	})

	t.Run("normal performance returns empty caution", func(t *testing.T) {
		player := resourceUsageSummary{ResourceTypeID: 3, WastePct: 3, GeneratedPerMinute: 98}
		caution := buildResourceCaution(player, 100, 3, strategy)
		if caution != "" {
			t.Errorf("expected no caution for normal performance, got %q", caution)
		}
	})
}

func TestRound2(t *testing.T) {
	tests := []struct {
		input float64
		want  float64
	}{
		{1.234, 1.23},
		{1.235, 1.24},
		{0.0, 0.0},
		{100.0, 100.0},
		{-1.235, -1.24},
		{1.0 / 3.0, 0.33},
	}

	for _, tt := range tests {
		got := round2(tt.input)
		if got != tt.want {
			t.Errorf("round2(%f) = %f, want %f", tt.input, got, tt.want)
		}
	}
}

func TestClampTimestamp(t *testing.T) {
	start := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	end := start.Add(5 * time.Minute)

	t.Run("within range unchanged", func(t *testing.T) {
		ts := start.Add(2 * time.Minute)
		got := clampTimestamp(ts, start, end)
		if !got.Equal(ts) {
			t.Errorf("expected %v, got %v", ts, got)
		}
	})

	t.Run("before start clamped to start", func(t *testing.T) {
		got := clampTimestamp(start.Add(-1*time.Minute), start, end)
		if !got.Equal(start) {
			t.Errorf("expected start, got %v", got)
		}
	})

	t.Run("after end clamped to end", func(t *testing.T) {
		got := clampTimestamp(end.Add(1*time.Minute), start, end)
		if !got.Equal(end) {
			t.Errorf("expected end, got %v", got)
		}
	})

	t.Run("zero value returns zero", func(t *testing.T) {
		got := clampTimestamp(time.Time{}, start, end)
		if !got.IsZero() {
			t.Errorf("expected zero time, got %v", got)
		}
	})

	t.Run("zero start returns zero", func(t *testing.T) {
		got := clampTimestamp(start, time.Time{}, end)
		if !got.IsZero() {
			t.Errorf("expected zero time for zero start, got %v", got)
		}
	})
}

func TestFilterTopResourceComparisons(t *testing.T) {
	t.Run("filters out all-zero entries", func(t *testing.T) {
		comparisons := []types.ResourceUsageComparison{
			{ResourceType: "Energy", PlayerGeneratedPerMinute: 10, CohortMedianGeneratedPerMinute: 8},
			{ResourceType: "Empty", PlayerGeneratedPerMinute: 0, CohortMedianGeneratedPerMinute: 0, PlayerWastePerMinute: 0, CohortMedianWastePerMinute: 0},
		}
		result := filterTopResourceComparisons(comparisons, 10)
		if len(result) != 1 {
			t.Errorf("expected 1 result after filtering zeros, got %d", len(result))
		}
		if result[0].ResourceType != "Energy" {
			t.Errorf("expected Energy, got %q", result[0].ResourceType)
		}
	})

	t.Run("respects limit", func(t *testing.T) {
		comparisons := []types.ResourceUsageComparison{
			{ResourceType: "A", PlayerGeneratedPerMinute: 10},
			{ResourceType: "B", PlayerGeneratedPerMinute: 10},
			{ResourceType: "C", PlayerGeneratedPerMinute: 10},
		}
		result := filterTopResourceComparisons(comparisons, 2)
		if len(result) != 2 {
			t.Errorf("expected 2 results at limit, got %d", len(result))
		}
	})

	t.Run("keeps waste-only entries", func(t *testing.T) {
		comparisons := []types.ResourceUsageComparison{
			{ResourceType: "Fury", PlayerWastePerMinute: 5, CohortMedianWastePerMinute: 3},
		}
		result := filterTopResourceComparisons(comparisons, 10)
		if len(result) != 1 {
			t.Errorf("expected 1 result with waste data, got %d", len(result))
		}
	})
}

func TestNormalizeAbilityName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Heart Strike", "heart strike"},
		{"  Shadowstrike  ", "shadowstrike"},
		{"DANCING RUNE WEAPON", "dancing rune weapon"},
		{"", ""},
	}

	for _, tt := range tests {
		got := normalizeAbilityName(tt.input)
		if got != tt.want {
			t.Errorf("normalizeAbilityName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
