package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

var globalUUID int64 = 0
var GlobalClientCount int64 = 0

func (h *Hub) handleWs(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize: 4096,
		WriteBufferSize: 4096,
		CheckOrigin: func(r *http.Request) bool {return true},
		HandshakeTimeout: 10 * time.Second,
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	atomic.AddInt64(&GlobalClientCount, 1)
	uuid := atomic.AddInt64(&globalUUID, 1)
	client := &Client{
		UUID: uuid,
		Hub: h,
		Send: make(chan []byte, 512),
		conn: conn,
		lastActive: time.Now(),
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
		Broadcast:  make(chan []byte, 10000000),
		Register:   make(chan *Client, 100000),
		Unregister: make(chan *Client, 100000),
		redisClient: redisClient,
		redisQueue: make(chan PixelUpdate, 100000),
	}
	for i := 0; i < ShardCount; i++ {
		hub.shards[i] = &Shard{Clients: make(map[*Client]bool)}
	}
	for i := 0; i < 100; i++ {
		go hub.redisWorker()
	}

	go hub.Run()
	go SaveCanvas(redisClient)

	r.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		hub.handleWs(w, r)
	})

	filesDir := http.Dir("./static/")
	FileServer(r, "/play", filesDir)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/play/", http.StatusSeeOther)
	})
	r.Get("/canvas", CanvasHandler(redisClient))

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout: 15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		log.Println("Server Starting on 8080")
		if err := srv.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Fatalf("ListenAndServe(): %v", err)
			}
		}
	} ()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), time.Second * 30)
	defer cancel()

	if err := srv.Shutdown(ctxShutDown); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")

}
