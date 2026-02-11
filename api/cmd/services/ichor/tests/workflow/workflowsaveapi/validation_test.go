package workflowsaveapi_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/workflow/workflowsaveapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

// =============================================================================
// Action Config Validation Tests
// =============================================================================

// validationActionConfig tests action config validation for various action types.
func validationActionConfig(sd SaveSeedData) []apitest.Table {
	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 {
		return nil
	}

	return []apitest.Table{
		// =========================================================
		// create_alert validation
		// =========================================================
		{
			Name:       "create-alert-valid",
			URL:        "/v1/workflow/rules/full",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          "Valid Create Alert",
				IsActive:      true,
				EntityID:      sd.Entities[0].ID.String(),
				TriggerTypeID: sd.TriggerTypes[0].ID.String(),
				Actions: []workflowsaveapp.SaveActionRequest{
					{
						Name:           "Create Alert",
						ActionType:     "create_alert",
						IsActive:       true,
						ActionConfig:   json.RawMessage(`{"alert_type":"test","severity":"warning","title":"Test Alert","message":"Test message"}`),
					},
				},
				Edges: []workflowsaveapp.SaveEdgeRequest{{TargetActionID: "temp:0", EdgeType: "start"}},
			},
			GotResp: &workflowsaveapp.SaveWorkflowResponse{},
			ExpResp: &workflowsaveapp.SaveWorkflowResponse{},
			CmpFunc: func(got any, exp any) string {
				gotResp, ok := got.(*workflowsaveapp.SaveWorkflowResponse)
				if !ok {
					return "failed to cast response"
				}
				if gotResp.ID == "" {
					return "expected workflow to be created"
				}
				return ""
			},
		},
		{
			Name:       "create-alert-missing-severity",
			URL:        "/v1/workflow/rules/full",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPost,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          "Invalid Create Alert",
				IsActive:      true,
				EntityID:      sd.Entities[0].ID.String(),
				TriggerTypeID: sd.TriggerTypes[0].ID.String(),
				Actions: []workflowsaveapp.SaveActionRequest{
					{
						Name:           "Create Alert",
						ActionType:     "create_alert",

						IsActive:       true,
						ActionConfig:   json.RawMessage(`{"alert_type":"test","title":"Test","message":"Test"}`), // Missing severity
					},
				},
				Edges: []workflowsaveapp.SaveEdgeRequest{{TargetActionID: "temp:0", EdgeType: "start"}},
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				gotErr, ok := got.(*errs.Error)
				if !ok {
					return "failed to cast to error"
				}
				if gotErr.Code != errs.InvalidArgument {
					return fmt.Sprintf("expected InvalidArgument, got %v", gotErr.Code)
				}
				return ""
			},
		},
		// =========================================================
		// send_email validation
		// =========================================================
		{
			Name:       "send-email-valid",
			URL:        "/v1/workflow/rules/full",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          "Valid Send Email",
				IsActive:      true,
				EntityID:      sd.Entities[0].ID.String(),
				TriggerTypeID: sd.TriggerTypes[0].ID.String(),
				Actions: []workflowsaveapp.SaveActionRequest{
					{
						Name:           "Send Email",
						ActionType:     "send_email",

						IsActive:       true,
						ActionConfig:   json.RawMessage(`{"recipients":["test@example.com"],"subject":"Test Subject","body":"Test body"}`),
					},
				},
				Edges: []workflowsaveapp.SaveEdgeRequest{{TargetActionID: "temp:0", EdgeType: "start"}},
			},
			GotResp: &workflowsaveapp.SaveWorkflowResponse{},
			ExpResp: &workflowsaveapp.SaveWorkflowResponse{},
			CmpFunc: func(got any, exp any) string {
				gotResp, ok := got.(*workflowsaveapp.SaveWorkflowResponse)
				if !ok {
					return "failed to cast response"
				}
				if gotResp.ID == "" {
					return "expected workflow to be created"
				}
				return ""
			},
		},
		{
			Name:       "send-email-missing-recipients",
			URL:        "/v1/workflow/rules/full",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPost,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          "Invalid Send Email",
				IsActive:      true,
				EntityID:      sd.Entities[0].ID.String(),
				TriggerTypeID: sd.TriggerTypes[0].ID.String(),
				Actions: []workflowsaveapp.SaveActionRequest{
					{
						Name:           "Send Email",
						ActionType:     "send_email",

						IsActive:       true,
						ActionConfig:   json.RawMessage(`{"subject":"Test Subject","body":"Test body"}`), // Missing recipients
					},
				},
				Edges: []workflowsaveapp.SaveEdgeRequest{{TargetActionID: "temp:0", EdgeType: "start"}},
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				gotErr, ok := got.(*errs.Error)
				if !ok {
					return "failed to cast to error"
				}
				if gotErr.Code != errs.InvalidArgument {
					return fmt.Sprintf("expected InvalidArgument, got %v", gotErr.Code)
				}
				return ""
			},
		},
		// =========================================================
		// evaluate_condition validation
		// =========================================================
		{
			Name:       "evaluate-condition-valid",
			URL:        "/v1/workflow/rules/full",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          "Valid Evaluate Condition",
				IsActive:      true,
				EntityID:      sd.Entities[0].ID.String(),
				TriggerTypeID: sd.TriggerTypes[0].ID.String(),
				Actions: []workflowsaveapp.SaveActionRequest{
					{
						Name:           "Evaluate Condition",
						ActionType:     "evaluate_condition",

						IsActive:       true,
						ActionConfig:   json.RawMessage(`{"conditions":[{"field":"status","operator":"equals","value":"active"}]}`),
					},
					{
						Name:           "True Path",
						ActionType:     "create_alert",

						IsActive:       true,
						ActionConfig:   json.RawMessage(`{"alert_type":"test","severity":"info","title":"True","message":"Condition was true"}`),
					},
					{
						Name:           "False Path",
						ActionType:     "create_alert",

						IsActive:       true,
						ActionConfig:   json.RawMessage(`{"alert_type":"test","severity":"info","title":"False","message":"Condition was false"}`),
					},
				},
				Edges: []workflowsaveapp.SaveEdgeRequest{
					{TargetActionID: "temp:0", EdgeType: "start"},
					{SourceActionID: "temp:0", TargetActionID: "temp:1", EdgeType: "sequence", SourceOutput: "true"},
					{SourceActionID: "temp:0", TargetActionID: "temp:2", EdgeType: "sequence", SourceOutput: "false"},
				},
			},
			GotResp: &workflowsaveapp.SaveWorkflowResponse{},
			ExpResp: &workflowsaveapp.SaveWorkflowResponse{},
			CmpFunc: func(got any, exp any) string {
				gotResp, ok := got.(*workflowsaveapp.SaveWorkflowResponse)
				if !ok {
					return "failed to cast response"
				}
				if gotResp.ID == "" {
					return "expected workflow to be created"
				}
				return ""
			},
		},
		{
			Name:       "evaluate-condition-missing-conditions",
			URL:        "/v1/workflow/rules/full",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPost,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          "Invalid Evaluate Condition",
				IsActive:      true,
				EntityID:      sd.Entities[0].ID.String(),
				TriggerTypeID: sd.TriggerTypes[0].ID.String(),
				Actions: []workflowsaveapp.SaveActionRequest{
					{
						Name:           "Evaluate Condition",
						ActionType:     "evaluate_condition",

						IsActive:       true,
						ActionConfig:   json.RawMessage(`{}`), // Missing conditions
					},
				},
				Edges: []workflowsaveapp.SaveEdgeRequest{{TargetActionID: "temp:0", EdgeType: "start"}},
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				gotErr, ok := got.(*errs.Error)
				if !ok {
					return "failed to cast to error"
				}
				if gotErr.Code != errs.InvalidArgument {
					return fmt.Sprintf("expected InvalidArgument, got %v", gotErr.Code)
				}
				return ""
			},
		},
		// =========================================================
		// update_field validation
		// =========================================================
		{
			Name:       "update-field-missing-target-entity",
			URL:        "/v1/workflow/rules/full",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPost,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          "Invalid Update Field",
				IsActive:      true,
				EntityID:      sd.Entities[0].ID.String(),
				TriggerTypeID: sd.TriggerTypes[0].ID.String(),
				Actions: []workflowsaveapp.SaveActionRequest{
					{
						Name:           "Update Field",
						ActionType:     "update_field",

						IsActive:       true,
						ActionConfig:   json.RawMessage(`{"target_field":"status"}`), // Missing target_entity
					},
				},
				Edges: []workflowsaveapp.SaveEdgeRequest{{TargetActionID: "temp:0", EdgeType: "start"}},
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				gotErr, ok := got.(*errs.Error)
				if !ok {
					return "failed to cast to error"
				}
				if gotErr.Code != errs.InvalidArgument {
					return fmt.Sprintf("expected InvalidArgument, got %v", gotErr.Code)
				}
				return ""
			},
		},
	}
}

