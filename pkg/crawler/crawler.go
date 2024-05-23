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

func safeSending(data *response.TournamentEndRequest, repeatTimes int) {
	for i := 0; i < repeatTimes; i++ {
		if utils.SendMatchResultToTournament(tournamentEndpoint, data) {
			log.Println("Match result sent to tournament")
			return
		}
	}
	log.Println("Failed to send match result to tournament")
}

func sendMatchFinished(match *models.MatchData) {
	outcome := match.Match.GetMatchOutcome()
	radiant := response.TeamEnd{
		Score:  int(*match.Match.RadiantTeamScore),
		TeamId: match.TourMatch.Team1Id,
	}
	dire := response.TeamEnd{
		Score:  int(*match.Match.DireTeamScore),
		TeamId: match.TourMatch.Team2Id,
	}

	resp := response.TournamentEndRequest{
		TourId:    match.TourMatch.TournamentId,
		MatchId:   match.TourMatch.MatchIdx,
		Iteration: match.TourMatch.Iteration,
		Cancelled: false,
	}

	if outcome == protocol.EMatchOutcome_k_EMatchOutcome_RadVictory {
		resp.Winner = radiant
		resp.Loser = dire
	} else {
		resp.Winner = dire
		resp.Loser = radiant
	}

	safeSending(&resp, 5)
}

func matchFinished(match *models.MatchData) {
	if match.IsTournamentMatch {
		// Send match finished to tournament service
		go sendMatchFinished(match)
	}

	// Save player history
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
}

func sendMatchCancelled(match *models.MatchCancel) {
	radiant := response.TeamEnd{
		TeamId: match.TourMatch.Team1Id,
		Score:  0,
	}
	dire := response.TeamEnd{
		TeamId: match.TourMatch.Team2Id,
		Score:  0,
	}

	resp := response.TournamentEndRequest{
		TourId:    match.TourMatch.TournamentId,
		MatchId:   match.TourMatch.MatchIdx,
		Iteration: match.TourMatch.Iteration,
		Cancelled: true,
	}

	switch match.TeamDidntShow {
	case "teamA":
		radiant.Score = -1
		resp.Winner = dire
		resp.Loser = radiant
	case "teamB":
		dire.Score = -1
		resp.Winner = radiant
		resp.Loser = dire
	default:
		radiant.Score = -1
		dire.Score = -1
		resp.Winner = radiant
		resp.Loser = dire
	}

	safeSending(&resp, 5)
}

func matchCancelled(match *models.MatchCancel) {
	log.Println("Match cancel: ", match)
	if match.IsTournamentMatch {
		go sendMatchCancelled(match)
	}
}

func crawlMatches() {
	toRemove := []string{}
	for _, matchIdx := range scheduled_matches.Get() {
		match, err := wires.Instance.MatchService.GetMatch(matchIdx)
		if err != nil {
			log.Println("Failed to get match: ", err)
			continue
		}

		switch match := match.(type) {
		case models.MatchCancel:
			log.Println("Match cancelled: ", match)
			matchCancelled(&match)
			toRemove = append(toRemove, matchIdx)
		case models.MatchData:
			log.Println("Match finished: ")
			log.Println("Outcome: ", protocol.EMatchOutcome_name[int32(match.Match.GetMatchOutcome())])
			matchFinished(&match)
			toRemove = append(toRemove, matchIdx)
		default:
			continue
		}
	}

	for _, idx := range toRemove {
		scheduled_matches.Remove(idx)
	}
}

var tournamentEndpoint string

func Init(config *config.Config) bool {
	ticker := time.NewTicker(time.Duration(config.Interval) * time.Second)
	tournamentEndpoint = config.Tournament.URL + "/tournaments/move-teams-to-next-round"
	quit := make(chan struct{})
	scheduled_matches.Init()
	go func() bool {
		for {
			select {
			case <-ticker.C:
				crawlMatches()
			case <-quit:
				ticker.Stop()
				return true
			}
		}
	}()

	return true
}
