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

// CreateAlertHandler handles create_alert actions
type CreateAlertHandler struct {
	log *logger.Logger
	db  *sqlx.DB
}

// NewCreateAlertHandler creates a new create alert handler
func NewCreateAlertHandler(log *logger.Logger, db *sqlx.DB) *CreateAlertHandler {
	return &CreateAlertHandler{
		log: log,
		db:  db,
	}
}

func (h *CreateAlertHandler) GetType() string {
	return "create_alert"
}

func (h *CreateAlertHandler) Validate(config json.RawMessage) error {
	var cfg struct {
		Message    string   `json:"message"`
		Recipients []string `json:"recipients"`
		Priority   string   `json:"priority"`
	}

	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid configuration format: %w", err)
	}

	if cfg.Message == "" {
		return fmt.Errorf("alert message is required")
	}

	if len(cfg.Recipients) == 0 {
		return fmt.Errorf("recipients list is required and must not be empty")
	}

	validPriorities := map[string]bool{
		"low": true, "medium": true, "high": true, "critical": true,
	}
	if !validPriorities[cfg.Priority] {
		return fmt.Errorf("invalid priority level")
	}

	return nil
}

func (h *CreateAlertHandler) Execute(ctx context.Context, config json.RawMessage, context workflow.ActionExecutionContext) (interface{}, error) {
	h.log.Info(ctx, "Executing create_alert action",
		"entityID", context.EntityID,
		"ruleName", context.RuleName)

	result := map[string]interface{}{
		"alert_id":   fmt.Sprintf("alert_%d", time.Now().Unix()),
		"status":     "created",
		"created_at": time.Now().Format(time.RFC3339),
	}

	return result, nil
}
