package websocket

import (
	"log"
	"sync"

	"github.com/google/uuid"
	gorilla "github.com/gorilla/websocket"
)

// Message is a payload destined for a specific connected user.
type Message struct {
	UserID  uuid.UUID
	Payload []byte
}

// Client represents one active WebSocket connection owned by a user.
type Client struct {
	UserID uuid.UUID
	conn   *gorilla.Conn
	send   chan []byte
}

// Hub manages all active WebSocket connections, keyed by user_id.
// A single user can have multiple simultaneous connections (e.g. two browser tabs).
type Hub struct {
	mu         sync.RWMutex
	clients    map[uuid.UUID]map[*Client]bool

	Register   chan *Client
	Unregister chan *Client
	Send       chan Message
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[uuid.UUID]map[*Client]bool),
		Register:   make(chan *Client, 64),
		Unregister: make(chan *Client, 64),
		Send:       make(chan Message, 256),
	}
}

// Run processes hub events. Must be called in its own goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			if h.clients[client.UserID] == nil {
				h.clients[client.UserID] = make(map[*Client]bool)
			}
			h.clients[client.UserID][client] = true
			h.mu.Unlock()

		case client := <-h.Unregister:
			h.mu.Lock()
			if conns, ok := h.clients[client.UserID]; ok {
				delete(conns, client)
				if len(conns) == 0 {
					delete(h.clients, client.UserID)
				}
			}
			h.mu.Unlock()
			close(client.send)

		case msg := <-h.Send:
			h.mu.RLock()
			conns := h.clients[msg.UserID]
			h.mu.RUnlock()
			for c := range conns {
				select {
				case c.send <- msg.Payload:
				default:
					log.Printf("[ws-hub] slow client %s, dropping message", msg.UserID)
				}
			}
		}
	}
}

// Broadcast sends a raw JSON payload to all connections owned by userID.
// It is non-blocking — if the send channel is full the message is dropped.
func (h *Hub) Broadcast(userID uuid.UUID, payload []byte) {
	select {
	case h.Send <- Message{UserID: userID, Payload: payload}:
	default:
		log.Printf("[ws-hub] send channel full, dropping broadcast for user %s", userID)
	}
}
