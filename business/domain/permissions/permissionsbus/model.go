package permissionsbus

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/permissions/tableaccessbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/userrolebus"
)

// UserPermissions represents all permissions for a specific user
type UserPermissions struct {
	UserID      uuid.UUID
	Username    string
	RoleNames   []string
	Roles       []userrolebus.UserRole
	TableAccess map[string]tableaccessbus.TableAccess
}

// UserRole represents a role assigned to a user and its associated permissions
type UserRole struct {
	RoleID uuid.UUID
	Name   string
	Tables []tableaccessbus.TableAccess
}

// ConsolidatedPermissions represents simplified permissions aggregated across all roles
type ConsolidatedPermissions struct {
	UserID   uuid.UUID
	Username string
	Tables   map[string]tableaccessbus.TableAccess
	Roles    []string
}

// For efficient checking, a flattened structure can be useful
type PermissionKey struct {
	TableName string
	Operation string // "create", "read", "update", "delete"
}

// PermissionCache is a flattened structure for quick permission lookups
type PermissionCache map[PermissionKey]bool
