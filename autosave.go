package main

import (
	"context"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

func SaveCanvas(redisClient *redis.Client) {
    ticker := time.NewTicker(time.Minute * 5)
    defer ticker.Stop()

    for {
        <-ticker.C
        ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)

        func (redisClient *redis.Client, ctx context.Context, cancel context.CancelFunc)  {
			defer cancel()
			bytes, err := redisClient.Get(ctx, "canvas").Bytes()
			if err != nil {
				return
			}

			backupDir := "backups"
			if err = os.MkdirAll(backupDir, 0755); err != nil {
				return
			}

			file, err := os.Create(backupDir + "/canvas_backup_" + time.Now().Format("20060102_150405") + ".bin")
			if err != nil {
				return
			}

			defer file.Close()

			_, err = file.Write(bytes)
			if err != nil{
				return
			}
		} (redisClient, ctx, cancel)
    }
}
