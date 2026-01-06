package permissionsbus

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/tableaccessbus"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

// UserPermissions represents all permissions for a specific user
type UserPermissions struct {
	UserID      uuid.UUID                             `json:"user_id"`
	Username    string                                `json:"username"`
	RoleNames   []string                              `json:"role_names"`
	Roles       []userrolebus.UserRole                `json:"roles"`
	TableAccess map[string]tableaccessbus.TableAccess `json:"table_access"`
}

// UserRole represents a role assigned to a user and its associated permissions
type UserRole struct {
	RoleID uuid.UUID                   `json:"role_id"`
	Name   string                      `json:"name"`
	Tables []tableaccessbus.TableAccess `json:"tables"`
}

// ConsolidatedPermissions represents simplified permissions aggregated across all roles
type ConsolidatedPermissions struct {
	UserID   uuid.UUID                             `json:"user_id"`
	Username string                                `json:"username"`
	Tables   map[string]tableaccessbus.TableAccess `json:"tables"`
	Roles    []string                              `json:"roles"`
}

// For efficient checking, a flattened structure can be useful
type PermissionKey struct {
	TableName string `json:"table_name"`
	Operation string `json:"operation"` // "create", "read", "update", "delete"
}

// PermissionCache is a flattened structure for quick permission lookups
type PermissionCache map[PermissionKey]bool
