package rule_test

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/domain/http/workflow/ruleapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

// =============================================================================
// Toggle Active Tests

func toggleActive200(sd RuleSeedData) []apitest.Table {
	if len(sd.Rules) == 0 {
		return nil
	}

	ruleID := sd.Rules[0].ID

	table := []apitest.Table{
		{
			Name:       "deactivate",
			URL:        "/v1/workflow/rules/" + ruleID.String() + "/active",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPatch,
			Input: ruleapi.ToggleActiveRequest{
				IsActive: false,
			},
			GotResp: &ruleapi.RuleResponse{},
			ExpResp: &ruleapi.RuleResponse{},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*ruleapi.RuleResponse)
				if !exists {
					return "error getting rule response"
				}

				if gotResp.IsActive {
					return "rule should be deactivated"
				}

				return ""
			},
		},
		{
			Name:       "activate",
			URL:        "/v1/workflow/rules/" + ruleID.String() + "/active",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPatch,
			Input: ruleapi.ToggleActiveRequest{
				IsActive: true,
			},
			GotResp: &ruleapi.RuleResponse{},
			ExpResp: &ruleapi.RuleResponse{},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*ruleapi.RuleResponse)
				if !exists {
					return "error getting rule response"
				}

				if !gotResp.IsActive {
					return "rule should be activated"
				}

				return ""
			},
		},
	}

	return table
}

func toggleActive404(sd RuleSeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "not-found",
			URL:        "/v1/workflow/rules/" + uuid.New().String() + "/active",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusNotFound,
			Method:     http.MethodPatch,
			Input: ruleapi.ToggleActiveRequest{
				IsActive: false,
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

func toggleActive401(sd RuleSeedData) []apitest.Table {
	if len(sd.Rules) == 0 {
		return nil
	}

	table := []apitest.Table{
		{
			Name:       "no-token",
			URL:        "/v1/workflow/rules/" + sd.Rules[0].ID.String() + "/active",
			Token:      "",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodPatch,
			Input: ruleapi.ToggleActiveRequest{
				IsActive: false,
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
