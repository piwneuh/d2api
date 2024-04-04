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
		v1.GET("/player/:steamId/matches", getPlayerHistory)
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
	limit := c.Query("limit")
	limitInt := 10
	if limit != "" {
		var err error
		limitInt, err = strconv.Atoi(limit)
		if err != nil {
			c.JSON(400, gin.H{"error": "need to send a valid limit"})
		}
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
