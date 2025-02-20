package organizationalunitdb

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/permissions/organizationalunitbus"
)

type organizationalUnit struct {
	ID                    uuid.UUID `db:"organizational_unit_id"`
	ParentID              uuid.UUID `db:"parent_id"`
	Name                  string    `db:"name"`
	Level                 int       `db:"level"`
	Path                  string    `db:"path"`
	CanInheritPermissions bool      `db:"can_inherit_permissions"`
	CanRollupData         bool      `db:"can_rollup_data"`
	UnitType              string    `db:"unit_type"`
	IsActive              bool      `db:"is_active"`
}

func toDBOrganizationalUnit(bus organizationalunitbus.OrganizationalUnit) organizationalUnit {
	return organizationalUnit{
		ID:                    bus.ID,
		ParentID:              bus.ParentID,
		Name:                  bus.Name,
		Level:                 bus.Level,
		Path:                  bus.Path,
		CanInheritPermissions: bus.CanInheritPermissions,
		CanRollupData:         bus.CanRollupData,
		UnitType:              bus.UnitType,
		IsActive:              bus.IsActive,
	}
}

func toBusOrganizationalUnit(db organizationalUnit) organizationalunitbus.OrganizationalUnit {
	return organizationalunitbus.OrganizationalUnit{
		ID:                    db.ID,
		ParentID:              db.ParentID,
		Name:                  db.Name,
		Level:                 db.Level,
		Path:                  db.Path,
		CanInheritPermissions: db.CanInheritPermissions,
		CanRollupData:         db.CanRollupData,
		UnitType:              db.UnitType,
		IsActive:              db.IsActive,
	}
}

func toBusOrganizationalUnits(dbs []organizationalUnit) []organizationalunitbus.OrganizationalUnit {
	organizationalUnits := make([]organizationalunitbus.OrganizationalUnit, len(dbs))
	for i, db := range dbs {
		organizationalUnits[i] = toBusOrganizationalUnit(db)
	}
	return organizationalUnits
}
