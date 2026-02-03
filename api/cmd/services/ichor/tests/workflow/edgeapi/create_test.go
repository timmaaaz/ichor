package edge_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/domain/http/workflow/edgeapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
)

// =============================================================================
// Create Edge Tests - Success Cases

func createEdgeStart200(sd EdgeSeedData) []apitest.Table {
	if len(sd.Rules) == 0 || len(sd.Actions) < 3 {
		return nil
	}

	ruleID := sd.Rules[0].ID
	// Use action[2] which is not yet connected
	targetActionID := sd.Actions[2].ID

	table := []apitest.Table{
		{
			Name:       "start-edge",
			URL:        "/v1/workflow/rules/" + ruleID.String() + "/edges",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			Input: edgeapi.CreateEdgeRequest{
				SourceActionID: nil, // Start edges have no source
				TargetActionID: targetActionID,
				EdgeType:       workflow.EdgeTypeStart,
				EdgeOrder:      10,
			},
			GotResp: &edgeapi.EdgeResponse{},
			ExpResp: &edgeapi.EdgeResponse{
				RuleID:         ruleID,
				SourceActionID: nil,
				TargetActionID: targetActionID,
				EdgeType:       workflow.EdgeTypeStart,
				EdgeOrder:      10,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*edgeapi.EdgeResponse)
				if !exists {
					return "error occurred"
				}
				expResp := exp.(*edgeapi.EdgeResponse)

				// Copy server-generated fields
				expResp.ID = gotResp.ID
				expResp.CreatedDate = gotResp.CreatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func createEdgeSequence200(sd EdgeSeedData) []apitest.Table {
	if len(sd.Rules) == 0 || len(sd.Actions) < 3 {
		return nil
	}

	ruleID := sd.Rules[0].ID
	sourceActionID := sd.Actions[1].ID
	targetActionID := sd.Actions[2].ID

	table := []apitest.Table{
		{
			Name:       "sequence-edge",
			URL:        "/v1/workflow/rules/" + ruleID.String() + "/edges",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			Input: edgeapi.CreateEdgeRequest{
				SourceActionID: &sourceActionID,
				TargetActionID: targetActionID,
				EdgeType:       workflow.EdgeTypeSequence,
				EdgeOrder:      20,
			},
			GotResp: &edgeapi.EdgeResponse{},
			ExpResp: &edgeapi.EdgeResponse{
				RuleID:         ruleID,
				SourceActionID: &sourceActionID,
				TargetActionID: targetActionID,
				EdgeType:       workflow.EdgeTypeSequence,
				EdgeOrder:      20,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*edgeapi.EdgeResponse)
				if !exists {
					return "error occurred"
				}
				expResp := exp.(*edgeapi.EdgeResponse)

				// Copy server-generated fields
				expResp.ID = gotResp.ID
				expResp.CreatedDate = gotResp.CreatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func createEdgeBranch200(sd EdgeSeedData) []apitest.Table {
	if len(sd.Rules) == 0 || len(sd.Actions) < 4 {
		return nil
	}

	ruleID := sd.Rules[0].ID
	sourceActionID := sd.Actions[2].ID
	targetActionID := sd.Actions[3].ID

	table := []apitest.Table{
		{
			Name:       "true-branch-edge",
			URL:        "/v1/workflow/rules/" + ruleID.String() + "/edges",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			Input: edgeapi.CreateEdgeRequest{
				SourceActionID: &sourceActionID,
				TargetActionID: targetActionID,
				EdgeType:       workflow.EdgeTypeTrueBranch,
				EdgeOrder:      30,
			},
			GotResp: &edgeapi.EdgeResponse{},
			ExpResp: &edgeapi.EdgeResponse{
				RuleID:         ruleID,
				SourceActionID: &sourceActionID,
				TargetActionID: targetActionID,
				EdgeType:       workflow.EdgeTypeTrueBranch,
				EdgeOrder:      30,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*edgeapi.EdgeResponse)
				if !exists {
					return "error occurred"
				}
				expResp := exp.(*edgeapi.EdgeResponse)

				expResp.ID = gotResp.ID
				expResp.CreatedDate = gotResp.CreatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
		{
			Name:       "false-branch-edge",
			URL:        "/v1/workflow/rules/" + ruleID.String() + "/edges",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			Input: edgeapi.CreateEdgeRequest{
				SourceActionID: &sourceActionID,
				TargetActionID: targetActionID,
				EdgeType:       workflow.EdgeTypeFalseBranch,
				EdgeOrder:      31,
			},
			GotResp: &edgeapi.EdgeResponse{},
			ExpResp: &edgeapi.EdgeResponse{
				RuleID:         ruleID,
				SourceActionID: &sourceActionID,
				TargetActionID: targetActionID,
				EdgeType:       workflow.EdgeTypeFalseBranch,
				EdgeOrder:      31,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*edgeapi.EdgeResponse)
				if !exists {
					return "error occurred"
				}
				expResp := exp.(*edgeapi.EdgeResponse)

				expResp.ID = gotResp.ID
				expResp.CreatedDate = gotResp.CreatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
		{
			Name:       "always-edge",
			URL:        "/v1/workflow/rules/" + ruleID.String() + "/edges",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			Input: edgeapi.CreateEdgeRequest{
				SourceActionID: &sourceActionID,
				TargetActionID: targetActionID,
				EdgeType:       workflow.EdgeTypeAlways,
				EdgeOrder:      32,
			},
			GotResp: &edgeapi.EdgeResponse{},
			ExpResp: &edgeapi.EdgeResponse{
				RuleID:         ruleID,
				SourceActionID: &sourceActionID,
				TargetActionID: targetActionID,
				EdgeType:       workflow.EdgeTypeAlways,
				EdgeOrder:      32,
			},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*edgeapi.EdgeResponse)
				if !exists {
					return "error occurred"
				}
				expResp := exp.(*edgeapi.EdgeResponse)

				expResp.ID = gotResp.ID
				expResp.CreatedDate = gotResp.CreatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

// =============================================================================
// Create Edge Tests - Validation Errors (400)

func createEdgeInvalidType400(sd EdgeSeedData) []apitest.Table {
	if len(sd.Rules) == 0 || len(sd.Actions) == 0 {
		return nil
	}

	ruleID := sd.Rules[0].ID
	targetActionID := sd.Actions[0].ID

	table := []apitest.Table{
		{
			Name:       "invalid-edge-type",
			URL:        "/v1/workflow/rules/" + ruleID.String() + "/edges",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPost,
			Input: edgeapi.CreateEdgeRequest{
				SourceActionID: nil,
				TargetActionID: targetActionID,
				EdgeType:       "invalid_type",
				EdgeOrder:      0,
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}
				if !gotResp.Code.Equal(errs.FailedPrecondition) {
					return "expected FailedPrecondition error code"
				}
				return ""
			},
		},
	}

	return table
}

func createEdgeMissingTarget400(sd EdgeSeedData) []apitest.Table {
	if len(sd.Rules) == 0 {
		return nil
	}

	ruleID := sd.Rules[0].ID

	table := []apitest.Table{
		{
			Name:       "missing-target",
			URL:        "/v1/workflow/rules/" + ruleID.String() + "/edges",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPost,
			Input: edgeapi.CreateEdgeRequest{
				SourceActionID: nil,
				TargetActionID: uuid.Nil, // Missing target
				EdgeType:       workflow.EdgeTypeStart,
				EdgeOrder:      0,
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}
				if !gotResp.Code.Equal(errs.FailedPrecondition) {
					return "expected FailedPrecondition error code"
				}
				return ""
			},
		},
	}

	return table
}

func createEdgeStartWithSource400(sd EdgeSeedData) []apitest.Table {
	if len(sd.Rules) == 0 || len(sd.Actions) < 2 {
		return nil
	}

	ruleID := sd.Rules[0].ID
	sourceActionID := sd.Actions[0].ID
	targetActionID := sd.Actions[1].ID

	table := []apitest.Table{
		{
			Name:       "start-with-source",
			URL:        "/v1/workflow/rules/" + ruleID.String() + "/edges",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPost,
			Input: edgeapi.CreateEdgeRequest{
				SourceActionID: &sourceActionID, // Start edges must not have source
				TargetActionID: targetActionID,
				EdgeType:       workflow.EdgeTypeStart,
				EdgeOrder:      0,
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}
				if !gotResp.Code.Equal(errs.FailedPrecondition) {
					return "expected FailedPrecondition error code"
				}
				return ""
			},
		},
	}

	return table
}

func createEdgeNonStartWithoutSource400(sd EdgeSeedData) []apitest.Table {
	if len(sd.Rules) == 0 || len(sd.Actions) == 0 {
		return nil
	}

	ruleID := sd.Rules[0].ID
	targetActionID := sd.Actions[0].ID

	table := []apitest.Table{
		{
			Name:       "sequence-without-source",
			URL:        "/v1/workflow/rules/" + ruleID.String() + "/edges",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPost,
			Input: edgeapi.CreateEdgeRequest{
				SourceActionID: nil, // Non-start edges must have source
				TargetActionID: targetActionID,
				EdgeType:       workflow.EdgeTypeSequence,
				EdgeOrder:      0,
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}
				if !gotResp.Code.Equal(errs.FailedPrecondition) {
					return "expected FailedPrecondition error code"
				}
				return ""
			},
		},
	}

	return table
}

func createEdgeTargetNotInRule400(sd EdgeSeedData) []apitest.Table {
	if len(sd.Rules) == 0 {
		return nil
	}

	ruleID := sd.Rules[0].ID
	// Use action from other rule
	targetActionID := sd.OtherAction.ID

	table := []apitest.Table{
		{
			Name:       "target-not-in-rule",
			URL:        "/v1/workflow/rules/" + ruleID.String() + "/edges",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPost,
			Input: edgeapi.CreateEdgeRequest{
				SourceActionID: nil,
				TargetActionID: targetActionID, // Action from different rule
				EdgeType:       workflow.EdgeTypeStart,
				EdgeOrder:      0,
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}
				if !gotResp.Code.Equal(errs.FailedPrecondition) && !gotResp.Code.Equal(errs.InvalidArgument) {
					return "expected FailedPrecondition or InvalidArgument error code"
				}
				return ""
			},
		},
	}

	return table
}

func createEdgeSourceNotInRule400(sd EdgeSeedData) []apitest.Table {
	if len(sd.Rules) == 0 || len(sd.Actions) == 0 {
		return nil
	}

	ruleID := sd.Rules[0].ID
	// Use action from other rule as source
	sourceActionID := sd.OtherAction.ID
	targetActionID := sd.Actions[0].ID

	table := []apitest.Table{
		{
			Name:       "source-not-in-rule",
			URL:        "/v1/workflow/rules/" + ruleID.String() + "/edges",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPost,
			Input: edgeapi.CreateEdgeRequest{
				SourceActionID: &sourceActionID, // Action from different rule
				TargetActionID: targetActionID,
				EdgeType:       workflow.EdgeTypeSequence,
				EdgeOrder:      0,
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got, exp any) string {
				gotResp, exists := got.(*errs.Error)
				if !exists {
					return "error occurred"
				}
				if !gotResp.Code.Equal(errs.FailedPrecondition) && !gotResp.Code.Equal(errs.InvalidArgument) {
					return "expected FailedPrecondition or InvalidArgument error code"
				}
				return ""
			},
		},
	}

	return table
}

// =============================================================================
// Create Edge Tests - Not Found Errors (404)

func createEdgeTargetNotFound404(sd EdgeSeedData) []apitest.Table {
	if len(sd.Rules) == 0 {
		return nil
	}

	ruleID := sd.Rules[0].ID

	table := []apitest.Table{
		{
			Name:       "target-not-found",
			URL:        "/v1/workflow/rules/" + ruleID.String() + "/edges",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusNotFound,
			Method:     http.MethodPost,
			Input: edgeapi.CreateEdgeRequest{
				SourceActionID: nil,
				TargetActionID: uuid.New(), // Non-existent action
				EdgeType:       workflow.EdgeTypeStart,
				EdgeOrder:      0,
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
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

func createEdgeSourceNotFound404(sd EdgeSeedData) []apitest.Table {
	if len(sd.Rules) == 0 || len(sd.Actions) == 0 {
		return nil
	}

	ruleID := sd.Rules[0].ID
	nonExistentID := uuid.New()
	targetActionID := sd.Actions[0].ID

	table := []apitest.Table{
		{
			Name:       "source-not-found",
			URL:        "/v1/workflow/rules/" + ruleID.String() + "/edges",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusNotFound,
			Method:     http.MethodPost,
			Input: edgeapi.CreateEdgeRequest{
				SourceActionID: &nonExistentID, // Non-existent action
				TargetActionID: targetActionID,
				EdgeType:       workflow.EdgeTypeSequence,
				EdgeOrder:      0,
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
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

func createEdgeRuleNotFound404(sd EdgeSeedData) []apitest.Table {
	if len(sd.Actions) == 0 {
		return nil
	}

	targetActionID := sd.Actions[0].ID

	table := []apitest.Table{
		{
			Name:       "rule-not-found",
			URL:        "/v1/workflow/rules/" + uuid.New().String() + "/edges",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusNotFound,
			Method:     http.MethodPost,
			Input: edgeapi.CreateEdgeRequest{
				SourceActionID: nil,
				TargetActionID: targetActionID,
				EdgeType:       workflow.EdgeTypeStart,
				EdgeOrder:      0,
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
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
// Create Edge Tests - Auth Errors (401)

func createEdge401(sd EdgeSeedData) []apitest.Table {
	if len(sd.Rules) == 0 || len(sd.Actions) == 0 {
		return nil
	}

	ruleID := sd.Rules[0].ID
	targetActionID := sd.Actions[0].ID

	table := []apitest.Table{
		{
			Name:       "no-token",
			URL:        "/v1/workflow/rules/" + ruleID.String() + "/edges",
			Token:      "",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodPost,
			Input: edgeapi.CreateEdgeRequest{
				SourceActionID: nil,
				TargetActionID: targetActionID,
				EdgeType:       workflow.EdgeTypeStart,
				EdgeOrder:      0,
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
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
