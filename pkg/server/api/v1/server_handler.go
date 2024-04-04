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
		v1.POST("/match", postMatch)
		v1.GET("/match/:matchIdx", getMatch)
		v1.GET("/player/:steamId/history", getPlayerHistory)
	}
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
	steamIdUint, err := strconv.ParseUint(steamId, 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "need to send a valid steamId"})
	}

	history, err := wires.Instance.MatchService.GetPlayerHistory(uint32(steamIdUint))
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, history)
}
