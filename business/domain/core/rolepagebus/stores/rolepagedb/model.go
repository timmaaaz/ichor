package rolepagedb

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/rolepagebus"
)

type rolePage struct {
	ID         uuid.UUID `db:"id"`
	RoleID     uuid.UUID `db:"role_id"`
	PageID     uuid.UUID `db:"page_id"`
	CanAccess  bool      `db:"can_access"`
	ShowInMenu bool      `db:"show_in_menu"`
}

func toDBRolePage(bus rolepagebus.RolePage) rolePage {
	return rolePage{
		ID:         bus.ID,
		RoleID:     bus.RoleID,
		PageID:     bus.PageID,
		CanAccess:  bus.CanAccess,
		ShowInMenu: bus.ShowInMenu,
	}
}

func toBusRolePage(db rolePage) rolepagebus.RolePage {
	return rolepagebus.RolePage{
		ID:         db.ID,
		RoleID:     db.RoleID,
		PageID:     db.PageID,
		CanAccess:  db.CanAccess,
		ShowInMenu: db.ShowInMenu,
	}
}

func toBusRolePages(dbs []rolePage) []rolepagebus.RolePage {
	rolePages := make([]rolepagebus.RolePage, len(dbs))
	for i, db := range dbs {
		rolePages[i] = toBusRolePage(db)
	}
	return rolePages
}
