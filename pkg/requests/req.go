package requests

type InviteLobbyReq struct {
	SteamId uint64 `uri:"steamId"`
}

type CreateMatchReq struct {
	TeamA       []uint64    `json:"teamA"`
	TeamB       []uint64    `json:"teamB"`
	LobbyConfig LobbyConfig `json:"lobbyConfig"`
	StartTime   string      `json:"startTime"`
}

type LobbyConfig struct {
	GameName     string `json:"gameName"`
	PassKey      string `json:"passKey"`
	ServerRegion uint32 `json:"serverRegion"`
	GameMode     string `json:"gameMode"`
	FillWithBots bool   `json:"fillWithBots,omitempty"`
}

type TourMatch struct {
	Region                 string                `json:"region"`
	MatchDB                string                `json:"matchDb"`
	MatchId                int                   `json:"matchId"`
	MatchIdx               int                   `json:"matchIdx"`
	TournamentId           int                   `json:"tournamentId"`
	TournamentName         string                `json:"tournamentName"`
	TournamentLogo         string                `json:"tournamentLogo"`
	TournamentOwnerId      string                `json:"tournamentOwnerId"`
	EstimatedTournamentEnd int64                 `json:"estimatedTournamentEnd"`
	Team1Id                int                   `json:"team1Id"`
	Team2Id                int                   `json:"team2Id"`
	Map                    string                `json:"map"`
	Team1                  TeamForMiddleware     `json:"team1"`
	Team2                  TeamForMiddleware     `json:"team2"`
	Players                []PlayerForMiddleware `json:"players"`
	Currency               string                `json:"currency"`
	DeadlineEpoch          int64                 `json:"deadlineEpoch"`
	StartEpoch             int64                 `json:"startEpoch"`
	NumberOfRounds         int                   `json:"numberOfRounds"`
	Prize                  int                   `json:"prize"`
	RoundTime              int                   `json:"roundTime"`
	Cancelled              bool                  `json:"cancelled"`
	Iteration              int                   `json:"iteration"`
}

type TeamForMiddleware struct {
	Name string `json:"name"`
	Logo string `json:"logo"`
}

type PlayerForMiddleware struct {
	Team          string `json:"team"`
	SteamId       string `json:"steam_id_64"`
	WalletAddress string `json:"walletAddress"`
}

type ScheduleMatchRequest struct {
	Matches []TourMatch `json:"matches"`
}

type ReinvitePlayersReq struct {
	MatchIdx int      `json:"matchIdx"`
	Players  []uint64 `json:"players"`
}

type Notification struct {
	Content  string   `json:"content"`
	Metadata Metadata `json:"metadata"`
	UserIds  []string `json:"user_ids"`
	Type     string   `json:"type"`
	Subtype  string   `json:"subtype"`
	RefId    string   `json:"ref_id"`
	Service  string   `json:"service"`
}

type Metadata struct {
	Ids map[string]string `json:"ids"`
}
