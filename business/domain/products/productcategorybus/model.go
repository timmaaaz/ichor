package productcategorybus

import (
	"time"

	"github.com/google/uuid"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

type ProductCategory struct {
	ProductCategoryID uuid.UUID `json:"product_category_id"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	CreatedDate       time.Time `json:"created_date"`
	UpdatedDate       time.Time `json:"updated_date"`
}

type NewProductCategory struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateProductCategory struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}
