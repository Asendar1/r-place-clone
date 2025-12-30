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
		//check if a backup file exists
		files, err := os.ReadDir("backups")
		if err == nil && len(files) > 0 {
			fileName := files[len(files) - 1].Name()
			filePath := "backups/" + fileName
			data, err := os.ReadFile(filePath)
			if err == nil {
				redisClient.Do(ctx, "SET", "canvas", data)
			}
			// then clean up old backups
			for _, file := range files {
				if file.Name() != fileName {
					os.Remove("backups/" + file.Name())
				}
			}
		} else {
			redisClient.Do(ctx, "BITFIELD", "canvas", "SET", "u4", "#999999", "0")
		}
	}
	return redisClient
}
