package purchaseorderdb

import "github.com/timmaaaz/ichor/business/sdk/workflow/protected"

// RegisterProtected declares the procurement.purchase_orders columns that generic
// workflow writes must not set directly. The status field carries the approve/reject
// state-machine invariants (purchaseorderbus.Approve/Reject, ErrAlreadyApproved/Rejected)
// which raw SQL bypasses. Sourced from the `protected:"true"` db-model tags.
func RegisterProtected(reg *protected.Registry) {
	protected.CollectStructTags(reg, "procurement.purchase_orders", "approve_purchase_order / reject_purchase_order", purchaseOrder{})
}
