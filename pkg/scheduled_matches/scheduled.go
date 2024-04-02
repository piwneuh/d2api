package scheduled_matches

import (
	"d2api/pkg/utils"
)

var scheduledMatches []string

func Get() []string {
	return scheduledMatches
}

func Remove(i int) {
	if i == len(scheduledMatches)-1 {
		scheduledMatches = scheduledMatches[:i]
		return
	}
	scheduledMatches = append(scheduledMatches[:i], scheduledMatches[i+1:]...)
}

func Init() {
	scheduledMatches = []string{}
	matchIdxs, err := utils.GetAllMatchIdxs()
	if err != nil {
		return
	}

	for _, idx := range matchIdxs {
		match, err := utils.GetMatchRedis(idx)
		if err != nil {
			continue
		}

		if match.Status == "scheduled" || match.Status == "started" {
			scheduledMatches = append(scheduledMatches, idx)
		}
	}
}

func Add(matchIdx string) {
	scheduledMatches = append(scheduledMatches, matchIdx)
}
