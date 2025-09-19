package productbus

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ProductID            uuid.UUID
	SKU                  string
	BrandID              uuid.UUID
	ProductCategoryID    uuid.UUID
	Name                 string
	Description          string
	ModelNumber          string
	UpcCode              string
	Status               string
	IsActive             bool
	IsPerishable         bool
	HandlingInstructions string
	UnitsPerCase         int
	CreatedDate          time.Time
	UpdatedDate          time.Time
}

type NewProduct struct {
	SKU                  string
	BrandID              uuid.UUID
	ProductCategoryID    uuid.UUID
	Name                 string
	Description          string
	ModelNumber          string
	UpcCode              string
	Status               string
	IsActive             bool
	IsPerishable         bool
	HandlingInstructions string
	UnitsPerCase         int
}

type UpdateProduct struct {
	SKU                  *string
	BrandID              *uuid.UUID
	ProductCategoryID    *uuid.UUID
	Name                 *string
	Description          *string
	ModelNumber          *string
	UpcCode              *string
	Status               *string
	IsActive             *bool
	IsPerishable         *bool
	HandlingInstructions *string
	UnitsPerCase         *int
}
