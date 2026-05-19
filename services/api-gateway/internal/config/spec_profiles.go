package config

import (
	"strings"
	"sync"
)

type SpecProfile struct {
	Label          string             `json:"label,omitempty"`
	SourceURL      string             `json:"sourceUrl,omitempty"`
	Role           string             `json:"role,omitempty"`
	StatPriorities []string           `json:"statPriorities,omitempty"`
	KeyMechanics   []string           `json:"keyMechanics,omitempty"`
	Rotation       []SpecGuideSection `json:"rotation,omitempty"`
	Opener         []SpecGuideSection `json:"opener,omitempty"`
}

type SpecGuideSection struct {
	Context    string          `json:"context,omitempty"`
	HeroTalent string          `json:"heroTalent,omitempty"`
	Steps      []SpecGuideStep `json:"steps,omitempty"`
}

type SpecGuideStep struct {
	Text     string   `json:"text"`
	SpellIDs []string `json:"spellIds,omitempty"`
}

type manualSpecProfile struct {
	Class       string
	Spec        string
	Role        string
	Mechanics   []string
	Screenshots []manualSpecScreenshot
}

type manualSpecScreenshot struct {
	HeroTalent string
	File       string
}

var (
	specProfilesOnce sync.Once
	specProfiles     map[string]SpecProfile
)

func SpecProfileFor(characterClass, characterSpec string) (SpecProfile, bool) {
	specProfilesOnce.Do(func() {
		specProfiles = loadSpecProfiles()
	})
	profile, ok := specProfiles[specProfileKey(characterClass, characterSpec)]
	return profile, ok
}

func loadSpecProfiles() map[string]SpecProfile {
	profiles := make(map[string]SpecProfile, len(manualSpecProfiles))
	for _, entry := range manualSpecProfiles {
		profiles[specProfileKey(entry.Class, entry.Spec)] = buildManualSpecProfile(entry)
	}
	return profiles
}

func buildManualSpecProfile(entry manualSpecProfile) SpecProfile {
	sections := make([]SpecGuideSection, 0, len(entry.Screenshots))
	for _, screenshot := range entry.Screenshots {
		sections = append(sections, SpecGuideSection{
			Context:    "Guide reference screenshot",
			HeroTalent: screenshot.HeroTalent,
			Steps: []SpecGuideStep{
				{Text: "Manual guide reference: " + screenshot.File},
			},
		})
	}

	sourceURL := ""
	if len(entry.Screenshots) > 0 {
		sourceURL = entry.Screenshots[0].File
	}

	return SpecProfile{
		Label:        entry.Spec + " " + entry.Class,
		SourceURL:    sourceURL,
		Role:         entry.Role,
		KeyMechanics: entry.Mechanics,
		Rotation:     sections,
	}
}

func screenshot(heroTalent, file string) manualSpecScreenshot {
	return manualSpecScreenshot{
		HeroTalent: heroTalent,
		File:       "guide reference screenshots/" + file,
	}
}

