package tableaccessbus

import "github.com/google/uuid"

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

type TableAccess struct {
	ID        uuid.UUID `json:"id"`
	RoleID    uuid.UUID `json:"role_id"`
	TableName string    `json:"table_name"`
	CanCreate bool      `json:"can_create"`
	CanRead   bool      `json:"can_read"`
	CanUpdate bool      `json:"can_update"`
	CanDelete bool      `json:"can_delete"`
}

type NewTableAccess struct {
	RoleID    uuid.UUID `json:"role_id"`
	TableName string    `json:"table_name"`
	CanCreate bool      `json:"can_create"`
	CanRead   bool      `json:"can_read"`
	CanUpdate bool      `json:"can_update"`
	CanDelete bool      `json:"can_delete"`
}

type UpdateTableAccess struct {
	RoleID    *uuid.UUID `json:"role_id,omitempty"`
	TableName *string    `json:"table_name,omitempty"`
	CanCreate *bool      `json:"can_create,omitempty"`
	CanRead   *bool      `json:"can_read,omitempty"`
	CanUpdate *bool      `json:"can_update,omitempty"`
	CanDelete *bool      `json:"can_delete,omitempty"`
}
