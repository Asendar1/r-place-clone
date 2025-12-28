package main

import (
	"context"
	"os"

	"github.com/redis/go-redis/v9"
)


func InitCanvasRedis(ctx context.Context) *redis.Client {
	REDIS_ADDR := os.Getenv("REDIS_ADDR")
	if REDIS_ADDR == "" {
		REDIS_ADDR = "localhost:6379"
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     REDIS_ADDR,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	if err := redisClient.Ping(ctx).Err(); err != nil {
		panic(err)
	}

	exists, _ := redisClient.Exists(ctx, "canvas").Result()
	if exists == 0 {
		redisClient.Do(ctx, "BITFIELD", "canvas", "SET", "u4", "#999999", "0")
	}
	return redisClient
}
