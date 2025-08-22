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

// SendEmailHandler handles send_email actions
type SendEmailHandler struct {
	log *logger.Logger
	db  *sqlx.DB
}

// NewSendEmailHandler creates a new send email handler
func NewSendEmailHandler(log *logger.Logger, db *sqlx.DB) *SendEmailHandler {
	return &SendEmailHandler{
		log: log,
		db:  db,
	}
}

func (h *SendEmailHandler) GetType() string {
	return "send_email"
}

func (h *SendEmailHandler) Validate(config json.RawMessage) error {
	var cfg struct {
		Recipients []string `json:"recipients"`
		Subject    string   `json:"subject"`
	}

	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid configuration format: %w", err)
	}

	if len(cfg.Recipients) == 0 {
		return fmt.Errorf("email recipients list is required and must not be empty")
	}

	if cfg.Subject == "" {
		return fmt.Errorf("email subject is required")
	}

	return nil
}

func (h *SendEmailHandler) Execute(ctx context.Context, config json.RawMessage, context workflow.ActionExecutionContext) (interface{}, error) {
	h.log.Info(ctx, "Executing send_email action",
		"entityID", context.EntityID,
		"ruleName", context.RuleName)

	result := map[string]interface{}{
		"email_id": fmt.Sprintf("email_%d", time.Now().Unix()),
		"status":   "sent",
		"sent_at":  time.Now().Format(time.RFC3339),
	}

	return result, nil
}
