package services

import (
	"d2api/config"
	"d2api/pkg/handlers"
	"d2api/pkg/models"
	"d2api/pkg/repository"
	"d2api/pkg/requests"
	"d2api/pkg/response"
	"d2api/pkg/scheduled_matches"
	"d2api/pkg/utils"
	"errors"
	"fmt"
	"strconv"
	"time"
)

type TournamentService struct {
	Handlers []*handlers.Handler
	Config   *config.Config
	Repo     *repository.Repository
}

func NewTournamentService(handlers []*handlers.Handler, config *config.Config, repository *repository.Repository) TournamentService {
	return TournamentService{
		Handlers: handlers,
		Config:   config,
		Repo:     repository,
	}
}

func (t *TournamentService) ScheduleRound(round []requests.TourMatch) ([]response.TournamentEndRequest, error) {
	var cancelled []response.TournamentEndRequest
	for _, match := range round {
		if match.Cancelled {
			winner := -1
			loser := -1
			score := -1

			var playerIds []string

			if match.Team1Id != -1 {
				winner = match.Team1Id
				loser = match.Team2Id

				score = 0
			} else if match.Team2Id != -1 {
				winner = match.Team2Id
				loser = match.Team1Id

				score = 0
			}

			moveTeams := response.TournamentEndRequest{
				TourId:    match.TournamentId,
				MatchId:   match.MatchIdx,
				Cancelled: true,
				Iteration: match.Iteration,
				Winner: response.TeamEnd{
					TeamId: winner,
					Score:  score,
				},
				Loser: response.TeamEnd{
					TeamId: loser,
					Score:  -1,
				},
			}

			playerIds = append(playerIds, match.TournamentOwnerId)

			metadata := make(map[string]string)
			metadata["match_id"] = strconv.Itoa(match.MatchIdx)
			metadata["team1_clan_id"] = strconv.Itoa(match.Team1Id)
			metadata["team2_clan_id"] = strconv.Itoa(match.Team2Id)
			metadata["tournament_name"] = match.TournamentName
			metadata["tournament_image"] = match.TournamentLogo
			utils.SendNotification(int64(match.TournamentId), playerIds, fmt.Sprintf("Match Cancelled: %s vs %s", match.Team1.Name, match.Team2.Name), "tournament", "MATCH_CANCELLED", metadata)

			cancelled = append(cancelled, moveTeams)
		} else {
			matchIdx := strconv.Itoa(match.MatchIdx)

			matchDetails := models.MatchDetails{
				MatchStatus: models.MatchStatus{Status: "scheduled", MatchId: 0, IsTournamentMatch: true, TourMatch: match},
			}

			utils.SetMatchRedis(matchIdx, matchDetails)

			scheduled_matches.Add(matchIdx)

			req, err := createTournamentMatch(&match)
			if err != nil {
				return make([]response.TournamentEndRequest, 0), nil
			}

			go utils.MatchScheduleThread(&t.Handlers, *req, matchIdx, t.Config.TimeToCancel)
		}
	}

	return cancelled, nil
}

func createTournamentMatch(match *requests.TourMatch) (*requests.CreateMatchReq, error) {
	var req requests.CreateMatchReq
	key := strconv.Itoa(match.TournamentId) + "_" + strconv.Itoa(match.MatchIdx)
	region, err := strconv.ParseUint(match.Region, 10, 32)
	if err != nil {
		region = 3
	}

	req.LobbyConfig = requests.LobbyConfig{
		GameName:     key,
		PassKey:      key,
		ServerRegion: uint32(region),
		GameMode:     "AP",
	}

	for _, player := range match.Players {
		id, err := strconv.ParseUint(player.SteamId, 10, 64)
		if err != nil {
			return nil, errors.New("wrong steam id for player " + player.SteamId)
		}

		if player.Team == "team1" {
			req.TeamA = append(req.TeamA, id)
		} else {
			req.TeamB = append(req.TeamB, id)
		}
	}

	startTime := time.UnixMilli(match.StartEpoch)
	req.StartTime = startTime.Format(time.RFC3339)
	return &req, nil
}
