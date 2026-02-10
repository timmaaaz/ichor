package edge_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/domain/http/workflow/edgeapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

// =============================================================================
// Query Edges Tests

func queryEdges200(sd EdgeSeedData) []apitest.Table {
	if len(sd.Rules) == 0 {
		return nil
	}

	ruleID := sd.Rules[0].ID

	// Build expected edge list from seed data
	expEdges := make(edgeapi.EdgeList, len(sd.Edges))
	for i, edge := range sd.Edges {
		expEdges[i] = toEdgeResponse(edge)
	}

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/workflow/rules/" + ruleID.String() + "/edges",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &edgeapi.EdgeList{},
			ExpResp:    &expEdges,
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*edgeapi.EdgeList)
				if !exists {
					return "error occurred"
				}
				expResp := exp.(*edgeapi.EdgeList)

				return cmp.Diff(*gotResp, *expResp)
			},
		},
	}

	return table
}

func queryEdgesRuleNotFound404(sd EdgeSeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "rule-not-found",
			URL:        "/v1/workflow/rules/" + uuid.New().String() + "/edges",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusNotFound,
			Method:     http.MethodGet,
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

func queryEdges401(sd EdgeSeedData) []apitest.Table {
	if len(sd.Rules) == 0 {
		return nil
	}

	table := []apitest.Table{
		{
			Name:       "no-token",
			URL:        "/v1/workflow/rules/" + sd.Rules[0].ID.String() + "/edges",
			Token:      "",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodGet,
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
// Query Edge by ID Tests

func queryEdgeByID200(sd EdgeSeedData) []apitest.Table {
	if len(sd.Rules) == 0 || len(sd.Edges) == 0 {
		return nil
	}

	ruleID := sd.Rules[0].ID
	edge := sd.Edges[0]
	expResp := toEdgeResponse(edge)

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/workflow/rules/" + ruleID.String() + "/edges/" + edge.ID.String(),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &edgeapi.EdgeResponse{},
			ExpResp:    &expResp,
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*edgeapi.EdgeResponse)
				if !exists {
					return "error occurred"
				}
				expResp := exp.(*edgeapi.EdgeResponse)

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func queryEdgeByIDNotFound404(sd EdgeSeedData) []apitest.Table {
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
			Method:     http.MethodGet,
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

func queryEdgeByIDWrongRule404(sd EdgeSeedData) []apitest.Table {
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
			Method:     http.MethodGet,
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
