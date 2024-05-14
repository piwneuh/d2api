package models

import "d2api/pkg/requests"

type MatchModel struct {
	requests.MatchForMiddleware `bson:",inline"`
	D2MatchId                   string `bson:"d2_match_id"`
	Finished                    bool   `bson:"finished"`
}

type TournamentRound struct {
	Matches      []MatchModel `bson:"matches"`
	Finished     bool         `bson:"finished"`
	TournamentId int          `bson:"tournament_id"`
}
