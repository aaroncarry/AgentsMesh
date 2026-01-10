package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 64 * 1024 // 64KB
)

// MessageType defines the type of WebSocket message
type MessageType string

const (
	MessageTypeTerminalInput  MessageType = "terminal:input"
	MessageTypeTerminalOutput MessageType = "terminal:output"
	MessageTypeTerminalResize MessageType = "terminal:resize"
	MessageTypePodStatus      MessageType = "pod:status"
	MessageTypeAgentStatus    MessageType = "agent:status"
	MessageTypeChannelMessage MessageType = "channel:message"
	MessageTypePing           MessageType = "ping"
	MessageTypePong           MessageType = "pong"
	MessageTypeError          MessageType = "error"
)

// Message represents a WebSocket message
type Message struct {
	Type      MessageType     `json:"type"`
	PodKey    string          `json:"pod_key,omitempty"`
	ChannelID int64           `json:"channel_id,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
	Timestamp int64           `json:"timestamp"`
}

// TerminalInputData represents terminal input
type TerminalInputData struct {
	Data string `json:"data"`
}

// TerminalOutputData represents terminal output
type TerminalOutputData struct {
	Data string `json:"data"`
}

// TerminalResizeData represents terminal resize
type TerminalResizeData struct {
	Cols int `json:"cols"`
	Rows int `json:"rows"`
}

// PodStatusData represents pod status update
type PodStatusData struct {
	Status      string `json:"status"`
	AgentStatus string `json:"agent_status,omitempty"`
}

// Client represents a WebSocket client
type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	send      chan []byte
	userID    int64
	orgID     int64
	podKey    string // Empty if not connected to a pod
	channelID int64  // Non-zero if subscribed to a channel
	mu        sync.Mutex
}

// Hub manages WebSocket connections
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Clients by pod
	podClients map[string]map[*Client]bool

	// Clients by channel
	channelClients map[int64]map[*Client]bool

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast to pod
	podBroadcast chan *PodMessage

	// Broadcast to channel
	channelBroadcast chan *ChannelMessage

	mu sync.RWMutex
}

// PodMessage represents a message to broadcast to a pod
type PodMessage struct {
	PodKey  string
	Message []byte
}

// ChannelMessage represents a message to broadcast to a channel
type ChannelMessage struct {
	ChannelID int64
	Message   []byte
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		clients:          make(map[*Client]bool),
		podClients:       make(map[string]map[*Client]bool),
		channelClients:   make(map[int64]map[*Client]bool),
		register:         make(chan *Client),
		unregister:       make(chan *Client),
		podBroadcast:     make(chan *PodMessage, 256),
		channelBroadcast: make(chan *ChannelMessage, 256),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true

			if client.podKey != "" {
				if h.podClients[client.podKey] == nil {
					h.podClients[client.podKey] = make(map[*Client]bool)
				}
				h.podClients[client.podKey][client] = true
			}

			if client.channelID != 0 {
				if h.channelClients[client.channelID] == nil {
					h.channelClients[client.channelID] = make(map[*Client]bool)
				}
				h.channelClients[client.channelID][client] = true
			}
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)

				if client.podKey != "" {
					delete(h.podClients[client.podKey], client)
					if len(h.podClients[client.podKey]) == 0 {
						delete(h.podClients, client.podKey)
					}
				}

				if client.channelID != 0 {
					delete(h.channelClients[client.channelID], client)
					if len(h.channelClients[client.channelID]) == 0 {
						delete(h.channelClients, client.channelID)
					}
				}
			}
			h.mu.Unlock()

		case msg := <-h.podBroadcast:
			h.mu.RLock()
			clients := h.podClients[msg.PodKey]
			h.mu.RUnlock()

			for client := range clients {
				select {
				case client.send <- msg.Message:
				default:
					h.unregister <- client
				}
			}

		case msg := <-h.channelBroadcast:
			h.mu.RLock()
			clients := h.channelClients[msg.ChannelID]
			h.mu.RUnlock()

			for client := range clients {
				select {
				case client.send <- msg.Message:
				default:
					h.unregister <- client
				}
			}
		}
	}
}

// BroadcastToPod sends a message to all clients connected to a pod
func (h *Hub) BroadcastToPod(podKey string, msg *Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.podBroadcast <- &PodMessage{
		PodKey:  podKey,
		Message: data,
	}
}

// BroadcastToChannel sends a message to all clients subscribed to a channel
func (h *Hub) BroadcastToChannel(channelID int64, msg *Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.channelBroadcast <- &ChannelMessage{
		ChannelID: channelID,
		Message:   data,
	}
}

// GetPodClientCount returns the number of clients connected to a pod
func (h *Hub) GetPodClientCount(podKey string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.podClients[podKey])
}

// Register registers a client with the hub
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister unregisters a client from the hub
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// NewClient creates a new client
func NewClient(hub *Hub, conn *websocket.Conn, userID, orgID int64) *Client {
	return &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: userID,
		orgID:  orgID,
	}
}

// SetPod sets the pod for this client
func (c *Client) SetPod(podKey string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove from old pod
	if c.podKey != "" {
		c.hub.mu.Lock()
		delete(c.hub.podClients[c.podKey], c)
		if len(c.hub.podClients[c.podKey]) == 0 {
			delete(c.hub.podClients, c.podKey)
		}
		c.hub.mu.Unlock()
	}

	c.podKey = podKey

	// Add to new pod
	if podKey != "" {
		c.hub.mu.Lock()
		if c.hub.podClients[podKey] == nil {
			c.hub.podClients[podKey] = make(map[*Client]bool)
		}
		c.hub.podClients[podKey][c] = true
		c.hub.mu.Unlock()
	}
}

// SetChannel sets the channel subscription for this client
func (c *Client) SetChannel(channelID int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove from old channel
	if c.channelID != 0 {
		c.hub.mu.Lock()
		delete(c.hub.channelClients[c.channelID], c)
		if len(c.hub.channelClients[c.channelID]) == 0 {
			delete(c.hub.channelClients, c.channelID)
		}
		c.hub.mu.Unlock()
	}

	c.channelID = channelID

	// Add to new channel
	if channelID != 0 {
		c.hub.mu.Lock()
		if c.hub.channelClients[channelID] == nil {
			c.hub.channelClients[channelID] = make(map[*Client]bool)
		}
		c.hub.channelClients[channelID][c] = true
		c.hub.mu.Unlock()
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump(onMessage func(*Client, *Message)) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		if onMessage != nil {
			onMessage(c, &msg)
		}
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
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
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Send sends a message to the client
func (c *Client) Send(msg *Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	select {
	case c.send <- data:
		return nil
	default:
		return nil
	}
}
