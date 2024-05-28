package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Redis         RedisConfig
	Server        ServerConfig
	Mongo         MongoConfig
	InventoryPath string
	TimeToCancel  uint32
	Interval      uint32
	Tournament    TournamentConfig
	Stats         StatsConfig
	SteamWebApi   SteamWebApiConfig
}
type SteamWebApiConfig struct {
	Key string
	URL string
}
type TournamentConfig struct {
	URL string
}

type StatsConfig struct {
	URL string
}

type ServerConfig struct {
	Port string
}

type MongoConfig struct {
	URL      string
	Database string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

var GlobalConfig *Config

func NewConfig() *Config {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Configuration error: ", err)
	}
	db, err := strconv.Atoi(readEnvVar("REDIS_DB"))
	if err != nil {
		db = 0
	}

	timeToCancel, err := strconv.ParseUint(readEnvVar("TIME_TO_CANCEL"), 10, 32)
	if err != nil {
		timeToCancel = 300
	}

	interval, err := strconv.ParseUint(readEnvVar("CRAWLER_INTERVAL"), 10, 32)
	if err != nil {
		interval = 60
	}

	GlobalConfig = &Config{
		Redis: RedisConfig{
			Host:     readEnvVar("REDIS_HOST"),
			Port:     readEnvVar("REDIS_PORT"),
			Password: readEnvVar("REDIS_PASSWORD"),
			DB:       db,
		},
		Server: ServerConfig{
			Port: readEnvVar("SERVER_PORT"),
		},
		Mongo: MongoConfig{
			URL:      readEnvVar("MONGO_URL"),
			Database: readEnvVar("MONGO_DATABASE"),
		},
		InventoryPath: readEnvVar("INVENTORY_PATH"),
		TimeToCancel:  uint32(timeToCancel),
		Interval:      uint32(interval),
		Tournament: TournamentConfig{
			URL: readEnvVar("TOURNAMENT_URL"),
		},
		Stats: StatsConfig{
			URL: readEnvVar("STATS_URL"),
		},
		SteamWebApi: SteamWebApiConfig{
			URL: readEnvVar("STEAM_WEB_API_URL"),
			Key: readEnvVar("STEAM_WEB_API_KEY"),
		},
	}

	return GlobalConfig
}

func readEnvVar(name string) string {
	return os.Getenv(name)
}
