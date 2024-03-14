package requests

type InviteLobbyReq struct {
	SteamId uint64 `uri:"steamId"`
}

type CreateMatchReq struct {
	TeamA       []uint64    `json:"teamA"`
	TeamB       []uint64    `json:"teamB"`
	LobbyConfig LobbyConfig `json:"lobbyConfig"`
	StartTime   string      `json:"startTime"`
}

type LobbyConfig struct {
	GameName     string `json:"gameName"`
	PassKey      string `json:"passKey"`
	ServerRegion uint32 `json:"serverRegion"`
	GameMode     string `json:"gameMode"`
}
