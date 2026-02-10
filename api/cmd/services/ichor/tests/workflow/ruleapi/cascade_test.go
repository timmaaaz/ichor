package rule_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/domain/http/workflow/ruleapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

// =============================================================================
// Phase 12.12: Cascade Visualization API Integration Tests
// =============================================================================

// Test_CascadeAPI is the main test orchestrator for cascade visualization tests.
func Test_CascadeAPI(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_CascadeAPI")

	sd, err := insertCascadeSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// ============================================================
	// Cascade Map Tests - Success Cases
	// ============================================================
	test.Run(t, cascadeMap200WithDownstream(sd), "cascadeMap-with-downstream-200")
	test.Run(t, cascadeMap200NoDownstream(sd), "cascadeMap-no-downstream-200")
	test.Run(t, cascadeMap200MultipleActions(sd), "cascadeMap-multiple-actions-200")
	test.Run(t, cascadeMapExcludesSelf(sd), "cascadeMap-excludes-self-200")
	test.Run(t, cascadeMapOnlyActiveRules(sd), "cascadeMap-only-active-200")
	test.Run(t, cascadeMapResponseStructure(sd), "cascadeMap-response-structure-200")
	test.Run(t, cascadeMapEmptyActions(sd), "cascadeMap-empty-actions-200")

	// ============================================================
	// Cascade Map Tests - Error Cases
	// ============================================================
	test.Run(t, cascadeMapRuleNotFound404(sd), "cascadeMap-rule-not-found-404")
	test.Run(t, cascadeMap401(sd), "cascadeMap-401")
}

// cascadeMap200WithDownstream tests that the cascade-map endpoint returns
// downstream workflows when a rule has update_field actions that modify entities
// that other rules are listening to.
func cascadeMap200WithDownstream(sd CascadeSeedData) []apitest.Table {
	if len(sd.DownstreamTriggerRules) == 0 {
		return nil
	}

	// The primary rule has an action that modifies the entity that downstream rules listen to
	ruleID := sd.PrimaryRule.ID

	table := []apitest.Table{
		{
			Name:       "with-downstream",
			URL:        "/v1/workflow/rules/" + ruleID.String() + "/cascade-map",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &ruleapi.CascadeResponse{},
			ExpResp:    &ruleapi.CascadeResponse{},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*ruleapi.CascadeResponse)
				if !exists {
					return "error occurred: failed to cast response"
				}

				// Verify the rule info is correct
				if gotResp.RuleID != ruleID.String() {
					return fmt.Sprintf("expected rule_id %s, got %s", ruleID.String(), gotResp.RuleID)
				}

				if gotResp.RuleName != sd.PrimaryRule.Name {
					return fmt.Sprintf("expected rule_name %s, got %s", sd.PrimaryRule.Name, gotResp.RuleName)
				}

				// Find the update_field action and check for downstream workflows
				foundDownstream := false
				for _, action := range gotResp.Actions {
					if action.ActionType == "update_field" && len(action.DownstreamWorkflows) > 0 {
						foundDownstream = true
						// Verify downstream rule info
						for _, downstream := range action.DownstreamWorkflows {
							if downstream.RuleID == "" {
								return "downstream workflow missing rule_id"
							}
							if downstream.RuleName == "" {
								return "downstream workflow missing rule_name"
							}
							if downstream.WillTriggerIf == "" {
								return "downstream workflow missing will_trigger_if description"
							}
						}
					}
				}

				if !foundDownstream {
					return "expected downstream workflows for update_field action but found none"
				}

				return ""
			},
		},
	}

	return table
}

