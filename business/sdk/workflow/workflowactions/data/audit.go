package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// AuditLogConfig represents configuration for audit log entries.
type AuditLogConfig struct {
	EntityName string         `json:"entity_name"`
	EntityID   string         `json:"entity_id"`
	Action     string         `json:"action"`
	Message    string         `json:"message"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

// AuditLogHandler handles log_audit_entry actions.
type AuditLogHandler struct {
	log          *logger.Logger
	db           *sqlx.DB
	templateProc *workflow.TemplateProcessor
}

// NewAuditLogHandler creates a new audit log handler.
func NewAuditLogHandler(log *logger.Logger, db *sqlx.DB) *AuditLogHandler {
	return &AuditLogHandler{
		log:          log,
		db:           db,
		templateProc: workflow.NewTemplateProcessor(workflow.DefaultTemplateProcessingOptions()),
	}
}

// GetType returns the action type.
func (h *AuditLogHandler) GetType() string {
	return "log_audit_entry"
}

// SupportsManualExecution returns true - audit logging can be run manually.
func (h *AuditLogHandler) SupportsManualExecution() bool {
	return true
}

// IsAsync returns false - audit logging completes inline.
func (h *AuditLogHandler) IsAsync() bool {
	return false
}

// GetDescription returns a human-readable description.
func (h *AuditLogHandler) GetDescription() string {
	return "Write an audit trail entry to the workflow audit log"
}

// Validate validates the audit log configuration.
func (h *AuditLogHandler) Validate(config json.RawMessage) error {
	var cfg AuditLogConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid configuration format: %w", err)
	}

	if cfg.EntityName == "" {
		return errors.New("entity_name is required")
	}

	if cfg.Action == "" {
		return errors.New("action is required")
	}

	if cfg.Message == "" {
		return errors.New("message is required")
	}

	return nil
}

// Execute writes an audit log entry.
func (h *AuditLogHandler) Execute(ctx context.Context, config json.RawMessage, execContext workflow.ActionExecutionContext) (any, error) {
	var cfg AuditLogConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	templateContext := buildTemplateContext(execContext)

	// Process templates
	resolvedMessage := processTemplateValue(h.templateProc, ctx, h.log, cfg.Message, templateContext)
	resolvedEntityID := processTemplateValue(h.templateProc, ctx, h.log, cfg.EntityID, templateContext)

	// Process metadata templates
	var metadataJSON []byte
	if cfg.Metadata != nil {
		processedMeta := make(map[string]any, len(cfg.Metadata))
		for k, v := range cfg.Metadata {
			processedMeta[k] = processTemplateValue(h.templateProc, ctx, h.log, v, templateContext)
		}
		var err error
		metadataJSON, err = json.Marshal(processedMeta)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	// Parse entity_id to UUID
	var entityID uuid.UUID
	if str, ok := resolvedEntityID.(string); ok && str != "" {
		var err error
		entityID, err = uuid.Parse(str)
		if err != nil {
			// If entity_id isn't a valid UUID, use the execution context entity_id
			entityID = execContext.EntityID
		}
	} else {
		entityID = execContext.EntityID
	}

	// Build insert
	auditID := uuid.New()
	now := time.Now().UTC()

	query := `INSERT INTO workflow.audit_log
		(id, entity_name, entity_id, action, message, metadata, rule_id, execution_id, user_id, created_date)
		VALUES (:id, :entity_name, :entity_id, :action, :message, :metadata, :rule_id, :execution_id, :user_id, :created_date)`

	args := map[string]any{
		"id":           auditID,
		"entity_name":  cfg.EntityName,
		"entity_id":    entityID,
		"action":       cfg.Action,
		"message":      resolvedMessage,
		"metadata":     metadataJSON,
		"rule_id":      execContext.RuleID,
		"execution_id": execContext.ExecutionID,
		"user_id":      execContext.UserID,
		"created_date": now,
	}

	if _, err := sqldb.NamedExecContextWithCount(ctx, h.log, h.db, query, args); err != nil {
		return nil, fmt.Errorf("audit log insert failed: %w", err)
	}

	result := map[string]any{
		"audit_id": auditID,
		"status":   "logged",
	}

	h.log.Info(ctx, "log_audit_entry completed",
		"audit_id", auditID,
		"entity_name", cfg.EntityName,
		"action", cfg.Action)

	return result, nil
}
