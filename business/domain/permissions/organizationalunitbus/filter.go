package organizationalunitbus

import "github.com/google/uuid"

// QueryFilter holds the available fields a query can be filtered on.
// We are using pointer semantics because the With API mutates the value.
type QueryFilter struct {
	ID                    *uuid.UUID
	ParentID              *uuid.UUID
	Name                  *string
	Level                 *int
	Path                  *string
	CanInheritPermissions *bool
	CanRollupData         *bool
	UnitType              *string
	IsActive              *bool
}
