// Package websocket provides a generic WebSocket connection registry (Hub)
// and client wrapper for real-time communication.
//
// This is a foundation-layer package that uses string-based IDs for maximum
// flexibility. Application-specific semantics (e.g., "user:{uuid}", "role:{uuid}")
// are defined by higher layers.
//
// Architecture:
//
//	Hub manages all WebSocket connections and provides broadcast methods.
//	Client wraps individual WebSocket connections with read/write pumps.
//	Message defines the envelope format for WebSocket messages.
//
// Usage:
//
//	// Create hub
//	hub := websocket.NewHub(log)
//	go hub.Run(ctx)
//
//	// On connection, create client and register
//	client := websocket.NewClient(hub, conn, log)
//	hub.Register(ctx, client, []string{"user:uuid", "role:admin"})
//
//	// Broadcast to specific ID
//	hub.BroadcastToID("user:uuid", messageBytes)
//
//	// Broadcast to all
//	hub.BroadcastAll(messageBytes)
package websocket
