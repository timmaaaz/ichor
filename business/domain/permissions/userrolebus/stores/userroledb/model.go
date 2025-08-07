package userroledb

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/permissions/userrolebus"
)

type userRole struct {
	ID     uuid.UUID `db:"id"`
	UserID uuid.UUID `db:"user_id"`
	RoleID uuid.UUID `db:"role_id"`
}

func toDBUserRole(bus userrolebus.UserRole) userRole {
	return userRole{
		ID:     bus.ID,
		UserID: bus.UserID,
		RoleID: bus.RoleID,
	}
}

func toBusUserRole(db userRole) userrolebus.UserRole {
	return userrolebus.UserRole{
		ID:     db.ID,
		UserID: db.UserID,
		RoleID: db.RoleID,
	}
}

func toBusUserRoles(dbs []userRole) []userrolebus.UserRole {
	userRoles := make([]userrolebus.UserRole, len(dbs))
	for i, db := range dbs {
		userRoles[i] = toBusUserRole(db)
	}
	return userRoles
}
