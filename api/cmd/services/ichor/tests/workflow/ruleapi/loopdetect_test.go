package rule_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/domain/http/workflow/ruleapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions"
)

// =============================================================================
// P2 — Static cascade-loop detector
// =============================================================================

// Test_DetectCascadeLoops exercises the DB-backed detector (the full assembly: load active
// rules + their action manifests, overlay the in-flight candidate, classify) against real
// seeded rules. The pure tier logic is covered exhaustively in the package unit tests
// (business/sdk/workflow/cascade_detect_test.go); this proves the wiring + value-awareness
// against actual Postgres data.
func Test_DetectCascadeLoops(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_DetectCascadeLoops")

	sd, err := insertCascadeSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("seeding: %s", err)
	}
	if len(sd.Entities) < 7 {
		t.Skipf("need >= 7 workflow entities, have %d", len(sd.Entities))
	}

	ctx := context.Background()
	bus := test.DB.BusDomain.Workflow
	createdBy := sd.Users[0].ID

	reg := workflow.NewActionRegistry()
	workflowactions.RegisterCoreActions(reg, test.DB.Log, test.DB.DB)

	var onUpdate workflow.TriggerType
	for _, tt := range sd.TriggerTypes {
		if tt.Name == "on_update" {
			onUpdate = tt
		}
	}
	if onUpdate.ID == uuid.Nil {
		t.Fatal("on_update trigger type not seeded")
	}
	entityTypeID := sd.EntityTypes[0].ID

	tmpl, err := bus.CreateActionTemplate(ctx, workflow.NewActionTemplate{
		Name:          "loopdetect update_field",
		Description:   "update_field template for loop-detection tests",
		ActionType:    "update_field",
		DefaultConfig: createUpdateFieldActionConfig("x.y", "status", "z"),
		CreatedBy:     createdBy,
	})
	if err != nil {
		t.Fatalf("create template: %s", err)
	}

	full := func(e workflow.Entity) string { return e.SchemaName + "." + e.Name }

	// seedActiveUFRule creates an ACTIVE rule with one update_field action (no edges needed —
	// the inter-rule detector reads action manifests, not the intra-rule action DAG).
	seedActiveUFRule := func(name string, triggerEntity workflow.Entity, conds *json.RawMessage, target workflow.Entity, value string) {
		t.Helper()
		rule, err := bus.CreateRule(ctx, workflow.NewAutomationRule{
			Name: name, Description: name, EntityID: triggerEntity.ID, EntityTypeID: entityTypeID,
			TriggerTypeID: onUpdate.ID, TriggerConditions: conds, IsActive: true, CreatedBy: createdBy,
		})
		if err != nil {
			t.Fatalf("seed rule %q: %s", name, err)
		}
		if _, err := bus.CreateRuleAction(ctx, workflow.NewRuleAction{
			AutomationRuleID: rule.ID, Name: name + " action",
			ActionConfig: createUpdateFieldActionConfig(full(target), "status", value),
			IsActive:     true, TemplateID: &tmpl.ID,
		}); err != nil {
			t.Fatalf("seed action %q: %s", name, err)
		}
	}

	// candidate is the in-flight rule under evaluation (an update_field setter).
	candidate := func(active bool, triggerEntity workflow.Entity, conds *json.RawMessage, target workflow.Entity, value string) workflow.CandidateRule {
		return workflow.CandidateRule{
			RuleID: uuid.New(), Name: "candidate", IsActive: active,
			EntityID: triggerEntity.ID, TriggerTypeID: onUpdate.ID, TriggerConditions: conds,
			Actions: []workflow.CandidateAction{{
				ActionType: "update_field",
				Config:     createUpdateFieldActionConfig(full(target), "status", value),
			}},
		}
	}

	e := sd.Entities

	// Seed the active counterparts used by the convergent + cross-loop scenarios.
	seedActiveUFRule("convergent-B", e[3], changedTo("status", "PROCESSING"), e[2], "ALLOCATED")
	seedActiveUFRule("crossloop-B", e[5], nil /* auto-match */, e[4], "z")

	t.Run("self-loop-auto-match-blocks", func(t *testing.T) {
		// auto-match on e1 writing e1 → always re-arms → provable loop.
		res, err := bus.DetectCascadeLoops(ctx, reg, candidate(true, e[1], nil, e[1], "x"))
		if err != nil {
			t.Fatalf("detect: %s", err)
		}
		if !res.HasErrors() {
			t.Fatalf("auto-match self-loop must ERROR, got %+v", res)
		}
	})

	t.Run("convergent-sync-allowed", func(t *testing.T) {
		// A: e2 changed_to ALLOCATED → writes e3=PROCESSING; B (seeded): e3 changed_to
		// PROCESSING → writes e2=ALLOCATED. The changed_to latches converge → not a loop.
		res, err := bus.DetectCascadeLoops(ctx, reg, candidate(true, e[2], changedTo("status", "ALLOCATED"), e[3], "PROCESSING"))
		if err != nil {
			t.Fatalf("detect: %s", err)
		}
		if res.HasErrors() {
			t.Fatalf("convergent sync must NOT be blocked, got errors: %+v", res.Errors)
		}
		if len(res.Info) == 0 {
			t.Fatalf("convergent sync should surface INFO, got %+v", res)
		}
	})

	t.Run("cross-rule-auto-match-loop-blocks", func(t *testing.T) {
		// candidate: auto-match e4 → writes e5; B (seeded): auto-match e5 → writes e4.
		res, err := bus.DetectCascadeLoops(ctx, reg, candidate(true, e[4], nil, e[5], "z"))
		if err != nil {
			t.Fatalf("detect: %s", err)
		}
		if !res.HasErrors() {
			t.Fatalf("cross-rule auto-match loop must ERROR, got %+v", res)
		}
	})

	t.Run("forward-only-info-no-error", func(t *testing.T) {
		// candidate writes e6.status='approved'; seed a consumer on e6 that writes nothing cyclic.
		consumer, err := bus.CreateRule(ctx, workflow.NewAutomationRule{
			Name: "fwd-consumer", Description: "fwd", EntityID: e[6].ID, EntityTypeID: entityTypeID,
			TriggerTypeID: onUpdate.ID, TriggerConditions: changedTo("status", "approved"), IsActive: true, CreatedBy: createdBy,
		})
		if err != nil {
			t.Fatalf("seed consumer: %s", err)
		}
		_ = consumer // no actions → no outgoing edge → cannot loop

		res, err := bus.DetectCascadeLoops(ctx, reg, candidate(true, e[1], nil, e[6], "approved"))
		if err != nil {
			t.Fatalf("detect: %s", err)
		}
		if res.HasErrors() {
			t.Fatalf("forward-only must not ERROR, got %+v", res.Errors)
		}
		if len(res.Info) == 0 {
			t.Fatalf("forward-only should surface an INFO datapoint, got %+v", res)
		}
	})

	t.Run("inactive-candidate-skipped", func(t *testing.T) {
		res, err := bus.DetectCascadeLoops(ctx, reg, candidate(false, e[1], nil, e[1], "x"))
		if err != nil {
			t.Fatalf("detect: %s", err)
		}
		if res.HasErrors() || len(res.Warnings) != 0 || len(res.Info) != 0 {
			t.Fatalf("inactive candidate must be a no-op, got %+v", res)
		}
	})
}

