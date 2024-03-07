package server

import (
	"context"
	"d2api/config"
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

	go func() {
		r := gin.Default()
		r.Use(gin.Logger())
		r.Use(gin.Recovery())
		r.Use(CORSMiddleware())

		api.RegisterVersion(r, context.Background())
		redis.Init(s.config, context.Background())

		err := r.Run(":" + s.config.Server.Port)
		if err != nil {
			log.Fatal("Could not start the server" + err.Error())
			return
		}

		log.Println("Server started on port " + s.config.Server.Port)
	}()
	wires.Instance.MatchService.InitSteamConnection(s.config)
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
