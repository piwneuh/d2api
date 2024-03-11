package wires

import (
	"d2api/pkg/handlers"
	"d2api/pkg/services"
)

type Wires struct {
	MatchService services.MatchService
}

var Instance *Wires

func Init(inventoryPath string) {
	handlers, err := handlers.LoadHandlers(inventoryPath)
	if err != nil {
		panic(err)
	}

	Instance = &Wires{
		MatchService: services.NewMatchService(handlers),
	}
}
