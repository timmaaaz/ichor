package communication

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/rabbitmq"
)

// BuildAlertPayload is the single source of truth for the WebSocket JSON
// payload shape used by every alert publisher. Key names and inclusion rules
// must match the alertapi HTTP response (see api/domain/http/workflow/alertapi/model.go:toAppAlert)
// so frontend subscribers receive the same shape on either channel.
//
// When adding a new field to alertbus.Alert that should reach the frontend,
// add it here — not at individual publish sites — to prevent twin-site drift.
func BuildAlertPayload(alert alertbus.Alert) map[string]interface{} {
	payload := map[string]interface{}{
		"id":          alert.ID.String(),
		"alertType":   alert.AlertType,
		"severity":    alert.Severity,
		"title":       alert.Title,
		"message":     alert.Message,
		"status":      alert.Status,
		"createdDate": alert.CreatedDate.Format(time.RFC3339),
		"updatedDate": alert.UpdatedDate.Format(time.RFC3339),
	}

	if len(alert.Context) > 0 {
		payload["context"] = alert.Context
	}
	if alert.SourceEntityName != "" {
		payload["sourceEntityName"] = alert.SourceEntityName
	}
	if alert.SourceEntityID != uuid.Nil {
		payload["sourceEntityId"] = alert.SourceEntityID.String()
	}
	if alert.SourceRuleID != uuid.Nil {
		payload["sourceRuleId"] = alert.SourceRuleID.String()
	}
	if alert.SourceRuleName != "" {
		payload["sourceRuleName"] = alert.SourceRuleName
	}
	if alert.ActionURL != "" {
		payload["actionUrl"] = alert.ActionURL
	}
	if alert.ExpiresDate != nil {
		payload["expiresDate"] = alert.ExpiresDate.Format(time.RFC3339)
	}

	return payload
}

// buildRecipientAlertMessage constructs the RabbitMQ Message for a single
// alert recipient. Users are targeted via msg.UserID; roles via a "role_id"
// key in the payload (the consumer reads it from there).
func buildRecipientAlertMessage(alert alertbus.Alert, alertData map[string]interface{}, recipient alertbus.AlertRecipient) *rabbitmq.Message {
	payload := map[string]interface{}{"alert": alertData}
	msg := &rabbitmq.Message{
		Type:       "alert",
		EntityName: "workflow.alerts",
		EntityID:   alert.ID,
	}
	switch recipient.RecipientType {
	case "user":
		msg.UserID = recipient.RecipientID
	case "role":
		payload["role_id"] = recipient.RecipientID.String()
	}
	msg.Payload = payload
	return msg
}

// PublishAlertToRecipients publishes an alert to RabbitMQ (QueueTypeAlert) for
// WebSocket delivery — one message per recipient. This is the single
// alert-publishing seam; every site that creates a workflow alert should call
// this instead of hand-rolling publish loops. Publish errors are logged but
// not returned — the alert is already persisted and WS delivery is
// best-effort. A nil publisher is a no-op (graceful degradation for tests and
// core registrations).
func PublishAlertToRecipients(ctx context.Context, publisher *rabbitmq.WorkflowQueue, log *logger.Logger, alert alertbus.Alert, recipients []alertbus.AlertRecipient) {
	if publisher == nil {
		return
	}
	alertData := BuildAlertPayload(alert)
	for _, recipient := range recipients {
		msg := buildRecipientAlertMessage(alert, alertData, recipient)
		if err := publisher.Publish(ctx, rabbitmq.QueueTypeAlert, msg); err != nil {
			log.Error(ctx, "failed to publish alert to rabbitmq",
				"alert_id", alert.ID,
				"recipient_type", recipient.RecipientType,
				"recipient_id", recipient.RecipientID,
				"error", err)
		}
	}
}
