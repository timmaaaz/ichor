package websocket

import (
	"encoding/json"
	"time"
)

// MessageType identifies the type of WebSocket message.
type MessageType string

const (
	// MessageTypeAlert is an alert notification message.
	MessageTypeAlert MessageType = "alert"
	// MessageTypePing is a heartbeat ping message.
	MessageTypePing MessageType = "ping"
	// MessageTypePong is a heartbeat pong response message.
	MessageTypePong MessageType = "pong"
)

// Message is the envelope format for WebSocket messages.
// All messages sent through the WebSocket are wrapped in this format.
type Message struct {
	Type      MessageType     `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
}
