package crawler

import (
	"d2api/config"
	"d2api/pkg/models"
	"d2api/pkg/scheduled_matches"
	"d2api/pkg/wires"
	"log"
	"time"
)

func start() bool {
	for i, matchIdx := range scheduled_matches.Get() {
		match, err := wires.Instance.MatchService.GetMatch(matchIdx)
		if err != nil {
			log.Println("Failed to get match: ", err)
			continue
		}

		switch match.(type) {
		case models.MatchDetails:
			log.Println("Match details: ", match)
		case models.MatchLobby:
			log.Println("Match lobby: ", match)
		case models.MatchCancel:
			log.Println("Match cancel: ", match)
			scheduled_matches.Remove(i)
		case models.MatchData:
			log.Println("Match data: ", match)
			scheduled_matches.Remove(i)
		default:
			log.Println("Unknown match type: ", match)
		}
	}
	return true
}

func Init(config *config.Config) bool {
	ticker := time.NewTicker(time.Duration(config.Interval) * time.Second)
	quit := make(chan struct{})
	scheduled_matches.Init()
	go func() bool {
		for {
			select {
			case <-ticker.C:
				flag := start()

				if !flag {
					return false
				}
			case <-quit:
				ticker.Stop()
				return true
			}
		}
	}()

	return true
}
