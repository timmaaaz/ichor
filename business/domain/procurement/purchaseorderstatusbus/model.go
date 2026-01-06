package purchaseorderstatusbus

import "github.com/google/uuid"

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

// PurchaseOrderStatus represents information about a purchase order status.
type PurchaseOrderStatus struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	SortOrder   int       `json:"sort_order"`
}

// NewPurchaseOrderStatus contains information needed to create a new purchase order status.
type NewPurchaseOrderStatus struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

// UpdatePurchaseOrderStatus contains information needed to update a purchase order status.
type UpdatePurchaseOrderStatus struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	SortOrder   *int    `json:"sort_order,omitempty"`
}