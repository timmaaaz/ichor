package websocket

import (
	"context"

	"github.com/timmaaaz/ichor/foundation/rabbitmq"
)

// MessageHandler processes messages from the queue for real-time delivery via WebSocket.
// Implementations handle specific message types (alerts, inventory updates, etc.)
type MessageHandler interface {
	// MessageType returns the type string this handler processes (e.g., "alert").
	MessageType() string

	// HandleMessage processes a single message for WebSocket delivery.
	HandleMessage(ctx context.Context, msg *rabbitmq.Message) error
}
