package workflow

import (
	"time"

	"github.com/google/uuid"
)

// ExecutionFilter provides query filtering for automation executions.
type ExecutionFilter struct {
	ID            *uuid.UUID       // Filter by specific execution ID
	RuleID        *uuid.UUID       // Filter by automation rule (maps to automation_rules_id column)
	Status        *ExecutionStatus // Filter by status (completed, failed, running, etc.)
	TriggerSource *string          // Filter by trigger source ("automation" or "manual")
	DateFrom      *time.Time       // Filter executions after this date
	DateTo        *time.Time       // Filter executions before this date
}
