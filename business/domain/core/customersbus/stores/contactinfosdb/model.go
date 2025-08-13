package customersdb

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/customersbus"
)

type customers struct {
	ID                uuid.UUID `db:"id"`
	Name              string    `db:"name"`
	ContactID         uuid.UUID `db:"contact_id"`
	DeliveryAddressID uuid.UUID `db:"delivery_address_id"`
	Notes             string    `db:"notes"`
	CreatedBy         uuid.UUID `db:"created_by"`
	UpdatedBy         uuid.UUID `db:"updated_by"`
	CreatedDate       time.Time `db:"created_date"`
	UpdatedDate       time.Time `db:"updated_date"`
}

func toDBCustomers(bus customersbus.Customers) customers {
	return customers{
		ID:                bus.ID,
		Name:              bus.Name,
		ContactID:         bus.ContactID,
		DeliveryAddressID: bus.DeliveryAddressID,
		Notes:             bus.Notes,
		CreatedBy:         bus.CreatedBy,
		UpdatedBy:         bus.UpdatedBy,
		CreatedDate:       bus.CreatedDate,
		UpdatedDate:       bus.UpdatedDate,
	}
}

func toBusCustomers(db customers) customersbus.Customers {
	return customersbus.Customers{
		ID:                db.ID,
		Name:              db.Name,
		ContactID:         db.ContactID,
		DeliveryAddressID: db.DeliveryAddressID,
		Notes:             db.Notes,
		CreatedBy:         db.CreatedBy,
		UpdatedBy:         db.UpdatedBy,
		CreatedDate:       db.CreatedDate,
		UpdatedDate:       db.UpdatedDate,
	}
}

func toBusCustomerss(dbs []customers) []customersbus.Customers {
	bus := make([]customersbus.Customers, len(dbs))
	for i, db := range dbs {
		bus[i] = toBusCustomers(db)
	}
	return bus
}
