package api

import (
	"context"
	"github.com/golang/protobuf/proto"
	"github.com/paralin/go-dota2"
	"github.com/paralin/go-dota2/protocol"
	"github.com/paralin/go-steam"
	"github.com/paralin/go-steam/netutil"
	"github.com/paralin/go-steam/protocol/steamlang"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"log"
	"net/http"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) HandleHello(w http.ResponseWriter, r *http.Request) {
	println("Hello")
	myLoginInfo := new(steam.LogOnDetails)
	myLoginInfo.Username = "yungko_"
	myLoginInfo.Password = "iplaywow123"

	steamClient := steam.NewClient()
	steamClient.ConnectTo(netutil.ParsePortAddr("162.254.196.43:27021"))
	for event := range steamClient.Events() {
		switch e := event.(type) {
		case *steam.ConnectedEvent:
			println("Connected to Steam network")
			steamClient.Auth.LogOn(myLoginInfo)
		case *steam.MachineAuthUpdateEvent:
			ioutil.WriteFile("sentry", e.Hash, 0666)
		case *steam.LoggedOnEvent:
			steamClient.Social.SetPersonaState(steamlang.EPersonaState_Online)
		case steam.FatalErrorEvent:
			log.Print(e)
		case error:
			log.Print(e)
		}
		println("Hello2")
	}
	println("Hello1")

	dotaClient := dota2.New(steamClient, logrus.New())

	cMsgDOTACreateTeam := &protocol.CMsgDOTACreateTeam{
		Name: proto.String("4glory"),
		Tag:  proto.String("4g"),
	}
	team, err := dotaClient.CreateTeam(context.Background(), cMsgDOTACreateTeam)
	if err != nil {
		return
	}
	println("team: ", team)

	//defaultCMsgPracticeLobbySetDetails := createLobby()
	//dotaClient.CreateLobby(defaultCMsgPracticeLobbySetDetails)

	println("World")
}

func createLobby() *protocol.CMsgPracticeLobbySetDetails {
	cLobbyTeamDetails1 := protocol.CLobbyTeamDetails{
		TeamId:                     proto.Uint32(0), //mora se pooppuniti
		TeamName:                   proto.String("radiant"),
		TeamTag:                    proto.String("radiant"),
		TeamLogo:                   proto.Uint64(0),
		TeamBaseLogo:               proto.Uint64(0),
		TeamBannerLogo:             proto.Uint64(0),
		TeamComplete:               proto.Bool(false),
		Rank:                       proto.Uint32(0),
		RankChange:                 proto.Int32(0),
		IsHomeTeam:                 proto.Bool(false),
		IsChallengeMatch:           proto.Bool(false),
		ChallengeMatchTokenAccount: proto.Uint64(0),
		TeamLogoUrl:                proto.String("https://steamcdn-a.akamaihd.net/apps/dota2/images/team_logos/5.png"),
		TeamAbbreviation:           proto.String("radiant"),
	}

	cLobbyTeamDetails2 := protocol.CLobbyTeamDetails{
		TeamId:                     proto.Uint32(0), //mora se pooppuniti
		TeamName:                   proto.String("4glory"),
		TeamTag:                    proto.String("4glory"),
		TeamLogo:                   proto.Uint64(0),
		TeamBaseLogo:               proto.Uint64(0),
		TeamBannerLogo:             proto.Uint64(0),
		TeamComplete:               proto.Bool(false),
		Rank:                       proto.Uint32(0),
		RankChange:                 proto.Int32(0),
		IsHomeTeam:                 proto.Bool(false),
		IsChallengeMatch:           proto.Bool(false),
		ChallengeMatchTokenAccount: proto.Uint64(0),
		TeamLogoUrl:                proto.String("https://steamcdn-a.akamaihd.net/apps/dota2/images/team_logos/6.png"),
		TeamAbbreviation:           proto.String("4glory"),
	}

	defaultCMsgPracticeLobbySetDetails := &protocol.CMsgPracticeLobbySetDetails{
		LobbyId:                     proto.Uint64(0), // mora se dodati
		GameName:                    proto.String("nasa igra"),
		TeamDetails:                 []*protocol.CLobbyTeamDetails{&cLobbyTeamDetails1, &cLobbyTeamDetails2},
		ServerRegion:                proto.Uint32(3),
		GameMode:                    proto.Uint32(2),
		CmPick:                      (*protocol.DOTA_CM_PICK)(proto.Int32(0)),
		BotDifficultyRadiant:        (*protocol.DOTABotDifficulty)(proto.Int32(0)),
		AllowCheats:                 proto.Bool(false),
		FillWithBots:                proto.Bool(false),
		IntroMode:                   proto.Bool(false),
		AllowSpectating:             proto.Bool(true),
		PassKey:                     proto.String(""), // dodati posle
		Leagueid:                    proto.Uint32(1),
		PenaltyLevelRadiant:         proto.Uint32(0),
		PenaltyLevelDire:            proto.Uint32(0),
		LoadGameId:                  proto.Uint32(0),
		SeriesType:                  proto.Uint32(0),
		RadiantSeriesWins:           proto.Uint32(0),
		DireSeriesWins:              proto.Uint32(0),
		Allchat:                     proto.Bool(false),
		DotaTvDelay:                 (*protocol.LobbyDotaTVDelay)(proto.Int32(0)),
		Lan:                         proto.Bool(false),
		CustomGameMode:              proto.String("custom"),
		CustomMapName:               proto.String("inferno"),
		CustomDifficulty:            proto.Uint32(0),
		CustomGameId:                proto.Uint64(0),
		CustomMinPlayers:            proto.Uint32(2),
		CustomMaxPlayers:            proto.Uint32(10),
		Visibility:                  (*protocol.DOTALobbyVisibility)(proto.Int32(0)),
		CustomGameCrc:               proto.Uint64(0),
		CustomGameTimestamp:         proto.Uint32(0),
		PreviousMatchOverride:       proto.Uint64(0),
		PauseSetting:                (*protocol.LobbyDotaPauseSetting)(proto.Int32(0)),
		BotDifficultyDire:           (*protocol.DOTABotDifficulty)(proto.Int32(0)),
		BotRadiant:                  proto.Uint64(0),
		BotDire:                     proto.Uint64(0),
		SelectionPriorityRules:      (*protocol.DOTASelectionPriorityRules)(proto.Int32(0)),
		CustomGamePenalties:         proto.Bool(false),
		LanHostPingLocation:         proto.String("Serbia"),
		LeagueNodeId:                proto.Uint32(0),
		RequestedHeroIds:            []uint32{},
		ScenarioSave:                nil,
		AbilityDraftSpecificDetails: nil,
	}
	return defaultCMsgPracticeLobbySetDetails
}
