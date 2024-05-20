package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"

	"github.com/paralin/go-dota2"
	"github.com/paralin/go-dota2/cso"
	"github.com/paralin/go-steam"
	"github.com/paralin/go-steam/protocol/steamlang"
	"github.com/sirupsen/logrus"
)

type SteamConfig struct {
	Username string
	Password string
}
type Handler struct {
	SteamClient *steam.Client
	DotaClient  *dota2.Dota2
	SteamConfig
	Occupied bool
}

func NewHandler(steamConfig SteamConfig) *Handler {
	return &Handler{
		SteamConfig: steamConfig,
		Occupied:    false,
	}
}

func LoadHandlers(inventoryPath string) ([]*Handler, error) {
	text, err := os.ReadFile("./" + inventoryPath)
	if err != nil {
		log.Fatalln("could not read inventory file")
		return nil, err
	}

	var steamConfigs []SteamConfig
	err = json.Unmarshal(text, &steamConfigs)
	if err != nil {
		log.Fatalln("could not unmarshal inventory file")
		return nil, err
	}

	handlers := make([]*Handler, 0)
	for _, steamConfig := range steamConfigs {
		handlers = append(handlers, NewHandler(steamConfig))
	}

	return handlers, nil
}

func (h *Handler) InitSteamConnection() {
	// Grab the steam credentials from the environment
	details := &steam.LogOnDetails{
		Username: h.Username,
		Password: h.Password,
	}

	// Grab actual server list
	err := steam.InitializeSteamDirectory()
	if err != nil {
		panic(err)
	}

	// Initialize the steam client
	h.SteamClient = steam.NewClient()
	h.SteamClient.Connect()

	// Listen to events happening on steam client
	for event := range h.SteamClient.Events() {
		switch e := event.(type) {
		case *steam.ConnectedEvent:
			log.Println("Connected to steam network, trying to log in...")
			h.SteamClient.Auth.LogOn(details)
		case *steam.LoggedOnEvent:
			log.Println("Successfully logged on to steam")
			// Set account state to online
			h.SteamClient.Social.SetPersonaState(steamlang.EPersonaState_Online)

			// Once logged in, we can initialize the dota2 client
			h.DotaClient = dota2.New(h.SteamClient, logrus.New())
			h.DotaClient.SetPlaying(true)

			time.Sleep(1 * time.Second)

			// Try to get a session
			h.DotaClient.SayHello()

			eventCh, _, err := h.DotaClient.GetCache().SubscribeType(cso.Lobby) // Listen to lobby cache
			if err != nil {
				log.Println("Failed to subscribe to lobby cache:", err)
			}

			lobbyEvent := <-eventCh
			lobby := lobbyEvent.Object.String()
			log.Printf("Lobby: %v", lobby)

		case steam.FatalErrorEvent:
			log.Println("Fatal error occurred: ", e.Error())
		}
	}
}

func GetFreeHandler(handlers []*Handler) (*Handler, uint16, error) {
	for i, handler := range handlers {
		if !handler.Occupied {
			handler.Occupied = true
			if handler.SteamClient == nil || handler.DotaClient == nil {
				go func() {
					handler.InitSteamConnection()
				}()
				time.Sleep(2 * time.Second)
			}
			return handler, uint16(i), nil
		}
	}

	return nil, 0, errors.New("no available bot")
}

func GetFirstHandler(handlers []*Handler) (*Handler, uint16, error) {
	if len(handlers) == 0 {
		return nil, 0, errors.New("no available bot")
	}

	handler := handlers[0]
	if handler.SteamClient == nil || handler.DotaClient == nil {
		go func() {
			handler.InitSteamConnection()
		}()
		time.Sleep(3 * time.Second)
	}
	return handler, 0, nil
}
