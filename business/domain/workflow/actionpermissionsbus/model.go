// Package actionpermissionsbus provides business logic for managing action permissions
// that control which roles can execute workflow actions manually.
package actionpermissionsbus

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ActionPermission represents permission for a role to execute a specific action type.
// JSON tags are included for workflow event serialization.
type ActionPermission struct {
	ID          uuid.UUID       `json:"id"`
	RoleID      uuid.UUID       `json:"role_id"`
	ActionType  string          `json:"action_type"`
	IsAllowed   bool            `json:"is_allowed"`
	Constraints json.RawMessage `json:"constraints"` // Stubbed for future constraint implementation
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// NewActionPermission contains information needed to create a new action permission.
type NewActionPermission struct {
	RoleID      uuid.UUID
	ActionType  string
	IsAllowed   bool
	Constraints json.RawMessage
}

// UpdateActionPermission contains information needed to update an action permission.
// Pointer semantics are used to distinguish "not provided" from "provided as zero value".
type UpdateActionPermission struct {
	IsAllowed   *bool
	Constraints *json.RawMessage
}
