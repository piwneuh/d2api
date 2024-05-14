package services

import (
	"d2api/config"
	"d2api/pkg/handlers"
	"d2api/pkg/models"
	"d2api/pkg/repository"
	"d2api/pkg/requests"
	"d2api/pkg/scheduled_matches"
	"d2api/pkg/utils"
	"errors"
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

func (t *TournamentService) ScheduleRound(round []requests.MatchForMiddleware) error {
	tourRound := models.TournamentRound{
		Finished:     false,
		TournamentId: round[0].TournamentId,
	}

	for _, match := range round {
		tourRound.Matches = append(tourRound.Matches, models.MatchModel{
			MatchForMiddleware: match,
			Finished:           false,
		})
	}

	_, err := t.Repo.Insert("round", tourRound)
	if err != nil {
		return err
	}

	for _, match := range round {
		matchIdx := strconv.Itoa(match.MatchIdx)

		utils.SetMatchRedis(matchIdx, models.MatchDetails{
			MatchStatus: models.MatchStatus{Status: "scheduled", MatchId: 0, IsTournamentMatch: true},
		})

		scheduled_matches.Add(matchIdx)

		req, err := createTournamentMatch(&match)
		if err != nil {
			return err
		}

		go utils.MatchScheduleThread(&t.Handlers, *req, matchIdx, t.Config.TimeToCancel)
	}

	return nil
}

func createTournamentMatch(match *requests.MatchForMiddleware) (*requests.CreateMatchReq, error) {
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

		if player.Team == match.Team1Id {
			req.TeamA = append(req.TeamA, id)
		} else {
			req.TeamB = append(req.TeamB, id)
		}
	}

	startTime := time.Unix(match.StartEpoch, 0)
	req.StartTime = startTime.Format(time.RFC3339)
	return &req, nil
}