// =============================================================================
// Graph Validation Tests
// =============================================================================

// validationGraph tests graph structure validation (cycles, start edges, reachability).
func validationGraph(sd SaveSeedData) []apitest.Table {
	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 {
		return nil
	}

	return []apitest.Table{
		// =========================================================
		// Start edge validation
		// =========================================================
		{
			Name:       "multi-start-edges",
			URL:        "/v1/workflow/rules/full",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPost,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          "Multi Start Edges",
				IsActive:      true,
				EntityID:      sd.Entities[0].ID.String(),
				TriggerTypeID: sd.TriggerTypes[0].ID.String(),
				Actions: []workflowsaveapp.SaveActionRequest{
					{Name: "Action 1", ActionType: "create_alert", IsActive: true,
						ActionConfig: json.RawMessage(`{"alert_type":"test","severity":"info","title":"T","message":"M"}`)},
					{Name: "Action 2", ActionType: "create_alert", IsActive: true,
						ActionConfig: json.RawMessage(`{"alert_type":"test","severity":"info","title":"T","message":"M"}`)},
				},
				Edges: []workflowsaveapp.SaveEdgeRequest{
					{TargetActionID: "temp:0", EdgeType: "start"},
					{TargetActionID: "temp:1", EdgeType: "start"}, // Second start edge - not allowed
				},
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				gotErr, ok := got.(*errs.Error)
				if !ok {
					return "failed to cast to error"
				}
				if gotErr.Code != errs.InvalidArgument {
					return fmt.Sprintf("expected InvalidArgument, got %v", gotErr.Code)
				}
				return ""
			},
		},
		{
			Name:       "no-start-edge",
			URL:        "/v1/workflow/rules/full",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPost,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          "No Start Edge",
				IsActive:      true,
				EntityID:      sd.Entities[0].ID.String(),
				TriggerTypeID: sd.TriggerTypes[0].ID.String(),
				Actions: []workflowsaveapp.SaveActionRequest{
					{Name: "Action 1", ActionType: "create_alert", IsActive: true,
						ActionConfig: json.RawMessage(`{"alert_type":"test","severity":"info","title":"T","message":"M"}`)},
					{Name: "Action 2", ActionType: "create_alert", IsActive: true,
						ActionConfig: json.RawMessage(`{"alert_type":"test","severity":"info","title":"T","message":"M"}`)},
				},
				Edges: []workflowsaveapp.SaveEdgeRequest{
					// No start edge, only sequence
					{SourceActionID: "temp:0", TargetActionID: "temp:1", EdgeType: "sequence"},
				},
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				gotErr, ok := got.(*errs.Error)
				if !ok {
					return "failed to cast to error"
				}
				if gotErr.Code != errs.InvalidArgument {
					return fmt.Sprintf("expected InvalidArgument, got %v", gotErr.Code)
				}
				return ""
			},
		},
		// =========================================================
		// Cycle detection
		// =========================================================
		{
			Name:       "graph-cycle",
			URL:        "/v1/workflow/rules/full",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPost,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          "Cycle Test",
				IsActive:      true,
				EntityID:      sd.Entities[0].ID.String(),
				TriggerTypeID: sd.TriggerTypes[0].ID.String(),
				Actions: []workflowsaveapp.SaveActionRequest{
					{Name: "Action 1", ActionType: "create_alert", IsActive: true,
						ActionConfig: json.RawMessage(`{"alert_type":"test","severity":"info","title":"T","message":"M"}`)},
					{Name: "Action 2", ActionType: "create_alert", IsActive: true,
						ActionConfig: json.RawMessage(`{"alert_type":"test","severity":"info","title":"T","message":"M"}`)},
				},
				Edges: []workflowsaveapp.SaveEdgeRequest{
					{TargetActionID: "temp:0", EdgeType: "start"},
					{SourceActionID: "temp:0", TargetActionID: "temp:1", EdgeType: "sequence"},
					{SourceActionID: "temp:1", TargetActionID: "temp:0", EdgeType: "sequence"}, // Creates cycle
				},
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				gotErr, ok := got.(*errs.Error)
				if !ok {
					return "failed to cast to error"
				}
				if gotErr.Code != errs.InvalidArgument {
					return fmt.Sprintf("expected InvalidArgument, got %v", gotErr.Code)
				}
				return ""
			},
		},
		{
			Name:       "graph-self-cycle",
			URL:        "/v1/workflow/rules/full",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPost,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          "Self Cycle Test",
				IsActive:      true,
				EntityID:      sd.Entities[0].ID.String(),
				TriggerTypeID: sd.TriggerTypes[0].ID.String(),
				Actions: []workflowsaveapp.SaveActionRequest{
					{Name: "Action 1", ActionType: "create_alert", IsActive: true,
						ActionConfig: json.RawMessage(`{"alert_type":"test","severity":"info","title":"T","message":"M"}`)},
				},
				Edges: []workflowsaveapp.SaveEdgeRequest{
					{TargetActionID: "temp:0", EdgeType: "start"},
					{SourceActionID: "temp:0", TargetActionID: "temp:0", EdgeType: "sequence"}, // Self-loop
				},
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				gotErr, ok := got.(*errs.Error)
				if !ok {
					return "failed to cast to error"
				}
				if gotErr.Code != errs.InvalidArgument {
					return fmt.Sprintf("expected InvalidArgument, got %v", gotErr.Code)
				}
				return ""
			},
		},
		// =========================================================
		// Reachability validation
		// =========================================================
		{
			Name:       "graph-unreachable",
			URL:        "/v1/workflow/rules/full",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPost,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          "Unreachable Action Test",
				IsActive:      true,
				EntityID:      sd.Entities[0].ID.String(),
				TriggerTypeID: sd.TriggerTypes[0].ID.String(),
				Actions: []workflowsaveapp.SaveActionRequest{
					{Name: "Action 1", ActionType: "create_alert", IsActive: true,
						ActionConfig: json.RawMessage(`{"alert_type":"test","severity":"info","title":"T","message":"M"}`)},
					{Name: "Action 2 - Unreachable", ActionType: "create_alert", IsActive: true,
						ActionConfig: json.RawMessage(`{"alert_type":"test","severity":"info","title":"T","message":"M"}`)},
				},
				Edges: []workflowsaveapp.SaveEdgeRequest{
					{TargetActionID: "temp:0", EdgeType: "start"},
					// No edge to temp:1, making it unreachable
				},
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				gotErr, ok := got.(*errs.Error)
				if !ok {
					return "failed to cast to error"
				}
				if gotErr.Code != errs.InvalidArgument {
					return fmt.Sprintf("expected InvalidArgument, got %v", gotErr.Code)
				}
				return ""
			},
		},
		// =========================================================
		// Transaction rollback test
		// =========================================================
		{
			Name:       "rollback-on-edge-failure",
			URL:        "/v1/workflow/rules/full",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPost,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          "Rollback Test",
				IsActive:      true,
				EntityID:      sd.Entities[0].ID.String(),
				TriggerTypeID: sd.TriggerTypes[0].ID.String(),
				Actions: []workflowsaveapp.SaveActionRequest{
					{Name: "Action", ActionType: "create_alert", IsActive: true,
						ActionConfig: json.RawMessage(`{"alert_type":"test","severity":"info","title":"T","message":"M"}`)},
				},
				Edges: []workflowsaveapp.SaveEdgeRequest{
					{TargetActionID: "temp:999", EdgeType: "start"}, // Invalid temp ID - will fail
				},
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				gotErr, ok := got.(*errs.Error)
				if !ok {
					return "failed to cast to error"
				}
				if gotErr.Code != errs.InvalidArgument {
					return fmt.Sprintf("expected InvalidArgument, got %v", gotErr.Code)
				}
				return ""
			},
		},
	}
}

