package main

import (
	"log"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/paralin/go-dota2"
	"github.com/paralin/go-dota2/cso"
	"github.com/paralin/go-dota2/protocol"
	"github.com/paralin/go-steam"
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

	// Initialize the handler with empty clients
	handler := handler{}

	initGinServer(&handler)

	initSteamConnection(&handler)

}

func initSteamConnection(handler *handler) {
	
	// Grab the steam credentials from the environment
	details := &steam.LogOnDetails{
		Username: os.Getenv("STEAM_USERNAME"),
		Password: os.Getenv("STEAM_PASSWORD"),
	}
	
	// Grab actual server list
	err := steam.InitializeSteamDirectory()
	if err != nil {
		panic(err)
	}

	// Initialize the steam client
	handler.steamClient = steam.NewClient()
	handler.steamClient.Connect()

	// Listen to events happening on steam client
	for event := range handler.steamClient.Events() {
		switch e := event.(type) {
		case *steam.ConnectedEvent:
			log.Println("Connected to steam network, trying to log in...")
			handler.steamClient.Auth.LogOn(details)
		case *steam.LoggedOnEvent:
			log.Println("Successfully logged on to steam")		
			// Set account state to online
			handler.steamClient.Social.SetPersonaState(steamlang.EPersonaState_Online)

			// Once logged in, we can initialize the dota2 client
			handler.dotaClient = dota2.New(handler.steamClient, logrus.New())
			handler.dotaClient.SetPlaying(true)

			// Try to get a session
			handler.dotaClient.SayHello()

			eventCh, _, err := handler.dotaClient.GetCache().SubscribeType(cso.Lobby) // Listen to lobby cache
			if err != nil {
				log.Fatalf("Failed to subscribe to lobby cache: %v", err)
			}
	
			lobbyEvent := <- eventCh
			lobby := lobbyEvent.Object.String()
			log.Printf("Lobby: %v", lobby)

		case steam.FatalErrorEvent:
			log.Println("Fatal error occurred: ", e.Error()) 
		}
	}
}


func initGinServer(handler *handler) {
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
			PassKey: 		   	 proto.String("test"),
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
}

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}
