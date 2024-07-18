package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"os"
	"sync"
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
	Broken   bool
}

type Handlers struct {
	Handlers []*Handler
	Mutex    *sync.Mutex
}

type ListHander struct {
	Id       int    `json:"id"`
	SignedIn bool   `json:"signed_in"`
	Username string `json:"username"`
	Occupied bool   `json:"occupied"`
	Broken   bool   `json:"broken"`
}

var Hs *Handlers

func LoadHandlers(inventoryPath string) {
	handlers, err := loadHandlers(inventoryPath)
	if err != nil {
		panic(err)
	}
	Hs = &Handlers{
		Mutex:    &sync.Mutex{},
		Handlers: handlers,
	}
}

func (hs *Handlers) GetMatchHandler(username string) (*Handler, error) {
	hs.Mutex.Lock()
	defer hs.Mutex.Unlock()
	for _, handler := range hs.Handlers {
		if handler.Username == username {
			if handler.SteamClient == nil || handler.DotaClient == nil {
				go func() {
					handler.InitSteamConnection()
				}()
				time.Sleep(3 * time.Second)
			}
			return handler, nil
		}
	}
	return nil, errors.New("no bot with that username")
}

func (hs *Handlers) GetFreeHandler() (*Handler, error) {
	hs.Mutex.Lock()
	defer hs.Mutex.Unlock()
	checkedIds := make(map[int]bool, 0)
	for i := 0; i < len(hs.Handlers); i++ {
		var num int
		for {
			num = rand.Intn(len(hs.Handlers))
			if checked, ok := checkedIds[num]; !ok || !checked {
				break
			}
		}

		checkedIds[num] = true
		handler := hs.Handlers[num]
		if !handler.Occupied && !handler.Broken {
			handler.Occupied = true
			if handler.SteamClient == nil || handler.DotaClient == nil {
				go func() {
					handler.InitSteamConnection()
				}()
				time.Sleep(2 * time.Second)
			}
			return handler, nil
		}
	}

	return nil, errors.New("no available bot")
}

func (hs *Handlers) GetFirstHandler() (*Handler, error) {
	if len(hs.Handlers) == 0 {
		return nil, errors.New("no available bot")
	}

	handler := hs.Handlers[rand.Intn(len(hs.Handlers))]
	if handler.SteamClient == nil || handler.DotaClient == nil {
		hs.Mutex.Lock()
		defer hs.Mutex.Unlock()
		go func() {
			handler.InitSteamConnection()
		}()
		time.Sleep(3 * time.Second)
	}
	return handler, nil
}

func NewHandler(steamConfig SteamConfig) *Handler {
	return &Handler{
		SteamConfig: steamConfig,
		Occupied:    false,
		Broken:      false,
	}
}

func loadHandlers(inventoryPath string) ([]*Handler, error) {
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

func (hs *Handlers) GetAllBots() []*ListHander {
	hs.Mutex.Lock()
	defer hs.Mutex.Unlock()
	bots := make([]*ListHander, 0)
	for i, handler := range hs.Handlers {
		bots = append(bots, &ListHander{
			Id:       i,
			SignedIn: handler.SteamClient != nil && handler.DotaClient != nil,
			Username: handler.Username,
			Occupied: handler.Occupied,
			Broken:   handler.Broken,
		})
	}
	return bots
}

func (hs *Handlers) LeaveLobby(username string) error {
	hs.Mutex.Lock()
	defer hs.Mutex.Unlock()
	for _, handler := range hs.Handlers {
		if handler.Username == username {
			handler.DotaClient.DestroyLobby(context.Background())
			time.Sleep(1 * time.Second)

			handler.Occupied = false
			handler.Broken = false
			return nil
		}
	}

	return errors.New("no bot with that username")
}
