package crawler

import (
	"d2api/config"
	"d2api/pkg/models"
	"d2api/pkg/scheduled_matches"
	"d2api/pkg/wires"
	"log"
	"time"

	"github.com/paralin/go-dota2/protocol"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// func matchFinished(match *models.MatchData, tournamentEndpoint string) {
// if match.IsTournamentMatch {
// 	res := wires.Repo.Get("round", bson.M{"tournamentId": match.TournamentId, "finished": false})
// 	if res.Err() != nil {
// 		log.Println("Failed to get round: ", res.Err())
// 		return
// 	}

// 	var round models.TournamentRound
// 	if err := res.Decode(&round); err != nil {
// 		log.Println("Failed to decode round: ", err)
// 		return
// 	}

// 	round.Finished = true
// 	for _, roundMatch := range round.Matches {
// 		if match.MatchId == uint64(roundMatch.MatchIdx) {
// 			roundMatch.Finished = true
// 		}

// 		if !roundMatch.Finished {
// 			round.Finished = false
// 		}
// 	}

// 	_, err := wires.Repo.Update("round", bson.M{"tournamentId": match.TournamentId, "finished": false}, round)
// 	if err != nil {
// 		log.Println("Failed to update round: ", err)
// 		return
// 	}

// 	var winner, loser response.TeamEnd
// 	outcome := match.Match.GetMatchOutcome()
// 	if outcome == protocol.EMatchOutcome_k_EMatchOutcome_RadVictory {
// 		winner.TeamId = round.
// 	} else {

// 	}
// }

// 	opts := options.Update().SetUpsert(true)
// 	for _, player := range match.Match.Players {
// 		_, err := wires.Repo.Update("players", player.AccountId, bson.M{
// 			"$push": bson.M{"matches": bson.M{
// 				"$each":     bson.A{match.Match.MatchId},
// 				"$position": 0,
// 			}},
// 		}, opts)
// 		if err != nil {
// 			log.Println("Failed to update player: ", err)
// 			continue
// 		}
// 	}
// }

func crawlMatches(cfg *config.Config) {
	// tournamentEndpoint := cfg.Tournament.URL + "/tournaments/move-teams-to-next-round"
	_ = cfg

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
