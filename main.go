package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
)

func main() {
	ctx := context.Background()
	redisClient := InitCanvasRedis(ctx)

	defer redisClient.Close()

	r := chi.NewRouter()
	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(time.Second * 60))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {

		time_started := time.Now()
		defer func() {
			log.Printf("Request processed in %s", time.Since(time_started))
		} ()

		val, err := redisClient.Get(r.Context(), "canvas").Bytes()
		if err == redis.Nil {
			http.Error(w, "Canvas not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(val)
	})



	log.Fatal(http.ListenAndServe(":8080", r))
}
