package organizationalunitbus

import "github.com/google/uuid"

type OrganizationalUnit struct {
	ID                    uuid.UUID
	ParentID              uuid.UUID
	Name                  string
	Level                 int
	Path                  string
	CanInheritPermissions bool
	CanRollupData         bool
	UnitType              string
	IsActive              bool
}

type NewOrganizationalUnit struct {
	ParentID              uuid.UUID
	Name                  string
	Level                 int
	Path                  string
	CanInheritPermissions bool
	CanRollupData         bool
	UnitType              string
	IsActive              bool
}

type UpdateOrganizationalUnit struct {
	ParentID              *uuid.UUID
	Name                  *string
	Level                 *int
	Path                  *string
	CanInheritPermissions *bool
	CanRollupData         *bool
	UnitType              *string
	IsActive              *bool
}
