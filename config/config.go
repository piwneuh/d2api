package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Steam  SteamConfig
	Redis  RedisConfig
	Server ServerConfig
}

type ServerConfig struct {
	Port string
}

type SteamConfig struct {
	Username string
	Password string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func NewConfig() *Config {
	db, err := strconv.Atoi(readEnvVar("REDIS_DB"))
	if err != nil {
		db = 0
	}

	return &Config{
		Steam: SteamConfig{
			Username: readEnvVar("STEAM_USERNAME"),
			Password: readEnvVar("STEAM_PASSWORD"),
		},
		Redis: RedisConfig{
			Host:     readEnvVar("REDIS_HOST"),
			Port:     readEnvVar("REDIS_PORT"),
			Password: readEnvVar("REDIS_PASSWORD"),
			DB:       db,
		},
		Server: ServerConfig{
			Port: readEnvVar("SERVER_PORT"),
		},
	}
}

func readEnvVar(name string) string {
	godotenv.Load(".env")
	return os.Getenv(name)
}
