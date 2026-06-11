package purchaseorderlineitemdb

import "github.com/timmaaaz/ichor/business/sdk/workflow/protected"

// RegisterProtected declares the procurement.purchase_order_line_items columns generic
// workflow writes must not set directly. quantity_received/quantity_cancelled accumulate
// via purchaseorderlineitembus.ReceiveQuantity (accumulate-not-replace); a generic write
// would clobber the running total. No typed workflow action wraps it yet (FOLLOW_UP F3).
func RegisterProtected(reg *protected.Registry) {
	protected.CollectStructTags(reg, "procurement.purchase_order_line_items", "", purchaseOrderLineItem{})
}
