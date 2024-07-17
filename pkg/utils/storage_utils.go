package utils

import (
	"context"
	"d2api/pkg/handlers"
	m "d2api/pkg/models"
	"d2api/pkg/redis"
	"encoding/json"
	"errors"
	"log"
)

func getAndSetMatchToRedis(matchIdx string, handler *handlers.Handler, teamA, teamB []uint64) bool {
	matchDetails, err := GetMatchRedis(matchIdx)
	if err != nil {
		log.Println("Failed to get match:", err)
		return false
	}
	matchDetails.Status = "scheduled"
	matchDetails.Handler = handler.Username
	matchDetails.TeamA = teamA
	matchDetails.TeamB = teamB

	if err = SetMatchRedis(matchIdx, *matchDetails); err != nil {
		log.Println("Failed to set match:", err)
		return false
	}

	return true
}

func GetMatchRedis(matchIdx string) (*m.MatchDetails, error) {
	marshalled, err := redis.RedisClient.Get(context.Background(), matchIdx).Result()
	if err != nil {
		return nil, err
	}

	var match m.MatchDetails
	err = json.Unmarshal([]byte(marshalled), &match)
	if err != nil {
		return nil, errors.New("match not found")
	}

	return &match, nil
}

func SetMatchRedis(matchIdx string, match m.MatchDetails) error {
	marshalled, err := json.Marshal(match)
	if err != nil {
		return err
	}

	err = redis.RedisClient.Set(context.Background(), matchIdx, marshalled, 0).Err()
	if err != nil {
		return err
	}

	return nil
}

func GetAllMatchIdxs() ([]string, error) {
	keys, err := redis.RedisClient.Keys(context.Background(), "*").Result()
	if err != nil {
		log.Println("Failed to get keys:", err)
		return nil, err
	}

	return keys, nil
}
