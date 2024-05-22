package v1

import (
	"context"
	"d2api/pkg/requests"
	"d2api/pkg/wires"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RegisterServer(router *gin.Engine, ctx context.Context) {
	v1 := router.Group("/v1")
	{
		match := v1.Group("/match")
		{
			match.POST("", postMatch)
			match.POST("/reinvite-players", postReinvitePlayers)
			match.GET("/:matchIdx", getMatch)
			match.GET("/info/:matchIdx", getMatchInfo)
		}

		v1.POST("/schedule-tournament", postScheduleTournament)

		player := v1.Group("/player")
		{
			player.GET("/:steamId/matches", getPlayerHistory)
			player.GET("/od/:steamId/matches", getPlayerHistoryOD)
		}
	}
}

func postReinvitePlayers(c *gin.Context) {
	var req requests.ReinvitePlayersReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	err := wires.Instance.MatchService.ReinvitePlayers(req)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "players reinvited"})
}

func postScheduleTournament(c *gin.Context) {
	var req requests.ScheduleMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	res, err := wires.Instance.TournamentService.ScheduleRound(req.Matches)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, res)
}

func postMatch(c *gin.Context) {
	var req requests.CreateMatchReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	matchIdx, err := wires.Instance.MatchService.ScheduleMatch(req)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"matchIdx": matchIdx})
}

func getMatch(c *gin.Context) {
	matchIdx := c.Param("matchIdx")
	match, err := wires.Instance.MatchService.GetMatch(matchIdx)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, match)
}

func getMatchInfo(c *gin.Context) {
	matchIdx := c.Param("matchIdx")
	match, err := wires.Instance.MatchService.GetMatchInfo(matchIdx)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, match)
}

func getPlayerHistory(c *gin.Context) {
	steamId := c.Param("steamId")
	limit := c.DefaultQuery("limit", "10")
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		c.JSON(400, gin.H{"error": "need to send a valid limit"})
	}

	steamIdUint, err := strconv.ParseInt(steamId, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "need to send a valid steamId"})
	}

	history, err := wires.Instance.MatchService.GetPlayerHistory(steamIdUint, limitInt)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, history)
}

func getPlayerHistoryOD(c *gin.Context) {
	steamId := c.Param("steamId")
	limit := c.DefaultQuery("limit", "10")
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		c.JSON(400, gin.H{"error": "need to send a valid limit"})
	}

	steamIdUint, err := strconv.ParseInt(steamId, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "need to send a valid steamId"})
	}

	history, err := wires.Instance.MatchService.GetPlayerHistoryOpenDota(steamIdUint, limitInt)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, history)
}
