package services

import (
	"context"
	"errors"
	"log"
	"strconv"
	"time"

	"d2api/config"
	"d2api/pkg/handlers"
	"d2api/pkg/redis"
	"d2api/pkg/requests"
	"d2api/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/proto"
	"github.com/paralin/go-dota2"
	"github.com/paralin/go-dota2/cso"
	"github.com/paralin/go-dota2/protocol"
	"github.com/paralin/go-steam"
	"github.com/paralin/go-steam/protocol/steamlang"
	steamId "github.com/paralin/go-steam/steamid"
	"github.com/sirupsen/logrus"
)

type MatchService struct {
	Handler *handlers.Handler
}

func NewMatchService(handler *handlers.Handler) MatchService {
	return MatchService{
		Handler: handler,
	}
}

func (s *MatchService) InitSteamConnection(config *config.Config) {
	// Grab the steam credentials from the environment
	details := &steam.LogOnDetails{
		Username: config.Steam.Username,
		Password: config.Steam.Password,
	}

	// Grab actual server list
	err := steam.InitializeSteamDirectory()
	if err != nil {
		panic(err)
	}

	// Initialize the steam client
	s.Handler.SteamClient = steam.NewClient()
	s.Handler.SteamClient.Connect()

	// Listen to events happening on steam client
	for event := range s.Handler.SteamClient.Events() {
		switch e := event.(type) {
		case *steam.ConnectedEvent:
			log.Println("Connected to steam network, trying to log in...")
			s.Handler.SteamClient.Auth.LogOn(details)
		case *steam.LoggedOnEvent:
			log.Println("Successfully logged on to steam")
			// Set account state to online
			s.Handler.SteamClient.Social.SetPersonaState(steamlang.EPersonaState_Online)

			// Once logged in, we can initialize the dota2 client
			s.Handler.DotaClient = dota2.New(s.Handler.SteamClient, logrus.New())
			s.Handler.DotaClient.SetPlaying(true)

			time.Sleep(1 * time.Second)

			// Try to get a session
			s.Handler.DotaClient.SayHello()

			eventCh, _, err := s.Handler.DotaClient.GetCache().SubscribeType(cso.Lobby) // Listen to lobby cache
			if err != nil {
				log.Fatalf("Failed to subscribe to lobby cache: %v", err)
			}

			lobbyEvent := <-eventCh
			lobby := lobbyEvent.Object.String()
			log.Printf("Lobby: %v", lobby)

		case steam.FatalErrorEvent:
			log.Println("Fatal error occurred: ", e.Error())
		}
	}
}

func (s *MatchService) ScheduleMatch(c *gin.Context, req requests.CreateMatchReq) string {
	s.Handler.DotaClient.DestroyLobby(c)
	time.Sleep(1 * time.Second)

	// Create the lobby
	lobbyVisibility := protocol.DOTALobbyVisibility_DOTALobbyVisibility_Public

	lobbyDetails := &protocol.CMsgPracticeLobbySetDetails{
		GameName:     proto.String(req.LobbyConfig.GameName),
		Visibility:   &lobbyVisibility,
		PassKey:      proto.String(req.LobbyConfig.PassKey),
		ServerRegion: proto.Uint32(req.LobbyConfig.ServerRegion),
	}

	s.Handler.DotaClient.CreateLobby(lobbyDetails)
	time.Sleep(1 * time.Second)

	s.Handler.DotaClient.SetLobbyCoach(protocol.DOTA_GC_TEAM_DOTA_GC_TEAM_GOOD_GUYS)
	time.Sleep(1 * time.Second)

	// Invite the teamA
	for _, id := range req.TeamA {
		s.Handler.DotaClient.InviteLobbyMember(steamId.SteamId(id))
	}

	// Invite the teamB
	for _, id := range req.TeamB {
		s.Handler.DotaClient.InviteLobbyMember(steamId.SteamId(id))
	}

	matchIdx := strconv.FormatInt(time.Now().UnixNano(), 10)
	redis.RedisClient.Set(context.Background(), matchIdx, "", 0)

	go runningThread(s.Handler, req, matchIdx)
	return matchIdx
}

func (s *MatchService) GetMatchDetails(c *gin.Context, matchIdx string) (*protocol.CMsgGCMatchDetailsResponse, error) {
	matchId, err := redis.RedisClient.Get(context.Background(), matchIdx).Result()
	if err != nil {
		return nil, errors.New("match not found")
	}

	if matchId == "" {
		return nil, errors.New("match hasn't started yet")
	}

	matchIdNum, err := strconv.ParseUint(matchId, 10, 64)
	if err != nil {
		return nil, err
	}

	details, err := s.Handler.DotaClient.RequestMatchDetails(c, matchIdNum)
	if err != nil {
		return nil, err
	}

	return details, nil
}

func runningThread(handler *handlers.Handler, req requests.CreateMatchReq, matchIdx string) {
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		log.Fatalf("Failed to parse start time: %v", err)
		return
	}

	time.Sleep(time.Until(startTime))

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

		if lobby.MatchId != nil {
			continue
		}

		redis.RedisClient.Set(context.Background(), matchIdx, *lobby.MatchId, 0)
		break
	}

	//Abandon the lobby
	handler.DotaClient.AbandonLobby()
}

func (s *MatchService) GetLobby(c *gin.Context) (*protocol.CSODOTALobby, error) {
	lobby, err := utils.GetCurrentLobby(s.Handler)
	if err != nil {
		log.Fatalf("Failed to get lobby: %v", err)
		return nil, err
	}
	return lobby, nil
}
