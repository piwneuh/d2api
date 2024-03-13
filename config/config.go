package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Redis         RedisConfig
	Server        ServerConfig
	InventoryPath string
	TimeToCancel  uint32
}

type ServerConfig struct {
	Port string
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

	timeToCancel, err := strconv.ParseUint(readEnvVar("TIME_TO_CANCEL"), 10, 32)
	if err != nil {
		timeToCancel = 300
	}

	return &Config{
		Redis: RedisConfig{
			Host:     readEnvVar("REDIS_HOST"),
			Port:     readEnvVar("REDIS_PORT"),
			Password: readEnvVar("REDIS_PASSWORD"),
			DB:       db,
		},
		Server: ServerConfig{
			Port: readEnvVar("SERVER_PORT"),
		},
		InventoryPath: readEnvVar("INVENTORY_PATH"),
		TimeToCancel:  uint32(timeToCancel),
	}
}

func readEnvVar(name string) string {
	godotenv.Load(".env")
	return os.Getenv(name)
}
