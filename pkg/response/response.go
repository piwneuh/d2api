package response

type TeamEnd struct {
	TeamId int `json:"teamId"`
	Score  int `json:"score"`
}

type TournamentEndRequest struct {
	TourId    int     `json:"tourId"`
	Winner    TeamEnd `json:"winnerTeam"`
	Loser     TeamEnd `json:"loserTeam"`
	MatchId   int     `json:"matchId"`
	Cancelled bool    `json:"cancelled"`
	Iteration int     `json:"iteration"`
}

type Player struct {
	SteamId       uint64 `json:"steamId"`
	IsInLobby     bool   `json:"isInLobby"`
	IsInRightTeam bool   `json:"isInRightTeam"`
}

type MatchInfo struct {
	Status         string   `json:"status"`
	Cancelled      bool     `json:"cancelled,omitempty"`
	RadiantPlayers []Player `json:"radiantPlayers,omitempty"`
	DirePlayers    []Player `json:"direPlayers,omitempty"`
}
