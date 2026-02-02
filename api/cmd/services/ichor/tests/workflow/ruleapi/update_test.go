package rule_test

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/domain/http/workflow/ruleapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

// =============================================================================
// Update Rule Tests

func updateRule200(sd RuleSeedData) []apitest.Table {
	if len(sd.Rules) == 0 {
		return nil
	}

	ruleID := sd.Rules[0].ID
	newName := "Updated Rule Name"
	newDescription := "Updated description"

	table := []apitest.Table{
		{
			Name:       "update-name",
			URL:        "/v1/workflow/rules/" + ruleID.String(),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPut,
			Input: ruleapi.UpdateRuleRequest{
				Name: &newName,
			},
			GotResp: &ruleapi.RuleResponse{},
			ExpResp: &ruleapi.RuleResponse{},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*ruleapi.RuleResponse)
				if !exists {
					return "error getting rule response"
				}

				if gotResp.Name != newName {
					return "rule name was not updated"
				}

				return ""
			},
		},
		{
			Name:       "update-description",
			URL:        "/v1/workflow/rules/" + ruleID.String(),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPut,
			Input: ruleapi.UpdateRuleRequest{
				Description: &newDescription,
			},
			GotResp: &ruleapi.RuleResponse{},
			ExpResp: &ruleapi.RuleResponse{},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*ruleapi.RuleResponse)
				if !exists {
					return "error getting rule response"
				}

				if gotResp.Description != newDescription {
					return "rule description was not updated"
				}

				return ""
			},
		},
		{
			Name:       "update-multiple-fields",
			URL:        "/v1/workflow/rules/" + ruleID.String(),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPut,
			Input: ruleapi.UpdateRuleRequest{
				Name:        &newName,
				Description: &newDescription,
			},
			GotResp: &ruleapi.RuleResponse{},
			ExpResp: &ruleapi.RuleResponse{},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*ruleapi.RuleResponse)
				if !exists {
					return "error getting rule response"
				}

				if gotResp.ID != ruleID {
					return "wrong rule ID"
				}

				return ""
			},
		},
	}

	return table
}

func updateRule404(sd RuleSeedData) []apitest.Table {
	newName := "Updated Name"

	table := []apitest.Table{
		{
			Name:       "not-found",
			URL:        "/v1/workflow/rules/" + uuid.New().String(),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusNotFound,
			Method:     http.MethodPut,
			Input: ruleapi.UpdateRuleRequest{
				Name: &newName,
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

func updateRule401(sd RuleSeedData) []apitest.Table {
	if len(sd.Rules) == 0 {
		return nil
	}

	newName := "Updated Name"

	table := []apitest.Table{
		{
			Name:       "no-token",
			URL:        "/v1/workflow/rules/" + sd.Rules[0].ID.String(),
			Token:      "",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodPut,
			Input: ruleapi.UpdateRuleRequest{
				Name: &newName,
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
