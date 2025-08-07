package productcategorydb

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/core/productcategorybus"
)

type productCategory struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	CreatedDate time.Time `db:"created_date"`
	UpdatedDate time.Time `db:"updated_date"`
}

func toDBProductCategory(bus productcategorybus.ProductCategory) productCategory {
	return productCategory{
		ID:          bus.ProductCategoryID,
		Name:        bus.Name,
		Description: bus.Description,
		CreatedDate: bus.CreatedDate,
		UpdatedDate: bus.UpdatedDate,
	}
}

func toBusProductCategory(db productCategory) productcategorybus.ProductCategory {
	return productcategorybus.ProductCategory{
		ProductCategoryID: db.ID,
		Name:              db.Name,
		Description:       db.Description,
		CreatedDate:       db.CreatedDate.Local(),
		UpdatedDate:       db.UpdatedDate.Local(),
	}
}

func toBusProductCategories(dbProductCategories []productCategory) []productcategorybus.ProductCategory {
	busProductCategories := make([]productcategorybus.ProductCategory, len(dbProductCategories))
	for i, dbProductCategory := range dbProductCategories {
		busProductCategories[i] = toBusProductCategory(dbProductCategory)
	}
	return busProductCategories
}