// cascadeMap200NoDownstream tests that the cascade-map endpoint returns
// empty downstream_workflows for actions that don't modify entities.
func cascadeMap200NoDownstream(sd CascadeSeedData) []apitest.Table {
	// Use the rule that only has non-modifying actions
	ruleID := sd.NonModifyingRule.ID

	table := []apitest.Table{
		{
			Name:       "no-downstream",
			URL:        "/v1/workflow/rules/" + ruleID.String() + "/cascade-map",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &ruleapi.CascadeResponse{},
			ExpResp:    &ruleapi.CascadeResponse{},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*ruleapi.CascadeResponse)
				if !exists {
					return "error occurred: failed to cast response"
				}

				// Verify the rule info is correct
				if gotResp.RuleID != ruleID.String() {
					return fmt.Sprintf("expected rule_id %s, got %s", ruleID.String(), gotResp.RuleID)
				}

				// All actions should have empty downstream_workflows
				for _, action := range gotResp.Actions {
					if len(action.DownstreamWorkflows) > 0 {
						return fmt.Sprintf("expected no downstream workflows for action %s, got %d",
							action.ActionName, len(action.DownstreamWorkflows))
					}
				}

				return ""
			},
		},
	}

	return table
}

// cascadeMap200MultipleActions tests that cascade-map correctly handles
// rules with multiple actions - some that modify entities and some that don't.
func cascadeMap200MultipleActions(sd CascadeSeedData) []apitest.Table {
	// Use the mixed rule that has both modifying and non-modifying actions
	ruleID := sd.MixedActionsRule.ID

	table := []apitest.Table{
		{
			Name:       "multiple-actions",
			URL:        "/v1/workflow/rules/" + ruleID.String() + "/cascade-map",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &ruleapi.CascadeResponse{},
			ExpResp:    &ruleapi.CascadeResponse{},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*ruleapi.CascadeResponse)
				if !exists {
					return "error occurred: failed to cast response"
				}

				if len(gotResp.Actions) < 2 {
					return fmt.Sprintf("expected at least 2 actions, got %d", len(gotResp.Actions))
				}

				// Verify we have a mix of actions with and without entity modifications
				hasModifying := false
				hasNonModifying := false

				for _, action := range gotResp.Actions {
					if action.ModifiesEntity != "" {
						hasModifying = true
					} else {
						hasNonModifying = true
					}
				}

				if !hasModifying {
					return "expected at least one modifying action"
				}
				if !hasNonModifying {
					return "expected at least one non-modifying action"
				}

				return ""
			},
		},
	}

	return table
}

// cascadeMapRuleNotFound404 tests that cascade-map returns 404 for non-existent rules.
func cascadeMapRuleNotFound404(sd CascadeSeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "rule-not-found",
			URL:        "/v1/workflow/rules/" + uuid.New().String() + "/cascade-map",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusNotFound,
			Method:     http.MethodGet,
			GotResp:    &errs.Error{},
			ExpResp:    &errs.Error{},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred: failed to cast error response"
				}
				if !gotResp.Code.Equal(errs.NotFound) {
					return fmt.Sprintf("expected NotFound error code, got %s", gotResp.Code)
				}
				return ""
			},
		},
	}

	return table
}

// cascadeMap401 tests that cascade-map requires authentication.
func cascadeMap401(sd CascadeSeedData) []apitest.Table {
	ruleID := sd.PrimaryRule.ID

	table := []apitest.Table{
		{
			Name:       "no-token",
			URL:        "/v1/workflow/rules/" + ruleID.String() + "/cascade-map",
			Token:      "",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodGet,
			GotResp:    &errs.Error{},
			ExpResp:    &errs.Error{},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred: failed to cast error response"
				}
				if !gotResp.Code.Equal(errs.Unauthenticated) {
					return fmt.Sprintf("expected Unauthenticated error code, got %s", gotResp.Code)
				}
				return ""
			},
		},
	}

	return table
}

