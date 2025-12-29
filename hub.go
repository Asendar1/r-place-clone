package main

import (
	"context"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

//#region Client

type Client struct {
	UUID   int64
	Hub    *Hub
	Send   chan []byte
	conn   *websocket.Conn
	closed sync.Once
}

func (c *Client) Close() {
	c.closed.Do(func() {
		c.Hub.Unregister <- c
		c.conn.Close()
	})
}

func (c *Client) DisplayRefresh() {
	defer c.Close()

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
	defer c.Close()

	for {
		_, p, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		if len(p) == 4 {
			x := int(p[0])<<8 | int(p[1])
			y := (int(p[2])<<8 | int(p[3])) >> 4
			color := int(p[3] & 0x0F)

			if x >= 1000 || y >= 1000 || color > 15 || color < 0 {
				continue
			}
			offset := y*1000 + x
			select {
			case c.Hub.redisQueue <- PixelUpdate{offset: offset, color: color}:
			default:
			}
			c.Hub.Broadcast <- p
		}
	}
}

//#endregion

//#region Hub

const ShardCount = 12

type Shard struct {
	Clients map[*Client]bool
	sync.RWMutex
}

type PixelUpdate struct {
	offset int
	color  int
}

type Hub struct {
	shards      [ShardCount]*Shard
	buffer      []byte
	Broadcast   chan []byte
	Register    chan *Client
	Unregister  chan *Client
	redisClient *redis.Client
	redisQueue  chan PixelUpdate
}

func (h *Hub) Run() {
	timer := time.NewTicker(time.Millisecond * 100)
	defer timer.Stop()

	for {
		select {
		case client := <-h.Register:
			i := client.UUID % ShardCount
			h.shards[i].Lock()
			h.shards[i].Clients[client] = true
			h.shards[i].Unlock()
		case client := <-h.Unregister:
			i := client.UUID % ShardCount
			h.shards[i].Lock()
			if _, ok := h.shards[i].Clients[client]; ok {
				delete(h.shards[i].Clients, client)
				close(client.Send)
			}
			h.shards[i].Unlock()
		case msg := <-h.Broadcast:
			h.buffer = append(h.buffer, msg...)
		case <-timer.C:

			totalClients := 0
			for i := 0; i < ShardCount; i++ {
				h.shards[i].RLock()
				totalClients += len(h.shards[i].Clients)
				h.shards[i].RUnlock()
			}

			currentBuffer := h.buffer
			h.buffer = make([]byte, 0, 4096)
			payload := h.makePayLoad(totalClients, currentBuffer)
			for i := 0; i < ShardCount; i++ {
				go func(s *Shard) {
					s.RLock()
					defer s.RUnlock()

					for client := range s.Clients {
						select {
						case client.Send <- payload:
						default:
							continue
						}
					}
				}(h.shards[i])
			}
		}
	}
}

func (h *Hub) makePayLoad(totalClients int, currentBuffer []byte) []byte {
	payload := make([]byte, len(currentBuffer)+5)

	payload[0] = 255
	payload[1] = byte(totalClients >> 24)
	payload[2] = byte(totalClients >> 16)
	payload[3] = byte(totalClients >> 8)
	payload[4] = byte(totalClients)

	copy(payload[5:], currentBuffer)
	return payload
}

func (h *Hub) redisWorker() {
	for update := range h.redisQueue {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		offsetStr := strconv.Itoa(update.offset)
		err := h.redisClient.Do(ctx, "BITFIELD", "canvas", "SET", "u4", "#"+offsetStr, update.color).Err()
		if err != nil {
			log.Printf("Redis Error: %v", err)
		}
		cancel()
	}
}

//#endregion
