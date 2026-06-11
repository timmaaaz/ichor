package transferorderdb

import "github.com/timmaaaz/ichor/business/sdk/workflow/protected"

// RegisterProtected declares the inventory.transfer_orders columns generic workflow
// writes must not set directly. quantity/claimed_by/completed_by belong to the claim/
// execute state machine (transferorderbus.Claim/Execute) — no typed workflow action
// wraps them yet (FOLLOW_UP F3), so they are blocked-with-no-route until one is built.
// (status is already protected via the approve/reject_transfer_order manifest claims.)
func RegisterProtected(reg *protected.Registry) {
	protected.CollectStructTags(reg, "inventory.transfer_orders", "", transferOrder{})
}
