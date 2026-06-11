package ordersdb

import "github.com/timmaaaz/ichor/business/sdk/workflow/protected"

// RegisterProtected declares the sales.orders columns generic workflow writes must not
// set directly. order_fulfillment_status_id is recomputed by the picking flow; setting
// it directly desyncs order state, so it is protect-only (no generic write, no route).
func RegisterProtected(reg *protected.Registry) {
	protected.CollectStructTags(reg, "sales.orders", "", dbOrder{})
}
