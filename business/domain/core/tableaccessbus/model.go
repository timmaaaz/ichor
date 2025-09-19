package tableaccessbus

import "github.com/google/uuid"

type TableAccess struct {
	ID        uuid.UUID
	RoleID    uuid.UUID
	TableName string
	CanCreate bool
	CanRead   bool
	CanUpdate bool
	CanDelete bool
}

type NewTableAccess struct {
	RoleID    uuid.UUID
	TableName string
	CanCreate bool
	CanRead   bool
	CanUpdate bool
	CanDelete bool
}

type UpdateTableAccess struct {
	RoleID    *uuid.UUID
	TableName *string
	CanCreate *bool
	CanRead   *bool
	CanUpdate *bool
	CanDelete *bool
}
