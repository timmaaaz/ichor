package websocket

import (
	"encoding/json"
	"time"
)

// MessageType identifies the type of WebSocket message.
type MessageType string

const (
	// MessageTypeAlert is a new alert notification message.
	MessageTypeAlert MessageType = "alert"
	// MessageTypeAlertUpdated signals that an existing alert's status changed.
	MessageTypeAlertUpdated MessageType = "alert_updated"
	// MessageTypeApprovalUpdated signals that an approval request was resolved.
	MessageTypeApprovalUpdated MessageType = "approval_updated"
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
