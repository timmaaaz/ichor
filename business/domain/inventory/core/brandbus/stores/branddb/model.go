package branddb

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/brandbus"
)

type brand struct {
	ID            uuid.UUID `db:"id"`
	Name          string    `db:"name"`
	ContactInfoID uuid.UUID `db:"contact_info_id"`
	CreatedDate   time.Time `db:"created_date"`
	UpdatedDate   time.Time `db:"updated_date"`
}

func toDBBrand(bus brandbus.Brand) brand {
	return brand{
		ID:            bus.BrandID,
		Name:          bus.Name,
		ContactInfoID: bus.ContactInfoID,
		CreatedDate:   bus.CreatedDate,
		UpdatedDate:   bus.UpdatedDate,
	}
}

func toBusBrand(db brand) brandbus.Brand {
	return brandbus.Brand{
		BrandID:       db.ID,
		Name:          db.Name,
		ContactInfoID: db.ContactInfoID,
		CreatedDate:   db.CreatedDate.Local(),
		UpdatedDate:   db.UpdatedDate.Local(),
	}
}

func toBusBrands(dbBrands []brand) []brandbus.Brand {
	busBrands := make([]brandbus.Brand, len(dbBrands))
	for i, dbBrand := range dbBrands {
		busBrands[i] = toBusBrand(dbBrand)
	}
	return busBrands
}
