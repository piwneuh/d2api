package api

import (
	"github.com/gin-gonic/gin"
	"github.com/piwneuh/d2api/internal/api/dtos"
	"net/http"
)

// This function's name is a must. App Engine uses it to drive the requests properly.
func init() {
	// Starts a new Gin instance with no middle-ware
	r := gin.New()

	// Define your handlers
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World!")
	})
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	r.POST("/lobby/invite", func(c *gin.Context) {
		req := inviteLobbyReq{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.String(http.StatusBadRequest, "Invalid request body")
			return
		}

		c.String(http.StatusOK, "Create user with name %s, age %d successfully", req.Name, req.Age)
	})

	// Handle all requests using net/http
	http.Handle("/", r)
}
