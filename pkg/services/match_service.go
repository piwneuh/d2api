package services

import (
	"context"
	"errors"
	"log"
	"strconv"
	"time"

	"d2api/pkg/handlers"
	"d2api/pkg/requests"
	"d2api/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/proto"
	"github.com/paralin/go-dota2/protocol"
	steamId "github.com/paralin/go-steam/steamid"
)

type MatchService struct {
	Handlers []*handlers.Handler
}

func NewMatchService(handlers []*handlers.Handler) MatchService {
	return MatchService{
		Handlers: handlers,
	}
}

func (s *MatchService) ScheduleMatch(c *gin.Context, req requests.CreateMatchReq) (string, error) {
	handler, handlerId, err := handlers.GetFreeHandler(s.Handlers)
	if err != nil {
		return "", err
	}

	handler.DotaClient.DestroyLobby(context.Background())
	time.Sleep(1 * time.Second)

	// Create the lobby
	lobbyVisibility := protocol.DOTALobbyVisibility_DOTALobbyVisibility_Public

	lobbyDetails := &protocol.CMsgPracticeLobbySetDetails{
		GameName:     proto.String(req.LobbyConfig.GameName),
		Visibility:   &lobbyVisibility,
		PassKey:      proto.String(req.LobbyConfig.PassKey),
		ServerRegion: proto.Uint32(req.LobbyConfig.ServerRegion),
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

	matchIdx := strconv.FormatInt(time.Now().UnixNano(), 10)
	utils.SetMatchRedis(matchIdx, utils.MatchDetails{
		MatchId:   0,
		HandlerId: uint16(handlerId),
	})

	go runningThread(handler, req, matchIdx)
	return matchIdx, nil
}

func runningThread(handler *handlers.Handler, req requests.CreateMatchReq, matchIdx string) {
	if req.StartTime != "" {
		startTime, err := time.Parse(time.RFC3339, req.StartTime)
		if err != nil {
			log.Fatalf("Failed to parse start time: %v", err)
			return
		}

		time.Sleep(time.Until(startTime))
	}

	for {
		time.Sleep(2 * time.Second)
		lobby, err := utils.GetCurrentLobby(handler)
		if err != nil {
			log.Fatalf("Failed to get lobby: %v", err)
		}

		goodGuys, badGuys, err := utils.GetGoodAndBadGuys(lobby)
		if err != nil {
			log.Fatalf("Failed to get good and bad guys: %v", err)
		}

		if utils.AreAllPlayerHere(goodGuys, badGuys, &req) {
			break
		}
	}

	// Start the game
	handler.DotaClient.LaunchLobby()
	for {
		time.Sleep(2 * time.Second)
		lobby, err := utils.GetCurrentLobby(handler)
		if err != nil {
			log.Fatalf("Failed to get lobby: %v", err)
		}

		if lobby.MatchId == nil {
			continue
		}

		match, err := utils.GetMatchRedis(matchIdx)
		if err != nil {
			log.Fatalf("Failed to get match: %v", err)
		}

		match.MatchId = *lobby.MatchId
		err = utils.SetMatchRedis(matchIdx, *match)
		if err != nil {
			log.Fatalf("Failed to set match: %v", err)
		}

		break
	}

	//Abandon the lobby
	handler.DotaClient.AbandonLobby()
	handler.Occupied = false
}

func (s *MatchService) GetLobby(c *gin.Context, matchIdx string) (*protocol.CSODOTALobby, error) {
	match, err := utils.GetMatchRedis(matchIdx)
	if err != nil {
		return nil, err
	}

	if match.MatchId != 0 {
		return nil, errors.New("match has already started")
	}

	lobby, err := utils.GetCurrentLobby(s.Handlers[match.HandlerId])
	if err != nil {
		log.Fatalf("Failed to get lobby: %v", err)
		return nil, err
	}
	return lobby, nil
}

func (s *MatchService) GetMatchDetails(c *gin.Context, matchIdx string) (*protocol.CMsgGCMatchDetailsResponse, error) {
	match, err := utils.GetMatchRedis(matchIdx)
	if err != nil {
		return nil, err
	}

	if match.MatchId == 0 {
		return nil, errors.New("match hasn't started yet")
	}

	details, err := s.Handlers[match.HandlerId].DotaClient.RequestMatchDetails(c, match.MatchId)
	if err != nil {
		return nil, err
	}

	return details, nil
}
