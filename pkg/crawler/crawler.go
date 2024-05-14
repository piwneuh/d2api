package crawler

import (
	"d2api/config"
	"d2api/pkg/models"
	"d2api/pkg/scheduled_matches"
	"d2api/pkg/wires"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func crawlMatches() {
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
			log.Println("Match data: ", match)
			if match.IsTournamentMatch {
				log.Println("Match is tournament match, data")
				log.Println("Match", match)
			}
			scheduled_matches.Remove(i)

			opts := options.Update().SetUpsert(true)
			for _, player := range match.Match.Players {
				wires.Repo.Update("players", player.AccountId, bson.M{
					"$push": bson.M{"matches": bson.M{
						"$each":     bson.A{match.Match.MatchId},
						"$position": 0,
					}},
				}, opts)
			}
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
				crawlMatches()
			case <-quit:
				ticker.Stop()
				return true
			}
		}
	}()

	return true
}
