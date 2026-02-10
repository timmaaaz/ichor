package rule_test

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/domain/http/workflow/ruleapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

// =============================================================================
// Simulation Tests

func testRule200(sd RuleSeedData) []apitest.Table {
	if len(sd.Rules) == 0 {
		return nil
	}

	table := []apitest.Table{
		{
			Name:       "basic-simulation",
			URL:        fmt.Sprintf("/v1/workflow/rules/%s/test", sd.Rules[0].ID),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			Input: &ruleapi.TestRuleRequest{
				SampleData: map[string]interface{}{
					"status": "shipped",
					"total":  100.50,
				},
			},
			GotResp: &ruleapi.SimulationResult{},
			ExpResp: &ruleapi.SimulationResult{},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*ruleapi.SimulationResult)
				if !exists {
					return "error getting simulation result"
				}

				// Rule ID should match
				if gotResp.RuleID != sd.Rules[0].ID {
					return fmt.Sprintf("expected rule_id %s, got %s", sd.Rules[0].ID, gotResp.RuleID)
				}

				// Rule name should be populated
				if gotResp.RuleName == "" {
					return "expected rule_name to be populated"
				}

				// ActionsToExecute should include our seeded actions
				if len(sd.Actions) > 0 && len(gotResp.ActionsToExecute) == 0 {
					return "expected actions_to_execute to include seeded actions"
				}

				return ""
			},
		},
		{
			Name:       "simulation-with-matching-condition",
			URL:        fmt.Sprintf("/v1/workflow/rules/%s/test", sd.Rules[0].ID),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			Input: &ruleapi.TestRuleRequest{
				SampleData: map[string]interface{}{
					"order": map[string]interface{}{
						"status": "delivered",
						"total":  500.00,
					},
				},
			},
			GotResp: &ruleapi.SimulationResult{},
			ExpResp: &ruleapi.SimulationResult{},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*ruleapi.SimulationResult)
				if !exists {
					return "error getting simulation result"
				}

				// Should return a valid simulation result
				if gotResp.RuleID == uuid.Nil {
					return "rule_id should not be nil"
				}

				// TemplatePreview should be populated with flattened sample data
				// (may be empty if no templates in action configs)
				_ = gotResp.TemplatePreview

				return ""
			},
		},
		{
			Name:       "simulation-empty-sample-data",
			URL:        fmt.Sprintf("/v1/workflow/rules/%s/test", sd.Rules[0].ID),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			Input: &ruleapi.TestRuleRequest{
				SampleData: map[string]interface{}{},
			},
			GotResp: &ruleapi.SimulationResult{},
			ExpResp: &ruleapi.SimulationResult{},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*ruleapi.SimulationResult)
				if !exists {
					return "error getting simulation result"
				}

				// Should still return valid response even with empty data
				if gotResp.RuleID != sd.Rules[0].ID {
					return fmt.Sprintf("expected rule_id %s, got %s", sd.Rules[0].ID, gotResp.RuleID)
				}

				return ""
			},
		},
	}

	return table
}

func testRule404(sd RuleSeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "rule-not-found",
			URL:        fmt.Sprintf("/v1/workflow/rules/%s/test", uuid.New()),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusNotFound,
			Method:     http.MethodPost,
			Input: &ruleapi.TestRuleRequest{
				SampleData: map[string]interface{}{"test": "data"},
			},
			GotResp: &map[string]any{},
			ExpResp: &map[string]any{},
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		},
	}

	return table
}

func testRule401(sd RuleSeedData) []apitest.Table {
	if len(sd.Rules) == 0 {
		return nil
	}

	table := []apitest.Table{
		{
			Name:       "no-token",
			URL:        fmt.Sprintf("/v1/workflow/rules/%s/test", sd.Rules[0].ID),
			Token:      "",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodPost,
			Input: &ruleapi.TestRuleRequest{
				SampleData: map[string]interface{}{"test": "data"},
			},
			GotResp: &map[string]any{},
			ExpResp: &map[string]any{},
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		},
	}

	return table
}

// =============================================================================
// Rule Execution History Tests

func queryRuleExecutions200(sd RuleSeedData) []apitest.Table {
	if len(sd.Rules) == 0 {
		return nil
	}

	table := []apitest.Table{
		{
			Name:       "basic-query",
			URL:        fmt.Sprintf("/v1/workflow/rules/%s/executions", sd.Rules[0].ID),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &map[string]any{},
			ExpResp:    &map[string]any{},
			CmpFunc: func(got any, exp any) string {
				// Just verify we get a valid response (may have 0 executions)
				return ""
			},
		},
		{
			Name:       "query-with-pagination",
			URL:        fmt.Sprintf("/v1/workflow/rules/%s/executions?page=1&rows=10", sd.Rules[0].ID),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &map[string]any{},
			ExpResp:    &map[string]any{},
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		},
	}

	return table
}

func queryRuleExecutions404(sd RuleSeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "rule-not-found",
			URL:        fmt.Sprintf("/v1/workflow/rules/%s/executions", uuid.New()),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusNotFound,
			Method:     http.MethodGet,
			GotResp:    &map[string]any{},
			ExpResp:    &map[string]any{},
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		},
	}

	return table
}

func queryRuleExecutions401(sd RuleSeedData) []apitest.Table {
	if len(sd.Rules) == 0 {
		return nil
	}

	table := []apitest.Table{
		{
			Name:       "no-token",
			URL:        fmt.Sprintf("/v1/workflow/rules/%s/executions", sd.Rules[0].ID),
			Token:      "",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodGet,
			GotResp:    &map[string]any{},
			ExpResp:    &map[string]any{},
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		},
	}

	return table
}
