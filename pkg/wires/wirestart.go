package wires

import (
	"d2api/config"
	"d2api/pkg/mongodb"
	"d2api/pkg/repository"
	"d2api/pkg/services"
)

type Wires struct {
	MatchService      services.MatchService
	TournamentService services.TournamentService
}

var Instance *Wires
var Repo *repository.Repository

func Init(config *config.Config) {
	Repo = repository.NewRepository(mongodb.Instance.Database)

	Instance = &Wires{
		MatchService:      services.NewMatchService(config, Repo),
		TournamentService: services.NewTournamentService(config, Repo),
	}
}
