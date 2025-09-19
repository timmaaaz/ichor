package roledb

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/rolebus"
)

type role struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
}

func toDBRole(bus rolebus.Role) role {
	return role{
		ID:          bus.ID,
		Name:        bus.Name,
		Description: bus.Description,
	}
}

func toBusRole(db role) rolebus.Role {
	return rolebus.Role{
		ID:          db.ID,
		Name:        db.Name,
		Description: db.Description,
	}
}

func toBusRoles(dbs []role) []rolebus.Role {
	roles := make([]rolebus.Role, len(dbs))
	for i, db := range dbs {
		roles[i] = toBusRole(db)
	}
	return roles
}
