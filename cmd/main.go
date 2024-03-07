package main

import (
	"d2api/config"
	"d2api/pkg/server"
)

func main() {
	config := config.NewConfig()
	server := server.NewServer(config)
	server.Start()
}
