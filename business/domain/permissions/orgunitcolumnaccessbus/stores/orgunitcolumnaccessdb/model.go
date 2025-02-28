package orgunitcolumnaccessdb

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/permissions/orgunitcolumnaccessbus"
)

type orgUnitColumnAccess struct {
	ID                    uuid.UUID `db:"org_unit_column_access_id"`
	OrganizationalUnitID  uuid.UUID `db:"organizational_unit_id"`
	TableName             string    `db:"table_name"`
	ColumnName            string    `db:"column_name"`
	CanRead               bool      `db:"can_read"`
	CanUpdate             bool      `db:"can_update"`
	CanInheritPermissions bool      `db:"can_inherit_permissions"`
	CanRollupData         bool      `db:"can_rollup_data"`
}

func toDBOrgUnitColumnAccess(bus orgunitcolumnaccessbus.OrgUnitColumnAccess) orgUnitColumnAccess {
	return orgUnitColumnAccess{
		ID:                    bus.ID,
		OrganizationalUnitID:  bus.OrganizationalUnitID,
		TableName:             bus.TableName,
		ColumnName:            bus.ColumnName,
		CanRead:               bus.CanRead,
		CanUpdate:             bus.CanUpdate,
		CanInheritPermissions: bus.CanInheritPermissions,
		CanRollupData:         bus.CanRollupData,
	}
}

func toBusOrgUnitColumnAccess(db orgUnitColumnAccess) orgunitcolumnaccessbus.OrgUnitColumnAccess {
	return orgunitcolumnaccessbus.OrgUnitColumnAccess{
		ID:                    db.ID,
		OrganizationalUnitID:  db.OrganizationalUnitID,
		TableName:             db.TableName,
		ColumnName:            db.ColumnName,
		CanRead:               db.CanRead,
		CanUpdate:             db.CanUpdate,
		CanInheritPermissions: db.CanInheritPermissions,
		CanRollupData:         db.CanRollupData,
	}
}

func toBusOrgUnitColumnAccesses(db []orgUnitColumnAccess) []orgunitcolumnaccessbus.OrgUnitColumnAccess {
	bus := make([]orgunitcolumnaccessbus.OrgUnitColumnAccess, len(db))
	for i, v := range db {
		bus[i] = toBusOrgUnitColumnAccess(v)
	}
	return bus
}
