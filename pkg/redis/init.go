package redis

import (
	"context"
	"d2api/config"
	"log"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func Init(config *config.Config, ctx context.Context) {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     config.Redis.Host + ":" + config.Redis.Port,
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
	})

	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %s", err)
	}
}
