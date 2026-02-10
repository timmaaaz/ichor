package edge_test

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

// =============================================================================
// Delete Edge Tests - Success Cases

func deleteEdge200(sd EdgeSeedData) []apitest.Table {
	if len(sd.Rules) == 0 || len(sd.Edges) == 0 {
		return nil
	}

	ruleID := sd.Rules[0].ID
	// Use the sequence edge (index 1) for deletion test
	// Leave start edge (index 0) intact for other tests
	edgeID := sd.Edges[1].ID

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/workflow/rules/" + ruleID.String() + "/edges/" + edgeID.String(),
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

// =============================================================================
// Delete Edge Tests - Not Found Errors (404)

func deleteEdgeNotFound404(sd EdgeSeedData) []apitest.Table {
	if len(sd.Rules) == 0 {
		return nil
	}

	ruleID := sd.Rules[0].ID

	table := []apitest.Table{
		{
			Name:       "edge-not-found",
			URL:        "/v1/workflow/rules/" + ruleID.String() + "/edges/" + uuid.New().String(),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusNotFound,
			Method:     http.MethodDelete,
			GotResp:    &errs.Error{},
			ExpResp:    &errs.Error{},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}
				if !gotResp.Code.Equal(errs.NotFound) {
					return "expected NotFound error code"
				}
				return ""
			},
		},
	}

	return table
}

func deleteEdgeWrongRule404(sd EdgeSeedData) []apitest.Table {
	if len(sd.Edges) == 0 {
		return nil
	}

	// Use other rule's ID with an edge that belongs to the primary rule
	edgeID := sd.Edges[0].ID
	wrongRuleID := sd.OtherRule.ID

	table := []apitest.Table{
		{
			Name:       "edge-wrong-rule",
			URL:        "/v1/workflow/rules/" + wrongRuleID.String() + "/edges/" + edgeID.String(),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusNotFound,
			Method:     http.MethodDelete,
			GotResp:    &errs.Error{},
			ExpResp:    &errs.Error{},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}
				if !gotResp.Code.Equal(errs.NotFound) {
					return "expected NotFound error code"
				}
				return ""
			},
		},
	}

	return table
}

func deleteEdgeRuleNotFound404(sd EdgeSeedData) []apitest.Table {
	if len(sd.Edges) == 0 {
		return nil
	}

	edgeID := sd.Edges[0].ID

	table := []apitest.Table{
		{
			Name:       "rule-not-found",
			URL:        "/v1/workflow/rules/" + uuid.New().String() + "/edges/" + edgeID.String(),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusNotFound,
			Method:     http.MethodDelete,
			GotResp:    &errs.Error{},
			ExpResp:    &errs.Error{},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}
				if !gotResp.Code.Equal(errs.NotFound) {
					return "expected NotFound error code"
				}
				return ""
			},
		},
	}

	return table
}

// =============================================================================
// Delete Edge Tests - Auth Errors (401)

func deleteEdge401(sd EdgeSeedData) []apitest.Table {
	if len(sd.Rules) == 0 || len(sd.Edges) == 0 {
		return nil
	}

	ruleID := sd.Rules[0].ID
	edgeID := sd.Edges[0].ID

	table := []apitest.Table{
		{
			Name:       "no-token",
			URL:        "/v1/workflow/rules/" + ruleID.String() + "/edges/" + edgeID.String(),
			Token:      "",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodDelete,
			GotResp:    &errs.Error{},
			ExpResp:    &errs.Error{},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}
				if !gotResp.Code.Equal(errs.Unauthenticated) {
					return "expected Unauthenticated error code"
				}
				return ""
			},
		},
	}

	return table
}

// =============================================================================
// Delete All Edges Tests

func deleteAllEdges200(sd EdgeSeedData) []apitest.Table {
	// Note: We use the OtherRule for this test to not interfere with other tests
	// that depend on the primary rule's edges
	ruleID := sd.OtherRule.ID

	table := []apitest.Table{
		{
			Name:       "delete-all",
			URL:        "/v1/workflow/rules/" + ruleID.String() + "/edges-all",
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

func deleteAllEdgesRuleNotFound404(sd EdgeSeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "rule-not-found",
			URL:        "/v1/workflow/rules/" + uuid.New().String() + "/edges-all",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusNotFound,
			Method:     http.MethodDelete,
			GotResp:    &errs.Error{},
			ExpResp:    &errs.Error{},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}
				if !gotResp.Code.Equal(errs.NotFound) {
					return "expected NotFound error code"
				}
				return ""
			},
		},
	}

	return table
}
