package workflowsaveapi_test

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/workflow/workflowsaveapp"
)

// =============================================================================
// Dry-Run Validation Tests (POST /v1/workflow/rules/full?dry_run=true)
// =============================================================================

// dryRunValid200 tests that a valid workflow payload returns valid=true
// with correct action and edge counts, without persisting anything.
func dryRunValid200(sd SaveSeedData) []apitest.Table {
	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 {
		return nil
	}

	return []apitest.Table{
		{
			Name:       "valid-workflow",
			URL:        "/v1/workflow/rules/full?dry_run=true",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          "Dry Run Valid Workflow",
				IsActive:      true,
				EntityID:      sd.Entities[0].ID.String(),
				TriggerTypeID: sd.TriggerTypes[0].ID.String(),
				Actions: []workflowsaveapp.SaveActionRequest{
					{
						Name:         "Alert Action",
						ActionType:   "create_alert",
						IsActive:     true,
						ActionConfig: json.RawMessage(`{"alert_type":"info","severity":"info","title":"Test","message":"Dry run"}`),
					},
					{
						Name:         "Second Action",
						ActionType:   "create_alert",
						IsActive:     true,
						ActionConfig: json.RawMessage(`{"alert_type":"info","severity":"info","title":"Test 2","message":"Dry run 2"}`),
					},
				},
				Edges: []workflowsaveapp.SaveEdgeRequest{
					{TargetActionID: "temp:0", EdgeType: "start"},
					{SourceActionID: "temp:0", TargetActionID: "temp:1", EdgeType: "sequence"},
				},
			},
			GotResp: &workflowsaveapp.ValidationResult{},
			ExpResp: &workflowsaveapp.ValidationResult{},
			CmpFunc: func(got any, exp any) string {
				gotResp, ok := got.(*workflowsaveapp.ValidationResult)
				if !ok {
					return "error casting response"
				}
				if !gotResp.Valid {
					return fmt.Sprintf("expected valid=true, got errors: %v", gotResp.Errors)
				}
				if gotResp.ActionCount != 2 {
					return fmt.Sprintf("expected action_count=2, got %d", gotResp.ActionCount)
				}
				if gotResp.EdgeCount != 2 {
					return fmt.Sprintf("expected edge_count=2, got %d", gotResp.EdgeCount)
				}

				return ""
			},
		},
	}
}

// dryRunInvalid200 tests that invalid workflows return valid=false with
// meaningful errors, without persisting anything.
func dryRunInvalid200(sd SaveSeedData) []apitest.Table {
	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 {
		return nil
	}

	return []apitest.Table{
		{
			Name:       "no-start-edge",
			URL:        "/v1/workflow/rules/full?dry_run=true",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          "Dry Run No Start Edge",
				IsActive:      true,
				EntityID:      sd.Entities[0].ID.String(),
				TriggerTypeID: sd.TriggerTypes[0].ID.String(),
				Actions: []workflowsaveapp.SaveActionRequest{
					{
						Name:         "Action 1",
						ActionType:   "create_alert",
						IsActive:     true,
						ActionConfig: json.RawMessage(`{"alert_type":"info","severity":"info","title":"T","message":"M"}`),
					},
					{
						Name:         "Action 2",
						ActionType:   "create_alert",
						IsActive:     true,
						ActionConfig: json.RawMessage(`{"alert_type":"info","severity":"info","title":"T","message":"M"}`),
					},
				},
				Edges: []workflowsaveapp.SaveEdgeRequest{
					// No start edge — only a sequence edge
					{SourceActionID: "temp:0", TargetActionID: "temp:1", EdgeType: "sequence"},
				},
			},
			GotResp: &workflowsaveapp.ValidationResult{},
			ExpResp: &workflowsaveapp.ValidationResult{},
			CmpFunc: func(got any, exp any) string {
				gotResp, ok := got.(*workflowsaveapp.ValidationResult)
				if !ok {
					return "error casting response"
				}
				if gotResp.Valid {
					return "expected valid=false for workflow with no start edge"
				}
				if len(gotResp.Errors) == 0 {
					return "expected validation errors"
				}

				return ""
			},
		},
		{
			Name:       "cycle",
			URL:        "/v1/workflow/rules/full?dry_run=true",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          "Dry Run Cycle",
				IsActive:      true,
				EntityID:      sd.Entities[0].ID.String(),
				TriggerTypeID: sd.TriggerTypes[0].ID.String(),
				Actions: []workflowsaveapp.SaveActionRequest{
					{
						Name:         "Action 1",
						ActionType:   "create_alert",
						IsActive:     true,
						ActionConfig: json.RawMessage(`{"alert_type":"info","severity":"info","title":"T","message":"M"}`),
					},
					{
						Name:         "Action 2",
						ActionType:   "create_alert",
						IsActive:     true,
						ActionConfig: json.RawMessage(`{"alert_type":"info","severity":"info","title":"T","message":"M"}`),
					},
				},
				Edges: []workflowsaveapp.SaveEdgeRequest{
					{TargetActionID: "temp:0", EdgeType: "start"},
					{SourceActionID: "temp:0", TargetActionID: "temp:1", EdgeType: "sequence"},
					{SourceActionID: "temp:1", TargetActionID: "temp:0", EdgeType: "sequence"}, // Creates cycle
				},
			},
			GotResp: &workflowsaveapp.ValidationResult{},
			ExpResp: &workflowsaveapp.ValidationResult{},
			CmpFunc: func(got any, exp any) string {
				gotResp, ok := got.(*workflowsaveapp.ValidationResult)
				if !ok {
					return "error casting response"
				}
				if gotResp.Valid {
					return "expected valid=false for workflow with cycle"
				}
				if len(gotResp.Errors) == 0 {
					return "expected validation errors"
				}

				return ""
			},
		},
	}
}
