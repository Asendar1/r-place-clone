package main

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
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

func CanvasHandler(redisClient *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := redisClient.Get(context.Background(), "canvas").Bytes()
		if err != nil {
			http.Error(w, "Failed to get canvas data", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(data)
	}
}

func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path + "/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
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
		redisClient: redisClient,
	}

	go hub.Run()
	r.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		hub.handleWs(w, r)
	})

	filesDir := http.Dir("./static/")
	FileServer(r, "/play", filesDir)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/play/", http.StatusSeeOther)
	})
	r.Get("/canvas", CanvasHandler(redisClient))

	log.Fatal(http.ListenAndServe(":8080", r))
}
