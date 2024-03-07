package v1

import (
	"context"
	"d2api/pkg/requests"
	"d2api/pkg/wires"

	"github.com/gin-gonic/gin"
)

func RegisterServer(router *gin.Engine, ctx context.Context) {
	v1 := router.Group("/v1")
	{
		v1.POST("/ScheduleMatch", scheduleMatch)
		v1.GET("/GetMatchDetails", getMatchDetails)
	}
}

func scheduleMatch(c *gin.Context) {
	var req requests.CreateMatchReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	matchIdx := wires.Instance.MatchService.ScheduleMatch(c, req)
	c.JSON(200, gin.H{"matchIdx": matchIdx})
}

func getMatchDetails(c *gin.Context) {
	matchIdx := c.Query("matchIdx")
	details, err := wires.Instance.MatchService.GetMatchDetails(c, matchIdx)
	if err != nil {
		c.JSON(400, gin.H{"msg": err.Error()})
		return
	}

	c.JSON(200, gin.H{"details": details})
}
