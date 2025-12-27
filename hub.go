package main

import (
	"context"
	"log"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	Hub  *Hub
	Send chan []byte
	conn *websocket.Conn
}

func (c *Client) DisplayRefresh() {
	defer func() {
		c.conn.Close()
	}()
	for {
		msg, ok := <-c.Send
		if !ok {
			return
		}

		err := c.conn.WriteMessage(websocket.BinaryMessage, msg)
		if err != nil {
			return
		}
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.conn.Close()
	}()
	for {
		_, p, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		if len(p) == 4 {
			redisClient := c.Hub.redisClient
			err := redisClient.Set(context.Background(), "canvas", p, 0).Err()
			if err != nil {
				log.Println("Redis set error:", err)
			}
			c.Hub.Broadcast <- p
		}

	}
}

type Hub struct {
	Clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	redisClient *redis.Client
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true
		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}
		case msg := <-h.Broadcast:
			for client := range h.Clients {
				select {
				case client.Send <- msg:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
		}
	}
}
