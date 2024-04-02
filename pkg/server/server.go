package server

import (
	"context"
	"d2api/config"
	"d2api/pkg/crawler"
	"d2api/pkg/redis"
	"d2api/pkg/server/api"
	"d2api/pkg/wires"
	"log"

	"github.com/gin-gonic/gin"
)

type Server struct {
	config *config.Config
}

func NewServer(config *config.Config) *Server {
	return &Server{
		config: config,
	}
}

func (s *Server) Start() {

	r := gin.Default()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(CORSMiddleware())

	wires.Init(s.config)
	redis.Init(s.config, context.Background())
	crawler.Init(s.config)
	api.RegisterVersion(r, context.Background())

	err := r.Run(":" + s.config.Server.Port)
	if err != nil {
		log.Fatal("Could not start the server" + err.Error())
		return
	}

	log.Println("Server started on port " + s.config.Server.Port)
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length")
		c.Writer.Header().Set("Access-Allow-Methods", "POST, GET")

		c.Next()
	}
}
