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
}

type MatchDetails struct {
	MatchStatus
	HandlerId    uint16
	CancelReason string
}

type MatchCancel struct {
	MatchStatus
	Reason string
}

type MatchLobby struct {
	MatchStatus
	Lobby *protocol.CSODOTALobby
}

type MatchData struct {
	MatchStatus
	Match *protocol.CMsgDOTAMatch
}
