package types

import "time"

// NormalizedEvent represents a normalized combat event
type NormalizedEvent struct {
	Timestamp time.Time              `json:"timestamp"`
	EventType string                 `json:"eventType"` // "cast", "damage", "heal", "buff", "debuff", "cooldown"
	SourceID  int                    `json:"sourceId"`
	TargetID  int                    `json:"targetId,omitempty"`
	Ability   Ability                `json:"ability"`
	EventData map[string]interface{} `json:"eventData,omitempty"` // additional event-specific data
}

// Ability represents a spell/ability
type Ability struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	School    string `json:"school,omitempty"`    // "physical", "holy", "fire", etc.
	IsMajorCD bool   `json:"isMajorCd,omitempty"` // whether this is a major cooldown
	IsBuff    bool   `json:"isBuff,omitempty"`    // whether this is a buff
}

// CastEvent represents a spell cast
type CastEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Ability   Ability   `json:"ability"`
	SourceID  int       `json:"sourceId"`
}

// DamageEvent represents damage dealt
type DamageEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Ability   Ability   `json:"ability"`
	SourceID  int       `json:"sourceId"`
	TargetID  int       `json:"targetId"`
	Amount    int       `json:"amount"`
}

// HealEvent represents healing done
type HealEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Ability   Ability   `json:"ability"`
	SourceID  int       `json:"sourceId"`
	TargetID  int       `json:"targetId"`
	Amount    int       `json:"amount"`
}

// BuffEvent represents a buff application/removal
type BuffEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Ability   Ability   `json:"ability"`
	SourceID  int       `json:"sourceId"`
	TargetID  int       `json:"targetId"`
	EventType string    `json:"eventType"` // "apply", "remove", "refresh"
}

// CooldownEvent represents a cooldown usage
type CooldownEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Ability   Ability   `json:"ability"`
	SourceID  int       `json:"sourceId"`
	EventType string    `json:"eventType"` // "start", "end"
}

type ResourceEvent struct {
	Timestamp      time.Time `json:"timestamp"`
	SourceID       int       `json:"sourceId"`
	TargetID       int       `json:"targetId"`
	ResourceTypeID int       `json:"resourceTypeId"`
	ResourceType   string    `json:"resourceType"`
	Amount         float64   `json:"amount,omitempty"`
	Change         float64   `json:"change"`
	Waste          float64   `json:"waste"`
	MaxAmount      float64   `json:"maxAmount,omitempty"`
}

// PlayerFightData contains all events for a single player's fight
type PlayerFightData struct {
	PlayerID            int             `json:"playerId"`
	FightID             int             `json:"fightId"`
	FightStart          time.Time       `json:"fightStart"`
	FightEnd            time.Time       `json:"fightEnd"`
	TalentImportCode    string          `json:"talentImportCode,omitempty"`
	TalentCalculatorURL string          `json:"talentCalculatorUrl,omitempty"`
	CastEvents          []CastEvent     `json:"castEvents"`
	DamageEvents        []DamageEvent   `json:"damageEvents"`
	HealEvents          []HealEvent     `json:"healEvents"`
	BuffEvents          []BuffEvent     `json:"buffEvents"`
	CooldownEvents      []CooldownEvent `json:"cooldownEvents"`
	ResourceEvents      []ResourceEvent `json:"resourceEvents"`
}
