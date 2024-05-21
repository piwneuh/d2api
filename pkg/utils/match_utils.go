package utils

import (
	"context"
	h "d2api/pkg/handlers"
	"d2api/pkg/requests"
	"errors"
	"log"
	"strconv"
	"time"

	steamId "github.com/paralin/go-steam/steamid"

	"github.com/paralin/go-dota2/protocol"
	"google.golang.org/protobuf/proto"
)

func MatchScheduleThread(hrs *[]*h.Handler, req requests.CreateMatchReq, matchIdx string, timeToCancel uint32) {
	waitForTimeToStart(req)
	handler, handlerId, err := getHandler(hrs)
	if err != nil {
		log.Println("Failed to get handler:", err)
		return
	}

	if ok := getAndSetMatchToRedis(matchIdx, handlerId); !ok {
		return
	}

	// Create the lobby
	lobby, err := createLobby(handler, req)
	if err != nil {
		log.Println("Failed to create lobby:", err)
		return
	}

	handler.DotaClient.SetLobbyCoach(protocol.DOTA_GC_TEAM_DOTA_GC_TEAM_GOOD_GUYS)
	inviteTeams(req, handler)

	channelResponse := joinLobbyChannel(lobby, handler)
	lobbyExpirationTime := time.Now().Add(time.Duration(timeToCancel) * time.Second)

	for {
		time.Sleep(2 * time.Second)
		lobby, err := GetCurrentLobby(handler)
		if err != nil {
			log.Println("Failed to get lobby: ", err)
			continue
		}

		missingTeamA, missingTeamB, err := GetMissingPlayers(lobby, &req)
		if err != nil {
			log.Println("Failed to get missing players:", err)
			continue
		}

		if len(missingTeamA) == 0 && len(missingTeamB) == 0 {
			break
		}

		sendMissingPlayersMessages(handler, channelResponse, missingTeamA, missingTeamB)

		if time.Now().After(lobbyExpirationTime) {
			lobbyExpired(matchIdx, missingTeamA, missingTeamB, handler)
			return
		}
	}

	startMatch(handler, channelResponse, matchIdx)
}

func startMatch(handler *h.Handler, channelResponse *protocol.CMsgDOTAJoinChatChannelResponse, matchIdx string) {
	handler.DotaClient.SendChannelMessage(*channelResponse.ChannelId, "All players joined the lobby")
	time.Sleep(2 * time.Second)

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

	handler.DotaClient.AbandonLobby()
	handler.Occupied = false
}

func sendMissingPlayersMessages(handler *h.Handler, channelResponse *protocol.CMsgDOTAJoinChatChannelResponse, missingTeamA []uint64, missingTeamB []uint64) {
	handler.DotaClient.SendChannelMessage(*channelResponse.ChannelId, "Waiting for players to join the lobby")
	if len(missingTeamA) > 0 {
		missingRadiants := "Missing radiant players: "
		for _, id := range missingTeamA {
			missingRadiants += strconv.FormatUint(id, 10) + ", "
		}
		handler.DotaClient.SendChannelMessage(*channelResponse.ChannelId, missingRadiants[:len(missingRadiants)-2])
	}

	if len(missingTeamB) > 0 {
		missingDire := "Missing dire players: "
		for _, id := range missingTeamB {
			missingDire += strconv.FormatUint(id, 10) + ", "
		}
		handler.DotaClient.SendChannelMessage(*channelResponse.ChannelId, missingDire[:len(missingDire)-2])
	}
}

func lobbyExpired(matchIdx string, missingTeamA []uint64, missingTeamB []uint64, handler *h.Handler) {
	log.Println("Cancelling match due to timeout", matchIdx)
	match, err := GetMatchRedis(matchIdx)
	if err != nil {
		log.Println("Failed to get match:", err)
	}

	match.Status = "cancelled"
	match.CancelReason = "reason: players didn't join in time."

	match.CancelReason = match.CancelReason[:len(match.CancelReason)-2]
	if len(missingTeamA) > 0 && len(missingTeamB) > 0 {
		match.TeamDidntShow = "both"
	} else if len(missingTeamA) > 0 {
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
}

func inviteTeams(req requests.CreateMatchReq, handler *h.Handler) {
	for _, id := range req.TeamA {
		handler.DotaClient.InviteLobbyMember(steamId.SteamId(id))
	}

	for _, id := range req.TeamB {
		handler.DotaClient.InviteLobbyMember(steamId.SteamId(id))
	}
}

func joinLobbyChannel(lobby *protocol.CSODOTALobby, handler *h.Handler) *protocol.CMsgDOTAJoinChatChannelResponse {
	lobbyKey := "Lobby_" + strconv.FormatUint(*lobby.LobbyId, 10)
	channelResponse, err := handler.DotaClient.JoinChatChannel(context.Background(), lobbyKey, protocol.DOTAChatChannelTypeT_DOTAChannelType_Lobby, false)
	if err != nil {
		log.Println("Failed to join chat channel:", err, channelResponse)
	}
	return channelResponse
}

func createLobby(handler *h.Handler, req requests.CreateMatchReq) (*protocol.CSODOTALobby, error) {
	time.Sleep(1 * time.Second)
	if res, err := handler.DotaClient.DestroyLobby(context.Background()); err != nil {
		log.Println("Failed to destroy lobby: ", err, res)
		return nil, err
	}
	time.Sleep(2 * time.Second)

	lobbyVisibility := protocol.DOTALobbyVisibility_DOTALobbyVisibility_Public

	lobbyDetails := &protocol.CMsgPracticeLobbySetDetails{
		GameName:     proto.String(req.LobbyConfig.GameName),
		Visibility:   &lobbyVisibility,
		PassKey:      proto.String(req.LobbyConfig.PassKey),
		ServerRegion: proto.Uint32(req.LobbyConfig.ServerRegion),
		GameMode:     proto.Uint32(GetGameModeFromString(req.LobbyConfig.GameMode)),
	}

	handler.DotaClient.CreateLobby(lobbyDetails)
	for i := 0; i < 5; i++ {
		time.Sleep(2 * time.Second)
		lobby, err := GetCurrentLobby(handler)
		if err != nil {
			log.Println("Failed to get lobby:", err)
			continue
		}
		return lobby, nil
	}

	return nil, errors.New("failed to create lobby")
}

func getHandler(hrs *[]*h.Handler) (*h.Handler, uint16, error) {
	for i := 0; i < 15; i++ {
		handler, handlerId, err := h.GetFreeHandler(*hrs)
		if err != nil {
			time.Sleep(5 * time.Second)
			log.Println("No available bot, retrying in 5 seconds", err)
			continue
		}

		return handler, handlerId, nil
	}

	return nil, 0, errors.New("no available bot")
}

func waitForTimeToStart(req requests.CreateMatchReq) {
	if req.StartTime != "" {
		startTime, err := time.Parse(time.RFC3339, req.StartTime)
		if err != nil {
			log.Println("Failed to parse start time:", err)
		} else {
			time.Sleep(time.Until(startTime))
		}
	}
}
