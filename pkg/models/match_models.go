package models

import (
	"d2api/pkg/requests"

	"github.com/paralin/go-dota2/protocol"
)

type MatchStatus struct {
	MatchId           uint64
	Status            string
	IsTournamentMatch bool
	TourMatch         requests.TourMatch
	TeamA             []uint64
	TeamB             []uint64
}

type MatchDetails struct {
	MatchStatus
	Handler       string
	CancelReason  string
	TeamDidntShow string
}

type MatchCancel struct {
	MatchStatus
	Reason        string
	TeamDidntShow string
}

type MatchLobby struct {
	MatchStatus
	Lobby *protocol.CSODOTALobby
}

type MatchData struct {
	MatchStatus
	Match *protocol.CMsgDOTAMatch
}

type MatchMongo struct {
	Id    string                  `bson:"_id" json:"id"`
	Match *protocol.CMsgDOTAMatch `bson:"match" json:"match"`
}
