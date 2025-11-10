package pagecontentdb

import (
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/config/pagecontentbus"
	"github.com/timmaaaz/ichor/business/sdk/sqldb/nulltypes"
)

// dbPageContent represents the database model for page content with nullable fields
type dbPageContent struct {
	ID            uuid.UUID       `db:"id"`
	PageConfigID  uuid.UUID       `db:"page_config_id"`
	ContentType   string          `db:"content_type"`
	Label         string          `db:"label"`
	TableConfigID sql.NullString  `db:"table_config_id"`
	FormID        sql.NullString  `db:"form_id"`
	OrderIndex    int             `db:"order_index"`
	ParentID      sql.NullString  `db:"parent_id"`
	Layout        json.RawMessage `db:"layout"`
	IsVisible     bool            `db:"is_visible"`
	IsDefault     bool            `db:"is_default"`
}

// toDBPageContent converts a business PageContent to database model
func toDBPageContent(bus pagecontentbus.PageContent) dbPageContent {
	return dbPageContent{
		ID:            bus.ID,
		PageConfigID:  bus.PageConfigID,
		ContentType:   bus.ContentType,
		Label:         bus.Label,
		TableConfigID: nulltypes.ToNullableUUID(bus.TableConfigID),
		FormID:        nulltypes.ToNullableUUID(bus.FormID),
		OrderIndex:    bus.OrderIndex,
		ParentID:      nulltypes.ToNullableUUID(bus.ParentID),
		Layout:        bus.Layout,
		IsVisible:     bus.IsVisible,
		IsDefault:     bus.IsDefault,
	}
}

// toBusPageContent converts a database PageContent to business model
func toBusPageContent(db dbPageContent) pagecontentbus.PageContent {
	return pagecontentbus.PageContent{
		ID:            db.ID,
		PageConfigID:  db.PageConfigID,
		ContentType:   db.ContentType,
		Label:         db.Label,
		TableConfigID: nulltypes.FromNullableUUID(db.TableConfigID),
		FormID:        nulltypes.FromNullableUUID(db.FormID),
		OrderIndex:    db.OrderIndex,
		ParentID:      nulltypes.FromNullableUUID(db.ParentID),
		Layout:        db.Layout,
		IsVisible:     db.IsVisible,
		IsDefault:     db.IsDefault,
		Children:      []pagecontentbus.PageContent{}, // Initialized empty
	}
}

// toBusPageContents converts a slice of database PageContent to business models
func toBusPageContents(dbs []dbPageContent) []pagecontentbus.PageContent {
	bus := make([]pagecontentbus.PageContent, len(dbs))
	for i, db := range dbs {
		bus[i] = toBusPageContent(db)
	}
	return bus
}
