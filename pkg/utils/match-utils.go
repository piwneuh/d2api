package utils

import (
	h "d2api/pkg/handlers"
	"d2api/pkg/requests"
	"log"
	"slices"

	"github.com/paralin/go-dota2/cso"
	"github.com/paralin/go-dota2/protocol"
)

func GetGoodAndBadGuys(lobby *protocol.CSODOTALobby) ([]uint64, []uint64, error) {
	goodGuys := make([]uint64, 0)
	badGuys := make([]uint64, 0)

	for _, member := range lobby.AllMembers {
		if *member.Team == protocol.DOTA_GC_TEAM_DOTA_GC_TEAM_GOOD_GUYS {
			goodGuys = append(goodGuys, *member.Id)
		} else if *member.Team == protocol.DOTA_GC_TEAM_DOTA_GC_TEAM_BAD_GUYS {
			badGuys = append(badGuys, *member.Id)
		}
	}

	return goodGuys, badGuys, nil
}

func GetCurrentLobby(handler *h.Handler) (*protocol.CSODOTALobby, error) {
	lobby, err := handler.DotaClient.GetCache().GetContainerForTypeID(cso.Lobby)
	if err != nil {
		log.Fatalf("Failed to get lobby: %v", err)
		return nil, err
	}

	return lobby.GetOne().(*protocol.CSODOTALobby), nil
}

func AreAllPlayerHere(goodGuys []uint64, badGuys []uint64, req *requests.CreateMatchReq) bool {
	// Check if the goodGuys and badGuys are ready
	if len(goodGuys) != len(req.TeamA) || len(badGuys) != len(req.TeamB) {
		return false
	}

	for _, id := range req.TeamA {
		if !slices.Contains(goodGuys, id) {
			return false
		}
	}

	for _, id := range req.TeamB {
		if !slices.Contains(badGuys, id) {
			return false
		}
	}

	return true
}
