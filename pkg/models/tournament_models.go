package models

import "d2api/pkg/requests"

type MatchModel struct {
	requests.MatchForMiddleware `json:",inline"`
	D2MatchId                   string `json:"d2_match_id"`
}
