// A simple example that uses the modules from the gsbot package and go-steam to log on
// to the Steam network.
//
// The command expects log on data, optionally with an auth code:
//
//     gsbot [username] [password]
//     gsbot [username] [password] [authcode]
// Valid servers: https://api.steampowered.com/ISteamDirectory/GetCMList/v1/?cellId=0

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/paralin/go-dota2"
	"github.com/paralin/go-dota2/cso"
	"github.com/paralin/go-steam"
	"github.com/paralin/go-steam/gsbot"
	"github.com/paralin/go-steam/protocol/steamlang"
	"github.com/paralin/go-steam/steamid"
	"github.com/sirupsen/logrus"
)

func main() {
	
	if len(os.Args) < 3 {
		fmt.Println("gsbot example\nusage: \n\tgsbot [username] [password] [authcode]")
		return
	}

	// optional auth code
	authcode := ""
	if len(os.Args) > 3 {
		authcode = os.Args[3]
	}

	err := steam.InitializeSteamDirectory()
	if err != nil {
		panic(err)
	}

	details := &gsbot.LogOnDetails{
		Username: os.Args[1],
		Password: os.Args[2],
		AuthCode: authcode,
	}

	bot := gsbot.Default()
	client := bot.Client
	auth := gsbot.NewAuth(bot, details, "sentry.bin")
	debug, err := gsbot.NewDebug(bot, "debug")
	if err != nil {
		panic(err)
	}
	client.RegisterPacketHandler(debug)
	serverList := gsbot.NewServerList(bot, "serverlist.json")
	serverList.Connect()

	for event := range client.Events() {
		auth.HandleEvent(event)
		debug.HandleEvent(event)
		serverList.HandleEvent(event)

		switch e := event.(type) {
		case error:
			fmt.Printf("Error: %v", e)
		case *steam.LoggedOnEvent:
			client.Social.SetPersonaState(steamlang.EPersonaState_Online)

		case *steam.PersonaStateEvent:
			fmt.Printf("Successfully logged on as %s\n", e.Name) // Here it is connected to steam client
			Connect2Dota(client)
		}
	}
}

func Connect2Dota(client *steam.Client) {

	println("Connecting to dota2")

	dotaClient := dota2.New(client, logrus.New())
	dotaClient.SetPlaying(true)

	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)
		dotaClient.SayHello()
	}

	// SOCACHE MECHANISM
	eventCh, eventCancel, err:= dotaClient.GetCache().SubscribeType(cso.Lobby) // Listen to lobby cache
	if err != nil {
		log.Fatalf("Failed to subscribe to lobby cache: %v", err)
	}

	dotaClient.InviteLobbyMember(76561198153440660) 
	defer eventCancel()

	lobbyEvent := <-eventCh
	lobby := lobbyEvent.Object.String()
	log.Printf("Lobby: %v", lobby)

	time.Sleep(30 * time.Second)

	dotaClient.LaunchLobby()
	println("Launched lobby")

	time.Sleep(30 * time.Second)
}

func InviteToLobby(dotaClient *dota2.Dota2, steamId steamid.SteamId) {
	println("Inviting to lobby")
	dotaClient.InviteLobbyMember(steamId)
}

func LaunchGame(dotaClient *dota2.Dota2) {
	println("Launching lobby")
	dotaClient.LaunchLobby()
}
