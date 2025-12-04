package fulfillmentstatusdb

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/assets/fulfillmentstatusbus"
)

type fulfillmentStatus struct {
	ID             uuid.UUID      `db:"id"`
	Name           string         `db:"name"`
	IconID         uuid.UUID      `db:"icon_id"`
	PrimaryColor   sql.NullString `db:"primary_color"`
	SecondaryColor sql.NullString `db:"secondary_color"`
	Icon           sql.NullString `db:"icon"`
}

func toDBFulfillmentStatus(fs fulfillmentstatusbus.FulfillmentStatus) fulfillmentStatus {
	db := fulfillmentStatus{
		ID:     fs.ID,
		Name:   fs.Name,
		IconID: fs.IconID,
	}

	if fs.PrimaryColor != "" {
		db.PrimaryColor = sql.NullString{String: fs.PrimaryColor, Valid: true}
	}

	if fs.SecondaryColor != "" {
		db.SecondaryColor = sql.NullString{String: fs.SecondaryColor, Valid: true}
	}

	if fs.Icon != "" {
		db.Icon = sql.NullString{String: fs.Icon, Valid: true}
	}

	return db
}

func toBusFulfillmentStatus(dbFS fulfillmentStatus) fulfillmentstatusbus.FulfillmentStatus {
	bus := fulfillmentstatusbus.FulfillmentStatus{
		ID:     dbFS.ID,
		Name:   dbFS.Name,
		IconID: dbFS.IconID,
	}

	if dbFS.PrimaryColor.Valid {
		bus.PrimaryColor = dbFS.PrimaryColor.String
	}

	if dbFS.SecondaryColor.Valid {
		bus.SecondaryColor = dbFS.SecondaryColor.String
	}

	if dbFS.Icon.Valid {
		bus.Icon = dbFS.Icon.String
	}

	return bus
}

func toBusFulfillmentStatuses(dbFS []fulfillmentStatus) []fulfillmentstatusbus.FulfillmentStatus {
	fulfillmentStatuses := make([]fulfillmentstatusbus.FulfillmentStatus, len(dbFS))
	for i, fs := range dbFS {
		fulfillmentStatuses[i] = toBusFulfillmentStatus(fs)
	}

	return fulfillmentStatuses
}
