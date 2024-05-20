package scheduled_matches

import (
	"d2api/pkg/utils"
)

var scheduledMatches []string

func Get() []string {
	return scheduledMatches
}

func Remove(matchIdx string) {
	for i, idx := range scheduledMatches {
		if idx == matchIdx {
			scheduledMatches = append(scheduledMatches[:i], scheduledMatches[i+1:]...)
			return
		}
	}
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
