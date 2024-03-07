package handlers

import (
	"github.com/paralin/go-dota2"
	"github.com/paralin/go-steam"
)

type Handler struct {
	SteamClient *steam.Client
	DotaClient  *dota2.Dota2
}
