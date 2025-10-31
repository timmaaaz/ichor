package pagedb

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/pagebus"
)

type dbPage struct {
	ID         uuid.UUID `db:"id"`
	Path       string    `db:"path"`
	Name       string    `db:"name"`
	Module     string    `db:"module"`
	Icon       string    `db:"icon"`
	SortOrder  int       `db:"sort_order"`
	IsActive   bool      `db:"is_active"`
	ShowInMenu bool      `db:"show_in_menu"`
}

func toDBPage(bus pagebus.Page) dbPage {
	return dbPage{
		ID:         bus.ID,
		Path:       bus.Path,
		Name:       bus.Name,
		Module:     bus.Module,
		Icon:       bus.Icon,
		SortOrder:  bus.SortOrder,
		IsActive:   bus.IsActive,
		ShowInMenu: bus.ShowInMenu,
	}
}

func toBusPage(db dbPage) pagebus.Page {
	return pagebus.Page{
		ID:         db.ID,
		Path:       db.Path,
		Name:       db.Name,
		Module:     db.Module,
		Icon:       db.Icon,
		SortOrder:  db.SortOrder,
		IsActive:   db.IsActive,
		ShowInMenu: db.ShowInMenu,
	}
}

func toBusPages(dbs []dbPage) []pagebus.Page {
	pages := make([]pagebus.Page, len(dbs))
	for i, db := range dbs {
		pages[i] = toBusPage(db)
	}
	return pages
}
