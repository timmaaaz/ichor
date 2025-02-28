package orgunitcolumnaccessbus

import "github.com/google/uuid"

type OrgUnitColumnAccess struct {
	ID                    uuid.UUID
	OrganizationalUnitID  uuid.UUID
	TableName             string
	ColumnName            string
	CanRead               bool
	CanUpdate             bool
	CanInheritPermissions bool
	CanRollupData         bool
}

type NewOrgUnitColumnAccess struct {
	OrganizationalUnitID  uuid.UUID
	TableName             string
	ColumnName            string
	CanRead               bool
	CanUpdate             bool
	CanInheritPermissions bool
	CanRollupData         bool
}

type UpdateOrgUnitColumnAccess struct {
	OrganizationalUnitID  *uuid.UUID
	TableName             *string
	ColumnName            *string
	CanRead               *bool
	CanUpdate             *bool
	CanInheritPermissions *bool
	CanRollupData         *bool
}
