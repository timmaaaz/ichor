package purchaseorderlineitemstatusbus

import "github.com/google/uuid"

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

// PurchaseOrderLineItemStatus represents information about a purchase order line item status.
type PurchaseOrderLineItemStatus struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	SortOrder   int       `json:"sort_order"`
}

// NewPurchaseOrderLineItemStatus contains information needed to create a new purchase order line item status.
type NewPurchaseOrderLineItemStatus struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

// UpdatePurchaseOrderLineItemStatus contains information needed to update a purchase order line item status.
type UpdatePurchaseOrderLineItemStatus struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	SortOrder   *int    `json:"sort_order,omitempty"`
}