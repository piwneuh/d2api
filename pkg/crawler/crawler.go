package crawler

import (
	"d2api/config"
	"d2api/pkg/models"
	"d2api/pkg/response"
	"d2api/pkg/scheduled_matches"
	"d2api/pkg/utils"
	"d2api/pkg/wires"
	"log"
	"time"

	"github.com/paralin/go-dota2/protocol"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func matchFinished(match *models.MatchData, tournamentEndpoint string, i int) {
	if match.IsTournamentMatch {
		outcome := match.Match.GetMatchOutcome()
		radiant := response.TeamEnd{
			Score:  int(*match.Match.RadiantTeamScore),
			TeamId: match.TournamentMatchModel.Team1Id,
		}
		dire := response.TeamEnd{
			Score:  int(*match.Match.DireTeamScore),
			TeamId: match.TournamentMatchModel.Team2Id,
		}

		resp := response.TournamentEndRequest{
			TourId:    match.TournamentMatchModel.TournamentId,
			MatchId:   match.TournamentMatchModel.MatchIdx,
			Iteration: 1,
			Cancelled: false,
		}

		if outcome == protocol.EMatchOutcome_k_EMatchOutcome_RadVictory {
			resp.Winner = radiant
			resp.Loser = dire
		} else {
			resp.Winner = dire
			resp.Loser = radiant
		}

		sent := false
		for i := 0; i < 5; i++ {
			sent = utils.SendMatchResultToTournament(tournamentEndpoint, &resp)
			if sent {
				break
			}
		}

		if !sent {
			log.Println("Could not send match result to tournament")
		}
	}

	opts := options.Update().SetUpsert(true)
	for _, player := range match.Match.Players {
		playerId := *player.AccountId
		matchId := *match.Match.MatchId

		go func(playerId uint32, matchId uint64) {
			_, err := wires.Repo.Update("players", bson.M{"_id": playerId}, bson.M{
				"$push": bson.M{"matches": bson.M{
					"$each":     bson.A{matchId},
					"$position": 0,
				}},
			}, opts)

			if err != nil {
				log.Println("Failed to update player: ", err)
			}
		}(playerId, matchId)
	}

	scheduled_matches.Remove(i)
}

func crawlMatches(cfg *config.Config) {
	tournamentEndpoint := cfg.Tournament.URL + "/tournaments/move-teams-to-next-round"

	for i, matchIdx := range scheduled_matches.Get() {
		match, err := wires.Instance.MatchService.GetMatch(matchIdx)
		if err != nil {
			log.Println("Failed to get match: ", err)
			continue
		}

		// matchIdxInt, err := strconv.Atoi(matchIdx)
		// if err != nil {
		// 	log.Println("Failed to convert matchIdx to int: ", err)
		// 	continue
		// }

		switch match := match.(type) {
		case models.MatchCancel:
			log.Println("Match cancel: ", match)
			if match.IsTournamentMatch {
				log.Println("Match is tournament match, cancelled")
				log.Println("Match", match)
			}
			scheduled_matches.Remove(i)

		case models.MatchData:
			log.Println("Match finished: ")
			log.Println("Outcome: ", protocol.EMatchOutcome_name[int32(match.Match.GetMatchOutcome())])
			matchFinished(&match, tournamentEndpoint, i)
		default:
			continue
		}
	}
}

func Init(config *config.Config) bool {
	ticker := time.NewTicker(time.Duration(config.Interval) * time.Second)
	quit := make(chan struct{})
	scheduled_matches.Init()
	go func() bool {
		for {
			select {
			case <-ticker.C:
				crawlMatches(config)
			case <-quit:
				ticker.Stop()
				return true
			}
		}
	}()

	return true
}
