package services

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"d2api/config"
	"d2api/pkg/handlers"
	"d2api/pkg/models"
	"d2api/pkg/repository"
	"d2api/pkg/requests"
	"d2api/pkg/response"
	"d2api/pkg/scheduled_matches"
	"d2api/pkg/utils"

	"github.com/jasonodonnell/go-opendota"
	"github.com/paralin/go-dota2/protocol"
	"github.com/paralin/go-steam/steamid"
	"go.mongodb.org/mongo-driver/bson"
)

type MatchService struct {
	Config *config.Config
	Repo   *repository.Repository
}

func NewMatchService(config *config.Config, repo *repository.Repository) MatchService {
	return MatchService{
		Config: config,
		Repo:   repo,
	}
}

func (s *MatchService) ScheduleMatch(req requests.CreateMatchReq) (string, error) {

	matchIdx := strconv.FormatInt(time.Now().UnixNano(), 10)
	utils.SetMatchRedis(matchIdx, models.MatchDetails{
		MatchStatus: models.MatchStatus{Status: "scheduled", MatchId: 0, IsTournamentMatch: false},
	})

	scheduled_matches.Add(matchIdx)

	go utils.MatchScheduleThread(req, matchIdx, s.Config.TimeToCancel)
	return matchIdx, nil
}

func (s *MatchService) GetMatch(matchIdx string) (interface{}, error) {
	match, err := utils.GetMatchRedis(matchIdx)
	if err != nil {
		var match models.MatchMongo
		err := s.Repo.Get("matches", bson.M{"_id": matchIdx}).Decode(&match)
		if err != nil {
			return nil, err
		}

		return match, nil
	}

	handler, err := handlers.Hs.GetMatchHandler(match.Handler)
	if err != nil {
		return nil, err
	}

	if match.Status == "cancelled" {
		return models.MatchCancel{MatchStatus: match.MatchStatus, Reason: match.CancelReason, TeamDidntShow: match.TeamDidntShow}, nil
	} else if match.Status == "scheduled" {
		lobby, err := utils.GetCurrentLobby(handler)
		if err != nil {
			return nil, err
		}
		return models.MatchLobby{MatchStatus: match.MatchStatus, Lobby: lobby}, nil
	}

	details, err := handler.DotaClient.RequestMatchDetails(context.Background(), match.MatchId)
	if err != nil {
		return nil, err
	}

	if *details.Result == 1 {
		match.MatchStatus.Status = "finished"
		err = utils.SetMatchRedis(matchIdx, *match)
		if err != nil {
			log.Println("Failed to set match:", err)
		}

		return models.MatchData{MatchStatus: match.MatchStatus, Match: details.Match}, nil
	} else if *details.Result == 2 {
		return match.MatchStatus, nil
	} else {
		return nil, errors.New("match not found")
	}
}

func (s *MatchService) GetMatchInfo(matchIdx string) (*response.MatchInfo, error) {
	data, err := s.GetMatch(matchIdx)
	if err != nil {
		return nil, err
	}

	var matchInfo *response.MatchInfo
	switch match := data.(type) {
	case models.MatchLobby:
		log.Println("MatchLobby")
		matchInfo = &response.MatchInfo{Status: match.MatchStatus.Status}
		for _, player := range match.TeamA {
			playerInfo := response.Player{SteamId: player, IsInLobby: false, IsInRightTeam: false}
			for _, lobbyPlayer := range match.Lobby.AllMembers {
				if *lobbyPlayer.Id == playerInfo.SteamId {
					playerInfo.IsInLobby = true
					if *lobbyPlayer.Team == *protocol.DOTA_GC_TEAM_DOTA_GC_TEAM_GOOD_GUYS.Enum() {
						playerInfo.IsInRightTeam = true
					}
					break
				}
			}

			matchInfo.RadiantPlayers = append(matchInfo.RadiantPlayers, playerInfo)
		}

		for _, player := range match.TeamB {
			playerInfo := response.Player{SteamId: player, IsInLobby: false, IsInRightTeam: false}
			for _, lobbyPlayer := range match.Lobby.AllMembers {
				if *lobbyPlayer.Id == playerInfo.SteamId {
					playerInfo.IsInLobby = true
					if *lobbyPlayer.Team == *protocol.DOTA_GC_TEAM_DOTA_GC_TEAM_BAD_GUYS.Enum() {
						playerInfo.IsInRightTeam = true
					}
					break
				}
			}
			matchInfo.DirePlayers = append(matchInfo.DirePlayers, playerInfo)
		}

	case models.MatchData:
		log.Println("MatchData")
		matchInfo = &response.MatchInfo{Status: match.MatchStatus.Status}
	case models.MatchCancel:
		log.Println("MatchCancel")
		matchInfo = &response.MatchInfo{Status: match.MatchStatus.Status, Cancelled: true}
	case models.MatchDetails:
		log.Println("MatchDetails")
		matchInfo = &response.MatchInfo{Status: match.MatchStatus.Status}
	case models.MatchMongo:
		log.Println("MatchMongo")
		matchInfo = &response.MatchInfo{Status: "finished"}
	default:
		return nil, errors.New("unknown match type")
	}

	return matchInfo, nil
}

func (s *MatchService) GetPlayerHistoryOpenDota(steamId int64, limit int) (interface{}, error) {
	client := opendota.NewClient(http.DefaultClient)
	matches, _, err := client.PlayerService.Matches(steamId, &opendota.PlayerParam{Limit: limit})
	if err != nil {
		return nil, err
	}

	return matches, nil
}

func (s *MatchService) GetPlayerHistory(steamId int64, limit int) (interface{}, error) {
	player := s.Repo.Get("players", bson.M{"_id": steamId})
	if player.Err() != nil {
		return nil, player.Err()
	}

	var playerModel models.Player
	if err := player.Decode(&playerModel); err != nil {
		return nil, err
	}

	var matchIds []string
	if len(playerModel.Matches) < limit {
		matchIds = playerModel.Matches
	} else {
		matchIds = playerModel.Matches[:limit]
	}

	var matches []*models.MatchMongo
	channel := make(chan *models.MatchMongo)
	for _, matchId := range matchIds {
		go func(matchId string) {
			var match models.MatchMongo
			err := s.Repo.Get("matches", bson.M{"_id": matchId}).Decode(&match)
			if err != nil {
				log.Println("Failed to get match:", err)
				channel <- nil
				return
			}

			channel <- &match
		}(matchId)
	}

	for range matchIds {
		match := <-channel
		if match != nil {
			matches = append(matches, match)
		}
	}

	return matches, nil
}

func (s *MatchService) ReinvitePlayers(req requests.ReinvitePlayersReq) error {
	match, err := utils.GetMatchRedis(strconv.Itoa(req.MatchIdx))
	if err != nil {
		return err
	}

	handler, err := handlers.Hs.GetMatchHandler(match.Handler)
	if err != nil {
		return err
	}

	for _, player := range req.Players {
		handler.DotaClient.InviteLobbyMember(steamid.SteamId(player))
	}

	return nil
}
