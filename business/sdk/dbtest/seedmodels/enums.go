package seedmodels

// =============================================================================
// STATUS NAME SLICES
// These are extracted from seedFrontend.go to give them findable homes.
// =============================================================================

// ApprovalStatusNames are the approval status strings seeded into the database.
var ApprovalStatusNames = []string{"SUCCESS", "ERROR", "WAITING", "REJECTED", "IN_PROGRESS"}

// FulfillmentStatusNames are the fulfillment status strings seeded into the database.
var FulfillmentStatusNames = []string{"SUCCESS", "ERROR", "WAITING", "REJECTED", "IN_PROGRESS"}

// StatusEntry represents a status with a name and description.
type StatusEntry struct {
	Name        string
	Description string
}

// StatusOrderedEntry represents a status with name, description and sort order.
type StatusOrderedEntry struct {
	Name        string
	Description string
	SortOrder   int
}

// OrderFulfillmentStatusData is the order fulfillment status seed data.
var OrderFulfillmentStatusData = []StatusEntry{
	{"PENDING", "Order is pending"},
	{"PROCESSING", "Order is being processed"},
	{"PICKING", "Order is being picked from warehouse"},
	{"PACKING", "Order is being packed"},
	{"READY_TO_SHIP", "Order is packed and ready for shipment"},
	{"SHIPPED", "Order has been shipped"},
	{"DELIVERED", "Order has been delivered"},
	{"CANCELLED", "Order has been cancelled"},
}

// PurchaseOrderStatusData is the purchase order status seed data.
var PurchaseOrderStatusData = []StatusOrderedEntry{
	{"DRAFT", "Purchase order is being prepared", 100},
	{"PENDING_APPROVAL", "Awaiting approval", 200},
	{"APPROVED", "Purchase order has been approved", 300},
	{"SENT", "Purchase order sent to supplier", 400},
	{"PARTIALLY_RECEIVED", "Some items have been received", 500},
	{"RECEIVED", "All items have been received", 600},
	{"CANCELLED", "Purchase order has been cancelled", 700},
	{"CLOSED", "Purchase order is closed", 800},
}

// PurchaseOrderLineItemStatusData is the purchase order line item status seed data.
var PurchaseOrderLineItemStatusData = []StatusOrderedEntry{
	{"PENDING", "Line item is pending", 100},
	{"ORDERED", "Line item has been ordered", 200},
	{"PARTIALLY_RECEIVED", "Some quantity has been received", 300},
	{"RECEIVED", "All quantity has been received", 400},
	{"BACKORDERED", "Line item is on backorder", 500},
	{"CANCELLED", "Line item has been cancelled", 600},
}
