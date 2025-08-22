package approval

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// SeekApprovalHandler handles seek_approval actions
type SeekApprovalHandler struct {
	log *logger.Logger
	db  *sqlx.DB
}

// NewSeekApprovalHandler creates a new seek approval handler
func NewSeekApprovalHandler(log *logger.Logger, db *sqlx.DB) *SeekApprovalHandler {
	return &SeekApprovalHandler{
		log: log,
		db:  db,
	}
}

func (h *SeekApprovalHandler) GetType() string {
	return "seek_approval"
}

func (h *SeekApprovalHandler) Validate(config json.RawMessage) error {
	var cfg struct {
		Approvers    []string `json:"approvers"`
		ApprovalType string   `json:"approval_type"`
	}

	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid configuration format: %w", err)
	}

	if len(cfg.Approvers) == 0 {
		return fmt.Errorf("approvers list is required and must not be empty")
	}

	validTypes := map[string]bool{"any": true, "all": true, "majority": true}
	if !validTypes[cfg.ApprovalType] {
		return fmt.Errorf("invalid approval_type, must be: any, all, or majority")
	}

	return nil
}

func (h *SeekApprovalHandler) Execute(ctx context.Context, config json.RawMessage, context workflow.ActionExecutionContext) (interface{}, error) {
	h.log.Info(ctx, "Executing seek_approval action",
		"entityID", context.EntityID,
		"ruleName", context.RuleName)

	result := map[string]interface{}{
		"approval_id":    fmt.Sprintf("approval_%d", time.Now().Unix()),
		"status":         "pending",
		"requested_at":   time.Now().Format(time.RFC3339),
		"reference_id":   context.EntityID,
		"reference_type": fmt.Sprintf("%s_%s", context.EntityName, context.EventType),
	}

	return result, nil
}
