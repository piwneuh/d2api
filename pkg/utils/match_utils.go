package utils

import (
	"context"
	h "d2api/pkg/handlers"
	m "d2api/pkg/models"
	"d2api/pkg/redis"
	"d2api/pkg/requests"
	"encoding/json"
	"errors"
	"log"
	"slices"
	"strings"

	"github.com/paralin/go-dota2/cso"
	"github.com/paralin/go-dota2/protocol"
)

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

func GetGameModeFromString(gameMode string) uint32 {
	fullString := "DOTA_GAMEMODE_" + strings.ToUpper(gameMode)
	gameModeInt, ok := protocol.DOTA_GameMode_value[fullString]
	if !ok {
		return uint32(protocol.DOTA_GameMode_value["DOTA_GAMEMODE_AP"])
	}

	return uint32(gameModeInt)
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

func GetMissingPlayers(goodGuys []uint64, badGuys []uint64, req *requests.CreateMatchReq) []uint64 {
	missingPlayers := make([]uint64, 0)

	for _, id := range req.TeamA {
		if !slices.Contains(goodGuys, id) {
			missingPlayers = append(missingPlayers, id)
		}
	}

	for _, id := range req.TeamB {
		if !slices.Contains(badGuys, id) {
			missingPlayers = append(missingPlayers, id)
		}
	}

	return missingPlayers
}

func GetMatchRedis(matchIdx string) (*m.MatchDetails, error) {
	marshalled, err := redis.RedisClient.Get(context.Background(), matchIdx).Result()
	if err != nil {
		return nil, err
	}

	var match m.MatchDetails
	err = json.Unmarshal([]byte(marshalled), &match)
	if err != nil {
		return nil, errors.New("match not found")
	}

	return &match, nil
}

func SetMatchRedis(matchIdx string, match m.MatchDetails) error {
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

func GetAllMatchIdxs() ([]string, error) {
	keys, err := redis.RedisClient.Keys(context.Background(), "*").Result()
	if err != nil {
		log.Fatalf("Failed to get keys: %v", err)
		return nil, err
	}

	return keys, nil
}
