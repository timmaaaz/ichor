package productdb

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
)

type product struct {
	ProductID            uuid.UUID `db:"id"`
	SKU                  string    `db:"sku"`
	BrandID              uuid.UUID `db:"brand_id"`
	ProductCategoryID    uuid.UUID `db:"category_id"`
	Name                 string    `db:"name"`
	Description          string    `db:"description"`
	ModelNumber          string    `db:"model_number"`
	UpcCode              string    `db:"upc_code"`
	Status               string    `db:"status"`
	IsActive             bool      `db:"is_active"`
	IsPerishable         bool      `db:"is_perishable"`
	HandlingInstructions string    `db:"handling_instructions"`
	UnitsPerCase         int       `db:"units_per_case"`
	TrackingType         string    `db:"tracking_type"`
	CreatedDate          time.Time `db:"created_date"`
	UpdatedDate          time.Time `db:"updated_date"`
}

func toDBProduct(bus productbus.Product) product {
	return product{
		ProductID:            bus.ProductID,
		SKU:                  bus.SKU,
		BrandID:              bus.BrandID,
		ProductCategoryID:    bus.ProductCategoryID,
		Name:                 bus.Name,
		Description:          bus.Description,
		ModelNumber:          bus.ModelNumber,
		UpcCode:              bus.UpcCode,
		Status:               bus.Status,
		IsActive:             bus.IsActive,
		IsPerishable:         bus.IsPerishable,
		HandlingInstructions: bus.HandlingInstructions,
		UnitsPerCase:         bus.UnitsPerCase,
		TrackingType:         bus.TrackingType,
		CreatedDate:          bus.CreatedDate,
		UpdatedDate:          bus.UpdatedDate,
	}
}

func toBusProduct(db product) productbus.Product {
	return productbus.Product{
		ProductID:            db.ProductID,
		SKU:                  db.SKU,
		BrandID:              db.BrandID,
		ProductCategoryID:    db.ProductCategoryID,
		Name:                 db.Name,
		Description:          db.Description,
		ModelNumber:          db.ModelNumber,
		UpcCode:              db.UpcCode,
		Status:               db.Status,
		IsActive:             db.IsActive,
		IsPerishable:         db.IsPerishable,
		HandlingInstructions: db.HandlingInstructions,
		UnitsPerCase:         db.UnitsPerCase,
		TrackingType:         db.TrackingType,
		CreatedDate:          db.CreatedDate.Local(),
		UpdatedDate:          db.UpdatedDate.Local(),
	}
}

func toBusProducts(DBs []product) []productbus.Product {
	bus := make([]productbus.Product, len(DBs))

	for i, db := range DBs {
		bus[i] = toBusProduct(db)
	}

	return bus
}
