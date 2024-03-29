package wires

import (
	"d2api/config"
	"d2api/pkg/handlers"
	"d2api/pkg/services"
)

type Wires struct {
	MatchService services.MatchService
}

var Instance *Wires

func Init(config *config.Config) {
	handlers, err := handlers.LoadHandlers(config.InventoryPath)
	if err != nil {
		panic(err)
	}

	Instance = &Wires{
		MatchService: services.NewMatchService(handlers, config),
	}
}
