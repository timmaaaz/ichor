package dbtest

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// Test_Seed_OverOrderGraph asserts the default over-order remediation branch is
// wired into Rule 5 ("Line Item Created - Granular Inventory Pipeline") by the
// platform seed. It proves:
//   - the reserve_inventory -> success_alert edge is now gated on the "success"
//     output port (the old unconditional/nil-port edge is gone),
//   - reserve_inventory --[insufficient_stock]--> an over_order create_alert,
//   - that over_order alert --[success]--> a seek_approval (approval hold),
//   - reserve_inventory --[failure]--> a (critical) alert.
func Test_Seed_OverOrderGraph(t *testing.T) {
	t.Parallel()

	db := NewDatabase(t, "Test_Seed_OverOrderGraph")

	if err := InsertSeedDataWithDB(db.Log, db.DB); err != nil {
		t.Fatalf("InsertSeedDataWithDB: %v", err)
	}

	ctx := context.Background()

	// Locate Rule 5 by its real seeded name.
	const ruleName = "Line Item Created - Granular Inventory Pipeline"
	rules, err := db.BusDomain.Workflow.QueryActiveRules(ctx)
	if err != nil {
		t.Fatalf("QueryActiveRules: %v", err)
	}
	var rule workflow.AutomationRule
	for _, r := range rules {
		if r.Name == ruleName {
			rule = r
			break
		}
	}
	if rule.ID.String() == "00000000-0000-0000-0000-000000000000" {
		t.Fatalf("rule %q not found among %d active rules", ruleName, len(rules))
	}

	actions, err := db.BusDomain.Workflow.QueryActionsByRule(ctx, rule.ID)
	if err != nil {
		t.Fatalf("QueryActionsByRule: %v", err)
	}
	actionByID := make(map[string]workflow.RuleAction, len(actions))
	var reserveID string
	for _, a := range actions {
		actionByID[a.ID.String()] = a
		if a.Name == "Reserve Inventory" {
			reserveID = a.ID.String()
		}
	}
	if reserveID == "" {
		t.Fatal("Reserve Inventory action not found in Rule 5")
	}

	edges, err := db.BusDomain.Workflow.QueryEdgesByRuleID(ctx, rule.ID)
	if err != nil {
		t.Fatalf("QueryEdgesByRuleID: %v", err)
	}

	// Collect the outgoing edges from reserve_inventory by their output port.
	var reserveHasNilPort bool
	reservePortTarget := map[string]string{} // port -> target action id
	for _, e := range edges {
		if e.SourceActionID == nil || e.SourceActionID.String() != reserveID {
			continue
		}
		if e.SourceOutput == nil {
			reserveHasNilPort = true
			continue
		}
		reservePortTarget[*e.SourceOutput] = e.TargetActionID.String()
	}

	// (a) the old unconditional (nil-port) reserve edge must be gone.
	if reserveHasNilPort {
		t.Error("reserve_inventory still has an edge with a nil SourceOutput (old unconditional success edge not gated on the success port)")
	}

	// reserve must now route success, insufficient_stock, and failure ports.
	for _, port := range []string{"success", "insufficient_stock", "failure"} {
		if _, ok := reservePortTarget[port]; !ok {
			t.Errorf("reserve_inventory missing %q output-port edge", port)
		}
	}

	// (b) reserve --[insufficient_stock]--> an over_order create_alert.
	ooTargetID := reservePortTarget["insufficient_stock"]
	ooAlert, ok := actionByID[ooTargetID]
	if !ok {
		t.Fatalf("insufficient_stock edge target %s is not a Rule 5 action", ooTargetID)
	}
	if at := alertTypeOf(t, ooAlert.ActionConfig); at != "over_order" {
		t.Fatalf("insufficient_stock target alert_type = %q, want over_order", at)
	}

	// (c) the over_order alert --[success]--> a seek_approval (approvers present).
	var approvalTargetID string
	for _, e := range edges {
		if e.SourceActionID == nil || e.SourceActionID.String() != ooTargetID {
			continue
		}
		if e.SourceOutput != nil && *e.SourceOutput == "success" {
			approvalTargetID = e.TargetActionID.String()
		}
	}
	if approvalTargetID == "" {
		t.Fatal("over_order alert has no success-port edge into an approval hold")
	}
	approvalAction, ok := actionByID[approvalTargetID]
	if !ok {
		t.Fatalf("approval edge target %s is not a Rule 5 action", approvalTargetID)
	}
	if appr := approversOf(t, approvalAction.ActionConfig); len(appr) == 0 {
		t.Fatalf("approval-hold action %q has no approvers — not a seek_approval node", approvalAction.Name)
	}

	// (d) reserve --[failure]--> a critical alert (target exists as a Rule 5 action).
	failTargetID := reservePortTarget["failure"]
	if _, ok := actionByID[failTargetID]; !ok {
		t.Fatalf("failure edge target %s is not a Rule 5 action", failTargetID)
	}
}

func alertTypeOf(t *testing.T, cfg json.RawMessage) string {
	t.Helper()
	var m struct {
		AlertType string `json:"alert_type"`
	}
	if err := json.Unmarshal(cfg, &m); err != nil {
		t.Fatalf("parse alert config: %v", err)
	}
	return m.AlertType
}

func approversOf(t *testing.T, cfg json.RawMessage) []string {
	t.Helper()
	var m struct {
		Approvers []string `json:"approvers"`
	}
	if err := json.Unmarshal(cfg, &m); err != nil {
		t.Fatalf("parse approval config: %v", err)
	}
	return m.Approvers
}
