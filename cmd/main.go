package main

import (
	"log"
	"os"
	"strconv"
	"time"

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

			time.Sleep(1 * time.Second)

			// Try to get a session
			handler.dotaClient.SayHello()

			eventCh, _, err := handler.dotaClient.GetCache().SubscribeType(cso.Lobby) // Listen to lobby cache
			if err != nil {
				log.Fatalf("Failed to subscribe to lobby cache: %v", err)
			}

			lobbyEvent := <-eventCh
			lobby := lobbyEvent.Object.String()
			log.Printf("Lobby: %v", lobby)

		case steam.FatalErrorEvent:
			log.Println("Fatal error occurred: ", e.Error())
		}
	}
}

func getGoodAndBadGuys(lobby *protocol.CSODOTALobby) ([]protocol.CSODOTALobbyMember, []protocol.CSODOTALobbyMember, error) {
	goodGuys := make([]protocol.CSODOTALobbyMember, 0)
	badGuys := make([]protocol.CSODOTALobbyMember, 0)

	for _, member := range lobby.AllMembers {
		log.Println(*member.Team)
		if *member.Team == protocol.DOTA_GC_TEAM_DOTA_GC_TEAM_GOOD_GUYS {
			goodGuys = append(goodGuys, *member)
		} else if *member.Team == protocol.DOTA_GC_TEAM_DOTA_GC_TEAM_BAD_GUYS {
			badGuys = append(badGuys, *member)
		}
	}

	return goodGuys, badGuys, nil
}

func getCurrentLobby(handler *handler) (*protocol.CSODOTALobby, error) {
	lobby, err := handler.dotaClient.GetCache().GetContainerForTypeID(cso.Lobby)
	if err != nil {
		log.Fatalf("Failed to get lobby: %v", err)
		return nil, err
	}

	return lobby.GetOne().(*protocol.CSODOTALobby), nil
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
			GameName:     proto.String("CirkoBrat"),
			Visibility:   &lobbyVisibility,
			PassKey:      proto.String("1234"),
			ServerRegion: proto.Uint32(3),
		}
		handler.dotaClient.CreateLobby(lobbyDetails)

		c.JSON(200, gin.H{
			"message": "Lobby has been created",
		})
	})

	// Destory the lobby
	r.DELETE("/lobby", func(c *gin.Context) {
		res, err := handler.dotaClient.DestroyLobby(c)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
		} else {
			c.JSON(200, gin.H{
				"message":  "Lobby has been destroyed",
				"response": res,
			})
		}
	})

	r.POST("/move-to-coach", func(c *gin.Context) {
		team := protocol.DOTA_GC_TEAM_DOTA_GC_TEAM_GOOD_GUYS
		handler.dotaClient.SetLobbyCoach(team)
		c.JSON(200, gin.H{"message": "Moved to coach"})
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

	// Get Lobby information
	r.GET("/lobby", func(c *gin.Context) {
		lobby, err := getCurrentLobby(handler)
		if err != nil {
			c.JSON(500, gin.H{"error": "No current lobby found"})
			return
		} else {
			c.JSON(200, gin.H{"lobby": lobby})
		}
	})

	// Get if lobby is ready
	r.POST("/lobby/ready/", func(c *gin.Context) {
		// Get the list of goodGuys and badGuys from the request body
		var request struct {
			GoodGuys []uint64 `json:"goodGuys"`
			BadGuys  []uint64 `json:"badGuys"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		lobby, err := getCurrentLobby(handler)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		goodGuys, badGuys, err := getGoodAndBadGuys(lobby)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		// Check if the goodGuys and badGuys are ready
		for _, id := range request.GoodGuys {
			ready := false
			for _, goodGuy := range goodGuys {
				if *goodGuy.Id == id && goodGuy.CoachTeam == nil {
					ready = true
				}
			}
			if !ready {
				c.JSON(200, gin.H{"ready": false})
				return
			}
		}

		for _, id := range request.BadGuys {
			ready := false
			for _, badGuy := range badGuys {
				if *badGuy.Id == id && badGuy.CoachTeam == nil {
					ready = true
				}
			}
			if !ready {
				c.JSON(200, gin.H{"ready": false})
				return
			}
		}

		c.JSON(200, gin.H{"ready": true})
	})

	r.GET("/match-history/:accountId", func(c *gin.Context) {
		accountId := c.Param("accountId")
		parsedId, err := strconv.ParseUint(accountId, 10, 32)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		uintId := uint32(parsedId)
		log.Println("ACCOUNT ID", uintId)

		var req protocol.CMsgDOTAGetPlayerMatchHistory
		req.AccountId = &uintId
		req.IncludeCustomGames = proto.Bool(true)
		req.IncludeEventGames = proto.Bool(true)
		req.IncludePracticeMatches = proto.Bool(true)

		log.Println("REQRQERQ", &req)

		matchHistory, err := handler.dotaClient.GetPlayerMatchHistory(c, &req)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		} else {
			c.JSON(200, gin.H{"matchHistory": matchHistory})
		}
	})

	r.GET("/match-details/:matchId", func(c *gin.Context) {
		matchId := c.Param("matchId")
		parsedId, err := strconv.ParseUint(matchId, 10, 64)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		details, err := handler.dotaClient.RequestMatchDetails(c, uint64(parsedId))
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		} else {
			c.JSON(200, gin.H{"details": details})
		}
	})

	r.DELETE("/leave-lobby/", func(c *gin.Context) {
		handler.dotaClient.AbandonLobby()
		c.JSON(200, gin.H{"message": "Lobby has been abandoned"})
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
