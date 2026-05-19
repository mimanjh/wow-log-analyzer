package types

import "time"

type UserProfile struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Avatar    string `json:"avatar,omitempty"`
	BattleTag string `json:"battleTag,omitempty"`
}

type OwnedCharacter struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Class        string `json:"class"`
	ServerName   string `json:"serverName"`
	ServerRegion string `json:"serverRegion"`
	ServerSlug   string `json:"serverSlug,omitempty"`
}

type CharacterReportsPage struct {
	Reports    []CharacterReportSummary `json:"reports"`
	NextCursor string                   `json:"nextCursor,omitempty"`
	HasMore    bool                     `json:"hasMore"`
}

type CharacterFightSummary struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Difficulty  string `json:"difficulty"`
	Kill        bool   `json:"kill"`
	KillTime    int    `json:"killTime"`
	EncounterID int    `json:"encounterId"`
}

type CharacterReportSummary struct {
	Code      string                 `json:"code"`
	Title     string                 `json:"title"`
	ZoneName  string                 `json:"zoneName,omitempty"`
	BossNames []string               `json:"bossNames,omitempty"`
	StartTime time.Time              `json:"startTime"`
	EndTime   time.Time              `json:"endTime"`
	Fights    []CharacterFightSummary `json:"fights,omitempty"`
}
