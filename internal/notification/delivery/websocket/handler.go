package websocket

import (
	"log"
	"net/http"
	"time"

	authmw "github.com/eduaccess/eduaccess-api/internal/shared/middleware"
	gorilla "github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	maxMsgSize = 512
)

var upgrader = gorilla.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // allow all origins; tighten in production via CORS_ALLOW_ORIGINS
	},
}

// Handler handles WebSocket upgrade requests for the notification stream.
type Handler struct{ hub *Hub }

func NewHandler(hub *Hub) *Handler { return &Handler{hub: hub} }

// ServeWS upgrades the HTTP connection to WebSocket. The caller must authenticate
// by passing their Supabase JWT in the ?token= query parameter.
func (h *Handler) ServeWS(c echo.Context) error {
	tokenStr := c.QueryParam("token")
	if tokenStr == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing token"})
	}

	userID, err := authmw.ValidateToken(tokenStr)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid or expired token"})
	}

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Printf("[ws] upgrade error for user %s: %v", userID, err)
		return nil
	}

	client := &Client{
		UserID: userID,
		conn:   conn,
		send:   make(chan []byte, 64),
	}
	h.hub.Register <- client

	go client.writePump()
	go client.readPump(h.hub)

	return nil
}

// readPump keeps the connection alive (handles pong and detects disconnect).
func (c *Client) readPump(hub *Hub) {
	defer func() {
		hub.Unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMsgSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}

// writePump drains the client's send channel and forwards payloads over WebSocket.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(gorilla.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(gorilla.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(gorilla.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
