package purchaseorderstatusbus

import "github.com/google/uuid"

// PurchaseOrderStatus represents information about a purchase order status.
type PurchaseOrderStatus struct {
	ID          uuid.UUID
	Name        string
	Description string
	SortOrder   int
}

// NewPurchaseOrderStatus contains information needed to create a new purchase order status.
type NewPurchaseOrderStatus struct {
	Name        string
	Description string
	SortOrder   int
}

// UpdatePurchaseOrderStatus contains information needed to update a purchase order status.
type UpdatePurchaseOrderStatus struct {
	Name        *string
	Description *string
	SortOrder   *int
}