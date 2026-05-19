package client

// WCLReport represents the raw Warcraft Logs API report structure
type WCLReport struct {
	Code      string  `json:"code"`
	Title     string  `json:"title"`
	StartTime int64   `json:"startTime"`
	EndTime   int64   `json:"endTime"`
	Zone      WCLZone `json:"zone"`
}

// WCLZone represents the raw Warcraft Logs API zone structure
type WCLZone struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// WCLFight represents the raw Warcraft Logs API fight structure
type WCLFight struct {
	ID              int     `json:"id"`
	Name            string  `json:"name"`
	StartTime       int64   `json:"startTime"`
	EndTime         int64   `json:"endTime"`
	EncounterID     int     `json:"encounterID"`
	Difficulty      int     `json:"difficulty"`
	Kill            bool    `json:"kill"`
	BossPercentage  float64 `json:"bossPercentage"`
	FriendlyPlayers []int   `json:"friendlyPlayers"`
}

type WCLActor struct {
	GameID   float64 `json:"gameID"`
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	PetOwner int     `json:"petOwner"`
	Server   string  `json:"server"`
	SubType  string  `json:"subType"`
	Type     string  `json:"type"`
}

type WCLAbility struct {
	GameID int    `json:"gameID"`
	Name   string `json:"name"`
}

type WCLPlayerSpec struct {
	Spec string `json:"spec"`
}

type WCLPlayerDetail struct {
	ID    int             `json:"id"`
	Name  string          `json:"name"`
	Type  string          `json:"type"`
	Specs []WCLPlayerSpec `json:"specs"`
}

type WCLRankingReport struct {
	Code      string `json:"code"`
	FightID   int    `json:"fightID"`
	StartTime int64  `json:"startTime"`
}

type WCLRankingServer struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Region string `json:"region"`
}

type WCLRankingEntry struct {
	Name      string           `json:"name"`
	Class     string           `json:"class"`
	Spec      string           `json:"spec"`
	Amount    float64          `json:"amount"`
	Duration  int              `json:"duration"`
	StartTime int64            `json:"startTime"`
	Report    WCLRankingReport `json:"report"`
	Server    WCLRankingServer `json:"server"`
}

type WCLCharacterRankingsResponse struct {
	Page         int               `json:"page"`
	HasMorePages bool              `json:"hasMorePages"`
	Count        int               `json:"count"`
	Rankings     []WCLRankingEntry `json:"rankings"`
	Error        string            `json:"error"`
}

type WCLTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type WCLRegion struct {
	Name string `json:"name"`
}

type WCLServer struct {
	Name   string    `json:"name"`
	Slug   string    `json:"slug"`
	Region WCLRegion `json:"region"`
}

type WCLUserCharacter struct {
	CanonicalID   int                 `json:"canonicalID"`
	Name          string              `json:"name"`
	ClassID       int                 `json:"classID"`
	Server        WCLServer           `json:"server"`
	RecentReports WCLReportPagination `json:"recentReports"`
}

type WCLCurrentUser struct {
	ID         int                `json:"id"`
	Name       string             `json:"name"`
	Avatar     string             `json:"avatar"`
	BattleTag  string             `json:"battleTag"`
	Characters []WCLUserCharacter `json:"characters"`
}

type WCLRankedCharacter struct {
	CanonicalID int       `json:"canonicalID"`
	Name        string    `json:"name"`
	ClassID     int       `json:"classID"`
	Server      WCLServer `json:"server"`
}

type WCLReportMasterData struct {
	Actors []WCLActor `json:"actors"`
}

type WCLReportSummary struct {
	Code             string               `json:"code"`
	Title            string               `json:"title"`
	StartTime        int64                `json:"startTime"`
	EndTime          int64                `json:"endTime"`
	Zone             *WCLZone             `json:"zone"`
	Fights           []WCLFight           `json:"fights"`
	RankedCharacters []WCLRankedCharacter `json:"rankedCharacters"`
	MasterData       WCLReportMasterData  `json:"masterData"`
}

type WCLReportPagination struct {
	Data         []WCLReportSummary `json:"data"`
	CurrentPage  int                `json:"current_page"`
	LastPage     int                `json:"last_page"`
	HasMorePages bool               `json:"has_more_pages"`
}
