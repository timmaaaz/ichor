package communication

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// SendNotificationHandler handles send_notification actions
type SendNotificationHandler struct {
	log *logger.Logger
	db  *sqlx.DB
}

// NewSendNotificationHandler creates a new send notification handler
func NewSendNotificationHandler(log *logger.Logger, db *sqlx.DB) *SendNotificationHandler {
	return &SendNotificationHandler{
		log: log,
		db:  db,
	}
}

func (h *SendNotificationHandler) GetType() string {
	return "send_notification"
}

func (h *SendNotificationHandler) Validate(config json.RawMessage) error {
	var cfg struct {
		Recipients []string `json:"recipients"`
		Channels   []struct {
			Type string `json:"type"`
		} `json:"channels"`
		Priority string `json:"priority"`
	}

	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid configuration format: %w", err)
	}

	if len(cfg.Recipients) == 0 {
		return fmt.Errorf("recipients list is required and must not be empty")
	}

	if len(cfg.Channels) == 0 {
		return fmt.Errorf("at least one notification channel is required")
	}

	validPriorities := map[string]bool{
		"low": true, "medium": true, "high": true, "critical": true,
	}
	if !validPriorities[cfg.Priority] {
		return fmt.Errorf("invalid priority level")
	}

	return nil
}

func (h *SendNotificationHandler) Execute(ctx context.Context, config json.RawMessage, context workflow.ActionExecutionContext) (interface{}, error) {
	h.log.Info(ctx, "Executing send_notification action",
		"entityID", context.EntityID,
		"ruleName", context.RuleName)

	result := map[string]interface{}{
		"notification_id": fmt.Sprintf("notif_%d", time.Now().Unix()),
		"status":          "sent",
		"sent_at":         time.Now().Format(time.RFC3339),
	}

	return result, nil
}
