package customersdb

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/sales/customersbus"
)

type customers struct {
	ID                uuid.UUID `db:"id"`
	Name              string    `db:"name"`
	ContactID         uuid.UUID `db:"contact_id"`
	DeliveryAddressID uuid.UUID `db:"delivery_address_id"`
	Notes             sql.NullString `db:"notes"`
	CreatedBy         uuid.UUID `db:"created_by"`
	UpdatedBy         uuid.UUID `db:"updated_by"`
	CreatedDate       time.Time `db:"created_date"`
	UpdatedDate       time.Time `db:"updated_date"`
}

func toDBCustomers(bus customersbus.Customers) customers {
	db := customers{
		ID:                bus.ID,
		Name:              bus.Name,
		ContactID:         bus.ContactID,
		DeliveryAddressID: bus.DeliveryAddressID,
		CreatedBy:         bus.CreatedBy,
		UpdatedBy:         bus.UpdatedBy,
		CreatedDate:       bus.CreatedDate,
		UpdatedDate:       bus.UpdatedDate,
	}

	if bus.Notes != "" {
		db.Notes = sql.NullString{String: bus.Notes, Valid: true}
	}

	return db
}

func toBusCustomers(db customers) customersbus.Customers {
	bus := customersbus.Customers{
		ID:                db.ID,
		Name:              db.Name,
		ContactID:         db.ContactID,
		DeliveryAddressID: db.DeliveryAddressID,
		CreatedBy:         db.CreatedBy,
		UpdatedBy:         db.UpdatedBy,
		CreatedDate:       db.CreatedDate,
		UpdatedDate:       db.UpdatedDate,
	}

	if db.Notes.Valid {
		bus.Notes = db.Notes.String
	}

	return bus
}

func toBusCustomerss(dbs []customers) []customersbus.Customers {
	bus := make([]customersbus.Customers, len(dbs))
	for i, db := range dbs {
		bus[i] = toBusCustomers(db)
	}
	return bus
}
