package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"
	// "github.com/redis/go-redis/v9"
)

func (h *Hub) handleWs(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize: 1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {return true},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	client := &Client{
		Hub: h,
		Send: make(chan []byte),
		conn: conn,
	}

	h.Register <- client

	go client.DisplayRefresh()
	client.ReadPump()
}

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

	hub := &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}

	go hub.Run()
	r.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		hub.handleWs(w, r)
	})

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		val := r.URL.Query().Get("q")
		if val != "" {
			hub.Broadcast <- []byte(val)
		}
	})

	log.Fatal(http.ListenAndServe(":8080", r))
}
