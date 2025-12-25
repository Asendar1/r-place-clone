package main

import (
	"context"
	"strconv"

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
			c.Hub.Broadcast <- p
		}

	}
}

type Hub struct {
	Clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
}

func (h *Hub) Run(redisClient *redis.Client) {
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
			// TODO fix redis write, this doesn't work
			x := uint32(msg[0])<<8 | uint32(msg[1])
			y := uint32(msg[2])<<8 | uint32(msg[3]) >> 4 // remove color bits

			offset := uint32(y + 500000 * x)
			offset_str := strconv.FormatUint(uint64(offset), 10)
			color := uint32(msg[3] & 0x0F)
			color_str := strconv.FormatUint(uint64(color), 10)
			redisClient.BitField(context.Background(), "canvas", "SET", "u8", offset_str, color_str)
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
