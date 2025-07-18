package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"discord-user-api/models"
)

type WebSocketManager struct {
	clients    map[*Client]bool
	broadcast  chan models.WebSocketEvent
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
}

type Client struct {
	manager *WebSocketManager
	conn    *websocket.Conn
	send    chan []byte
	userID  string
	guildID string
}

func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan models.WebSocketEvent),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (manager *WebSocketManager) Start() {
	log.Printf("🔌 WebSocket Manager başlatıldı")
	
	for {
		select {
		case client := <-manager.register:
			manager.mutex.Lock()
			manager.clients[client] = true
			manager.mutex.Unlock()
			log.Printf("🔗 Yeni WebSocket bağlantısı: %s", client.conn.RemoteAddr())

		case client := <-manager.unregister:
			manager.mutex.Lock()
			if _, ok := manager.clients[client]; ok {
				delete(manager.clients, client)
				close(client.send)
			}
			manager.mutex.Unlock()
			log.Printf("🔌 WebSocket bağlantısı kapatıldı: %s", client.conn.RemoteAddr())

		case event := <-manager.broadcast:
			manager.mutex.RLock()
			for client := range manager.clients {
				if shouldSendToClient(event, client) {
					select {
					case client.send <- eventToJSON(event):
					default:
						close(client.send)
						delete(manager.clients, client)
					}
				}
			}
			manager.mutex.RUnlock()
		}
	}
}

func (manager *WebSocketManager) Broadcast(event models.WebSocketEvent) {
	manager.broadcast <- event
}

func (manager *WebSocketManager) BroadcastToGuild(guildID string, eventType string, data interface{}) {
	event := models.WebSocketEvent{
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		GuildID:   guildID,
	}
	manager.Broadcast(event)
}

func (manager *WebSocketManager) BroadcastToUser(userID string, eventType string, data interface{}) {
	event := models.WebSocketEvent{
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		UserID:    userID,
	}
	manager.Broadcast(event)
}

func (manager *WebSocketManager) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ WebSocket upgrade hatası: %v", err)
		return
	}

	client := &Client{
		manager: manager,
		conn:    conn,
		send:    make(chan []byte, 256),
		userID:  r.URL.Query().Get("user_id"),
		guildID: r.URL.Query().Get("guild_id"),
	}

	manager.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.manager.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("❌ WebSocket okuma hatası: %v", err)
			}
			break
		}

		c.handleMessage(message)
	}
}

func (c *Client) handleMessage(message []byte) {
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("❌ WebSocket mesaj parse hatası: %v", err)
		return
	}

	msgType, ok := msg["type"].(string)
	if !ok {
		return
	}

	switch msgType {
	case "subscribe":
		if guildID, ok := msg["guild_id"].(string); ok {
			c.guildID = guildID
			log.Printf("📡 Kullanıcı guild'e abone oldu: %s", guildID)
		}
	case "unsubscribe":
		c.guildID = ""
		log.Printf("📡 Kullanıcı abonelikten çıktı")
	case "ping":
		response := models.WebSocketEvent{
			Type:      "pong",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
		c.send <- eventToJSON(response)
	}
}

func shouldSendToClient(event models.WebSocketEvent, client *Client) bool {
	if event.GuildID != "" {
		return client.guildID == event.GuildID
	}

	if event.UserID != "" {
		return client.userID == event.UserID
	}

	return true
}

func eventToJSON(event models.WebSocketEvent) []byte {
	data, _ := json.Marshal(event)
	return data
}

func (manager *WebSocketManager) GetConnectedClientsCount() int {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()
	return len(manager.clients)
}

func (manager *WebSocketManager) GetConnectedClientsInfo() []map[string]interface{} {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	var clients []map[string]interface{}
	for client := range manager.clients {
		clients = append(clients, map[string]interface{}{
			"user_id":  client.userID,
			"guild_id": client.guildID,
			"address":  client.conn.RemoteAddr().String(),
		})
	}
	return clients
} 