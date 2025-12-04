package lineitemfulfillmentstatusdb

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/sales/lineitemfulfillmentstatusbus"
)

type orderFulfillmentStatus struct {
	ID             uuid.UUID      `db:"id"`
	Name           string         `db:"name"`
	Description    string         `db:"description"`
	PrimaryColor   sql.NullString `db:"primary_color"`
	SecondaryColor sql.NullString `db:"secondary_color"`
	Icon           sql.NullString `db:"icon"`
}

func toBusLineItemFulfillmentStatus(db orderFulfillmentStatus) lineitemfulfillmentstatusbus.LineItemFulfillmentStatus {
	bus := lineitemfulfillmentstatusbus.LineItemFulfillmentStatus{
		ID:          db.ID,
		Name:        db.Name,
		Description: db.Description,
	}

	if db.PrimaryColor.Valid {
		bus.PrimaryColor = db.PrimaryColor.String
	}

	if db.SecondaryColor.Valid {
		bus.SecondaryColor = db.SecondaryColor.String
	}

	if db.Icon.Valid {
		bus.Icon = db.Icon.String
	}

	return bus
}

func toBusLineItemFulfillmentStatuses(dbs []orderFulfillmentStatus) []lineitemfulfillmentstatusbus.LineItemFulfillmentStatus {
	app := make([]lineitemfulfillmentstatusbus.LineItemFulfillmentStatus, len(dbs))
	for i, db := range dbs {
		app[i] = toBusLineItemFulfillmentStatus(db)
	}
	return app
}

func toDBLineItemFulfillmentStatus(bus lineitemfulfillmentstatusbus.LineItemFulfillmentStatus) orderFulfillmentStatus {
	db := orderFulfillmentStatus{
		ID:          bus.ID,
		Name:        bus.Name,
		Description: bus.Description,
	}

	if bus.PrimaryColor != "" {
		db.PrimaryColor = sql.NullString{String: bus.PrimaryColor, Valid: true}
	}

	if bus.SecondaryColor != "" {
		db.SecondaryColor = sql.NullString{String: bus.SecondaryColor, Valid: true}
	}

	if bus.Icon != "" {
		db.Icon = sql.NullString{String: bus.Icon, Valid: true}
	}

	return db
}