// changedTo builds a single-condition trigger (status changed_to value) as JSON.
func changedTo(field, value string) *json.RawMessage {
	tc := workflow.TriggerConditions{
		FieldConditions: []workflow.FieldCondition{{FieldName: field, Operator: workflow.OperatorChangedTo, Value: value}},
	}
	b, _ := json.Marshal(tc)
	raw := json.RawMessage(b)
	return &raw
}

// Test_CascadeLoopEnforcement proves the HTTP activation hook blocks a rule that would close a
// provable loop (the seeded auto-match update_field self-loop) and allows a benign rule.
func Test_CascadeLoopEnforcement(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_CascadeLoopEnforcement")

	sd, err := insertCascadeSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("seeding: %s", err)
	}

	test.Run(t, activateSelfLoopBlocked(sd), "activate-self-loop-blocked-400")
	test.Run(t, activateBenignAllowed(sd), "activate-benign-allowed-200")
}

// activateSelfLoopBlocked deactivates the seeded auto-match update_field self-loop, then proves
// re-activating it is rejected by the static detector.
func activateSelfLoopBlocked(sd CascadeSeedData) []apitest.Table {
	ruleID := sd.SelfTriggerRule.ID
	return []apitest.Table{
		{
			Name:       "deactivate-ok",
			URL:        "/v1/workflow/rules/" + ruleID.String() + "/active",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPatch,
			Input:      ruleapi.ToggleActiveRequest{IsActive: false},
			GotResp:    &ruleapi.RuleResponse{},
			ExpResp:    &ruleapi.RuleResponse{},
			CmpFunc:    func(got, exp any) string { return "" },
		},
		{
			Name:       "reactivate-blocked",
			URL:        "/v1/workflow/rules/" + ruleID.String() + "/active",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPatch,
			Input:      ruleapi.ToggleActiveRequest{IsActive: true},
			GotResp:    &errs.Error{},
			ExpResp:    &errs.Error{},
			CmpFunc: func(got, exp any) string {
				gotResp, ok := got.(*errs.Error)
				if !ok {
					return "failed to cast error response"
				}
				if !gotResp.Code.Equal(errs.InvalidArgument) {
					return "expected InvalidArgument, got " + gotResp.Code.String()
				}
				return ""
			},
		},
	}
}

// activateBenignAllowed proves a rule whose action produces no entity mutations activates fine.
func activateBenignAllowed(sd CascadeSeedData) []apitest.Table {
	ruleID := sd.NonModifyingRule.ID
	return []apitest.Table{
		{
			Name:       "deactivate-ok",
			URL:        "/v1/workflow/rules/" + ruleID.String() + "/active",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPatch,
			Input:      ruleapi.ToggleActiveRequest{IsActive: false},
			GotResp:    &ruleapi.RuleResponse{},
			ExpResp:    &ruleapi.RuleResponse{},
			CmpFunc:    func(got, exp any) string { return "" },
		},
		{
			Name:       "reactivate-ok",
			URL:        "/v1/workflow/rules/" + ruleID.String() + "/active",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPatch,
			Input:      ruleapi.ToggleActiveRequest{IsActive: true},
			GotResp:    &ruleapi.RuleResponse{},
			ExpResp:    &ruleapi.RuleResponse{},
			CmpFunc: func(got, exp any) string {
				gotResp, ok := got.(*ruleapi.RuleResponse)
				if !ok {
					return "failed to cast rule response"
				}
				if !gotResp.IsActive {
					return "rule should be activated"
				}
				return ""
			},
		},
	}
}
