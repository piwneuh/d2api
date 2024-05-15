package v1

import (
	"context"
	"d2api/pkg/requests"
	"d2api/pkg/response"
	"d2api/pkg/wires"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RegisterServer(router *gin.Engine, ctx context.Context) {
	v1 := router.Group("/v1")
	{
		v1.POST("/match", postMatch)
		v1.POST("/schedule-tournament", postScheduleTournament)
		v1.GET("/match/:matchIdx", getMatch)
		v1.GET("/player/:steamId/matches", getPlayerHistory)
		v1.GET("/player/od/:steamId/matches", getPlayerHistoryOD)
	}
}

func postScheduleTournament(c *gin.Context) {
	var req []requests.MatchForMiddleware
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	err := wires.Instance.TournamentService.ScheduleRound(req)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, make([]response.TournamentEndRequest, 0))
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
