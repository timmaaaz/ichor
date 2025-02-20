package organizationalunitdb

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/permissions/organizationalunitbus"
)

type organizationalUnit struct {
	ID                    uuid.UUID  `db:"organizational_unit_id"`
	ParentID              *uuid.UUID `db:"parent_id"` // Has to be pointer to be nullable in the db for root level org units
	Name                  string     `db:"name"`
	Level                 int        `db:"level"`
	Path                  string     `db:"path"`
	CanInheritPermissions bool       `db:"can_inherit_permissions"`
	CanRollupData         bool       `db:"can_rollup_data"`
	UnitType              string     `db:"unit_type"`
	IsActive              bool       `db:"is_active"`
}

func toDBOrganizationalUnit(bus organizationalunitbus.OrganizationalUnit) organizationalUnit {
	var parentID *uuid.UUID
	if bus.ParentID != uuid.Nil {
		parentID = &bus.ParentID
	}

	return organizationalUnit{
		ID:                    bus.ID,
		ParentID:              parentID,
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
	var parentID uuid.UUID
	if db.ParentID == nil {
		parentID = uuid.Nil
	} else {
		parentID = *db.ParentID
	}

	return organizationalunitbus.OrganizationalUnit{
		ID:                    db.ID,
		ParentID:              parentID,
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