var manualSpecProfiles = []manualSpecProfile{
	{
		Class: "Death Knight",
		Spec:  "Blood",
		Role:  "Tank",
		Mechanics: []string{
			"Manage rune availability so core spenders are not delayed.",
			"Build and spend Runic Power deliberately.",
			"Plan defensive cooldowns around dangerous windows.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Deathbringer", "blood-deathbringer.png"),
			screenshot("Deathbringer", "blood-deathbringer2.png"),
			screenshot("San'layn", "blood-sanlayn.png"),
		},
	},
	{
		Class: "Death Knight",
		Spec:  "Frost",
		Role:  "DPS",
		Mechanics: []string{
			"Manage rune availability so core spenders are not delayed.",
			"Build and spend Runic Power deliberately.",
			"Align major cooldowns with high-value windows.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Deathbringer", "frost-deathbringer.png"),
			screenshot("Rider of the Apocalypse", "frost-rider.png"),
		},
	},
	{
		Class: "Death Knight",
		Spec:  "Unholy",
		Role:  "DPS",
		Mechanics: []string{
			"Manage rune availability so core spenders are not delayed.",
			"Build and spend Runic Power deliberately.",
			"Align major cooldowns with high-value windows.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Rider of the Apocalypse", "unholy-rider.png"),
			screenshot("San'layn", "unholy-sanlayn.png"),
		},
	},
	{
		Class: "Demon Hunter",
		Spec:  "Devourer",
		Role:  "DPS",
		Mechanics: []string{
			"Build and spend Fury around priority windows.",
			"Align major cooldowns with high-value windows.",
			"Plan defensive cooldowns around dangerous windows.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Annihilator", "devourer-annihilator.png"),
			screenshot("Voidscarred", "devourer-voidscarred.png"),
		},
	},
	{
		Class: "Demon Hunter",
		Spec:  "Havoc",
		Role:  "DPS",
		Mechanics: []string{
			"Build and spend Fury around priority windows.",
			"Align major cooldowns with high-value windows.",
			"Plan defensive cooldowns around dangerous windows.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Aldrachi Reaver", "havoc-aldrachi.png"),
			screenshot("Aldrachi Reaver", "havoc-aldrachi2.png"),
			screenshot("Fel-Scarred", "havoc-felscarred.png"),
		},
	},
	{
		Class: "Demon Hunter",
		Spec:  "Vengeance",
		Role:  "Tank",
		Mechanics: []string{
			"Build and spend Fury around priority windows.",
			"Plan defensive cooldowns around dangerous windows.",
			"Maintain important buffs, debuffs, or effects.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Aldrachi Reaver", "vengeance-aldrachi.png"),
			screenshot("Aldrachi Reaver", "vengeance-aldrachi2.png"),
			screenshot("Annihilator", "vengeance-annihilator.png"),
			screenshot("Annihilator", "vengeance-annihilator2.png"),
		},
	},
	{
		Class: "Druid",
		Spec:  "Balance",
		Role:  "DPS",
		Mechanics: []string{
			"Maintain important buffs, debuffs, or effects.",
			"Align major cooldowns with high-value windows.",
			"Plan movement so priority casts are not delayed.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Keeper of the Grove", "balance-keeper.png"),
			screenshot("Keeper of the Grove", "balance-keeper2.png"),
			screenshot("Elune's Chosen", "balance-elune.png"),
		},
	},
	{
		Class: "Druid",
		Spec:  "Feral",
		Role:  "DPS",
		Mechanics: []string{
			"Build and spend Combo Points efficiently.",
			"Manage Energy pooling and avoid starving key windows.",
			"Maintain important buffs, debuffs, or effects.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Druid of the Claw", "feral-druidoftheclaw.png"),
			screenshot("Druid of the Claw", "feral-druidoftheclaw2.png"),
			screenshot("Wildstalker", "feral-wildstalker.png"),
			screenshot("Wildstalker", "feral-wildstalker2.png"),
		},
	},
	{
		Class: "Druid",
		Spec:  "Guardian",
		Role:  "Tank",
		Mechanics: []string{
			"Maintain important buffs, debuffs, or effects.",
			"Plan defensive cooldowns around dangerous windows.",
			"Use self-healing tools when survivability requires it.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Druid of the Claw", "guardian-druidoftheclaw.png"),
			screenshot("Elune's Chosen", "guardian-eluner.png"),
		},
	},
	{
		Class: "Druid",
		Spec:  "Restoration",
		Role:  "Healer",
		Mechanics: []string{
			"Plan healing globals around damage patterns and resource limits.",
			"Maintain important buffs, debuffs, or effects.",
			"Plan major cooldowns around dangerous windows.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Keeper of the Grove", "estoration-keeper.png"),
			screenshot("Wildstalker", "restoration-wildstalker.png"),
		},
	},
	{
		Class: "Evoker",
		Spec:  "Augmentation",
		Role:  "DPS",
		Mechanics: []string{
			"Maintain important buffs, debuffs, or effects.",
			"Align major cooldowns with high-value windows.",
			"Plan support globals around group damage windows.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Chronowarden", "augmentation-chronowarden.png"),
			screenshot("Scalecommander", "augmentation-scalecommander.png"),
		},
	},
	{
		Class: "Evoker",
		Spec:  "Devastation",
		Role:  "DPS",
		Mechanics: []string{
			"Align major cooldowns with high-value windows.",
			"Plan empowered casts so priority windows are not delayed.",
			"Plan movement so priority casts are not delayed.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Flameshaper", "devastation-flameshaper.png"),
			screenshot("Scalecommander", "devastation-scalecommander.png"),
		},
	},
	{
		Class: "Evoker",
		Spec:  "Preservation",
		Role:  "Healer",
		Mechanics: []string{
			"Plan healing globals around damage patterns and resource limits.",
			"Plan empowered casts so priority windows are not delayed.",
			"Plan major cooldowns around dangerous windows.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Chronowarden", "preservation-chronowarden.png"),
			screenshot("Flameshaper", "preservation-flameshaper.png"),
		},
	},
	{
		Class: "Hunter",
		Spec:  "Beast Mastery",
		Role:  "DPS",
		Mechanics: []string{
			"Manage Focus so priority abilities stay available.",
			"Maintain important buffs, debuffs, or effects.",
			"Align major cooldowns with high-value windows.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Dark Ranger", "beastmastery-darkranger.png"),
			screenshot("Pack Leader", "beastmastery-packleader.png"),
		},
	},
	{
		Class: "Hunter",
		Spec:  "Marksmanship",
		Role:  "DPS",
		Mechanics: []string{
			"Manage Focus so priority abilities stay available.",
			"Align major cooldowns with high-value windows.",
			"Plan movement so priority casts are not delayed.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Dark Ranger", "marksmanship-darkranger.png"),
			screenshot("Dark Ranger", "marksmanship-darkranger2.png"),
			screenshot("Sentinel", "marksmanship-sentinel.png"),
			screenshot("Sentinel", "marksmanship-sentinel2.png"),
		},
	},
	{
		Class: "Hunter",
		Spec:  "Survival",
		Role:  "DPS",
		Mechanics: []string{
			"Manage Focus so priority abilities stay available.",
			"Align major cooldowns with high-value windows.",
			"Maintain important buffs, debuffs, or effects.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Pack Leader", "survival-packleader.png"),
			screenshot("Sentinel", "survival-sentinel.png"),
		},
	},
	{
		Class: "Mage",
		Spec:  "Arcane",
		Role:  "DPS",
		Mechanics: []string{
			"Manage mana and charges so priority windows are not delayed.",
			"Align major cooldowns with high-value windows.",
			"Plan movement so priority casts are not delayed.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Spellslinger", "arcane-spellslinger.png"),
			screenshot("Sunfury", "arcane-sunfury.png"),
		},
	},
	{
		Class: "Mage",
		Spec:  "Fire",
		Role:  "DPS",
		Mechanics: []string{
			"Align major cooldowns with high-value windows.",
			"Maintain important buffs, debuffs, or effects.",
			"Plan movement so priority casts are not delayed.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Frostfire", "fire-frostfire.png"),
			screenshot("Sunfury", "fire-sunfury.png"),
		},
	},
	{
		Class: "Mage",
		Spec:  "Frost",
		Role:  "DPS",
		Mechanics: []string{
			"Align major cooldowns with high-value windows.",
			"Maintain important buffs, debuffs, or effects.",
			"Plan movement so priority casts are not delayed.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Frostfire", "frost-frostfire.png"),
			screenshot("Spellslinger", "frost-spellslinger.png"),
		},
	},
	{
		Class: "Monk",
		Spec:  "Brewmaster",
		Role:  "Tank",
		Mechanics: []string{
			"Maintain important buffs, debuffs, or effects.",
			"Plan defensive cooldowns around dangerous windows.",
			"Use self-healing tools when survivability requires it.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Master of Harmony", "brewmaster-masterofharmony.png"),
			screenshot("Master of Harmony", "brewmaster-masterofharmony2.png"),
			screenshot("Shado-Pan", "brewmaster-shadopan.png"),
			screenshot("Shado-Pan", "brewmaster-shadopan2.png"),
		},
	},
	{
		Class: "Monk",
		Spec:  "Mistweaver",
		Role:  "Healer",
		Mechanics: []string{
			"Plan healing globals around damage patterns and resource limits.",
			"Maintain important buffs, debuffs, or effects.",
			"Plan major cooldowns around dangerous windows.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Conduit of the Celestials", "mistweaver-conduit.png"),
			screenshot("Master of Harmony", "mistweaver-masterofharmony.png"),
		},
	},
	{
		Class: "Monk",
		Spec:  "Windwalker",
		Role:  "DPS",
		Mechanics: []string{
			"Manage Energy and Chi so priority abilities stay available.",
			"Align major cooldowns with high-value windows.",
			"Maintain important buffs, debuffs, or effects.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Conduit of the Celestials", "windwalker-conduit.png"),
			screenshot("Conduit of the Celestials", "windwalker-conduit2.png"),
			screenshot("Shado-Pan", "windwalker-shadopan.png"),
		},
	},
	{
		Class: "Paladin",
		Spec:  "Holy",
		Role:  "Healer",
		Mechanics: []string{
			"Generate and spend Holy Power efficiently.",
			"Plan healing globals around damage patterns and resource limits.",
			"Plan major cooldowns around dangerous windows.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Herald of the Sun", "holy-herald.png"),
			screenshot("Lightsmith", "holy-lightsmith.png"),
		},
	},
	{
		Class: "Paladin",
		Spec:  "Protection",
		Role:  "Tank",
		Mechanics: []string{
			"Generate and spend Holy Power efficiently.",
			"Plan defensive cooldowns around dangerous windows.",
			"Maintain important buffs, debuffs, or effects.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Lightsmith", "protection-lightsmith.png"),
			screenshot("Templar", "protection-templar.png"),
		},
	},
	{
		Class: "Paladin",
		Spec:  "Retribution",
		Role:  "DPS",
		Mechanics: []string{
			"Generate and spend Holy Power efficiently.",
			"Align major cooldowns with high-value windows.",
			"Maintain important buffs, debuffs, or effects.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Herald of the Sun", "retribution-herald.png"),
			screenshot("Templar", "retribution-templar.png"),
		},
	},
	{
		Class: "Priest",
		Spec:  "Discipline",
		Role:  "Healer",
		Mechanics: []string{
			"Plan healing globals around damage patterns and resource limits.",
			"Maintain important buffs, debuffs, or effects.",
			"Plan major cooldowns around dangerous windows.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Oracle", "discipline-oracle.png"),
			screenshot("Voidweaver", "discipline-voidweaver.png"),
		},
	},
	{
		Class: "Priest",
		Spec:  "Holy",
		Role:  "Healer",
		Mechanics: []string{
			"Plan healing globals around damage patterns and resource limits.",
			"Maintain important buffs, debuffs, or effects.",
			"Plan major cooldowns around dangerous windows.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Archon", "holy-archon.png"),
			screenshot("Oracle", "holy-oracle.png"),
		},
	},
	{
		Class: "Priest",
		Spec:  "Shadow",
		Role:  "DPS",
		Mechanics: []string{
			"Build and spend Insanity around priority windows.",
			"Maintain important buffs, debuffs, or effects.",
			"Align major cooldowns with high-value windows.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Archon", "shadow-archon.png"),
			screenshot("Voidweaver", "shadow-voidweaver.png"),
		},
	},
	{
		Class: "Rogue",
		Spec:  "Assassination",
		Role:  "DPS",
		Mechanics: []string{
			"Build and spend Combo Points efficiently.",
			"Manage Energy pooling and avoid starving key windows.",
			"Maintain important buffs, debuffs, or effects.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Deathstalker", "assassination-deathstalker.png"),
			screenshot("Fatebound", "assassination-fatebound.png"),
		},
	},
	{
		Class: "Rogue",
		Spec:  "Outlaw",
		Role:  "DPS",
		Mechanics: []string{
			"Build and spend Combo Points efficiently.",
			"Manage Energy pooling and avoid starving key windows.",
			"Align major cooldowns with high-value windows.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Fatebound", "outlaw-fatebound.png"),
			screenshot("Trickster", "outlaw-trickster.png"),
		},
	},
	{
		Class: "Rogue",
		Spec:  "Subtlety",
		Role:  "DPS",
		Mechanics: []string{
			"Build and spend Combo Points efficiently.",
			"Manage Energy pooling and avoid starving key windows.",
			"Align major cooldowns with high-value windows.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Deathstalker", "subtlety-deathstalker.png"),
			screenshot("Trickster", "subtlety-trickster.png"),
		},
	},
	{
		Class: "Shaman",
		Spec:  "Elemental",
		Role:  "DPS",
		Mechanics: []string{
			"Use Maelstrom generation and spenders efficiently.",
			"Maintain important buffs, debuffs, or effects.",
			"Align major cooldowns with high-value windows.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Farseer", "elemental-farseer.png"),
			screenshot("Farseer", "elemental-farseer2.png"),
			screenshot("Stormbringer", "elemental-stormbringer.png"),
			screenshot("Stormbringer", "elemental-stormbringer2.png"),
		},
	},
	{
		Class: "Shaman",
		Spec:  "Enhancement",
		Role:  "DPS",
		Mechanics: []string{
			"Use Maelstrom generation and spenders efficiently.",
			"Maintain important buffs, debuffs, or effects.",
			"Align major cooldowns with high-value windows.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Stormbringer", "enhancement-stormbringer.png"),
			screenshot("Totemic", "enhancement-totemic.png"),
		},
	},
	{
		Class: "Shaman",
		Spec:  "Restoration",
		Role:  "Healer",
		Mechanics: []string{
			"Plan healing globals around damage patterns and resource limits.",
			"Use Maelstrom generation and spenders efficiently.",
			"Plan major cooldowns around dangerous windows.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Farseer", "restoration-farseer.png"),
			screenshot("Totemic", "restoration-totemic.png"),
		},
	},
	{
		Class: "Warlock",
		Spec:  "Affliction",
		Role:  "DPS",
		Mechanics: []string{
			"Generate and spend Soul Shards efficiently.",
			"Maintain important buffs, debuffs, or effects.",
			"Align major cooldowns with high-value windows.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Hellcaller", "affliction-hellcaller.png"),
			screenshot("Soul Harvester", "affliction-soulharvester.png"),
		},
	},
	{
		Class: "Warlock",
		Spec:  "Demonology",
		Role:  "DPS",
		Mechanics: []string{
			"Generate and spend Soul Shards efficiently.",
			"Align major cooldowns with high-value windows.",
			"Maintain important buffs, debuffs, or effects.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Diabolist", "demonology-diabolist.png"),
			screenshot("Soul Harvester", "demonology-soulharvester.png"),
		},
	},
	{
		Class: "Warlock",
		Spec:  "Destruction",
		Role:  "DPS",
		Mechanics: []string{
			"Generate and spend Soul Shards efficiently.",
			"Align major cooldowns with high-value windows.",
			"Maintain important buffs, debuffs, or effects.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Diabolist", "destruction-diabolist.png"),
			screenshot("Hellcaller", "destruction-hellcaller.png"),
		},
	},
	{
		Class: "Warrior",
		Spec:  "Arms",
		Role:  "DPS",
		Mechanics: []string{
			"Build and spend Rage around priority windows.",
			"Maintain important buffs, debuffs, or effects.",
			"Align major cooldowns with high-value windows.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Colossus", "arms-colossus.png"),
			screenshot("Slayer", "arms-slayer.png"),
		},
	},
	{
		Class: "Warrior",
		Spec:  "Fury",
		Role:  "DPS",
		Mechanics: []string{
			"Build and spend Rage around priority windows.",
			"Maintain important buffs, debuffs, or effects.",
			"Align major cooldowns with high-value windows.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Mountain Thane", "fury-mountainthane.png"),
			screenshot("Slayer", "fury-slayer.png"),
		},
	},
	{
		Class: "Warrior",
		Spec:  "Protection",
		Role:  "Tank",
		Mechanics: []string{
			"Build and spend Rage around priority windows.",
			"Plan defensive cooldowns around dangerous windows.",
			"Maintain important buffs, debuffs, or effects.",
		},
		Screenshots: []manualSpecScreenshot{
			screenshot("Colossus", "protection-colossus.png"),
			screenshot("Mountain Thane", "protection-mountainthane.png"),
		},
	},
}

func specProfileKey(characterClass, characterSpec string) string {
	return normalizeSpecProfilePart(characterSpec) + " " + normalizeSpecProfilePart(characterClass)
}

func normalizeSpecProfilePart(value string) string {
	value = strings.ReplaceAll(value, "-", " ")
	return strings.ToLower(strings.Join(strings.Fields(value), " "))
}
