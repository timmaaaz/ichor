package productuomdb

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/products/productuombus"
)

type productUOM struct {
	ID               uuid.UUID      `db:"id"`
	ProductID        uuid.UUID      `db:"product_id"`
	Name             string         `db:"name"`
	Abbreviation     sql.NullString `db:"abbreviation"`
	ConversionFactor float64        `db:"conversion_factor"`
	IsBase           bool           `db:"is_base"`
	IsApproximate    bool           `db:"is_approximate"`
	Notes            sql.NullString `db:"notes"`
	CreatedDate      time.Time      `db:"created_date"`
	UpdatedDate      time.Time      `db:"updated_date"`
}

func toDBProductUOM(bus productuombus.ProductUOM) productUOM {
	return productUOM{
		ID:               bus.ID,
		ProductID:        bus.ProductID,
		Name:             bus.Name,
		Abbreviation:     sql.NullString{String: bus.Abbreviation, Valid: bus.Abbreviation != ""},
		ConversionFactor: bus.ConversionFactor,
		IsBase:           bus.IsBase,
		IsApproximate:    bus.IsApproximate,
		Notes:            sql.NullString{String: bus.Notes, Valid: bus.Notes != ""},
		CreatedDate:      bus.CreatedDate.UTC(),
		UpdatedDate:      bus.UpdatedDate.UTC(),
	}
}

func toBusProductUOM(db productUOM) productuombus.ProductUOM {
	return productuombus.ProductUOM{
		ID:               db.ID,
		ProductID:        db.ProductID,
		Name:             db.Name,
		Abbreviation:     db.Abbreviation.String,
		ConversionFactor: db.ConversionFactor,
		IsBase:           db.IsBase,
		IsApproximate:    db.IsApproximate,
		Notes:            db.Notes.String,
		CreatedDate:      db.CreatedDate.Local(),
		UpdatedDate:      db.UpdatedDate.Local(),
	}
}

func toBusProductUOMs(dbs []productUOM) []productuombus.ProductUOM {
	bus := make([]productuombus.ProductUOM, len(dbs))
	for i, db := range dbs {
		bus[i] = toBusProductUOM(db)
	}
	return bus
}
