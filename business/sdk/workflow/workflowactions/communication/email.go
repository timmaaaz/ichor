package communication

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// SendEmailHandler handles send_email actions.
type SendEmailHandler struct {
	log         *logger.Logger
	db          *sqlx.DB
	emailClient EmailClient // nil = graceful degradation (log + skip)
	emailFrom   string      // sender address passed to EmailClient.Send
}

// NewSendEmailHandler creates a new send email handler.
// emailClient may be nil; if so, Execute logs a warning and returns a no-op result.
func NewSendEmailHandler(log *logger.Logger, db *sqlx.DB, emailClient EmailClient, emailFrom string) *SendEmailHandler {
	return &SendEmailHandler{
		log:         log,
		db:          db,
		emailClient: emailClient,
		emailFrom:   emailFrom,
	}
}

func (h *SendEmailHandler) GetType() string {
	return "send_email"
}

// SupportsManualExecution returns true — emails can be triggered manually.
func (h *SendEmailHandler) SupportsManualExecution() bool {
	return true
}

// IsAsync returns true — email sending is handled as an async Temporal activity.
func (h *SendEmailHandler) IsAsync() bool {
	return true
}

// GetDescription returns a human-readable description for discovery APIs.
func (h *SendEmailHandler) GetDescription() string {
	return "Send an email to specified recipients"
}

func (h *SendEmailHandler) Validate(config json.RawMessage) error {
	var cfg struct {
		Recipients      []string `json:"recipients"`
		Subject         string   `json:"subject"`
		SimulateFailure bool     `json:"simulate_failure,omitempty"`
	}

	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid configuration format: %w", err)
	}

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

func (h *SendEmailHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext) (interface{}, error) {
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

	if cfg.SimulateFailure {
		failureMsg := cfg.FailureMessage
		if failureMsg == "" {
			failureMsg = "simulated email delivery failure"
		}
		h.log.Warn(ctx, "Simulating email failure for testing",
			"entityID", execCtx.EntityID,
			"ruleName", execCtx.RuleName,
			"error", failureMsg)
		return nil, fmt.Errorf(failureMsg)
	}

	// Resolve template variables in subject and body.
	subject := resolveTemplateVars(cfg.Subject, execCtx.RawData)
	body := resolveTemplateVars(cfg.Body, execCtx.RawData)

	emailID := uuid.New().String()
	now := time.Now()

	if h.emailClient != nil {
		id, err := h.emailClient.Send(h.emailFrom, cfg.Recipients, subject, body)
		if err != nil {
			return nil, fmt.Errorf("send email: %w", err)
		}
		// Use the provider-assigned ID when available.
		if id != "" {
			emailID = id
		}
	} else {
		h.log.Warn(ctx, "send_email: no email client configured, skipping delivery",
			"recipients", cfg.Recipients, "subject", subject)
	}

	h.log.Info(ctx, "send_email executed",
		"email_id", emailID,
		"recipients", cfg.Recipients,
		"subject", subject)

	return map[string]interface{}{
		"email_id":   emailID,
		"status":     "sent",
		"sent_at":    now.Format(time.RFC3339),
		"recipients": cfg.Recipients,
		"subject":    subject,
	}, nil
}
