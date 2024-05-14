package wires

import (
	"d2api/config"
	"d2api/pkg/handlers"
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
	handlers, err := handlers.LoadHandlers(config.InventoryPath)
	if err != nil {
		panic(err)
	}

	Repo = repository.NewRepository(mongodb.Instance.Database)

	Instance = &Wires{
		MatchService:      services.NewMatchService(handlers, config, Repo),
		TournamentService: services.NewTournamentService(handlers, config, Repo),
	}
}
