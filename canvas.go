package main

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)


func InitCanvasRedis(ctx context.Context) *redis.Client {
	REDIS_ADDR := os.Getenv("REDIS_ADDR")

	redisClient := redis.NewClient(&redis.Options{
		Addr:     REDIS_ADDR,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	u_int4, err := redisClient.BitField(ctx, "canvas", "set", "u4", "#999999", "5").Result()
	if err != nil {
		log.Fatalf("Failed to set bitfield: %v", err)
	}
	log.Printf("BitField set result: %v", u_int4)
	return redisClient
}
