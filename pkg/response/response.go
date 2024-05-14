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
