package permissionsbus

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/permissions/crossunitpermissionsbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/organizationalunitbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/orgunitcolumnaccessbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/tableaccessbus"
)

// UserPermissions represents all permissions for a specific user
type UserPermissions struct {
	UserID                uuid.UUID
	Username              string
	Roles                 UserRole
	Organizations         []organizationalunitbus.OrganizationalUnit
	TableAccess           []tableaccessbus.TableAccess
	CrossUnitPermissions  []crossunitpermissionsbus.CrossUnitPermission
	OrgUnitColumnAccesses []orgunitcolumnaccessbus.OrgUnitColumnAccess
}

// UserRole represents a role assigned to a user and its associated permissions
type UserRole struct {
	RoleID uuid.UUID
	Name   string
	Tables []TableAccess
}

// TableAccess represents permissions for a specific table
type TableAccess struct {
	TableName string
	CanCreate bool
	CanRead   bool
	CanUpdate bool
	CanDelete bool
}

// ConsolidatedPermissions represents simplified permissions aggregated across all roles
type ConsolidatedPermissions struct {
	UserID   uuid.UUID
	Username string
	Tables   map[string]TableAccess
	Roles    []string
}

// For efficient checking, a flattened structure can be useful
type PermissionKey struct {
	TableName string
	Operation string // "create", "read", "update", "delete"
}

// PermissionCache is a flattened structure for quick permission lookups
type PermissionCache map[PermissionKey]bool
