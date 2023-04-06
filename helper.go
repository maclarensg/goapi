package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
)

var redisClient *redis.Client

func initRedis(host string, port int, password string, db int) error {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       db,
	})

	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		return err
	}

	return nil
}

func logQuery(q Query) {
	queryJSON, err := json.Marshal(q)
	if err != nil {
		fmt.Println("Error marshaling query:", err)
		return
	}

	err = redisClient.LPush(ctx, RedisKey, queryJSON).Err()

	if err != nil {
		log.Printf("Failed to log query: %v", err)
		return
	}
}

func isRunningInKubernetes() bool {
	_, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount")
	return err == nil
}
