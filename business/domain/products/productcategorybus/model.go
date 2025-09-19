package productcategorybus

import (
	"time"

	"github.com/google/uuid"
)

type ProductCategory struct {
	ProductCategoryID uuid.UUID
	Name              string
	Description       string
	CreatedDate       time.Time
	UpdatedDate       time.Time
}

type NewProductCategory struct {
	Name        string
	Description string
}

type UpdateProductCategory struct {
	Name        *string
	Description *string
}
