package orderlineitemsdb

import "github.com/timmaaaz/ichor/business/sdk/workflow/protected"

// RegisterProtected declares the sales.order_line_items columns generic workflow writes
// must not set directly. picked_quantity, backordered_quantity, quantity, and the
// line_item_fulfillment_statuses_id are written transactionally by the picking ledger;
// a direct generic write corrupts the ledger invariants, so they are protect-only.
func RegisterProtected(reg *protected.Registry) {
	protected.CollectStructTags(reg, "sales.order_line_items", "", orderLineItem{})
}
