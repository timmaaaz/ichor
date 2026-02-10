package workflow

import "github.com/google/uuid"

// AutomationRuleFilter provides query filtering for automation rules.
type AutomationRuleFilter struct {
	ID            *uuid.UUID
	Name          *string
	IsActive      *bool
	EntityID      *uuid.UUID
	EntityTypeID  *uuid.UUID
	TriggerTypeID *uuid.UUID
	CreatedBy     *uuid.UUID
}
