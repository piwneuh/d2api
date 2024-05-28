package utils

import (
	"d2api/config"
	h "d2api/pkg/handlers"
	"d2api/pkg/requests"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/paralin/go-dota2/cso"
	"github.com/paralin/go-dota2/protocol"
)

func GetMissingPlayers(lobby *protocol.CSODOTALobby, req *requests.CreateMatchReq) ([]uint64, []uint64, error) {
	goodGuys, badGuys, err := GetGoodAndBadGuys(lobby)
	if err != nil {
		log.Println("Failed to get good and bad guys:", err)
		return nil, nil, err
	}

	var missingTeamA, missingTeamB []uint64
	for _, id := range req.TeamA {
		if !slices.Contains(goodGuys, id) {
			missingTeamA = append(missingTeamA, id)
		}
	}

	for _, id := range req.TeamB {
		if !slices.Contains(badGuys, id) {
			missingTeamB = append(missingTeamB, id)
		}
	}

	return missingTeamA, missingTeamB, nil
}

func GetGoodAndBadGuys(lobby *protocol.CSODOTALobby) ([]uint64, []uint64, error) {
	goodGuys := make([]uint64, 0)
	badGuys := make([]uint64, 0)

	for _, member := range lobby.AllMembers {
		if member.Team == nil {
			continue
		} else if *member.Team == protocol.DOTA_GC_TEAM_DOTA_GC_TEAM_GOOD_GUYS {
			goodGuys = append(goodGuys, *member.Id)
		} else if *member.Team == protocol.DOTA_GC_TEAM_DOTA_GC_TEAM_BAD_GUYS {
			badGuys = append(badGuys, *member.Id)
		}
	}

	return goodGuys, badGuys, nil
}

func GetGameModeFromString(gameMode string) uint32 {
	fullString := "DOTA_GAMEMODE_" + strings.ToUpper(gameMode)
	gameModeInt, ok := protocol.DOTA_GameMode_value[fullString]
	if !ok {
		return uint32(protocol.DOTA_GameMode_value["DOTA_GAMEMODE_AP"])
	}

	return uint32(gameModeInt)
}

func GetCurrentLobby(handler *h.Handler) (*protocol.CSODOTALobby, error) {
	lobby, err := handler.DotaClient.GetCache().GetContainerForTypeID(cso.Lobby)
	if err != nil {
		log.Println("Failed to get lobby:", err)
		return nil, err
	}

	lobbyMessage := lobby.GetOne()
	if lobbyMessage == nil {
		return nil, errors.New("no lobby found")
	}

	return lobbyMessage.(*protocol.CSODOTALobby), nil
}

func getSteamUsernamesFromSteamIds(steamIds []uint64) map[uint64]string {
	steamApi := &config.GlobalConfig.SteamWebApi
	var steamIdsStr string
	for _, id := range steamIds {
		steamIdsStr = steamIdsStr + strconv.FormatUint(id, 10) + ","
	}
	steamIdsStr = strings.TrimSuffix(steamIdsStr, ",")

	var steamUsernamesResponse struct {
		Response struct {
			Players []struct {
				PersonaName string `json:"personaname"`
				SteamId     string `json:"steamid"`
			}
		} `json:"response"`
	}

	var steamUsernames = make(map[uint64]string)

	resp, err := http.Get(steamApi.URL + "key=" + steamApi.Key + "&steamids=" + steamIdsStr)
	if err != nil {
		log.Println("Failed to get steam usernames:", err)
		return steamUsernames
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println("Failed to get steam usernames: status code", resp.StatusCode)
		return steamUsernames
	}

	err = json.NewDecoder(resp.Body).Decode(&steamUsernamesResponse)
	if err != nil {
		log.Println("Failed to get steam usernames:", err)
		return steamUsernames
	}

	for _, player := range steamUsernamesResponse.Response.Players {
		steamId, err := strconv.ParseUint(player.SteamId, 10, 64)
		if err != nil {
			log.Println("Failed to parse steam id:", err)
			continue
		}

		steamUsernames[steamId] = player.PersonaName
	}

	return steamUsernames
}
