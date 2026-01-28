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

// SupportsManualExecution returns true - emails can be sent manually
func (h *SendEmailHandler) SupportsManualExecution() bool {
	return true
}

// IsAsync returns true - email sending is queued for async processing
func (h *SendEmailHandler) IsAsync() bool {
	return true
}

// GetDescription returns a human-readable description for discovery APIs
func (h *SendEmailHandler) GetDescription() string {
	return "Send an email to specified recipients"
}

func (h *SendEmailHandler) Validate(config json.RawMessage) error {
	var cfg struct {
		Recipients      []string `json:"recipients"`
		Subject         string   `json:"subject"`
		SimulateFailure bool     `json:"simulate_failure,omitempty"`
		FailureMessage  string   `json:"failure_message,omitempty"`
	}

	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid configuration format: %w", err)
	}

	// Skip validation if we're simulating failure for testing
	if cfg.SimulateFailure {
		return nil
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
	var cfg struct {
		Recipients      []string `json:"recipients"`
		Subject         string   `json:"subject"`
		Body            string   `json:"body"`
		SimulateFailure bool     `json:"simulate_failure,omitempty"`
		FailureMessage  string   `json:"failure_message,omitempty"`
	}

	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse email configuration: %w", err)
	}

	// Handle test failure simulation
	if cfg.SimulateFailure {
		failureMsg := cfg.FailureMessage
		if failureMsg == "" {
			failureMsg = "simulated email delivery failure: SMTP connection refused"
		}
		h.log.Warn(ctx, "Simulating email failure for testing",
			"entityID", context.EntityID,
			"ruleName", context.RuleName,
			"error", failureMsg)
		return nil, fmt.Errorf(failureMsg)
	}

	h.log.Info(ctx, "Executing send_email action",
		"entityID", context.EntityID,
		"ruleName", context.RuleName,
		"recipients", cfg.Recipients,
		"subject", cfg.Subject)

	// TODO: Implement actual email sending here
	// smtp.SendMail(server, auth, from, recipients, message)

	result := map[string]interface{}{
		"email_id":   fmt.Sprintf("email_%d", time.Now().Unix()),
		"status":     "sent",
		"sent_at":    time.Now().Format(time.RFC3339),
		"recipients": cfg.Recipients,
		"subject":    cfg.Subject,
	}

	return result, nil
}
