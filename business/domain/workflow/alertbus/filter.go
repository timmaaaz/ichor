package alertbus

import (
	"time"

	"github.com/google/uuid"
)

// QueryFilter holds the available fields a query can be filtered on.
// We are using pointer semantics because the With API mutates the value.
type QueryFilter struct {
	ID               *uuid.UUID
	AlertType        *string
	Severities       []string // Supports multiple severity values (e.g., "low,medium,high")
	Status           *string
	SourceEntityName *string
	SourceEntityID   *uuid.UUID
	SourceRuleID     *uuid.UUID
	CreatedAfter     *time.Time
	CreatedBefore    *time.Time
}
