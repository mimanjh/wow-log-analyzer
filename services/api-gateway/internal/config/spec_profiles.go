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

type heroTalentRotation struct {
	HeroTalent string
	Steps      []string
}

type manualSpecProfile struct {
	Class          string
	Spec           string
	Role           string
	Mechanics      []string
	StatPriorities []string
	HeroTalents    []heroTalentRotation
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
	sections := make([]SpecGuideSection, 0, len(entry.HeroTalents))
	for _, ht := range entry.HeroTalents {
		steps := make([]SpecGuideStep, 0, len(ht.Steps))
		for _, step := range ht.Steps {
			steps = append(steps, SpecGuideStep{Text: step})
		}
		sections = append(sections, SpecGuideSection{
			Context:    "Rotation",
			HeroTalent: ht.HeroTalent,
			Steps:      steps,
		})
	}

	return SpecProfile{
		Label:          entry.Spec + " " + entry.Class,
		Role:           entry.Role,
		StatPriorities: entry.StatPriorities,
		KeyMechanics:   entry.Mechanics,
		Rotation:       sections,
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
		StatPriorities: []string{"Haste", "Mastery", "Critical Strike", "Versatility"},
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Deathbringer",
				Steps: []string{
					"Reaper's Mark on cooldown.",
					"Death Strike if you need the healing or are above 75 Runic Power.",
					"Maintain at least 5 charges of Bone Shield.",
					"Blood Boil if Blood Plague from Dancing Rune Weapon is not on your target.",
					"Death and Decay if you have a Crimson Scourge proc.",
					"Heart Strike to spend runes.",
					"If you run out of runes, cast Blood Boil.",
				},
			},
			{
				HeroTalent: "San'layn",
				Steps: []string{
					"Death Strike if you need the healing or are above 75 Runic Power.",
					"You also need 45 RP — two to three full Heart Strikes — to overcap Runic Power.",
					"Maintain at least 5 charges of Bone Shield.",
					"Blood Boil if the Blood Plague copy from Dancing Rune Weapon is not currently on your target.",
					"Death and Decay if you have a Crimson Scourge proc.",
					"Fill with Vampiric Strike.",
					"If you run out of runes, cast Blood Boil.",
				},
			},
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
		StatPriorities: []string{"Haste", "Critical Strike", "Mastery", "Versatility"},
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Deathbringer",
				Steps: []string{
					"Empower Rune Weapon if you have 2 charges.",
					"Pillar of Frost.",
					"Breath of Sindragosa.",
					"Frostwyrm's Fury.",
					"Obliterate if you have 2 Killing Machine stacks or Exterminate.",
					"Howling Blast with Rime.",
					"Frost Strike to avoid Runic Power waste.",
					"Obliterate with Killing Machine.",
					"Empower Rune Weapon to generate Killing Machine.",
					"Frost Strike.",
					"Obliterate without Killing Machine.",
					"Howling Blast.",
				},
			},
			{
				HeroTalent: "Rider of the Apocalypse",
				Steps: []string{
					"Empower Rune Weapon if you have 2 charges.",
					"Pillar of Frost.",
					"Breath of Sindragosa.",
					"Frostwyrm's Fury.",
					"Obliterate if you have 2 Killing Machine stacks.",
					"Howling Blast with Rime.",
					"Frost Strike to avoid Runic Power waste.",
					"Obliterate with Killing Machine.",
					"Empower Rune Weapon to generate Killing Machine.",
					"Frost Strike.",
					"Obliterate without Killing Machine.",
					"Howling Blast.",
				},
			},
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
		StatPriorities: []string{"Mastery", "Critical Strike", "Haste", "Versatility"},
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Rider of the Apocalypse",
				Steps: []string{
					"Dark Transformation on cooldown.",
					"Putrefy if Dark Transformation is active.",
					"Death Coil if you have a Sudden Doom proc.",
					"Festering Strike if you have no Lesser Ghoul stacks.",
					"Scourge Strike if you have no Lesser Ghoul stacks.",
					"Death Coil.",
				},
			},
			{
				HeroTalent: "San'layn",
				Steps: []string{
					"Dark Transformation on cooldown.",
					"Putrefy if Dark Transformation is active.",
					"Death Coil if you have a Sudden Doom proc.",
					"Festering Strike if you have no Lesser Ghoul stacks.",
					"Scourge Strike if you have Lesser Ghoul stacks.",
					"Death Coil.",
				},
			},
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
		StatPriorities: []string{"Haste", "Mastery", "Critical Strike", "Versatility"},
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Annihilator",
				Steps: []string{
					"Reap at 76 Fury & 4 souls, or 60 Fury & 10 souls with Moment of Craving.",
					"Void Metamorphosis.",
					"Void Ray.",
					"Reap when it will proc Voidfall.",
					"Soul Immolation.",
					"Consume.",
				},
			},
			{
				HeroTalent: "Void-Scarred",
				Steps: []string{
					"Voidblade if you are about to enter Void Metamorphosis.",
					"The Hunt if you are about to enter Void Metamorphosis.",
					"Reap at 74 Fury or with a Moment of Craving proc.",
					"Void Metamorphosis.",
					"Void Ray.",
					"Soul Immolation.",
					"Consume.",
				},
			},
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
		StatPriorities: []string{"Haste", "Versatility"},
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Aldrachi Reaver",
				Steps: []string{
					"If Inertia buff is up, cast Felblade or Fel Rush to proc it.",
					"The Hunt.",
					"Death Sweep.",
					"Immolation Aura if capped on charges.",
					"Vengeful Retreat.",
					"Vengeful Retreat if Eye Beam is off cooldown.",
					"Eye Beam.",
					"Essence Break while in Metamorphosis.",
					"Metamorphosis.",
					"Blade Dance.",
					"Reaver's Glaive.",
					"Annihilation.",
					"Chaos Strike.",
					"Immolation Aura.",
					"Felblade.",
					"Fel Rush when no other abilities are available.",
					"Throw Glaive is not cast due to Screaming Brutality using the charges for free.",
				},
			},
			{
				HeroTalent: "Fel-Scarred",
				Steps: []string{
					"If Inertia buff is up, cast Felblade or Fel Rush to proc it.",
					"The Hunt.",
					"Death Sweep.",
					"Immolation Aura if capped on charges.",
					"Vengeful Retreat.",
					"Vengeful Retreat if Eye Beam is off cooldown.",
					"Eye Beam.",
					"Essence Break while in Metamorphosis.",
					"Metamorphosis.",
					"Blade Dance.",
					"Annihilation.",
					"Chaos Strike.",
					"Immolation Aura.",
					"Felblade.",
					"Throw Glaive or Fel Rush when no other abilities are available.",
				},
			},
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
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Aldrachi Reaver",
				Steps: []string{
					"Use Internal Strike off GCD if you are at or near 2 charges and will not need it for movement soon.",
					"Metamorphosis.",
					"Use Fracture if you are at or near 2 charges.",
					"Use Fracture with Rending Strike on your priority target to proc Reaver's Mark and generate Souls and Fury.",
					"Use Spirit Bomb with 4+ Souls if Fiery Brand is about to run out.",
					"Use Spirit Bomb if Fiery Brand is about to run out.",
					"Use Fiery Brand if the debuff is not currently active.",
					"Use Spirit Bomb with 6+ Souls.",
					"Use Soul Cleave with Glaive Flurry to proc Fury of the Aldrachi and trigger Rite of the Fight.",
					"Use Sigil of Spite to activate Art of the Glaive if you do not already have a Reaver's Glaive available.",
					"Use Reaver's Glaive to activate the Rending Strike and Glaive Flurry buffs.",
					"Use Immolation Aura.",
					"Use Sigil of Flame.",
					"Use Soul Carver.",
					"Use Soul Cleave with 1 or more Souls.",
					"Use Felblade if you have at least 50 Fury.",
					"Spend Fury with Soul Cleave at 0 Souls.",
					"Use Fracture if you want to cap Fury or Souls.",
					"Use Throw Glaive for filler or when kiting.",
				},
			},
			{
				HeroTalent: "Annihilator",
				Steps: []string{
					"Use Internal Strike off GCD if you are at or near 2 charges and will not need it for movement soon.",
					"Maintain 75 Fury if you are at 2 stacks of Voidtall or Metamorphosis is about to come off cooldown.",
					"Use Metamorphosis if you will not overcap duration and Spirit Bomb has more than 10s remaining on its CD.",
					"Use Fracture if you are at or near 2 charges.",
					"Use Spirit Bomb with 4+ Souls if Fiery Brand is about to run out.",
					"Use Spirit Bomb if Fiery Brand is about to run out.",
					"Use Spirit Bomb with 5+ Souls. If this activated Voidtall, cast Soul Cleave on the next GCD.",
					"Use Immolation Aura.",
					"Use Sigil of Flame.",
					"Use Soul Carver if you will not overcap Souls.",
					"Use Sigil of Spite if you will not overcap Souls.",
					"Use Fiery Brand if the debuff is not currently active.",
					"Use Soul Cleave with 1 or more Souls.",
					"Use Felblade if you have at least 50 Fury.",
					"Spend Fury with Soul Cleave at 0 Souls.",
					"Use Fracture if you won't cap Fury or Souls.",
					"Use Throw Glaive for filler or when kiting.",
				},
			},
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
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Keeper of the Grove",
				Steps: []string{
					"Apply Moonfire and refresh within pandemic.",
					"Apply Sunfire and refresh within pandemic or if your next cast is Force of Nature.",
					"Press Fury of Elune during an Eclipse or right before pressing Force of Nature.",
					"Enter Solar Eclipse if you are sitting on 2 charges of the cooldown.",
					"Press Force of Nature if you are not currently in an Eclipse and your next cast is either Solar Eclipse or Incarnation: Chosen of Elune.",
					"Press Celestial Alignment if you just used Force of Nature and you will have another charge available before the Cosmos Spirits is off cooldown.",
					"Convoke the Spirits if you are below 40 and Force of Nature is active.",
					"Enter Solar Eclipse with enough resources to spend all Ascendant Eclipses in your next 3 GCDs.",
					"Press Force of Nature to spend at the start of the cooldown.",
					"Starsurge to prevent capping or to consume Ascendant Eclipses at the start of the cooldown.",
				},
			},
			{
				HeroTalent: "Elune's Chosen",
				Steps: []string{
					"Apply Moonfire and refresh within pandemic.",
					"Apply Sunfire and refresh within pandemic.",
					"Press Incarnation: Chosen of Elune if you're not in Lunar Eclipse and Fury of Elune is off cooldown.",
					"Press Fury of Elune off cooldown.",
					"Enter Lunar Eclipse whenever Starlord expires and you have any procs of Starsurge or Touch of the Cosmos available.",
					"Enter Lunar Eclipse if you are above 90% and you're either about to overcap on charges or Fury of Elune is also available.",
					"Starlord to consume Touch the Cosmos.",
					"Starsurge to consume Touch the Cosmos.",
					"Starsurge to prevent capping on or to consume Ascendant Eclipses at the start of Lunar Eclipse.",
					"Starfire to generate.",
				},
			},
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
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Druid of the Claw",
				Steps: []string{
					"Ferocious Bite with Apex Predator's Craving procs.",
					"Rip: if you have 5 Combo Points, Rip is missing or in pandemic, and you have Tiger's Fury active.",
					"If you have 5 Combo Points and 50 energy, Rip is active, and Tiger's Fury will not be active before the dot expires.",
					"Ferocious Bite if you have 5 Combo Points and 50 energy, Rip is active, and Berserk is not active.",
					"Berserk: sync with the Spirits with Tiger's Fury, these should be synced.",
					"Feral Frenzy with Tiger's Fury active.",
					"When Convoke is coming off CD with 3 seconds left, dump energy so you have as little as possible.",
					"Feral Frenzy on cooldown.",
					"Rake if it is missing on the target or in pandemic.",
					"Moonfire has 2 seconds or less of duration and Tiger's Fury is not ready.",
					"Shred to generate Combo Points.",
				},
			},
			{
				HeroTalent: "Wildstalker",
				Steps: []string{
					"Ferocious Bite with Apex Predator's Craving procs.",
					"Rip: if you have 5 combo points, Rip is missing or in pandemic, and you have Tiger's Fury active.",
					"If you have 5 combo points, Rip is missing or in pandemic, and Tiger's Fury will not be active before the dot expires.",
					"Ferocious Bite if you have 5 Combo Points and 50 energy, Rip is active, and Berserk is not active.",
					"Berserk and Convoke the Spirits with Tiger's Fury, these should be synced.",
					"Feral Frenzy with Tiger's Fury active.",
					"When Convoke is coming off CD with 3 seconds left, dump energy so you have as little as possible.",
					"Feral Frenzy on cooldown.",
					"Rake if it is missing on the target or in pandemic.",
					"Moonfire has 2 seconds or less of duration and Tiger's Fury is not ready.",
					"Shred to Generate Combo Points.",
				},
			},
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
		StatPriorities: []string{"Mastery", "Versatility", "Critical Strike"},
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Druid of the Claw",
				Steps: []string{
					"Maintain Moonfire on your primary target.",
					"Maintain 3-5 stacks of Thrash.",
					"Red Moon on cooldown.",
					"Mangle on cooldown.",
					"Thrash on cooldown.",
					"Spend Rage on either Maul (offensively) or Ironfur (defensively).",
					"Frenzied Regeneration if your health dips low.",
					"Moonfire with Galactic Guardian procs.",
					"Use your cooldowns Barkskin/Incarnation: Guardian of Ursoc as frequently as possible.",
					"Swipe if you have nothing else to press.",
				},
			},
			{
				HeroTalent: "Elune's Chosen",
				Steps: []string{
					"Maintain Moonfire on your primary target.",
					"Maintain 3-5 stacks of Thrash.",
					"Red Moon on cooldown.",
					"Mangle on cooldown.",
					"Thrash on cooldown.",
					"Spend Rage on either Maul (offensively) or Ironfur (defensively).",
					"Frenzied Regeneration if your health dips low.",
					"Moonfire with Galactic Guardian procs.",
					"Use your cooldowns Barkskin/Incarnation: Guardian of Ursoc as frequently as possible.",
					"Swipe if you have nothing else to press.",
				},
			},
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
		StatPriorities: []string{"Versatility", "Critical Strike"},
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Keeper of the Grove",
				Steps: []string{
					"Keep Efflorescence active as frequently as possible.",
					"Lifebloom is crucial to keep up at all times; refreshing is automated with Lifeshrowing.",
					"Swiftmend and Wild Growth on cooldown if the raid has taken any damage. Follow Swiftmend with Rejuvenation.",
					"Rotate Rejuvenation and Regrowth; cast Rejuvenation until Abundance stacks are high then swap to Regrowth until they are low again.",
					"Fill with Wrath to restore mana via Master Shapeshifter.",
				},
			},
			{
				HeroTalent: "Wildstalker",
				Steps: []string{
					"Keep Efflorescence active as frequently as possible.",
					"Lifebloom is crucial to keep up at all times; refreshing is automated with Lifeshrowing.",
					"Swiftmend and Wild Growth on cooldown if the raid has taken any damage. Follow Swiftmend with Rejuvenation.",
					"Rotate Rejuvenation and Regrowth; cast Rejuvenation until Abundance stacks are high then swap to Regrowth until they are low again.",
					"Fill with Wrath to restore mana via Master Shapeshifter.",
				},
			},
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
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Chronowarden",
				Steps: []string{
					"Maintain Prescience on allied DPS players.",
					"Ebon Might if a few seconds or less remain on the active buff duration.",
					"Breath of Eons on cooldown.",
					"Tip the Scales on cooldown.",
					"Fire Breath at Max Rank whenever possible.",
					"Upheaval at Rank 1, unless increased radius is needed to hit relevant targets.",
					"Only if talented: Cast Time Skip on cooldown.",
					"Living Flame as filler in nearly all situations.",
					"Azure Strike as backup filler that can be cast while moving, or used to slow enemies.",
					"Maintain Blistering Scales on a tank.",
				},
			},
			{
				HeroTalent: "Scalecommander",
				Steps: []string{
					"Maintain Prescience on allied DPS players.",
					"Ebon Might if a few seconds or less remain on the active buff duration.",
					"Breath of Eons on cooldown.",
					"Tip the Scales on cooldown.",
					"Fire Breath at Max Rank whenever possible.",
					"Upheaval at Rank 1, unless increased radius is needed to hit relevant targets.",
					"Eruption.",
					"Living Flame as filler in nearly all situations.",
					"Azure Strike as backup filler that can be cast while moving, or used to slow enemies.",
					"Maintain Blistering Scales on a tank.",
				},
			},
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
		StatPriorities: []string{"Versatility", "Critical Strike", "Haste", "Mastery"},
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Flameshaper",
				Steps: []string{
					"Dragonrage.",
					"Tip the Scales if Eternity Surge is ready. Eternity Surge with Tip the Scales.",
					"Fire Breath to maintain the DoT debuff.",
					"Eternity Surge Rank 1.",
					"Channel Disintegrate.",
					"Azure Sweep.",
					"Living Flame with Leaping Flames or Burnout.",
					"Living Flame as filler.",
					"Azure Strike if needed for movement.",
				},
			},
			{
				HeroTalent: "Scalecommander",
				Steps: []string{
					"Hover if capped and Deep Breath is ready soon.",
					"Deep Breath.",
					"Deep Breath when Strafing Run is about to expire.",
					"Disintegrate.",
					"Tip the Scales.",
					"Eternity Surge Rank 1.",
					"Fire Breath Rank 1.",
					"Azure Sweep if not capped on Essence Burst.",
					"Channel Disintegrate with Mass Disintegrate.",
					"Mass Disintegrate.",
					"Living Flame with Leaping Flames or Burnout.",
					"Living Flame as filler.",
					"Azure Strike if needed for movement.",
				},
			},
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
		StatPriorities: []string{"Mastery", "Critical Strike", "Haste", "Versatility"},
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Chronowarden",
				Steps: []string{
					"Temporal Anomaly on cooldown.",
					"Dream Breath often; cast after consuming Echo buffs to avoid overwriting them.",
					"Echo if you have Essence available and either no Essence Burst buff or you are at two stacks of Echo.",
					"Emerald Blossom if you have an Essence Burst.",
					"Consume Echo buffs with Meitrha's Blessing (common) or Verdant Embrace (rare).",
					"Living Flame on the boss to farm Essence Burst and contribute a little damage.",
				},
			},
			{
				HeroTalent: "Flameshaper",
				Steps: []string{
					"Temporal Anomaly on cooldown.",
					"Keep one Dream Breath charge on cooldown; cast the second when there is dangerous boss damage.",
					"Echo if you have Essence available and either no Essence Burst buff or you are following an Emerald Blossom cast.",
					"Emerald Blossom if you have an Essence Burst.",
					"Consume Echo buffs with Reversion (common) or Verdant Embrace (rare).",
					"Living Flame on the boss to farm Essence Burst and contribute a little damage.",
				},
			},
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
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Dark Ranger",
				Steps: []string{
					"Bestial Wrath.",
					"Kill Command if you have the Nature's Ally buff active.",
					"Black Arrow if Withering Fire is active.",
					"Wailing Arrow.",
					"Black Arrow.",
					"Cobra Shot.",
				},
			},
			{
				HeroTalent: "Pack Leader",
				Steps: []string{
					"Barbed Shot if Bestial Wrath is about to come off cooldown in the next 3 seconds.",
					"Bestial Wrath.",
					"Kill Command if you have a Howl of the Pack Leader buff or Nature's Ally active.",
					"Barbed Shot.",
					"Cobra Shot.",
				},
			},
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
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Dark Ranger",
				Steps: []string{
					"Black Arrow at your primary Precise Shots spender; always cast Black Arrow immediately after every Aimed Shot or Rapid Fire if you have 2 charges and no Precise Shots to spend.",
					"Aimed Shot if it is about to reach 2 charges.",
					"During Trueshot, cast Aimed Shot if you have a Black Arrow proc available but no Precise Shots to spend.",
					"Pair Trueshot with Rapid Fire; carry a fresh Bulletstorm into your cooldown.",
					"Rapid Fire on cooldown.",
					"Wailing Arrow if you do not have a Black Arrow proc available.",
				},
			},
			{
				HeroTalent: "Sentinel",
				Steps: []string{
					"Trueshot on cooldown; pair with Rapid Fire and Aimed Shot.",
					"Aimed Shot as your primary spender during Precise Shots.",
					"Rapid Fire on cooldown.",
					"Multishot to apply Trick Shots on 3+ targets.",
					"Arcane Shot or Steady Shot as filler.",
				},
			},
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
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Pack Leader",
				Steps: []string{
					"Hunter's Mark if it is not yet active.",
					"Kill Command whenever you have any Howl of the Pack Leader beads available to transform, without waiting any Tip of the Spear stacks.",
					"Kill Command before entering Takedown to maximize Tip of the Spear stacks.",
					"Use Tip of the Spear spenders until you run out of stacks or Focus, then cast Takedown.",
					"Takedown.",
					"Flanking Pitch.",
					"Wildfire Bomb if you have a Tip of the Spear stack to spend.",
					"Boomerang Trick if you have a Tip of the Spear stack to spend.",
					"Raptor Swipe if you have a Tip of the Spear stack to spend.",
					"Raptor Strike with or without a Tip of the Spear to spend.",
					"Kill Command.",
				},
			},
			{
				HeroTalent: "Sentinel",
				Steps: []string{
					"Hunter's Mark if it is not yet active.",
					"Kill Command whenever you run out of Tip of the Spear stacks.",
					"Kill Command once before entering Takedown to maximize Tip of the Spear stacks.",
					"Takedown.",
					"Moonlight Chakram.",
					"Flanking Pitch.",
					"Raptor Swipe if you have a Tip of the Spear stack to spend.",
					"Raptor Strike with or without a Tip of the Spear stack.",
				},
			},
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
		StatPriorities: []string{"Mastery", "Haste", "Critical Strike", "Versatility"},
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Spellslinger",
				Steps: []string{
					"Arcane Surge.",
					"Touch of the Magi.",
					"Arcane Pulse when you have 3 or more targets.",
					"Arcane Blast.",
					"Arcane Barrage if you run out of mana.",
				},
			},
			{
				HeroTalent: "Sunfury",
				Steps: []string{
					"Arcane Surge.",
					"Touch of the Magi.",
					"Arcane Pulse when you have 3 or more targets.",
					"Arcane Blast.",
					"Arcane Barrage if you run out of mana.",
				},
			},
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
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Frostfire",
				Steps: []string{
					"Meteor in the last 8 seconds of Combustion.",
					"Meteor if it will land during Combustion.",
					"Pyroblast with Hot Streak.",
					"Pyroblast again if you just spent Hot Streak at the end of a Scorch cast.",
					"Hard-cast Pyroblast with Pyroclasm.",
					"Fire Blast to generate Hot Streak.",
					"Scorch with Heat Shimmer and Heating Up.",
					"Scorch as filler.",
					"Frostfire Bolt as filler.",
				},
			},
			{
				HeroTalent: "Sunfury",
				Steps: []string{
					"Meteor if it will land during Combustion and you have fewer than 2 Pyroclasm stacks.",
					"Meteor in the last 8 seconds of Combustion.",
					"Pyroblast with Hot Streak.",
					"Pyroblast again if you just spent Hot Streak at the end of a Scorch cast.",
					"Hard-cast Pyroblast with Pyroclasm.",
					"Fire Blast to generate Hot Streak.",
					"Scorch with Heat Shimmer and Heating Up.",
					"Scorch as filler.",
					"Fireball as filler.",
				},
			},
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
		StatPriorities: []string{"Haste", "Versatility"},
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Frostfire",
				Steps: []string{
					"Flurry if Brain Freeze is active and Thermal Void is not active.",
					"Frozen Orb.",
					"Glacial Spike.",
					"Comet Storm.",
					"Ice Lance if Fingers of Frost is active.",
					"Ice Lance if Freezing at 10 or more stacks.",
					"Flurry.",
					"Ray of Frost.",
					"Frostfire Bolt.",
				},
			},
			{
				HeroTalent: "Spellslinger",
				Steps: []string{
					"Flurry if Brain Freeze is active and Thermal Void is not active.",
					"Frozen Orb.",
					"Glacial Spike.",
					"Ice Lance if Fingers of Frost is active.",
					"Ice Lance if Freezing at 6 or more stacks.",
					"Ray of Frost if you have 3 or fewer icicles.",
					"Flurry.",
					"Frostbolt.",
				},
			},
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
		StatPriorities: []string{"Versatility", "Critical Strike", "Mastery", "Haste"},
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Shado-Pan",
				Steps: []string{
					"Use Black Ox Brew if minimal waste is possible.",
					"Execute an enemy with Touch of Death, when allowed.",
					"Use Blackout Kick to trigger Blackout Combo.",
					"Only if taking minimal damage, activate Celestial Brew/Celestial Infusion.",
					"Activate Invoke Niuzao, the Black Ox.",
					"Consume your Blackout Combo buff with Tiger Palm.",
					"Use your first charge of Keg Smash or an active Empty Barrel from Bring Me Another.",
					"Use Breath of Fire to gain or refresh the Charred Passions buff.",
					"Use Exploding Keg, ideally with Rushing Jade Wind active beforehand.",
					"Use your second charge of Keg Smash.",
					"Use Rushing Jade Wind.",
					"Use Tiger Palm when there are no other abilities to press.",
				},
			},
			{
				HeroTalent: "Master of Harmony",
				Steps: []string{
					"Use Black Ox Brew if minimal waste is possible.",
					"Execute an enemy with Touch of Death, when allowed.",
					"Use Blackout Kick to trigger Blackout Combo.",
					"Chi Burst.",
					"Only to deal additional damage, consume one charge of Celestial Brew/Celestial Infusion if at two charges.",
					"Activate Invoke Niuzao, the Black Ox.",
					"Consume your Blackout Combo buff with Tiger Palm.",
					"Consume an active Empty Barrel from Bring Me Another.",
					"Use Breath of Fire to gain or refresh the Charred Passions buff.",
					"Use Exploding Keg, ideally with Rushing Jade Wind active beforehand.",
					"Use a charge of Keg Smash.",
					"Use Rushing Jade Wind.",
					"Use Tiger Palm when there are no other abilities to press.",
				},
			},
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
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Conduit of the Celestials",
				Steps: []string{
					"Vivify to keep someone alive that would otherwise die without your intervention.",
					"Debuff enemies with Mystic Touch.",
					"Restoral when it is assigned.",
					"Heart of the Jade Serpent abilities while active: Priority Renewing Mist at 3 charges.",
					"Rushing Wind Kick.",
					"Vivify if at 2 Zen Pulse stacks.",
					"Enveloping Mist if at 2 Spiritfront stacks.",
					"Thunder Focus Tea.",
					"Renewing Mist.",
					"Life Cocoon.",
				},
			},
			{
				HeroTalent: "Master of Harmony",
				Steps: []string{
					"Vivify to keep someone alive that would otherwise die without your intervention.",
					"Debuff enemies with Mystic Touch.",
					"Restoral when it is assigned.",
					"Mana Tea if at 20 Mana Tea stacks.",
					"Rising Sun Kick.",
					"Awaken Yu'lon, the Jade Serpent when it is assigned.",
					"Enveloping Mist if at 2 Spiritfront stacks.",
					"Vivify if at 6 Harmonic Surge stacks.",
					"Thunder Focus Tea.",
					"Renewing Mist.",
					"Vivify.",
					"Mana Tea.",
					"Blackout Kick to fill.",
					"Tiger Palm to fill.",
				},
			},
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
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Conduit of the Celestials",
				Steps: []string{
					"Fists of Fury if Heart of the Jade Serpent tasks less than 1 second.",
					"Touch of Death.",
					"Celestial Conduit if no active Heart of the Jade Serpent.",
					"Whirling Dragon Punch.",
					"Tiger Palm if less than 4 Chi, less than 2 stacks of Combo Breaker, AND about to cap energy.",
					"Strike of the Windlord.",
					"Rushing Wind Kick.",
					"Spinning Crane Kick with less than 4 seconds remaining on Dance of Chi-Ji AND not at 2 stacks of Combo Breaker.",
					"Rising Sun Kick.",
					"Zenith Stomp.",
					"Tiger Palm if not enough Chi for above.",
					"Blackout Kick with Combo Breaker.",
					"Slicing Winds.",
					"Spinning Crane Kick with Dance of Chi-Ji proc.",
					"Blackout Kick.",
					"Tiger Palm.",
				},
			},
			{
				HeroTalent: "Shado-Pan",
				Steps: []string{
					"Touch of Death.",
					"Whirling Dragon Punch.",
					"Tiger Palm if less than 4 Chi, less than 2 stacks of Combo Breaker, AND about to cap energy.",
					"Strike of the Windlord.",
					"Fists of Fury.",
					"Rushing Wind Kick.",
					"Spinning Crane Kick with less than 4 seconds remaining on Dance of Chi-Ji AND not at 2 stacks of Combo Breaker.",
					"Rising Sun Kick.",
					"Tiger Palm if not enough Chi for the above.",
					"Blackout Kick with Combo Breaker or Shado-Pan active.",
					"Slicing Winds.",
					"Spinning Crane Kick with Dance of Chi-Ji.",
					"Blackout Kick.",
					"Tiger Palm.",
				},
			},
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
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Herald of the Sun",
				Steps: []string{
					"Divine Toll.",
					"Holy Bulwark or Sacred Weapon.",
					"Holy Light as needed and as mana allows.",
					"Spend Infusion of Light procs on Judgment and Flash of Light.",
					"Holy Shock.",
					"During Avenging Crusader cast Hammer of Wrath.",
					"During Avenging Crusader cast Crusader Strike.",
					"Flash of Light. Only cast Judgment if you need to move.",
					"Judgment.",
					"Flash of Light.",
				},
			},
			{
				HeroTalent: "Lightsmith",
				Steps: []string{
					"Divine Toll.",
					"Holy Bulwark or Sacred Weapon.",
					"Holy Light as needed and as mana allows.",
					"Spend Infusion of Light procs on Judgment and Flash of Light.",
					"Holy Shock.",
					"During Avenging Crusader cast Hammer of Wrath.",
					"With an Awakening proc, cast Judgment.",
					"During Avenging Crusader cast Crusader Strike.",
					"Flash of Light. Only cast Judgment if you need to move.",
					"Judgment.",
					"Flash of Light.",
				},
			},
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
		StatPriorities: []string{"Mastery", "Critical Strike", "Haste", "Versatility"},
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Templar",
				Steps: []string{
					"Avenging Wrath on cooldown.",
					"Judgment on cooldown.",
					"Shield of the Righteous when you have 3-5 Holy Power or if it's free. Try not to cap Holy Power.",
					"Divine Toll if you have 0 Holy Power. Try to use this on cooldown as much as possible.",
					"Blessed Hammer/Hammer of the Righteous on cooldown.",
					"Word of Glory if your health drops below 50% or to top yourself off if you are in danger of a big hit.",
					"Consecration as a filler.",
				},
			},
			{
				HeroTalent: "Lightsmith",
				Steps: []string{
					"Avenging Wrath on cooldown.",
					"Sacred Weapon if not inside Avenging Wrath.",
					"Shield of the Righteous when you have 3-5 Holy Power or if it's free. Try not to cap Holy Power.",
					"Blessed Hammer on cooldown.",
					"Judgment on cooldown.",
					"Divine Toll if you have 0 Holy Power. Try to use this on cooldown as much as possible.",
					"Blessed Hammer/Hammer of the Righteous on cooldown.",
					"Word of Glory if your health drops below 50% or to top yourself off if you are in danger of a big hit.",
					"Consecration as a filler.",
				},
			},
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
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Herald of the Sun",
				Steps: []string{
					"Avenging Wrath.",
					"Execution Sentence.",
					"Final Verdict with 5 Holy Power.",
					"Wake of Ashes.",
					"Divine Toll.",
					"Hammer of Wrath with an Art of War proc.",
					"Blade of Justice with an Art of War proc.",
					"Final Verdict.",
					"Hammer of Wrath.",
					"Blade of Justice.",
					"Judgment.",
					"Templar Strike / Templar Slash.",
				},
			},
			{
				HeroTalent: "Templar",
				Steps: []string{
					"Avenging Wrath.",
					"Execution Sentence.",
					"Hammer of Light if it's castable after using Wake of Ashes.",
					"Hammer of Light with a Light's Deliverance proc when Avenging Wrath or Execution Sentence is up but will end within a few seconds.",
					"Final Verdict with 5 Holy Power.",
					"Divine Toll.",
					"Blade of Justice with an Art of War proc.",
					"Final Verdict.",
					"Blade of Justice.",
					"Hammer of Wrath.",
					"Judgment.",
				},
			},
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
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Oracle",
				Steps: []string{
					"Maintain/Spread Shadow Word: Pain, especially on targets reaching execute level.",
					"Void Shield any time it is available.",
					"Watch for Shadow Mend procs and weave them in as needed for high spot-healing.",
					"Use Evangelism to apply Atonements; Power Word: Radiance to refresh everyone or defensive Penance for individual atonements.",
					"Shadow Word: Death on sub 20% hp targets or for a damage global on the move.",
					"Penance defensively on allies.",
					"Mind Blast.",
					"Pain Smite and use Penance on cooldown.",
					"Use Power Word: Shields as needed to supplement.",
				},
			},
			{
				HeroTalent: "Voidweaver",
				Steps: []string{
					"Maintain/Spread Shadow Word: Pain, especially on targets reaching execute level.",
					"Void Shield any time it is available.",
					"Watch for Shadow Mend procs and weave them in as needed for high spot-healing.",
					"Use Evangelism to apply Atonements; Power Word: Radiance to refresh everyone or defensive Penance for individual atonements.",
					"Shadow Word: Death on sub 20% hp targets or for a damage global on the move.",
					"Penance defensively on allies.",
					"Mind Blast.",
					"Penance on cooldown.",
					"Use Power Word: Shields as needed to supplement.",
				},
			},
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
		StatPriorities: []string{"Critical Strike", "Versatility", "Mastery", "Haste"},
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Archon",
				Steps: []string{
					"Halo as raid damage begins; keep it expanding and contracting as damage hits.",
					"Apotheosis to reset the cooldown of your Holy Words.",
					"Holy Word: Serenity for single target healing on cooldown.",
					"Consume any Surge of Light procs on Prayer of Healing whenever possible; you can hold up to two stacks.",
					"Keep Prayer of Mending on cooldown.",
					"Flash Heals to spot-heal as needed.",
					"Smite during downtime to farm for Surge of Light procs.",
					"Divine Hymn for raid-wide damage.",
				},
			},
			{
				HeroTalent: "Oracle",
				Steps: []string{
					"Halo as raid damage begins; keep it expanding and contracting as damage hits.",
					"Apotheosis to reset the cooldown of your Holy Words.",
					"Holy Word: Serenity for single target healing on cooldown.",
					"Consume any Surge of Light procs on Prayer of Healing whenever possible; you can hold up to two stacks.",
					"Keep Prayer of Mending on cooldown.",
					"Flash Heals to spot-heal as needed.",
					"Smite during downtime to farm for Surge of Light procs.",
					"Divine Hymn for raid-wide damage.",
				},
			},
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
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Archon",
				Steps: []string{
					"Shadow Word: Pain if it is not active.",
					"Maintain Vampiric Touch via Tentacle Slam or hard cast.",
					"Halo.",
					"Voidform to enter Voidform.",
					"Void Valley.",
					"Spend on Shadow Word: Madness if it is not active, about to fall off, or you are about to cap.",
					"Mind Flay: Insanity if Shadow Word: Madness is active.",
					"Shadow Word: Madness if Voidform is active.",
					"Tentacle Slam if there will be no additional targets to DoT or upcoming movement.",
					"Shadow Word: Death if the target is below 20% HP.",
					"Mind Flay, interrupting as soon as anything of higher priority becomes available.",
				},
			},
			{
				HeroTalent: "Voidweaver",
				Steps: []string{
					"Shadow Word: Pain to maintain it.",
					"Maintain Vampiric Touch via Tentacle Slam or hard cast.",
					"Voidform to enter Voidform.",
					"Power Infusion.",
					"Shadow Word: Death if the target has an absorb shield.",
					"Void Blast if Shadow Word: Madness is active and Entropic Rift is about to expire.",
					"Void Valley.",
					"Spend on Shadow Word: Madness if not active, about to fall off, or Entropic Rift is active.",
					"Void Torrent to activate Entropic Rift.",
					"Mind Blast.",
					"Shadow Word: Madness.",
					"Tentacle Slam if there will be no additional targets to DoT or upcoming movement.",
					"Shadow Word: Death if the target is below 20% HP.",
					"Mind Flay, interrupting as soon as anything of higher priority becomes available.",
				},
			},
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
		StatPriorities: []string{"Mastery", "Versatility"},
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Deathstalker",
				Steps: []string{
					"Maintain Garrote.",
					"Mutilate until having 5 or more combo points when Darkest Night is active.",
					"Maintain Rupture by casting at 5 or more combo points.",
					"Vanish followed by Garrote to apply Improved Garrote; line up with Deathmark.",
					"Deathmark on cooldown.",
					"Kingsbane on cooldown, immediately after Deathmark when applicable.",
					"Envenom at 5 or more combo points, unless at 80 stacks of Implacable.",
				},
			},
			{
				HeroTalent: "Fatebound",
				Steps: []string{
					"Maintain Garrote.",
					"Mutilate until having 5 or more combo points.",
					"Maintain Rupture by casting at 5 or more combo points.",
					"Vanish followed by Garrote to apply Improved Garrote; line up with Deathmark.",
					"Deathmark on cooldown.",
					"Kingsbane on cooldown, immediately after Deathmark when applicable.",
					"Envenom at 5 or more combo points, unless at 80 stacks of Implacable.",
				},
			},
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
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Fatebound",
				Steps: []string{
					"Roll the Bones on cooldown if at stage 1 or less.",
					"Keep It Rolling if you have a stage 2 or higher Roll the Bones active.",
					"Preparation whenever Between the Eyes, Adrenaline Rush, and Blade Rush are on cooldown.",
					"Adrenaline Rush on cooldown at 2 or fewer combo points.",
					"Blade Rush on cooldown.",
					"Between the Eyes at 6 or more combo points.",
					"Dispatch at 6 or more combo points.",
					"Pistol Shot if Opportunity has 4 stacks.",
					"Pistol Shot if Opportunity has 3 stacks and you are at 1-3 combo points.",
					"Sinister Strike at 5 or fewer combo points.",
				},
			},
			{
				HeroTalent: "Trickster",
				Steps: []string{
					"Roll the Bones on cooldown if at stage 1 or less.",
					"Keep It Rolling if you have a stage 3 or higher Roll the Bones active.",
					"Preparation whenever Between the Eyes, Adrenaline Rush, and Blade Rush are on cooldown.",
					"Adrenaline Rush on cooldown at 2 or fewer combo points.",
					"Blade Rush on cooldown.",
					"Between the Eyes at 6 or more combo points.",
					"Killing Spree at 5 or more combo points; cancel early if you are going to overcap energy.",
					"Dispatch at 6 or more combo points.",
					"Pistol Shot if Opportunity has 4 stacks.",
					"Pistol Shot if Opportunity has 3 stacks and you are at 1-3 combo points.",
					"Sinister Strike at 5 or fewer combo points.",
				},
			},
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
		StatPriorities: []string{"Mastery", "Haste", "Critical Strike", "Versatility"},
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Deathstalker",
				Steps: []string{
					"Secret Technique.",
					"Eviscerate.",
					"Shadowstrike.",
				},
			},
			{
				HeroTalent: "Trickster",
				Steps: []string{
					"Secret Technique.",
					"Eviscerate.",
					"Shadowstrike.",
				},
			},
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
		StatPriorities: []string{"Versatility", "Haste", "Critical Strike"},
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Farseer",
				Steps: []string{
					"Use Spiritwalk's Grace and Nature's Swiftness for movement events.",
					"Stormkeeper on cooldown.",
					"Ancestral Swiftness on cooldown.",
					"Ascendance roughly on cooldown, but always after Stormkeeper.",
					"Elemental Blast if you are less than 15 from capping.",
					"Elemental Blast if you have Master of the Elements active.",
					"Elemental Blast if possible.",
					"Earth Shock if you are less than 15 from capping.",
					"Earth Shock if you have Master of the Elements active.",
					"Earth Shock if possible.",
					"Lava Burst.",
					"Lightning Bolt to consume Master of the Elements.",
					"Refresh Flame Shock within the pandemic window (30% of its current maximum duration) if you do not have Master of the Elements active.",
					"Refresh Flame Shock with Voltaic Blaze within the pandemic window.",
					"Lightning Bolt as your filler.",
				},
			},
			{
				HeroTalent: "Stormbringer",
				Steps: []string{
					"Use Spiritwalk's Grace and Nature's Swiftness for movement events.",
					"Stormkeeper on cooldown.",
					"Ancestral Swiftness on cooldown.",
					"Ascendance roughly on cooldown, but always after Stormkeeper.",
					"Refresh Flame Shock within the pandemic window (30%) if you do not have Master of the Elements active.",
					"Refresh Flame Shock with Voltaic Blaze within the pandemic window.",
					"Elemental Blast if you are less than 15 from capping.",
					"Elemental Blast if you have Master of the Elements active.",
					"Earth Shock if you are less than 15 from capping.",
					"Lava Burst.",
					"Lightning Bolt as your filler.",
				},
			},
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
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Stormbringer",
				Steps: []string{
					"Primordial Storm with 10 Maelstrom Weapon stacks.",
					"Sundering.",
					"Crash Lightning.",
					"Ascendance.",
					"Windstrike during Ascendance.",
					"Stormstrike during Doom Winds.",
					"Tempest / Lightning Bolt with 10 Maelstrom Weapon stacks.",
					"Stormstrike.",
					"Lava Lash.",
					"Voltaic Blaze.",
					"Lightning Bolt with 5+ Maelstrom Weapon stacks.",
				},
			},
			{
				HeroTalent: "Totemic",
				Steps: []string{
					"Voltaic Blaze if Flame Shock isn't active.",
					"Surging Totem.",
					"Lava Lash with Hot Hand or Whirling Fire active.",
					"Sundering.",
					"Doom Winds.",
					"Crash Lightning.",
					"Primordial Storm with 10 Maelstrom Weapon stacks.",
					"Stormstrike during Doom Winds.",
					"Tempest / Lightning Bolt with 10 Maelstrom Weapon stacks.",
					"Lava Lash.",
					"Stormstrike.",
					"Voltaic Blaze.",
					"Lightning Bolt with 5+ Maelstrom Weapon stacks.",
				},
			},
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
		StatPriorities: []string{"Haste", "Critical Strike", "Versatility"},
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Farseer",
				Steps: []string{
					"Use all your Stormstream Totem procs.",
					"Keep Riptide on cooldown.",
					"Use Ancestral Swiftness.",
					"Unleash Life.",
					"Maintain Healing Rain.",
					"Keep Healing Stream Totem on cooldown.",
					"Chain Heal or Healing Wave.",
				},
			},
			{
				HeroTalent: "Totemic",
				Steps: []string{
					"Use all your Stormstream Totem procs.",
					"Keep Riptide on cooldown.",
					"Use Nature's Swiftness.",
					"Maintain Surging Totem.",
					"Use Downpour as much as possible.",
					"Keep Healing Stream Totem on cooldown.",
					"Unleash Life.",
					"Chain Heal or Healing Wave.",
				},
			},
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
		StatPriorities: []string{"Versatility", "Haste", "Critical Strike", "Mastery"},
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Hellcaller",
				Steps: []string{
					"Haunt.",
					"Agony.",
					"Wither.",
					"Dark Harvest.",
					"Summon Darkglare.",
					"Malevolence.",
					"As many Unstable Afflictions as possible.",
					"Malefic Grasp.",
				},
			},
			{
				HeroTalent: "Soul Harvester",
				Steps: []string{
					"Haunt.",
					"Agony.",
					"Corruption.",
					"Summon Darkglare.",
					"Unstable Affliction.",
					"Dark Harvest.",
					"Unstable Affliction.",
					"Malefic Grasp.",
				},
			},
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
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Diabolist",
				Steps: []string{
					"Power Siphon.",
					"Call Dreadstalkers.",
					"Hand of Gul'dan / Ruination.",
					"Grimoire: Imp Lord / Grimoire: Fel Ravager.",
					"Summon Demonic Tyrant.",
					"Summon Doomguard.",
					"Demonbolt with Demonic Core.",
					"Shadow Bolt / Infernal Bolt.",
				},
			},
			{
				HeroTalent: "Soul Harvester",
				Steps: []string{
					"Power Siphon.",
					"Call Dreadstalkers.",
					"Hand of Gul'dan.",
					"Grimoire: Imp Lord / Grimoire: Fel Ravager.",
					"Summon Demonic Tyrant.",
					"Summon Doomguard.",
					"Demonbolt with Demonic Core.",
					"Shadow Bolt.",
				},
			},
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
		StatPriorities: []string{"Versatility", "Haste", "Critical Strike", "Mastery"},
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Diabolist",
				Steps: []string{
					"Maintain Immolate.",
					"Shadowburn when available and if you are about to cap on soul shards.",
					"Chaos Bolt to avoid capping soul shards.",
					"Chaos Bolt if you have more than 4 soul shards and Internal Bolt ready.",
					"Soul Fire when available and you have less than 4 soul shards.",
					"Conflagrate to keep this below 2 stacks or when moving.",
					"Conflagrate to generate soul shards and Backdraft stacks; use on Chaos Bolt as much as possible.",
					"Incinerate to generate soul shards.",
				},
			},
			{
				HeroTalent: "Hellcaller",
				Steps: []string{
					"Maintain Wither.",
					"Shadowburn when available and if you are about to cap on soul shards.",
					"Chaos Bolt to avoid capping soul shards.",
					"Soul Fire when available and you have less than 4 soul shards.",
					"Conflagrate to keep this below 2 stacks or when moving.",
					"Conflagrate to generate soul shards and Backdraft stacks; use on Chaos Bolt as much as possible.",
					"Incinerate to generate soul shards.",
				},
			},
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
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Colossus",
				Steps: []string{
					"Rend to apply or refresh at less than 4 seconds remaining.",
					"Ravager just before or with Colossus Smash.",
					"Avatar shortly before or with Colossus Smash.",
					"Colossus Smash.",
					"Demolish during Colossus Smash.",
					"Heroic Strike.",
					"Mortal Strike.",
					"Overpower.",
					"Execute during Sudden Death.",
					"Slam to fill the rotation.",
				},
			},
			{
				HeroTalent: "Slayer",
				Steps: []string{
					"Avatar shortly before or with Colossus Smash.",
					"Colossus Smash.",
					"Bladestorm during Colossus Smash.",
					"Mortal Strike.",
					"Heroic Strike.",
					"Execute during Sudden Death.",
					"Overpower.",
					"Rend to apply or refresh at less than 4 seconds remaining.",
					"Slam when nothing else is available.",
				},
			},
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
		StatPriorities: []string{"Critical Strike", "Versatility", "Haste", "Mastery"},
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Mountain Thane",
				Steps: []string{
					"Odyn's Fury.",
					"Rampage over 100 rage.",
					"Thunder Blast with two stacks available.",
					"Bloodbath.",
					"Rampage.",
					"Thunder Blast.",
					"Execute.",
					"Crushing Blow.",
					"Thunder Clap when nothing else is available.",
				},
			},
			{
				HeroTalent: "Slayer",
				Steps: []string{
					"Rampage over 100 rage.",
					"Bladestorm.",
					"Odyn's Fury.",
					"Execute.",
					"Bloodbath.",
					"Rampage.",
					"Crushing Blow.",
				},
			},
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
		StatPriorities: []string{"Haste", "Critical Strike", "Versatility", "Mastery"},
		HeroTalents: []heroTalentRotation{
			{
				HeroTalent: "Colossus",
				Steps: []string{
					"Always Charge into combat.",
					"Demolish on cooldown.",
					"Shield Slam on cooldown.",
					"Thunder Clap on cooldown and to apply Rend.",
					"Revenge.",
					"Spend excess Rage on Ignore Pain.",
					"Execute targets at or below 20% (35% if you have the talent) health.",
					"Use Impending Victory if you have low HP.",
				},
			},
			{
				HeroTalent: "Mountain Thane",
				Steps: []string{
					"Always Charge into combat.",
					"Shield Slam on cooldown.",
					"Thunder Clap on cooldown and to apply Rend.",
					"Revenge.",
					"Spend excess Rage on Ignore Pain.",
					"Execute targets at or below 20% (35% if you have the talent) health.",
					"Use Impending Victory if you have low HP.",
				},
			},
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
