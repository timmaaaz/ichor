package userdb

import "github.com/timmaaaz/ichor/business/sdk/workflow/protected"

// RegisterProtected declares the core.users columns generic workflow writes must not set
// directly. The approval fields (user_approval_status_id, approved_by, date_approved)
// belong to the user approve/deny flow (userbus.Approve/Deny); no typed workflow action
// wraps them yet (FOLLOW_UP F3), so they are blocked-with-no-route until one is built.
func RegisterProtected(reg *protected.Registry) {
	protected.CollectStructTags(reg, "core.users", "", user{})
}
