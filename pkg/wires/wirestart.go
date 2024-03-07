package wires

import (
	"d2api/pkg/handlers"
	"d2api/pkg/services"
)

type Wires struct {
	MatchService services.MatchService
}

func Init() *Wires {
	return &Wires{
		MatchService: services.NewMatchService(&handlers.Handler{}),
	}
}

var Instance = Init()
