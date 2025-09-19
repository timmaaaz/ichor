package tableaccessdb

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/tableaccessbus"
)

type tableAccess struct {
	ID        uuid.UUID `db:"id"`
	RoleID    uuid.UUID `db:"role_id"`
	TableName string    `db:"table_name"`
	CanCreate bool      `db:"can_create"`
	CanRead   bool      `db:"can_read"`
	CanUpdate bool      `db:"can_update"`
	CanDelete bool      `db:"can_delete"`
}

func toDBTableAccess(bus tableaccessbus.TableAccess) tableAccess {
	return tableAccess{
		ID:        bus.ID,
		RoleID:    bus.RoleID,
		TableName: bus.TableName,
		CanCreate: bus.CanCreate,
		CanRead:   bus.CanRead,
		CanUpdate: bus.CanUpdate,
		CanDelete: bus.CanDelete,
	}
}

func toBusTableAccess(db tableAccess) tableaccessbus.TableAccess {
	return tableaccessbus.TableAccess{
		ID:        db.ID,
		RoleID:    db.RoleID,
		TableName: db.TableName,
		CanCreate: db.CanCreate,
		CanRead:   db.CanRead,
		CanUpdate: db.CanUpdate,
		CanDelete: db.CanDelete,
	}
}

func toBusTableAccesses(dbs []tableAccess) []tableaccessbus.TableAccess {
	tableAccesses := make([]tableaccessbus.TableAccess, len(dbs))
	for i, db := range dbs {
		tableAccesses[i] = toBusTableAccess(db)
	}
	return tableAccesses
}
