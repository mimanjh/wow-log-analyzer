package types

import "time"

// NormalizedReport represents a Warcraft Logs report in our internal format
type NormalizedReport struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	Zone      Zone      `json:"zone"`
}

// Zone represents a raid zone/dungeon
type Zone struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// NormalizedFight represents a fight in our internal format
type NormalizedFight struct {
	ID              int                `json:"id"`
	Name            string             `json:"name"`
	StartTime       time.Time          `json:"startTime"`
	EndTime         time.Time          `json:"endTime"`
	EncounterID     int                `json:"encounterId"`
	Difficulty      string             `json:"difficulty"`
	Kill            bool               `json:"kill"`
	BossPercent     float64            `json:"bossPercent"`
	FriendlyPlayers []FightParticipant `json:"friendlyPlayers,omitempty"`
}

type FightParticipant struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	ServerName string `json:"serverName,omitempty"`
	Class      string `json:"class,omitempty"`
}

// FightSummary is a simplified version for initial fight selection
type FightSummary struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Difficulty  string `json:"difficulty"`
	KillTime    int    `json:"killTime"` // in seconds
	EncounterID int    `json:"encounterId"`
}
