package http

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/pkg/api/principal"
	"github.com/realmkit/rk-backend/pkg/events/domain"
	"github.com/realmkit/rk-backend/pkg/events/port"
)

const (
	// socketContextKey stores the request context for upgraded sockets.
	socketContextKey = "realmkit.events.socket_context"

	// socketWriteTimeout bounds one WebSocket write.
	socketWriteTimeout = 5 * time.Second
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
func (hub *Hub) Publish(ctx context.Context, event domain.Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	hub.Broadcast(ctx, event)
	return nil
}

// Broadcast sends event to subscribed local clients.
func (hub *Hub) Broadcast(ctx context.Context, event domain.Event) {
	hub.mu.Lock()
	clients := make([]*client, 0, len(hub.clients))
	for _, client := range hub.clients {
		clients = append(clients, client)
	}
	hub.mu.Unlock()
	for _, client := range clients {
		if err := ctx.Err(); err != nil {
			return
		}
		if client.matches(event.Scopes) {
			_ = client.write(ctx, event)
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
		ctx:    socketContext(conn),
		userID: socketUserID(conn),
		scopes: map[string]domain.Scope{},
		authz:  handler.services.ScopeAuthorizer,
	}
	hub.add(client)
	defer hub.remove(client.id)
	if client.ctx != nil {
		go client.closeOnContext()
	}
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
	id       uuid.UUID
	conn     *websocket.Conn
	ctx      context.Context
	userID   uuid.UUID
	scopes   map[string]domain.Scope
	scopesMu sync.RWMutex
	writeMu  sync.Mutex
	authz    port.ScopeAuthorizer
}

// handle handles one client message.
func (client *client) handle(body []byte) {
	var message socketMessage
	if err := json.Unmarshal(body, &message); err != nil {
		_ = client.writeMessage(map[string]any{"type": "error", "code": "invalid_json"})
		return
	}
	switch message.Type {
	case "subscribe":
		if !client.canSubscribe(message.Scope) {
			_ = client.writeMessage(map[string]any{"type": "error", "code": "scope_forbidden"})
			return
		}
		client.scopesMu.Lock()
		client.scopes[scopeKey(message.Scope)] = message.Scope
		client.scopesMu.Unlock()
		_ = client.writeMessage(map[string]any{"type": "subscribed", "scope": message.Scope})
	case "unsubscribe":
		client.scopesMu.Lock()
		delete(client.scopes, scopeKey(message.Scope))
		client.scopesMu.Unlock()
	case "ping":
		_ = client.writeMessage(map[string]any{"type": "pong"})
	default:
		_ = client.writeMessage(map[string]any{"type": "error", "code": "unknown_message_type"})
	}
}

// matches reports whether client subscribes to any scope.
func (client *client) matches(scopes []domain.Scope) bool {
	client.scopesMu.RLock()
	defer client.scopesMu.RUnlock()
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
	case domain.ScopeSystem:
		return false
	default:
		if client.userID == uuid.Nil || client.authz == nil {
			return false
		}
		if client.ctx == nil {
			return false
		}
		allowed, err := client.authz.CanSubscribe(client.ctx, port.Principal{
			UserID:    client.userID,
			Anonymous: false,
		}, scope)
		return err == nil && allowed
	}
}

// write sends one event message.
func (client *client) write(ctx context.Context, event domain.Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return client.writeMessage(map[string]any{"type": "event", "event": event})
}

// writeMessage sends one WebSocket message.
func (client *client) writeMessage(message map[string]any) error {
	client.writeMu.Lock()
	defer client.writeMu.Unlock()
	if err := client.conn.SetWriteDeadline(time.Now().Add(socketWriteTimeout)); err != nil {
		return err
	}
	return client.conn.WriteJSON(message)
}

// closeOnContext closes the socket once the request context is cancelled.
func (client *client) closeOnContext() {
	<-client.ctx.Done()
	_ = client.conn.Close()
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

// socketUserID extracts the authenticated principal user ID from websocket locals.
func socketUserID(conn *websocket.Conn) uuid.UUID {
	current, ok := conn.Locals(principal.LocalKey).(principal.Principal)
	if !ok {
		return uuid.Nil
	}
	return current.UserID
}

// socketContext extracts the request context from websocket locals.
func socketContext(conn *websocket.Conn) context.Context {
	current, ok := conn.Locals(socketContextKey).(context.Context)
	if !ok {
		return nil
	}
	return current
}