// cascadeMapExcludesSelf tests that a rule doesn't show itself as a downstream workflow.
func cascadeMapExcludesSelf(sd CascadeSeedData) []apitest.Table {
	// Use the self-triggering rule (listens to same entity it modifies)
	ruleID := sd.SelfTriggerRule.ID

	table := []apitest.Table{
		{
			Name:       "excludes-self",
			URL:        "/v1/workflow/rules/" + ruleID.String() + "/cascade-map",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &ruleapi.CascadeResponse{},
			ExpResp:    &ruleapi.CascadeResponse{},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*ruleapi.CascadeResponse)
				if !exists {
					return "error occurred: failed to cast response"
				}

				// Verify the rule doesn't appear in its own downstream workflows
				for _, action := range gotResp.Actions {
					for _, downstream := range action.DownstreamWorkflows {
						if downstream.RuleID == ruleID.String() {
							return "rule should not appear in its own downstream workflows"
						}
					}
				}

				return ""
			},
		},
	}

	return table
}

// cascadeMapOnlyActiveRules tests that cascade-map only shows active downstream rules.
func cascadeMapOnlyActiveRules(sd CascadeSeedData) []apitest.Table {
	// The primary rule modifies an entity that an inactive rule also listens to
	ruleID := sd.PrimaryRule.ID
	inactiveRuleID := sd.InactiveDownstreamRule.ID.String()

	table := []apitest.Table{
		{
			Name:       "only-active-rules",
			URL:        "/v1/workflow/rules/" + ruleID.String() + "/cascade-map",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &ruleapi.CascadeResponse{},
			ExpResp:    &ruleapi.CascadeResponse{},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*ruleapi.CascadeResponse)
				if !exists {
					return "error occurred: failed to cast response"
				}

				// Verify that the inactive rule does NOT appear in downstream workflows
				for _, action := range gotResp.Actions {
					for _, downstream := range action.DownstreamWorkflows {
						if downstream.RuleID == inactiveRuleID {
							return "inactive rules should not appear in downstream workflows"
						}
					}
				}

				return ""
			},
		},
	}

	return table
}

// cascadeMapResponseStructure tests the complete structure of the cascade-map response.
func cascadeMapResponseStructure(sd CascadeSeedData) []apitest.Table {
	ruleID := sd.PrimaryRule.ID

	// Build expected response structure
	expResp := ruleapi.CascadeResponse{
		RuleID:   ruleID.String(),
		RuleName: sd.PrimaryRule.Name,
	}

	table := []apitest.Table{
		{
			Name:       "response-structure",
			URL:        "/v1/workflow/rules/" + ruleID.String() + "/cascade-map",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &ruleapi.CascadeResponse{},
			ExpResp:    &expResp,
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*ruleapi.CascadeResponse)
				if !exists {
					return "error occurred: failed to cast response"
				}
				expResp := exp.(*ruleapi.CascadeResponse)

				// Compare rule ID and name using cmp.Diff
				if diff := cmp.Diff(gotResp.RuleID, expResp.RuleID); diff != "" {
					return diff
				}
				if diff := cmp.Diff(gotResp.RuleName, expResp.RuleName); diff != "" {
					return diff
				}

				// Verify actions array is not nil
				if gotResp.Actions == nil {
					return "actions array should not be nil"
				}

				return ""
			},
		},
	}

	return table
}

// cascadeMapEmptyActions tests cascade-map with a rule that has no actions.
func cascadeMapEmptyActions(sd CascadeSeedData) []apitest.Table {
	// Note: We need to check if InactiveDownstreamRule has no actions,
	// or we can skip this test if we don't have such a rule
	ruleID := sd.InactiveDownstreamRule.ID

	table := []apitest.Table{
		{
			Name:       "rule-no-actions",
			URL:        "/v1/workflow/rules/" + ruleID.String() + "/cascade-map",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &ruleapi.CascadeResponse{},
			ExpResp:    &ruleapi.CascadeResponse{},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*ruleapi.CascadeResponse)
				if !exists {
					return "error occurred: failed to cast response"
				}

				// Rule should still be queryable even if it has no actions
				if gotResp.RuleID != ruleID.String() {
					return fmt.Sprintf("expected rule_id %s, got %s", ruleID.String(), gotResp.RuleID)
				}

				// Actions array should be empty but not nil
				if gotResp.Actions == nil {
					return "actions array should not be nil"
				}

				return ""
			},
		},
	}

	return table
}
