package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/paralin/go-dota2"
	"github.com/paralin/go-dota2/cso"
	"github.com/paralin/go-dota2/protocol"
	"github.com/paralin/go-steam"
	"github.com/paralin/go-steam/gsbot"
	"github.com/paralin/go-steam/protocol/steamlang"
	steamId "github.com/paralin/go-steam/steamid"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

type handler struct {
	steamClient *steam.Client
	dotaClient  *dota2.Dota2
}

func main() {
	loadEnv()

	// Initialize the handler with uninitialized clients
	handler := &handler{}

	// Initialize the bot
	details := &gsbot.LogOnDetails{
		Username: os.Getenv("STEAM_USERNAME"),
		Password: os.Getenv("STEAM_PASSWORD"),
	}

	// Start the web server
	r := gin.Default()

	// Health check
	r.GET("/", func(c *gin.Context) {
		
		c.JSON(200, gin.H{
			"message": "Alive it is!",
		})
	})

	// Create Lobby
	r.POST("/lobby", func(c *gin.Context) {

		lobbyVisibility := protocol.DOTALobbyVisibility_DOTALobbyVisibility_Public

		lobbyDetails := &protocol.CMsgPracticeLobbySetDetails{
			GameName:            proto.String("RELATIVE"),
			Visibility: 		 &lobbyVisibility, 
		}
		handler.dotaClient.CreateLobby(lobbyDetails)

		c.JSON(200, gin.H{
			"message": "Lobby has been created",
		})
	})

	// Invite a player to the lobby
	r.POST("/invite/:steamId", func(c *gin.Context) {
		id := c.Param("steamId")
		log.Printf("Inviting player with steamId: %v", id)
		uintId, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		handler.dotaClient.InviteLobbyMember(steamId.SteamId(uintId))

		c.JSON(200, gin.H{
			"message": "Player has been invited",
		})
	})

	// Start the lobby
	r.POST("/start", func(c *gin.Context) {
		handler.dotaClient.LaunchLobby()
		c.JSON(200, gin.H{
			"message": "Lobby has been started",
		})
	})


	// Start the web server
	go func() {
		err := r.Run(":8080")
		if err != nil {
			log.Fatal(err)
		}
	}()
	
	err := steam.InitializeSteamDirectory()
	if err != nil {
		panic(err)
	}

	bot := gsbot.Default()
	handler.steamClient = bot.Client
	auth := gsbot.NewAuth(bot, details, "sentry.bin")
	debug, err := gsbot.NewDebug(bot, "debug")
	if err != nil {
		panic(err)
	}
	handler.steamClient.RegisterPacketHandler(debug)
	serverList := gsbot.NewServerList(bot, "serverlist.json")
	serverList.Connect()

	for event := range handler.steamClient.Events() {
		auth.HandleEvent(event)
		debug.HandleEvent(event)
		serverList.HandleEvent(event)

		switch e := event.(type) {
		case error:
			fmt.Printf("Error: %v", e)
		case *steam.LoggedOnEvent:
			handler.steamClient.Social.SetPersonaState(steamlang.EPersonaState_Online)
		
		case *steam.LoggedOffEvent:
			fmt.Printf("Logged off: %v", e.Result)
			handler.steamClient.Disconnect()
			
		case *steam.PersonaStateEvent:
			fmt.Printf("Successfully logged on as %s\n", e.Name) // Here it is connected to steam client
			
			println("Connecting to dota2")

			handler.dotaClient = dota2.New(handler.steamClient, logrus.New())
			handler.dotaClient.SetPlaying(true)

			// SOCACHE MECHANISM
			eventCh, eventCancel, err := handler.dotaClient.GetCache().SubscribeType(cso.Lobby) // Listen to lobby cache
			if err != nil {
				log.Fatalf("Failed to subscribe to lobby cache: %v", err)
			}

			defer eventCancel()
			
			lobbyEvent := <- eventCh
			lobby := lobbyEvent.Object.String()
			log.Printf("Lobby: %v", lobby)
		}
	}
}

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}
