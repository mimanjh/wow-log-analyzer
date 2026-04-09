package services

import "strings"

type specPrioritySet struct {
	Offensives []string
	Defensives []string
}

var specPriorities = map[string]specPrioritySet{
	"Blood Death Knight": {
		Offensives: []string{"Dancing Rune Weapon"},
		Defensives: []string{"Vampiric Blood", "Anti-Magic Zone", "Icebound Fortitude", "Lichborne", "Anti-Magic Shell"},
	},
	"Vengeance Demon Hunter": {
		Offensives: []string{"Fiery Brand", "Immolation Aura", "Fel Devastation", "Sigil of Flames"},
		Defensives: []string{"Darkness", "Demon Spikes"},
	},
	"Guardian Druid": {
		Offensives: []string{"Incarnation: Guardian of Ursoc", "Convoke the Spirits", "Lunar Beam"},
		Defensives: []string{"Barkskin", "Survival Instincts", "Frenzied Regeneration"},
	},
	"Brewmaster Monk": {
		Offensives: []string{"Invoke Niuzao, the Black Ox", "Exploding Keg", "Touch of Death"},
		Defensives: []string{"Fortifying Brew", "Celestial Infusion"},
	},
	"Protection Paladin": {
		Offensives: []string{"Avenging Wrath", "Divine Toll"},
		Defensives: []string{"Ardent Defender", "Guardian of Ancient Kings", "Divine Shield"},
	},
	"Protection Warrior": {
		Offensives: []string{"Avatar", "Ravager", "Thunder Blast", "Shield Charge", "Rend"},
		Defensives: []string{"Shield Wall", "Demoralizing Shout", "Spell Reflection"},
	},
	"Frost Death Knight": {
		Offensives: []string{"Pillar of Frost", "Frostwyrm's Fury", "Breath of Sindragosa"},
		Defensives: []string{"Icebound Fortitude", "Lichborne", "Anti-Magic Shell"},
	},
	"Unholy Death Knight": {
		Offensives: []string{"Army of the Dead", "Raise Abomination", "Dark Transformation", "Putrefy", "Soul Reaper"},
		Defensives: []string{"Icebound Fortitude", "Lichborne", "Anti-Magic Shell"},
	},
	"Havoc Demon Hunter": {
		Offensives: []string{"Metamorphosis", "Eye Beam", "Essence Break", "The Hunt"},
		Defensives: []string{"Darkness"},
	},
	"Feral Druid": {
		Offensives: []string{"Berserk", "Tiger's Fury", "Feral Frenzy", "Convoke the Spirits"},
		Defensives: []string{"Barkskin", "Survival Instincts", "Frenzied Regeneration"},
	},
	"Survival Hunter": {
		Offensives: []string{"Aspect of the Eagle", "Takedown", "Boomstick"},
		Defensives: []string{"Exhilaration", "Survival of the Fittest", "Aspect of the Turtle"},
	},
	"Windwalker Monk": {
		Offensives: []string{"Zenith", "Touch of Death"},
		Defensives: []string{"Touch of Karma", "Fortifying Brew"},
	},
	"Retribution Paladin": {
		Offensives: []string{"Avenging Wrath", "Divine Toll", "Wake of Ashes", "Execution Sentence"},
		Defensives: []string{"Divine Protection", "Divine Shield"},
	},
	"Assassination Rogue": {
		Offensives: []string{"Kingsbane", "Shiv", "Deathmark", "Vanish"},
		Defensives: []string{"Feint", "Evasion", "Cloak of Shadows", "Crimson Vial"},
	},
	"Outlaw Rogue": {
		Offensives: []string{"Adrenaline Rush", "Keep It Rolling", "Vanish"},
		Defensives: []string{"Feint", "Evasion", "Cloak of Shadows", "Crimson Vial"},
	},
	"Subtlety Rogue": {
		Offensives: []string{"Shadow Blades", "Shadow Dance", "Secret Technique", "Vanish"},
		Defensives: []string{"Feint", "Evasion", "Cloak of Shadows", "Crimson Vial"},
	},
	"Enhancement Shaman": {
		Offensives: []string{"Sundering", "Doom Winds", "Primordial Storm"},
		Defensives: []string{"Astral Shift"},
	},
	"Arms Warrior": {
		Offensives: []string{"Colossus Smash", "Ravager", "Avatar", "Bladestorm", "Sweeping Strikes", "Execute", "Rend"},
		Defensives: []string{"Die by the Sword", "Spell Reflection", "Defensive Stance"},
	},
	"Fury Warrior": {
		Offensives: []string{"Recklessness", "Avatar", "Odyn's Fury", "Bladestorm", "Rend"},
		Defensives: []string{"Enraged Regeneration", "Spell Reflection", "Defensive Stance"},
	},
	"Devourer Demon Hunter": {
		Offensives: []string{"Void Metamorphosis", "Collapsing Star"},
		Defensives: []string{"Darkness"},
	},
	"Balance Druid": {
		Offensives: []string{"Celestial Alignment", "Force of Nature", "Fury of Elune", "Full Moon", "Convoke the Spirits", "Lunar Eclipse", "Solar Eclipse"},
		Defensives: []string{"Barkskin", "Survival Instincts", "Frenzied Regeneration"},
	},
	"Augmentation Evoker": {
		Offensives: []string{"Ebon Might", "Breath of Eons", "Time Skip", "Spatial Paradox", "Time Spiral", "Zephyr", "Tip the Scales"},
		Defensives: []string{"Obsidian Scales"},
	},
	"Devastation Evoker": {
		Offensives: []string{"Dragonrage", "Deep Breath", "Spatial Paradox", "Time Spiral", "Zephyr", "Tip the Scales"},
		Defensives: []string{"Obsidian Scales"},
	},
	"Beast Mastery Hunter": {
		Offensives: []string{"Bestial Wrath", "Barbed Shot"},
		Defensives: []string{"Exhilrataion", "Survival of the Fittest", "Aspect of the Turtle"},
	},
	"Marksmanship Hunter": {
		Offensives: []string{"Trueshot", "Volley"},
		Defensives: []string{"Exhilrataion", "Survival of the Fittest", "Aspect of the Turtle"},
	},
	"Arcane Mage": {
		Offensives: []string{"Arcane Surge", "Touch of the Magi", "Mirror Image"},
		Defensives: []string{"Alter Time", "Ice Cold"},
	},
	"Fire Mage": {
		Offensives: []string{"Combustion", "Pyroblast", "Fire Blast", "Flamestrike", "Meteor", "Scorch", "Mirror Image"},
		Defensives: []string{"Alter Time", "Ice Cold"},
	},
	"Frost Mage": {
		Offensives: []string{"Ray of Frost", "Frozen Orb", "Mirror Image"},
		Defensives: []string{"Alter Time", "Ice Cold"},
	},
	"Shadow Priest": {
		Offensives: []string{"Voidform", "Halo"},
		Defensives: []string{"Dispersion", "Vampiric Embrace", "Desperate Prayer", "Fade"},
	},
	"Elemental Shaman": {
		Offensives: []string{"Ascendance", "Stormkeeper"},
		Defensives: []string{"Astral Shift", "Spiritwalker's Grace"},
	},
	"Affliction Warlock": {
		Offensives: []string{"Summon Darkglare", "Dark Harvest"},
		Defensives: []string{"Unending Resolve", "Dark Pact"},
	},
	"Demonology Warlock": {
		Offensives: []string{"Summon Demonic Tyrant", "Grimoire: Imp Lord", "Grimoire: Fel Ravager"},
		Defensives: []string{"Unending Resolve", "Dark Pact"},
	},
	"Destruction Warlock": {
		Offensives: []string{"Summon Infernal", "Havoc"},
		Defensives: []string{"Unending Resolve", "Dark Pact"},
	},
}

func specPriorityFor(characterClass, characterSpec string) specPrioritySet {
	return specPriorities[strings.TrimSpace(characterSpec)+" "+strings.TrimSpace(characterClass)]
}

func normalizeTrackedCooldownName(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func buildCooldownCategoryMap(specKey specPrioritySet) map[string]string {
	categoryMap := make(map[string]string, len(specKey.Offensives)+len(specKey.Defensives))
	for _, value := range specKey.Offensives {
		categoryMap[normalizeTrackedCooldownName(value)] = "offensive"
	}
	for _, value := range specKey.Defensives {
		categoryMap[normalizeTrackedCooldownName(value)] = "defensive"
	}
	return categoryMap
}
