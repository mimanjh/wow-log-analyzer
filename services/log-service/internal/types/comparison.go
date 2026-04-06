package types

import "time"

type ComparisonDataRequest struct {
	FightID     int `json:"fightId"`
	CharacterID int `json:"characterId"`
}

type ComparisonDataResponse struct {
	ReportID   string            `json:"reportId"`
	Fight      FightSummary      `json:"fight"`
	PlayerData PlayerFightData   `json:"playerData"`
	CohortData []PlayerFightData `json:"cohortData"`
}

type Ability struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	IsMajorCD bool   `json:"isMajorCd,omitempty"`
	IsBuff    bool   `json:"isBuff,omitempty"`
}

type CastEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Ability   Ability   `json:"ability"`
	SourceID  int       `json:"sourceId"`
}

type DamageEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Ability   Ability   `json:"ability"`
	SourceID  int       `json:"sourceId"`
	TargetID  int       `json:"targetId"`
	Amount    int       `json:"amount"`
}

type HealEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Ability   Ability   `json:"ability"`
	SourceID  int       `json:"sourceId"`
	TargetID  int       `json:"targetId"`
	Amount    int       `json:"amount"`
}

type BuffEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Ability   Ability   `json:"ability"`
	SourceID  int       `json:"sourceId"`
	TargetID  int       `json:"targetId"`
	EventType string    `json:"eventType"`
}

type CooldownEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Ability   Ability   `json:"ability"`
	SourceID  int       `json:"sourceId"`
	EventType string    `json:"eventType"`
}

type PlayerFightData struct {
	PlayerID       int             `json:"playerId"`
	FightID        int             `json:"fightId"`
	FightStart     time.Time       `json:"fightStart"`
	FightEnd       time.Time       `json:"fightEnd"`
	CastEvents     []CastEvent     `json:"castEvents"`
	DamageEvents   []DamageEvent   `json:"damageEvents"`
	HealEvents     []HealEvent     `json:"healEvents"`
	BuffEvents     []BuffEvent     `json:"buffEvents"`
	CooldownEvents []CooldownEvent `json:"cooldownEvents"`
}
