package userorganizationdb

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/permissions/userorganizationbus"
)

type userOrganization struct {
	ID             uuid.UUID `db:"user_organization_id"`
	UserID         uuid.UUID `db:"user_id"`
	OrganizationID uuid.UUID `db:"organization_id"`
	RoleID         uuid.UUID `db:"role_id"`
	IsUnitManager  bool      `db:"is_unit_manager"`
	StartDate      time.Time `db:"start_date"`
	EndDate        time.Time `db:"end_date"`
	CreatedBy      uuid.UUID `db:"created_by"`
	CreatedAt      time.Time `db:"created_at"`
}

func toDBUserOrganization(bus userorganizationbus.UserOrganization) userOrganization {
	return userOrganization{
		ID:             bus.ID,
		UserID:         bus.UserID,
		OrganizationID: bus.OrganizationID,
		RoleID:         bus.RoleID,
		IsUnitManager:  bus.IsUnitManager,
		StartDate:      bus.StartDate,
		EndDate:        bus.EndDate,
		CreatedBy:      bus.CreatedBy,
		CreatedAt:      bus.CreatedAt,
	}
}

func toBusUserOrganization(db userOrganization) userorganizationbus.UserOrganization {
	return userorganizationbus.UserOrganization{
		ID:             db.ID,
		UserID:         db.UserID,
		OrganizationID: db.OrganizationID,
		RoleID:         db.RoleID,
		IsUnitManager:  db.IsUnitManager,
		StartDate:      db.StartDate,
		EndDate:        db.EndDate,
		CreatedBy:      db.CreatedBy,
		CreatedAt:      db.CreatedAt,
	}
}

func toBusUserOrganizations(db []userOrganization) []userorganizationbus.UserOrganization {
	bus := make([]userorganizationbus.UserOrganization, len(db))
	for i, v := range db {
		bus[i] = toBusUserOrganization(v)
	}
	return bus
}
