package types

import "time"

type ComparisonDataRequest struct {
	Fight       FightSelection `json:"fight"`
	CharacterID int            `json:"characterId"`
}

type PlayerDataRequest struct {
	Fight       FightSelection `json:"fight"`
	CharacterID int            `json:"characterId"`
}

type RankingCandidatesRequest struct {
	Fight          FightSelection `json:"fight"`
	CharacterClass string         `json:"characterClass"`
	CharacterSpec  string         `json:"characterSpec"`
	Limit          int            `json:"limit,omitempty"`
}

type CohortMemberRequest struct {
	Candidate RankingCandidate `json:"candidate"`
}

type FightSelection struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Difficulty  string    `json:"difficulty"`
	KillTime    int       `json:"killTime"`
	EncounterID int       `json:"encounterId"`
	StartTime   time.Time `json:"startTime"`
	EndTime     time.Time `json:"endTime"`
	BossPercent float64   `json:"bossPercent,omitempty"`
}

type ComparisonDataResponse struct {
	ReportID   string            `json:"reportId"`
	Fight      FightSummary      `json:"fight"`
	PlayerData PlayerFightData   `json:"playerData"`
	CohortData []PlayerFightData `json:"cohortData"`
}

type RankingCandidate struct {
	Name         string  `json:"name"`
	Class        string  `json:"class"`
	Spec         string  `json:"spec"`
	Server       string  `json:"server,omitempty"`
	ServerRegion string  `json:"serverRegion,omitempty"`
	ReportID     string  `json:"reportId"`
	FightID      int     `json:"fightId"`
	RankValue    float64 `json:"rankValue"`
	DurationMS   int     `json:"durationMs"`
}

type CharacterOption struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Class      string `json:"class"`
	Spec       string `json:"spec"`
	Role       string `json:"role"`
	ServerName string `json:"serverName,omitempty"`
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
