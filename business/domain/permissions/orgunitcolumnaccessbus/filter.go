package orgunitcolumnaccessbus

import "github.com/google/uuid"

// QueryFilter holds the availabwle fields a query can be filtered on.
// We are using pointer semantics because the With API mutates the value.
type QueryFilter struct {
	ID                    *uuid.UUID
	OrganizationalUnitID  *uuid.UUID
	TableName             *string
	ColumnName            *string
	CanRead               *bool
	CanUpdate             *bool
	CanInheritPermissions *bool
	CanRollupData         *bool
}
