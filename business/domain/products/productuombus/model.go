package productuombus

import (
	"time"

	"github.com/google/uuid"
)

// ProductUOM represents a unit of measure for a product.
type ProductUOM struct {
	ID               uuid.UUID `json:"id"`
	ProductID        uuid.UUID `json:"product_id"`
	Name             string    `json:"name"`
	Abbreviation     string    `json:"abbreviation"`
	ConversionFactor float64   `json:"conversion_factor"`
	IsBase           bool      `json:"is_base"`
	IsApproximate    bool      `json:"is_approximate"`
	Notes            string    `json:"notes"`
	CreatedDate      time.Time `json:"created_date"`
	UpdatedDate      time.Time `json:"updated_date"`
}

// NewProductUOM contains the data required to create a new UOM.
type NewProductUOM struct {
	ProductID        uuid.UUID `json:"product_id"`
	Name             string    `json:"name"`
	Abbreviation     string    `json:"abbreviation"`
	ConversionFactor float64   `json:"conversion_factor"`
	IsBase           bool      `json:"is_base"`
	IsApproximate    bool      `json:"is_approximate"`
	Notes            string    `json:"notes"`
}

// UpdateProductUOM contains the fields that can be updated on a UOM.
// Pointer fields are optional — nil means do not change.
type UpdateProductUOM struct {
	Name             *string  `json:"name"`
	Abbreviation     *string  `json:"abbreviation"`
	ConversionFactor *float64 `json:"conversion_factor"`
	IsBase           *bool    `json:"is_base"`
	IsApproximate    *bool    `json:"is_approximate"`
	Notes            *string  `json:"notes"`
}
