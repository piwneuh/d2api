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
	"github.com/paralin/go-dota2/protocol"
	"google.golang.org/protobuf/proto"
	"os"
	"time"

	"github.com/paralin/go-dota2"
	"github.com/paralin/go-steam"
	"github.com/paralin/go-steam/gsbot"
	"github.com/paralin/go-steam/protocol/steamlang"
	"github.com/sirupsen/logrus"
)

func main() {
	// d, err := ioutil.ReadFile("serverlist.json")
	// if err != nil {
	// 	panic(err)
	// }
	// println(string(d))

	// var addrs []*netutil.PortAddr
	// err = json.Unmarshal(d, &addrs)
	// if err != nil {
	// 	println(err.Error())
	// }
	// raddr := addrs[rand.Intn(len(addrs))]
	// println(raddr.String())

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
			println("Connecting to dota2")

			dotaClient := dota2.New(client, logrus.New())
			dotaClient.SetPlaying(true)

			for i := 0; i < 10; i++ {
				time.Sleep(1 * time.Second)
				dotaClient.SayHello()
			}

			//teamDetails := &protocol.CMsgDOTACreateTeam{
			//	Name: proto.String("Example Team1"),
			//	Tag:  proto.String("TAG1"),
			//}
			//
			//team, err := dotaClient.CreateTeam(context.Background(), teamDetails)
			//if err != nil {
			//	log.Print(err)
			//}
			//log.Printf("Team created: %v", team)

			cLobbyTeamDetails1 := protocol.CLobbyTeamDetails{
				TeamId:       proto.Uint32(9187802),
				TeamName:     proto.String("radiant"),
				TeamTag:      proto.String("radiant"),
				TeamComplete: proto.Bool(true),
			}

			cLobbyTeamDetails2 := protocol.CLobbyTeamDetails{
				TeamId:       proto.Uint32(9187811),
				TeamName:     proto.String("4glory"),
				TeamTag:      proto.String("4glory"),
				TeamComplete: proto.Bool(true),
			}

			lobbyDetails := &protocol.CMsgPracticeLobbySetDetails{
				TeamDetails:         []*protocol.CLobbyTeamDetails{&cLobbyTeamDetails1, &cLobbyTeamDetails2},
				GameName:            proto.String("Relative dota2 lobby"),
				ServerRegion:        proto.Uint32(1), // Replace with the desired server region ID
				GameMode:            proto.Uint32(2), // Replace with the desired game mode ID
				AllowCheats:         proto.Bool(false),
				FillWithBots:        proto.Bool(false),
				AllowSpectating:     proto.Bool(true),
				PassKey:             proto.String("mylobby123"),          // Replace with your desired passkey
				CustomGameMode:      proto.String("my_custom_game_mode"), // Replace with your custom game mode
				CustomMapName:       proto.String("my_custom_map"),       // Replace with your custom map name
				CustomMinPlayers:    proto.Uint32(2),                     // Replace with your desired minimum players
				CustomMaxPlayers:    proto.Uint32(10),                    // Replace with your desired maximum players
				CustomGameCrc:       proto.Uint64(123456789),             // Replace with your CRC value
				CustomGameTimestamp: proto.Uint32(1234567890),            // Replace with your timestamp
				LanHostPingLocation: proto.String("US West"),             // Replace with your desired LAN host ping location
			}

			dotaClient.CreateLobby(lobbyDetails)

			dotaClient.Close()
			return // exit
		}
	}

}
