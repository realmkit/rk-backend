package http

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/gofiber/contrib/websocket"
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/pkg/events/domain"
)

// Hub stores active local WebSocket clients.
type Hub struct {
	mu      sync.Mutex
	clients map[uuid.UUID]*client
}

// NewHub creates a WebSocket hub.
func NewHub() *Hub {
	return &Hub{clients: map[uuid.UUID]*client{}}
}

// Publish broadcasts one dispatched event to local WebSocket clients.
func (hub *Hub) Publish(_ context.Context, event domain.Event) error {
	hub.Broadcast(event)
	return nil
}

// Broadcast sends event to subscribed local clients.
func (hub *Hub) Broadcast(event domain.Event) {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	for _, client := range hub.clients {
		if client.matches(event.Scopes) {
			_ = client.write(event)
		}
	}
}

// add stores one client.
func (hub *Hub) add(client *client) {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	hub.clients[client.id] = client
}

// remove deletes one client.
func (hub *Hub) remove(id uuid.UUID) {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	delete(hub.clients, id)
}

// webSocket handles one WebSocket connection.
func (handler handler) webSocket(conn *websocket.Conn) {
	hub := handler.services.Hub
	if hub == nil {
		hub = NewHub()
	}
	client := &client{
		id:     uuid.New(),
		conn:   conn,
		userID: socketUserID(conn),
		scopes: map[string]domain.Scope{},
	}
	hub.add(client)
	defer hub.remove(client.id)
	_ = conn.WriteJSON(map[string]any{"type": "ready", "connection_id": client.id.String()})
	for {
		_, body, err := conn.ReadMessage()
		if err != nil {
			return
		}
		client.handle(body)
	}
}

// client is one WebSocket connection.
type client struct {
	id     uuid.UUID
	conn   *websocket.Conn
	userID uuid.UUID
	scopes map[string]domain.Scope
}

// handle handles one client message.
func (client *client) handle(body []byte) {
	var message socketMessage
	if err := json.Unmarshal(body, &message); err != nil {
		_ = client.conn.WriteJSON(map[string]any{"type": "error", "code": "invalid_json"})
		return
	}
	switch message.Type {
	case "subscribe":
		if !client.canSubscribe(message.Scope) {
			_ = client.conn.WriteJSON(map[string]any{"type": "error", "code": "scope_forbidden"})
			return
		}
		client.scopes[scopeKey(message.Scope)] = message.Scope
		_ = client.conn.WriteJSON(map[string]any{"type": "subscribed", "scope": message.Scope})
	case "unsubscribe":
		delete(client.scopes, scopeKey(message.Scope))
	case "ping":
		_ = client.conn.WriteJSON(map[string]any{"type": "pong"})
	default:
		_ = client.conn.WriteJSON(map[string]any{"type": "error", "code": "unknown_message_type"})
	}
}

// matches reports whether client subscribes to any scope.
func (client *client) matches(scopes []domain.Scope) bool {
	for _, scope := range scopes {
		if _, ok := client.scopes[scopeKey(scope)]; ok {
			return true
		}
	}
	return false
}

// canSubscribe authorizes one subscription scope.
func (client *client) canSubscribe(scope domain.Scope) bool {
	if err := scope.Validate(); err != nil {
		return false
	}
	switch scope.Type {
	case domain.ScopeGlobal:
		return true
	case domain.ScopeUser:
		return client.userID != uuid.Nil && scope.ID == client.userID.String()
	default:
		return false
	}
}

// write sends one event message.
func (client *client) write(event domain.Event) error {
	return client.conn.WriteJSON(map[string]any{"type": "event", "event": event})
}

// socketMessage is a client WebSocket message.
type socketMessage struct {
	Type  string       `json:"type"`
	Scope domain.Scope `json:"scope"`
}

// scopeKey returns a comparable scope key.
func scopeKey(scope domain.Scope) string {
	return string(scope.Type) + ":" + scope.ID + ":" + scope.Permission
}

// socketUserID extracts the authenticated user ID made available to the socket.
func socketUserID(conn *websocket.Conn) uuid.UUID {
	value := conn.Headers("X-GameHub-User-Id")
	if value == "" {
		value = conn.Query("user_id")
	}
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil
	}
	return id
}
