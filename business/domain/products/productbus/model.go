package productbus

import (
	"time"

	"github.com/google/uuid"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

type Product struct {
	ProductID            uuid.UUID `json:"product_id"`
	SKU                  string    `json:"sku"`
	BrandID              uuid.UUID `json:"brand_id"`
	ProductCategoryID    uuid.UUID `json:"product_category_id"`
	Name                 string    `json:"name"`
	Description          string    `json:"description"`
	ModelNumber          string    `json:"model_number"`
	UpcCode              string    `json:"upc_code"`
	Status               string    `json:"status"`
	IsActive             bool      `json:"is_active"`
	IsPerishable         bool      `json:"is_perishable"`
	HandlingInstructions string    `json:"handling_instructions"`
	UnitsPerCase         int       `json:"units_per_case"`
	TrackingType         string    `json:"tracking_type"`
	CreatedDate          time.Time `json:"created_date"`
	UpdatedDate          time.Time `json:"updated_date"`
}

type NewProduct struct {
	SKU                  string     `json:"sku"`
	BrandID              uuid.UUID  `json:"brand_id"`
	ProductCategoryID    uuid.UUID  `json:"product_category_id"`
	Name                 string     `json:"name"`
	Description          string     `json:"description"`
	ModelNumber          string     `json:"model_number"`
	UpcCode              string     `json:"upc_code"`
	Status               string     `json:"status"`
	IsActive             bool       `json:"is_active"`
	IsPerishable         bool       `json:"is_perishable"`
	HandlingInstructions string     `json:"handling_instructions"`
	UnitsPerCase         int        `json:"units_per_case"`
	TrackingType         string     `json:"tracking_type"`
	CreatedDate          *time.Time `json:"created_date,omitempty"` // Optional: if nil, uses time.Now(), otherwise explicit date for seeding
}

type UpdateProduct struct {
	SKU                  *string    `json:"sku,omitempty"`
	BrandID              *uuid.UUID `json:"brand_id,omitempty"`
	ProductCategoryID    *uuid.UUID `json:"product_category_id,omitempty"`
	Name                 *string    `json:"name,omitempty"`
	Description          *string    `json:"description,omitempty"`
	ModelNumber          *string    `json:"model_number,omitempty"`
	UpcCode              *string    `json:"upc_code,omitempty"`
	Status               *string    `json:"status,omitempty"`
	IsActive             *bool      `json:"is_active,omitempty"`
	IsPerishable         *bool      `json:"is_perishable,omitempty"`
	HandlingInstructions *string    `json:"handling_instructions,omitempty"`
	UnitsPerCase         *int       `json:"units_per_case,omitempty"`
	TrackingType         *string    `json:"tracking_type,omitempty"`
}
