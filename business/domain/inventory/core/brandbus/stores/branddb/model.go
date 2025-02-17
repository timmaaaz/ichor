package branddb

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/brandbus"
)

type brand struct {
	ID             uuid.UUID `db:"brand_id"`
	Name           string    `db:"name"`
	ManufacturerID uuid.UUID `db:"manufacturer_id"`
	ContactInfo    uuid.UUID `db:"contact_info"`
	CreatedDate    time.Time `db:"created_date"`
	UpdatedDate    time.Time `db:"updated_date"`
}

func toDBBrand(bus brandbus.Brand) brand {
	return brand{
		ID:             bus.BrandID,
		Name:           bus.Name,
		ManufacturerID: bus.ManufacturerID,
		ContactInfo:    bus.ContactInfo,
		CreatedDate:    bus.CreatedDate,
		UpdatedDate:    bus.UpdatedDate,
	}
}

func toBusBrand(db brand) brandbus.Brand {
	return brandbus.Brand{
		BrandID:        db.ID,
		Name:           db.Name,
		ManufacturerID: db.ManufacturerID,
		ContactInfo:    db.ContactInfo,
		CreatedDate:    db.CreatedDate.Local(),
		UpdatedDate:    db.UpdatedDate.Local(),
	}
}

func toBusBrands(dbBrands []brand) []brandbus.Brand {
	busBrands := make([]brandbus.Brand, len(dbBrands))
	for i, dbBrand := range dbBrands {
		busBrands[i] = toBusBrand(dbBrand)
	}
	return busBrands
}
