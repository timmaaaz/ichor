package purchaseorderlineitemstatusbus

import "github.com/google/uuid"

// PurchaseOrderLineItemStatus represents information about a purchase order line item status.
type PurchaseOrderLineItemStatus struct {
	ID          uuid.UUID
	Name        string
	Description string
	SortOrder   int
}

// NewPurchaseOrderLineItemStatus contains information needed to create a new purchase order line item status.
type NewPurchaseOrderLineItemStatus struct {
	Name        string
	Description string
	SortOrder   int
}

// UpdatePurchaseOrderLineItemStatus contains information needed to update a purchase order line item status.
type UpdatePurchaseOrderLineItemStatus struct {
	Name        *string
	Description *string
	SortOrder   *int
}