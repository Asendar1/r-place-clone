package main

import (
	"context"
	"log"
	"strconv"
	"time"

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
			x := int(p[0])<<8 | int(p[1])
			y := (int(p[2])<<8 | int(p[3])) >> 4
			color := int(p[3] & 0x0F)

			if x >= 1000 || y >= 1000 || color > 15 {
				continue
			}
			offset := y * 1000 + x
			go func(off int, col int) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
				defer cancel()

				offsetStr := strconv.Itoa(off)
				err := c.Hub.redisClient.Do(ctx, "BITFIELD", "canvas", "SET", "u4", "#"+offsetStr, col).Err()
				if err != nil {
					log.Printf("Redis Error: %v", err)
				}
			}(offset, color)
			c.Hub.Broadcast <- p
		}
	}
}

type Hub struct {
	Clients     map[*Client]bool
	Broadcast   chan []byte
	Register    chan *Client
	Unregister  chan *Client
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
