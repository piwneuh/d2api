package utils

import (
	"context"
	h "d2api/pkg/handlers"
	"d2api/pkg/redis"
	"d2api/pkg/requests"
	"encoding/json"
	"errors"
	"log"
	"slices"

	"github.com/paralin/go-dota2/cso"
	"github.com/paralin/go-dota2/protocol"
)

type MatchDetails struct {
	MatchId   uint64
	HandlerId uint16
}

func GetGoodAndBadGuys(lobby *protocol.CSODOTALobby) ([]uint64, []uint64, error) {
	goodGuys := make([]uint64, 0)
	badGuys := make([]uint64, 0)

	for _, member := range lobby.AllMembers {
		if member.Team == nil {
			continue
		} else if *member.Team == protocol.DOTA_GC_TEAM_DOTA_GC_TEAM_GOOD_GUYS {
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

func GetMatchRedis(matchIdx string) (*MatchDetails, error) {
	marshalled, err := redis.RedisClient.Get(context.Background(), matchIdx).Result()
	if err != nil {
		return nil, err
	}

	var match MatchDetails
	err = json.Unmarshal([]byte(marshalled), &match)
	if err != nil {
		return nil, errors.New("match not found")
	}

	return &match, nil
}

func SetMatchRedis(matchIdx string, match MatchDetails) error {
	marshalled, err := json.Marshal(match)
	if err != nil {
		return err
	}

	err = redis.RedisClient.Set(context.Background(), matchIdx, marshalled, 0).Err()
	if err != nil {
		return err
	}

	return nil
}
