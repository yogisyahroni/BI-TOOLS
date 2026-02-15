package handlers

import (
	"insight-engine-backend/services"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// Connection limits to prevent "Insufficient resources" error
const (
	maxTotalConnections   = 5000  // Increased for stability
	maxConnectionsPerUser = 50    // Increased to handle multiple tabs/reloads
	maxGoroutines         = 10000 // Safety limit
)

// Track global connection count
var globalConnectionCount int32 = 0

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	wsHub *services.WebSocketHub
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(wsHub *services.WebSocketHub) *WebSocketHandler {
	return &WebSocketHandler{
		wsHub: wsHub,
	}
}

// HandleConnection handles WebSocket connection upgrade and communication
func (h *WebSocketHandler) HandleConnection(c *websocket.Conn) {
	// Check global connection limit first
	currentCount := atomic.LoadInt32(&globalConnectionCount)
	if currentCount >= maxTotalConnections {
		services.LogWarn("websocket_global_limit", "Global connection limit reached", map[string]interface{}{
			"current": currentCount,
			"max":     maxTotalConnections,
		})
		c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseTryAgainLater, "Server busy"))
		c.Close()
		return
	}

	// Increment counter
	atomic.AddInt32(&globalConnectionCount, 1)
	defer atomic.AddInt32(&globalConnectionCount, -1)

	// Check goroutine limit
	if runtime.NumGoroutine() > maxGoroutines {
		services.LogWarn("websocket_goroutine_limit", "Too many goroutines", map[string]interface{}{
			"current": runtime.NumGoroutine(),
			"max":     maxGoroutines,
		})
		c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseTryAgainLater, "Server busy"))
		c.Close()
		return
	}

	// Get user ID from locals (set by auth middleware)
	userID := c.Locals("userID")
	if userID == nil {
		services.LogWarn("websocket_no_user", "No user ID in WebSocket context", nil)
		c.Close()
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		services.LogWarn("websocket_invalid_user_type", "Invalid user ID type in WebSocket context", nil)
		c.Close()
		return
	}

	// Check per-user connection limit
	userConns := h.wsHub.GetUserConnectionCount(userIDStr)
	if userConns >= maxConnectionsPerUser {
		services.LogWarn("websocket_user_limit", "User connection limit reached", map[string]interface{}{
			"user_id": userIDStr,
			"current": userConns,
			"max":     maxConnectionsPerUser,
		})
		c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "Too many connections"))
		c.Close()
		return
	}

	services.LogInfo("websocket_connecting", "New WebSocket connection attempt", map[string]interface{}{
		"user_id":            userIDStr,
		"global_connections": currentCount,
		"goroutines":         runtime.NumGoroutine(),
	})

	// Create client with smaller buffer to save memory
	client := &services.WebSocketClient{
		UserID: userIDStr,
		Conn:   c,
		Send:   make(chan []byte, 64), // Reduced from 256 to save memory
	}

	// Register client
	h.wsHub.Register(client)

	// Start goroutines for reading and writing
	go h.writePump(client)
	h.readPump(client)
}

// readPump reads messages from the WebSocket connection
func (h *WebSocketHandler) readPump(client *services.WebSocketClient) {
	defer func() {
		h.wsHub.Unregister(client)
		client.Conn.Close()
	}()

	// Set read deadline
	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				services.LogError("websocket_read_error", "Error reading WebSocket message", map[string]interface{}{"error": err, "user_id": client.UserID})
			}
			break
		}

		// Handle incoming messages (ping/pong, etc)
		services.LogDebug("websocket_message_received", "Message received from client", map[string]interface{}{"user_id": client.UserID, "message_length": len(message)})
	}
}

// writePump writes messages to the WebSocket connection
func (h *WebSocketHandler) writePump(client *services.WebSocketClient) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Channel closed
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				services.LogError("websocket_write_error", "Error writing WebSocket message", map[string]interface{}{"error": err, "user_id": client.UserID})
				return
			}

		case <-ticker.C:
			// Send ping
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// GetStats returns WebSocket connection statistics
func (h *WebSocketHandler) GetStats(c *fiber.Ctx) error {
	stats := fiber.Map{
		"connectedUsers":   h.wsHub.GetConnectedUsers(),
		"totalConnections": h.wsHub.GetConnectionCount(),
		"timestamp":        time.Now(),
	}

	return c.JSON(stats)
}
