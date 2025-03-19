package supplierdb

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/supplier/supplierbus"
	"github.com/timmaaaz/ichor/business/domain/supplier/supplierbus/types"
)

type supplier struct {
	SupplierID   uuid.UUID      `db:"supplier_id"`
	ContactID    uuid.UUID      `db:"contact_id"`
	Name         string         `db:"name"`
	PaymentTerms string         `db:"payment_terms"`
	LeadTimeDays int            `db:"lead_time_days"`
	Rating       sql.NullString `db:"rating"`
	IsActive     bool           `db:"is_active"`
	CreatedDate  time.Time      `db:"created_date"`
	UpdatedDate  time.Time      `db:"updated_date"`
}

func toDBSupplier(bus supplierbus.Supplier) supplier {
	return supplier{
		SupplierID:   bus.SupplierID,
		ContactID:    bus.ContactID,
		Name:         bus.Name,
		PaymentTerms: bus.PaymentTerms,
		LeadTimeDays: bus.LeadTimeDays,
		Rating:       bus.Rating.DBValue(),
		IsActive:     bus.IsActive,
		CreatedDate:  bus.CreatedDate,
		UpdatedDate:  bus.UpdatedDate,
	}
}

func toBusSupplier(db supplier) (supplierbus.Supplier, error) {

	rating, err := types.ParseRoundedFloat(db.Rating.String)
	if err != nil {
		return supplierbus.Supplier{}, fmt.Errorf("tobussupplier: %v", err)
	}

	return supplierbus.Supplier{
		SupplierID:   db.SupplierID,
		ContactID:    db.ContactID,
		Name:         db.Name,
		PaymentTerms: db.PaymentTerms,
		LeadTimeDays: db.LeadTimeDays,
		Rating:       rating,
		IsActive:     db.IsActive,
		CreatedDate:  db.CreatedDate,
		UpdatedDate:  db.UpdatedDate,
	}, nil
}

func toBusSuppliers(db []supplier) ([]supplierbus.Supplier, error) {
	bus := make([]supplierbus.Supplier, len(db))

	for i, d := range db {
		s, err := toBusSupplier(d)
		if err != nil {
			return nil, fmt.Errorf("tobussuppliers: %v", err)
		}

		bus[i] = s
	}

	return bus, nil
}
