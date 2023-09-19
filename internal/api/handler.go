package api

import (
	"github.com/gin-gonic/gin"
	"github.com/paralin/go-dota2"
	"github.com/paralin/go-steam"
	"github.com/paralin/go-steam/steamid"
	"github.com/piwneuh/d2api/internal/api/dtos"
	"net/http"
)

type Handler struct {
	steamClient *steam.Client
	dotaClient  *dota2.Dota2
}

func New(steamClient *steam.Client, dotaClient *dota2.Dota2) (*Handler, error) {
	handler := &Handler{
		steamClient: steamClient,
		dotaClient:  dotaClient,
	}

	handler.init()
	return handler, nil
}

// This function's name is a must. App Engine uses it to drive the requests properly.
func (h *Handler) init() {
	// Starts a new Gin instance with no middle-ware
	r := gin.New()

	// Define your handlers
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World!")
	})
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	r.POST("/lobby/invite", func(c *gin.Context) {
		req := dtos.InviteLobbyReq{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.String(http.StatusBadRequest, "Invalid request body")
			return
		}

		h.InviteToLobby(steamid.SteamId(req.SteamId))
		c.String(http.StatusOK, "Invite user with steamId %s successfully", req.SteamId)
	})

	println("Starting server")
	err := r.Run()
	if err != nil {
		println("Failed to start server")
		return
	}
}

func (h *Handler) InviteToLobby(steamId steamid.SteamId) {
	println("Inviting to lobby")
	h.dotaClient.InviteLobbyMember(steamId)
}
