package crossunitpermissionsdb

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/permissions/crossunitpermissionsbus"
)

type crossUnitPermission struct {
	ID           uuid.UUID `db:"cross_unit_permission_id"`
	SourceUnitID uuid.UUID `db:"source_unit_id"`
	TargetUnitID uuid.UUID `db:"target_unit_id"`
	CanRead      bool      `db:"can_read"`
	CanUpdate    bool      `db:"can_update"`
	GrantedBy    uuid.UUID `db:"granted_by"`
	ValidFrom    time.Time `db:"valid_from"`
	ValidUntil   time.Time `db:"valid_until"`
	Reason       string    `db:"reason"`
}

func toDBCrossUnitPermission(bus crossunitpermissionsbus.CrossUnitPermission) crossUnitPermission {
	return crossUnitPermission{
		ID:           bus.ID,
		SourceUnitID: bus.SourceUnitID,
		TargetUnitID: bus.TargetUnitID,
		CanRead:      bus.CanRead,
		CanUpdate:    bus.CanUpdate,
		GrantedBy:    bus.GrantedBy,
		ValidFrom:    bus.ValidFrom.UTC(),
		ValidUntil:   bus.ValidUntil.UTC(),
		Reason:       bus.Reason,
	}
}

func toBusCrossUnitPermission(db crossUnitPermission) crossunitpermissionsbus.CrossUnitPermission {
	return crossunitpermissionsbus.CrossUnitPermission{
		ID:           db.ID,
		SourceUnitID: db.SourceUnitID,
		TargetUnitID: db.TargetUnitID,
		CanRead:      db.CanRead,
		CanUpdate:    db.CanUpdate,
		GrantedBy:    db.GrantedBy,
		ValidFrom:    db.ValidFrom.In(time.Local),
		ValidUntil:   db.ValidUntil.In(time.Local),
		Reason:       db.Reason,
	}
}

func toBusCrossUnitPermissions(db []crossUnitPermission) []crossunitpermissionsbus.CrossUnitPermission {
	bus := make([]crossunitpermissionsbus.CrossUnitPermission, len(db))
	for i, v := range db {
		bus[i] = toBusCrossUnitPermission(v)
	}
	return bus
}