// =============================================================================
// Edge Requirement Validation Tests
// =============================================================================

// validationEdgeRequirement tests the Phase 1 validation matrix:
//   - Actions without edges → rejected (InvalidArgument)
//   - No actions, no edges  → allowed (draft workflow, forced inactive)
func validationEdgeRequirement(sd SaveSeedData) []apitest.Table {
	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 {
		return nil
	}

	return []apitest.Table{
		{
			Name:       "actions-without-edges-rejected",
			URL:        "/v1/workflow/rules/full",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPost,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          "Actions Without Edges",
				IsActive:      true,
				EntityID:      sd.Entities[0].ID.String(),
				TriggerTypeID: sd.TriggerTypes[0].ID.String(),
				Actions: []workflowsaveapp.SaveActionRequest{
					{
						Name:         "Orphan Action",
						ActionType:   "create_alert",
						IsActive:     true,
						ActionConfig: json.RawMessage(`{"alert_type":"test","severity":"info","title":"Test","message":"This should fail"}`),
					},
				},
				// No Edges field — actions exist but no edges
			},
			GotResp: &errs.Error{},
			ExpResp: &errs.Error{},
			CmpFunc: func(got any, exp any) string {
				gotErr, ok := got.(*errs.Error)
				if !ok {
					return "failed to cast to error"
				}
				if gotErr.Code != errs.InvalidArgument {
					return fmt.Sprintf("expected InvalidArgument, got %v", gotErr.Code)
				}
				if !strings.Contains(strings.ToLower(gotErr.Error()), "edge") {
					return "expected error message to mention 'edge'"
				}
				return ""
			},
		},
		{
			Name:       "no-actions-no-edges-allowed",
			URL:        "/v1/workflow/rules/full",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          "Draft Workflow No Actions",
				IsActive:      true, // User requests active — should be forced inactive
				EntityID:      sd.Entities[0].ID.String(),
				TriggerTypeID: sd.TriggerTypes[0].ID.String(),
				// No Actions, No Edges — draft workflow
			},
			GotResp: &workflowsaveapp.SaveWorkflowResponse{},
			ExpResp: &workflowsaveapp.SaveWorkflowResponse{},
			CmpFunc: func(got any, exp any) string {
				gotResp, ok := got.(*workflowsaveapp.SaveWorkflowResponse)
				if !ok {
					return "failed to cast response"
				}
				if gotResp.ID == "" {
					return "expected draft workflow to be created"
				}
				if gotResp.IsActive {
					return "expected draft workflow to be forced inactive (is_active should be false)"
				}
				if len(gotResp.Actions) != 0 {
					return fmt.Sprintf("expected zero actions, got %d", len(gotResp.Actions))
				}
				if len(gotResp.Edges) != 0 {
					return fmt.Sprintf("expected zero edges, got %d", len(gotResp.Edges))
				}
				return ""
			},
		},
	}
}
