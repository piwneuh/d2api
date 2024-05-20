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
	"strconv"
	"strings"
	"time"

	steamId "github.com/paralin/go-steam/steamid"

	"github.com/paralin/go-dota2/cso"
	"github.com/paralin/go-dota2/protocol"
	"google.golang.org/protobuf/proto"
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
		log.Println("Failed to get lobby:", err)
		return nil, err
	}

	lobbyMessage := lobby.GetOne()
	if lobbyMessage == nil {
		return nil, errors.New("no lobby found")
	}

	return lobbyMessage.(*protocol.CSODOTALobby), nil
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

func GetMissingPlayers(goodGuys []uint64, badGuys []uint64, req *requests.CreateMatchReq) ([]uint64, bool, bool) {
	missingPlayers := make([]uint64, 0)
	var missingTeamA, missingTeamB bool
	for _, id := range req.TeamA {
		if !slices.Contains(goodGuys, id) {
			missingPlayers = append(missingPlayers, id)
			missingTeamA = true
		}
	}

	for _, id := range req.TeamB {
		if !slices.Contains(badGuys, id) {
			missingPlayers = append(missingPlayers, id)
			missingTeamB = true
		}
	}

	return missingPlayers, missingTeamA, missingTeamB
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
		log.Println("Failed to get keys:", err)
		return nil, err
	}

	return keys, nil
}

func MatchScheduleThread(hrs *[]*h.Handler, req requests.CreateMatchReq, matchIdx string, timeToCancel uint32) {
	if req.StartTime != "" {
		startTime, err := time.Parse(time.RFC3339, req.StartTime)
		if err != nil {
			log.Println("Failed to parse start time:", err)
			return
		}

		time.Sleep(time.Until(startTime))
	}
	var handler *h.Handler
	var handlerId uint16
	var err error

	for {
		handler, handlerId, err = h.GetFreeHandler(*hrs)
		if err != nil {
			time.Sleep(5 * time.Second)
			log.Println("No available bot, retrying in 5 seconds")
			continue
		}

		break
	}

	matchDetails, err := GetMatchRedis(matchIdx)
	if err != nil {
		log.Println("Failed to get match:", err)
		return
	}
	matchDetails.Status = "scheduled"
	matchDetails.HandlerId = handlerId

	SetMatchRedis(matchIdx, *matchDetails)

	time.Sleep(1 * time.Second)
	res, err := handler.DotaClient.DestroyLobby(context.Background())
	if err != nil {
		log.Println("Failed to destroy lobby: ", err, res)
	}
	time.Sleep(1 * time.Second)

	// Create the lobby
	lobbyVisibility := protocol.DOTALobbyVisibility_DOTALobbyVisibility_Public

	lobbyDetails := &protocol.CMsgPracticeLobbySetDetails{
		GameName:     proto.String(req.LobbyConfig.GameName),
		Visibility:   &lobbyVisibility,
		PassKey:      proto.String(req.LobbyConfig.PassKey),
		ServerRegion: proto.Uint32(req.LobbyConfig.ServerRegion),
		GameMode:     proto.Uint32(GetGameModeFromString(req.LobbyConfig.GameMode)),
	}

	handler.DotaClient.CreateLobby(lobbyDetails)
	time.Sleep(1 * time.Second)

	handler.DotaClient.SetLobbyCoach(protocol.DOTA_GC_TEAM_DOTA_GC_TEAM_GOOD_GUYS)

	// Invite the teamA
	for _, id := range req.TeamA {
		handler.DotaClient.InviteLobbyMember(steamId.SteamId(id))
	}

	// Invite the teamB
	for _, id := range req.TeamB {
		handler.DotaClient.InviteLobbyMember(steamId.SteamId(id))
	}

	lobbyCreationTime := time.Now()
	lobbyExpirationTime := lobbyCreationTime.Add(time.Duration(timeToCancel) * time.Second)

	for {
		time.Sleep(2 * time.Second)
		lobby, err := GetCurrentLobby(handler)
		if err != nil {
			log.Println("Failed to get lobby: ", err)
		}

		goodGuys, badGuys, err := GetGoodAndBadGuys(lobby)
		if err != nil {
			log.Println("Failed to get good and bad guys: ", err)
		}

		if AreAllPlayerHere(goodGuys, badGuys, &req) {
			break
		}

		if time.Now().After(lobbyExpirationTime) {
			log.Println("Cancelling match due to timeout", matchIdx)
			match, err := GetMatchRedis(matchIdx)
			if err != nil {
				log.Println("Failed to get match:", err)
			}

			missingPlayers, missingTeamA, missingTeamB := GetMissingPlayers(goodGuys, badGuys, &req)

			match.Status = "cancelled"
			match.CancelReason = "reason: players didn't join in time. players: "
			for _, id := range missingPlayers {
				match.CancelReason += strconv.FormatUint(id, 10) + ", "
			}
			match.CancelReason = match.CancelReason[:len(match.CancelReason)-2]
			if missingTeamA && missingTeamB {
				match.TeamDidntShow = "both"
			} else if missingTeamA {
				match.TeamDidntShow = "teamA"
			} else {
				match.TeamDidntShow = "teamB"
			}

			log.Println("Match cancelled due to timeout", matchIdx)

			err = SetMatchRedis(matchIdx, *match)
			if err != nil {
				log.Println("Failed to set match:", err)
			}

			handler.DotaClient.DestroyLobby(context.Background())
			handler.Occupied = false
			return
		}
	}

	// Start the game
	handler.DotaClient.LaunchLobby()
	for {
		time.Sleep(2 * time.Second)
		lobby, err := GetCurrentLobby(handler)
		if err != nil {
			log.Println("Failed to get lobby:", err)
		}

		if lobby.MatchId == nil {
			continue
		}

		match, err := GetMatchRedis(matchIdx)
		if err != nil {
			log.Println("Failed to get match:", err)
		}

		match.MatchId = *lobby.MatchId
		match.Status = "started"
		err = SetMatchRedis(matchIdx, *match)
		if err != nil {
			log.Println("Failed to set match:", err)
		}

		break
	}

	//Abandon the lobby
	handler.DotaClient.AbandonLobby()
	handler.Occupied = false
}
