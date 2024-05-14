package models

import "github.com/paralin/go-dota2/protocol"

type MatchStatus struct {
	MatchId           uint64
	Status            string
	IsTournamentMatch bool
	TournamentId      int
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
