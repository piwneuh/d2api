package models

type MatchDetails struct {
	MatchId      uint64
	HandlerId    uint16
	Status       string
	CancelReason string
}
