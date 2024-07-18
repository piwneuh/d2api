package models

import (
	"d2api/pkg/requests"

	"github.com/paralin/go-dota2/protocol"
)

type MatchStatus struct {
	MatchId           uint64             `json:"matchId"`
	Status            string             `json:"status"`
	IsTournamentMatch bool               `json:"isTournamentMatch"`
	TourMatch         requests.TourMatch `json:"tourMatch"`
	TeamA             []uint64           `json:"teamA"`
	TeamB             []uint64           `json:"teamB"`
}

type MatchDetails struct {
	MatchStatus   `json:",inline"`
	Handler       string `json:"handler"`
	CancelReason  string `json:"cancelReason"`
	TeamDidntShow string `json:"teamDidntShow"`
}

type MatchCancel struct {
	MatchStatus   `json:",inline"`
	Reason        string `json:"reason"`
	TeamDidntShow string `json:"teamDidntShow"`
}

type MatchLobby struct {
	MatchStatus
	Lobby *protocol.CSODOTALobby
}

type MatchData struct {
	MatchStatus `json:",inline"`
	Match       *protocol.CMsgDOTAMatch `json:"match"`
}

type MatchMongo struct {
	Id     string                  `bson:"_id" json:"id"`
	Status string                  `bson:"status" json:"status"`
	Match  *protocol.CMsgDOTAMatch `bson:"match" json:"match"`
}
