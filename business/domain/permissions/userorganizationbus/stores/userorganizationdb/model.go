package userorganizationdb

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/permissions/userorganizationbus"
)

type userOrganization struct {
	ID                   uuid.UUID    `db:"user_organization_id"`
	UserID               uuid.UUID    `db:"user_id"`
	OrganizationalUnitID uuid.UUID    `db:"organizational_unit_id"`
	RoleID               uuid.UUID    `db:"role_id"`
	IsUnitManager        bool         `db:"is_unit_manager"`
	StartDate            time.Time    `db:"start_date"`
	EndDate              sql.NullTime `db:"end_date"`
	CreatedBy            uuid.UUID    `db:"created_by"`
	CreatedAt            time.Time    `db:"created_at"`
}

func toDBUserOrganization(bus userorganizationbus.UserOrganization) userOrganization {
	endDate := sql.NullTime{
		Time:  bus.EndDate,
		Valid: !bus.EndDate.IsZero(),
	}

	return userOrganization{
		ID:                   bus.ID,
		UserID:               bus.UserID,
		OrganizationalUnitID: bus.OrganizationalUnitID,
		RoleID:               bus.RoleID,
		IsUnitManager:        bus.IsUnitManager,
		StartDate:            bus.StartDate,
		EndDate:              endDate,
		CreatedBy:            bus.CreatedBy,
		CreatedAt:            bus.CreatedAt,
	}
}

func toBusUserOrganization(db userOrganization) userorganizationbus.UserOrganization {
	var endDate time.Time
	if db.EndDate.Valid {
		endDate = db.EndDate.Time
	}

	return userorganizationbus.UserOrganization{
		ID:                   db.ID,
		UserID:               db.UserID,
		OrganizationalUnitID: db.OrganizationalUnitID,
		RoleID:               db.RoleID,
		IsUnitManager:        db.IsUnitManager,
		StartDate:            db.StartDate,
		EndDate:              endDate,
		CreatedBy:            db.CreatedBy,
		CreatedAt:            db.CreatedAt,
	}
}

func toBusUserOrganizations(db []userOrganization) []userorganizationbus.UserOrganization {
	bus := make([]userorganizationbus.UserOrganization, len(db))
	for i, v := range db {
		bus[i] = toBusUserOrganization(v)
	}
	return bus
}
