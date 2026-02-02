package rule_test

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

// =============================================================================
// Delete Rule Tests

func deleteRule200(sd RuleSeedData) []apitest.Table {
	if len(sd.Rules) < 2 {
		return nil
	}

	// Use the last rule for deletion test to avoid affecting other tests
	ruleID := sd.Rules[len(sd.Rules)-1].ID

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/workflow/rules/" + ruleID.String(),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusNoContent,
			Method:     http.MethodDelete,
			GotResp:    nil,
			ExpResp:    nil,
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		},
	}

	return table
}

func deleteRule404(sd RuleSeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "not-found",
			URL:        "/v1/workflow/rules/" + uuid.New().String(),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusNotFound,
			Method:     http.MethodDelete,
			GotResp:    &map[string]any{},
			ExpResp:    &map[string]any{},
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		},
	}

	return table
}

func deleteRule401(sd RuleSeedData) []apitest.Table {
	if len(sd.Rules) == 0 {
		return nil
	}

	table := []apitest.Table{
		{
			Name:       "no-token",
			URL:        "/v1/workflow/rules/" + sd.Rules[0].ID.String(),
			Token:      "",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodDelete,
			GotResp:    &map[string]any{},
			ExpResp:    &map[string]any{},
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		},
	}

	return table
}
