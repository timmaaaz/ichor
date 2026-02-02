package rule_test

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/domain/http/workflow/ruleapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

// =============================================================================
// Query Rules Tests

func queryRules200(sd RuleSeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/workflow/rules",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[ruleapi.RuleResponse]{},
			ExpResp:    &query.Result[ruleapi.RuleResponse]{},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*query.Result[ruleapi.RuleResponse])
				if !exists {
					return "error getting rules response"
				}

				// Should have at least the seeded rules
				if len(gotResp.Items) < len(sd.Rules) {
					return "expected at least seeded rules"
				}

				return ""
			},
		},
		{
			Name:       "with-pagination",
			URL:        "/v1/workflow/rules?page=1&rows=2",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[ruleapi.RuleResponse]{},
			ExpResp:    &query.Result[ruleapi.RuleResponse]{},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*query.Result[ruleapi.RuleResponse])
				if !exists {
					return "error getting rules response"
				}

				if len(gotResp.Items) > 2 {
					return "expected max 2 items with pagination"
				}

				return ""
			},
		},
		{
			Name:       "filter-by-active",
			URL:        "/v1/workflow/rules?is_active=true",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[ruleapi.RuleResponse]{},
			ExpResp:    &query.Result[ruleapi.RuleResponse]{},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*query.Result[ruleapi.RuleResponse])
				if !exists {
					return "error getting rules response"
				}

				for _, r := range gotResp.Items {
					if !r.IsActive {
						return "expected only active rules"
					}
				}

				return ""
			},
		},
	}

	return table
}

func queryRules401(sd RuleSeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "no-token",
			URL:        "/v1/workflow/rules",
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

// =============================================================================
// Query Rule by ID Tests

func queryRuleByID200(sd RuleSeedData) []apitest.Table {
	if len(sd.Rules) == 0 {
		return nil
	}

	ruleID := sd.Rules[0].ID

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/workflow/rules/" + ruleID.String(),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &ruleapi.RuleResponse{},
			ExpResp:    &ruleapi.RuleResponse{},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*ruleapi.RuleResponse)
				if !exists {
					return "error getting rule response"
				}

				if gotResp.ID != ruleID {
					return "wrong rule returned"
				}

				if gotResp.Name == "" {
					return "rule name is empty"
				}

				return ""
			},
		},
	}

	return table
}

func queryRuleByID404(sd RuleSeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "not-found",
			URL:        "/v1/workflow/rules/" + uuid.New().String(),
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

func queryRuleByID401(sd RuleSeedData) []apitest.Table {
	if len(sd.Rules) == 0 {
		return nil
	}

	table := []apitest.Table{
		{
			Name:       "no-token",
			URL:        "/v1/workflow/rules/" + sd.Rules[0].ID.String(),
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
