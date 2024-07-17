package utils

import (
	"context"
	h "d2api/pkg/handlers"
	"d2api/pkg/models"
	"d2api/pkg/requests"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	steamId "github.com/paralin/go-steam/steamid"

	"github.com/paralin/go-dota2/protocol"
	"google.golang.org/protobuf/proto"
)

func MatchScheduleThread(req requests.CreateMatchReq, matchIdx string, timeToCancel uint32) {
	waitForTimeToStart(req)
	handler, err := getHandler()
	if err != nil {
		log.Println("Failed to get handler:", err)
		return
	}

	if ok := getAndSetMatchToRedis(matchIdx, handler, req.TeamA, req.TeamB); !ok {
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
	steamIdsUsernames := getSteamUsernamesFromSteamIds(append(req.TeamA, req.TeamB...))

	lobbyExpirationTime := time.Now().Add(time.Duration(timeToCancel) * time.Second)
	i := 0
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

		if i = (i + 1) % 5; i == 1 {
			sendMissingPlayersMessages(handler, channelResponse, missingTeamA, missingTeamB, steamIdsUsernames)
		}

		if time.Now().After(lobbyExpirationTime) {
			match, err := GetMatchRedis(matchIdx)
			if err != nil {
				log.Println("Failed to get match:", err)
			}

			var players []string
			for _, player := range match.TourMatch.Players {
				players = append(players, player.SteamId)
			}
			players = append(players, match.TourMatch.TournamentOwnerId)

			metadata := make(map[string]string)
			metadata["match_id"] = strconv.Itoa(match.TourMatch.MatchIdx)
			metadata["team1_clan_id"] = strconv.Itoa(match.TourMatch.Team1Id)
			metadata["team1_clan_id"] = strconv.Itoa(match.TourMatch.Team2Id)
			metadata["tournament_name"] = match.TourMatch.TournamentName
			metadata["tournament_logo"] = match.TourMatch.TournamentLogo

			if len(missingTeamA) > 0 && len(missingTeamB) > 0 {
				metadata["abandoning_team1_name"] = match.TourMatch.Team1.Name
				metadata["abandoning_team1_id"] = strconv.Itoa(match.TourMatch.Team1Id)
				metadata["abandoning_team2_name"] = match.TourMatch.Team2.Name
				metadata["abandoning_team2_id"] = strconv.Itoa(match.TourMatch.Team2Id)
			} else if len(missingTeamA) > 0 {
				metadata["abandoning_team_name"] = match.TourMatch.Team1.Name
				metadata["abandoning_team_id"] = strconv.Itoa(match.TourMatch.Team1Id)
			} else if len(missingTeamB) > 0 {
				metadata["abandoning_team_name"] = match.TourMatch.Team2.Name
				metadata["abandoning_team_id"] = strconv.Itoa(match.TourMatch.Team2Id)
			}

			lobbyExpired(matchIdx, missingTeamA, missingTeamB, handler)
			message := fmt.Sprintf("Match Abandoned: %s vs %s", match.TourMatch.Team1.Name, match.TourMatch.Team2.Name)
			SendNotification(int64(match.TourMatch.TournamentId), players, message, "tournament", "MATCH_ABANDONED", metadata)
			return
		}
	}

	startMatch(handler, channelResponse, matchIdx)
}

func startMatch(handler *h.Handler, channelResponse *protocol.CMsgDOTAJoinChatChannelResponse, matchIdx string) {
	handler.DotaClient.SendChannelMessage(*channelResponse.ChannelId, "All players joined the lobby")
	time.Sleep(2 * time.Second)

	handler.DotaClient.LaunchLobby()

	var match *models.MatchDetails
	for {
		time.Sleep(2 * time.Second)
		lobby, err := GetCurrentLobby(handler)
		if err != nil {
			log.Println("Failed to get lobby:", err)
		}

		if lobby.MatchId == nil {
			continue
		}

		match, err = GetMatchRedis(matchIdx)
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

	metadata := make(map[string]string)
	metadata["match_id"] = strconv.Itoa(match.TourMatch.MatchIdx)
	metadata["team1_clan_id"] = strconv.Itoa(match.TourMatch.Team1Id)
	metadata["team2_clan_id"] = strconv.Itoa(match.TourMatch.Team2Id)
	metadata["team1_name"] = match.TourMatch.Team1.Name
	metadata["team2_name"] = match.TourMatch.Team2.Name
	metadata["team1_image"] = match.TourMatch.Team1.Logo
	metadata["team2_image"] = match.TourMatch.Team2.Logo
	metadata["tournament_name"] = match.TourMatch.TournamentName
	metadata["tournament_logo"] = match.TourMatch.TournamentLogo
	metadata["tournament_creator_steamId"] = match.TourMatch.TournamentOwnerId

	var playerIds []string
	for _, player := range match.TourMatch.Players {
		playerIds = append(playerIds, player.SteamId)
	}
	if match.IsTournamentMatch {
		message := fmt.Sprintf("Match starting: %s vs %s", match.TourMatch.Team1.Name, match.TourMatch.Team2.Name)
		SendNotification(int64(match.TourMatch.TournamentId), append(playerIds, match.TourMatch.TournamentOwnerId), message, "tournament", "MATCH_STARTING", metadata)
	}
	handler.DotaClient.AbandonLobby()
	handler.Occupied = false
}

func sendMissingPlayersMessages(handler *h.Handler, channelResponse *protocol.CMsgDOTAJoinChatChannelResponse, missingTeamA []uint64, missingTeamB []uint64, steamIdsUsernames map[uint64]string) {
	handler.DotaClient.SendChannelMessage(*channelResponse.ChannelId, "Waiting for players to join the lobby")
	sendMissingTeamMessage := func(handler *h.Handler, channelResponse *protocol.CMsgDOTAJoinChatChannelResponse, missingTeamA []uint64, steamIdsUsernames map[uint64]string, team string) {
		handler.DotaClient.SendChannelMessage(*channelResponse.ChannelId, "Missing "+team+" players:")
		handler.DotaClient.SendChannelMessage(*channelResponse.ChannelId, "-----------------------------")
		for _, id := range missingTeamA {
			var username string
			if val, ok := steamIdsUsernames[id]; ok {
				username = val
			} else {
				username = strconv.FormatUint(id, 10)
			}
			handler.DotaClient.SendChannelMessage(*channelResponse.ChannelId, "| "+username+" |")
		}
		handler.DotaClient.SendChannelMessage(*channelResponse.ChannelId, "-----------------------------")
	}

	if len(missingTeamA) > 0 {
		sendMissingTeamMessage(handler, channelResponse, missingTeamA, steamIdsUsernames, "radiant")
	}

	if len(missingTeamB) > 0 {
		sendMissingTeamMessage(handler, channelResponse, missingTeamB, steamIdsUsernames, "dire")
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
		handler.Broken = true
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

func getHandler() (*h.Handler, error) {
	for i := 0; i < 15; i++ {
		handler, err := h.Hs.GetFreeHandler()
		if err != nil {
			time.Sleep(5 * time.Second)
			log.Println("No available bot, retrying in 5 seconds", err)
			continue
		}

		return handler, nil
	}

	return nil, errors.New("no available bot")
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
