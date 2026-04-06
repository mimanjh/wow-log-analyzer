package client

// WCLReport represents the raw Warcraft Logs API report structure
type WCLReport struct {
	Code      string `json:"code"`
	Title     string `json:"title"`
	StartTime int64  `json:"startTime"`
	EndTime   int64  `json:"endTime"`
	Zone      WCLZone `json:"zone"`
}

// WCLZone represents the raw Warcraft Logs API zone structure
type WCLZone struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// WCLFight represents the raw Warcraft Logs API fight structure
type WCLFight struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	StartTime      int64  `json:"startTime"`
	EndTime        int64  `json:"endTime"`
	EncounterID    int    `json:"encounterID"`
	Difficulty     int    `json:"difficulty"`
	Kill           bool   `json:"kill"`
	BossPercentage int    `json:"bossPercentage"`
}