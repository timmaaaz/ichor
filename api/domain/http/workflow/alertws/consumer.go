package alertws

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/rabbitmq"
	"github.com/timmaaaz/ichor/foundation/websocket"
)

// AlertConsumer bridges RabbitMQ alerts to WebSocket clients.
type AlertConsumer struct {
	alertHub *AlertHub
	wq       *rabbitmq.WorkflowQueue
	log      *logger.Logger
	consumer *rabbitmq.Consumer
}

// NewAlertConsumer creates a new AlertConsumer.
func NewAlertConsumer(alertHub *AlertHub, wq *rabbitmq.WorkflowQueue, log *logger.Logger) *AlertConsumer {
	return &AlertConsumer{
		alertHub: alertHub,
		wq:       wq,
		log:      log,
	}
}

// Start begins consuming alerts from RabbitMQ.
// Blocks until context is cancelled.
// Note: When using SetHandlerRegistry on QueueManager, Start() is not needed
// as the QueueManager will route alert messages directly to HandleMessage.
func (ac *AlertConsumer) Start(ctx context.Context) error {
	consumer, err := ac.wq.Consume(ctx, rabbitmq.QueueTypeAlert, ac.handleAlert)
	if err != nil {
		return fmt.Errorf("starting alert consumer: %w", err)
	}
	ac.consumer = consumer

	<-ctx.Done()
	consumer.Stop()
	return ctx.Err()
}

// =============================================================================
// websocket.MessageHandler interface implementation
// =============================================================================

// MessageType returns the message type this handler processes.
// Implements websocket.MessageHandler.
func (ac *AlertConsumer) MessageType() string {
	return "alert"
}

// HandleMessage processes a single alert message for WebSocket delivery.
// Implements websocket.MessageHandler.
func (ac *AlertConsumer) HandleMessage(ctx context.Context, msg *rabbitmq.Message) error {
	return ac.handleAlert(ctx, msg)
}

// handleAlert processes a single alert message from RabbitMQ.
func (ac *AlertConsumer) handleAlert(ctx context.Context, msg *rabbitmq.Message) error {
	// Check for final failure (max retries exceeded)
	if finalFailure, ok := msg.Payload["_final_failure"].(bool); ok && finalFailure {
		ac.log.Error(ctx, "alert delivery failed after retries",
			"message_id", msg.ID,
			"entity_id", msg.EntityID)
		return nil // Don't retry - already at max
	}

	// Extract alert payload - it may be a map or RawMessage
	var alertPayload json.RawMessage
	if alertData, ok := msg.Payload["alert"]; ok {
		// Convert alert data to json.RawMessage
		data, err := json.Marshal(alertData)
		if err != nil {
			ac.log.Error(ctx, "failed to marshal alert data",
				"message_id", msg.ID,
				"error", err)
			return nil // Don't retry marshal errors
		}
		alertPayload = data
	} else {
		// If no "alert" key, use the entire payload
		data, err := json.Marshal(msg.Payload)
		if err != nil {
			ac.log.Error(ctx, "failed to marshal payload",
				"message_id", msg.ID,
				"error", err)
			return nil // Don't retry marshal errors
		}
		alertPayload = data
	}

	// Build WebSocket message envelope
	wsMsg := websocket.Message{
		Type:      websocket.MessageTypeAlert,
		Payload:   alertPayload,
		Timestamp: time.Now(),
	}

	msgBytes, err := json.Marshal(wsMsg)
	if err != nil {
		ac.log.Error(ctx, "failed to marshal websocket message", "error", err)
		return nil // Don't retry marshal errors
	}

	// Determine targeting and broadcast using AlertHub (business semantics)
	var delivered int

	// User-targeted alert
	if msg.UserID != uuid.Nil {
		delivered = ac.alertHub.BroadcastToUser(msg.UserID, msgBytes)
		ac.log.Info(ctx, "broadcast alert to user",
			"user_id", msg.UserID,
			"connections", delivered)
		// Role-targeted alert
	} else if roleID, ok := msg.Payload["role_id"].(string); ok && roleID != "" {
		rid, err := uuid.Parse(roleID)
		if err != nil {
			ac.log.Error(ctx, "invalid role_id in alert payload",
				"role_id", roleID,
				"error", err)
			return nil // Don't retry invalid role ID
		}
		delivered = ac.alertHub.BroadcastToRole(rid, msgBytes)
		ac.log.Info(ctx, "broadcast alert to role",
			"role_id", rid,
			"connections", delivered)
		// Broadcast to all connected clients
	} else {
		delivered = ac.alertHub.BroadcastAll(msgBytes)
		ac.log.Info(ctx, "broadcast alert to all",
			"connections", delivered)
	}

	return nil
}
