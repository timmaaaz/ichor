package workflowsaveapi_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/workflow/workflowsaveapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

// =============================================================================
// Create Workflow Tests (POST /v1/workflow/rules/full)
// =============================================================================

// create200Basic tests basic workflow creation with 1 action and 1 start edge.
func create200Basic(sd SaveSeedData) []apitest.Table {
	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 {
		return nil
	}

	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/workflow/rules/full",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          "Test Workflow Basic",
				Description:   "A test workflow",
				IsActive:      true,
				EntityID:      sd.Entities[0].ID.String(),
				TriggerTypeID: sd.TriggerTypes[0].ID.String(),
				Actions: []workflowsaveapp.SaveActionRequest{
					{
						Name:           "Create Alert",
						Description:    "Creates an alert",
						ActionType:     "create_alert",
						IsActive:       true,
						ActionConfig:   json.RawMessage(`{"alert_type":"test","severity":"info","title":"Test Alert","message":"Test message"}`),
					},
				},
				Edges: []workflowsaveapp.SaveEdgeRequest{
					{TargetActionID: "temp:0", EdgeType: "start", EdgeOrder: 0},
				},
			},
			GotResp: &workflowsaveapp.SaveWorkflowResponse{},
			ExpResp: &workflowsaveapp.SaveWorkflowResponse{
				Name:        "Test Workflow Basic",
				Description: "A test workflow",
				IsActive:    true,
				Actions: []workflowsaveapp.SaveActionResponse{
					{
						Name:           "Create Alert",
						Description:    "Creates an alert",
						ActionType:     "create_alert",
						IsActive:       true,
					},
				},
				Edges: []workflowsaveapp.SaveEdgeResponse{
					{EdgeType: "start", EdgeOrder: 0},
				},
			},
			CmpFunc: cmpCreateResponse,
		},
	}
}

// create200WithSequence tests workflow creation with 3 actions in sequence.
func create200WithSequence(sd SaveSeedData) []apitest.Table {
	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 {
		return nil
	}

	return []apitest.Table{
		{
			Name:       "with-sequence",
			URL:        "/v1/workflow/rules/full",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          "Sequence Workflow",
				Description:   "A workflow with sequential actions",
				IsActive:      true,
				EntityID:      sd.Entities[0].ID.String(),
				TriggerTypeID: sd.TriggerTypes[0].ID.String(),
				Actions: []workflowsaveapp.SaveActionRequest{
					{
						Name:           "Action 1",
						ActionType:     "create_alert",
						IsActive:       true,
						ActionConfig:   json.RawMessage(`{"alert_type":"step1","severity":"info","title":"Step 1","message":"First step"}`),
					},
					{
						Name:           "Action 2",
						ActionType:     "create_alert",
						IsActive:       true,
						ActionConfig:   json.RawMessage(`{"alert_type":"step2","severity":"info","title":"Step 2","message":"Second step"}`),
					},
					{
						Name:           "Action 3",
						ActionType:     "create_alert",
						IsActive:       true,
						ActionConfig:   json.RawMessage(`{"alert_type":"step3","severity":"info","title":"Step 3","message":"Third step"}`),
					},
				},
				Edges: []workflowsaveapp.SaveEdgeRequest{
					{TargetActionID: "temp:0", EdgeType: "start", EdgeOrder: 0},
					{SourceActionID: "temp:0", TargetActionID: "temp:1", EdgeType: "sequence", EdgeOrder: 1},
					{SourceActionID: "temp:1", TargetActionID: "temp:2", EdgeType: "sequence", EdgeOrder: 2},
				},
			},
			GotResp: &workflowsaveapp.SaveWorkflowResponse{},
			ExpResp: &workflowsaveapp.SaveWorkflowResponse{
				Name:        "Sequence Workflow",
				Description: "A workflow with sequential actions",
				IsActive:    true,
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, ok := got.(*workflowsaveapp.SaveWorkflowResponse)
				if !ok {
					return "failed to cast got response"
				}

				// Verify 3 actions created
				if len(gotResp.Actions) != 3 {
					return fmt.Sprintf("expected 3 actions, got %d", len(gotResp.Actions))
				}

				// Verify 3 edges created
				if len(gotResp.Edges) != 3 {
					return fmt.Sprintf("expected 3 edges, got %d", len(gotResp.Edges))
				}

				// Verify IDs are assigned
				if gotResp.ID == "" {
					return "rule ID should not be empty"
				}
				for i, action := range gotResp.Actions {
					if action.ID == "" {
						return fmt.Sprintf("action[%d] ID should not be empty", i)
					}
				}

				return ""
			},
		},
	}
}

// create200WithBranch tests workflow creation with evaluate_condition branching.
func create200WithBranch(sd SaveSeedData) []apitest.Table {
	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 {
		return nil
	}

	return []apitest.Table{
		{
			Name:       "with-branch",
			URL:        "/v1/workflow/rules/full",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          "Branching Workflow",
				Description:   "A workflow with conditional branching",
				IsActive:      true,
				EntityID:      sd.Entities[0].ID.String(),
				TriggerTypeID: sd.TriggerTypes[0].ID.String(),
				Actions: []workflowsaveapp.SaveActionRequest{
					{
						Name:           "Evaluate Amount",
						ActionType:     "evaluate_condition",
						IsActive:       true,
						ActionConfig:   json.RawMessage(`{"conditions":[{"field":"amount","operator":"greater_than","value":1000}]}`),
					},
					{
						Name:           "High Value Alert",
						ActionType:     "create_alert",
						IsActive:       true,
						ActionConfig:   json.RawMessage(`{"alert_type":"high_value","severity":"warning","title":"High Value","message":"Amount exceeds threshold"}`),
					},
					{
						Name:           "Normal Alert",
						ActionType:     "create_alert",
						IsActive:       true,
						ActionConfig:   json.RawMessage(`{"alert_type":"normal","severity":"info","title":"Normal","message":"Standard processing"}`),
					},
				},
				Edges: []workflowsaveapp.SaveEdgeRequest{
					{TargetActionID: "temp:0", EdgeType: "start", EdgeOrder: 0},
					{SourceActionID: "temp:0", TargetActionID: "temp:1", EdgeType: "sequence", SourceOutput: "true", EdgeOrder: 1},
					{SourceActionID: "temp:0", TargetActionID: "temp:2", EdgeType: "sequence", SourceOutput: "false", EdgeOrder: 2},
				},
			},
			GotResp: &workflowsaveapp.SaveWorkflowResponse{},
			ExpResp: &workflowsaveapp.SaveWorkflowResponse{},
			CmpFunc: func(got any, exp any) string {
				gotResp, ok := got.(*workflowsaveapp.SaveWorkflowResponse)
				if !ok {
					return "failed to cast got response"
				}

				// Verify 3 actions created
				if len(gotResp.Actions) != 3 {
					return fmt.Sprintf("expected 3 actions, got %d", len(gotResp.Actions))
				}

				// Verify branching edges exist (sequence edges with source_output)
				hasTrueOutput := false
				hasFalseOutput := false
				for _, edge := range gotResp.Edges {
					if edge.SourceOutput == "true" {
						hasTrueOutput = true
					}
					if edge.SourceOutput == "false" {
						hasFalseOutput = true
					}
				}
				if !hasTrueOutput {
					return "expected edge with source_output=true"
				}
				if !hasFalseOutput {
					return "expected edge with source_output=false"
				}

				return ""
			},
		},
	}
}

// create200WithCanvasLayout tests that canvas_layout is saved and returned.
func create200WithCanvasLayout(sd SaveSeedData) []apitest.Table {
	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 {
		return nil
	}

	canvasLayout := json.RawMessage(`{"viewport":{"x":0,"y":0,"zoom":1},"node_positions":{"action1":{"x":100,"y":100}}}`)

	return []apitest.Table{
		{
			Name:       "with-canvas-layout",
			URL:        "/v1/workflow/rules/full",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          "Canvas Layout Workflow",
				Description:   "A workflow with canvas layout",
				IsActive:      true,
				EntityID:      sd.Entities[0].ID.String(),
				TriggerTypeID: sd.TriggerTypes[0].ID.String(),
				CanvasLayout:  canvasLayout,
				Actions: []workflowsaveapp.SaveActionRequest{
					{
						Name:           "Create Alert",
						ActionType:     "create_alert",
						IsActive:       true,
						ActionConfig:   json.RawMessage(`{"alert_type":"test","severity":"info","title":"Test","message":"Test"}`),
					},
				},
				Edges: []workflowsaveapp.SaveEdgeRequest{
					{TargetActionID: "temp:0", EdgeType: "start", EdgeOrder: 0},
				},
			},
			GotResp: &workflowsaveapp.SaveWorkflowResponse{},
			ExpResp: &workflowsaveapp.SaveWorkflowResponse{},
			CmpFunc: func(got any, exp any) string {
				gotResp, ok := got.(*workflowsaveapp.SaveWorkflowResponse)
				if !ok {
					return "failed to cast got response"
				}

				// Verify canvas layout is returned
				if len(gotResp.CanvasLayout) == 0 {
					return "canvas_layout should not be empty"
				}

				// Verify it contains expected structure
				var layout map[string]interface{}
				if err := json.Unmarshal(gotResp.CanvasLayout, &layout); err != nil {
					return fmt.Sprintf("failed to parse canvas_layout: %v", err)
				}

				if _, ok := layout["viewport"]; !ok {
					return "canvas_layout should contain viewport"
				}

				return ""
			},
		},
	}
}

// create200TempIDResolution tests that temp:N references are resolved to real UUIDs.
func create200TempIDResolution(sd SaveSeedData) []apitest.Table {
	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 {
		return nil
	}

	return []apitest.Table{
		{
			Name:       "temp-id-resolution",
			URL:        "/v1/workflow/rules/full",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          "Temp ID Test",
				Description:   "Tests temp ID resolution",
				IsActive:      true,
				EntityID:      sd.Entities[0].ID.String(),
				TriggerTypeID: sd.TriggerTypes[0].ID.String(),
				Actions: []workflowsaveapp.SaveActionRequest{
					{
						Name:           "First Action",
						ActionType:     "create_alert",
						IsActive:       true,
						ActionConfig:   json.RawMessage(`{"alert_type":"test","severity":"info","title":"Test","message":"First"}`),
					},
					{
						Name:           "Second Action",
						ActionType:     "create_alert",
						IsActive:       true,
						ActionConfig:   json.RawMessage(`{"alert_type":"test","severity":"info","title":"Test","message":"Second"}`),
					},
				},
				Edges: []workflowsaveapp.SaveEdgeRequest{
					{TargetActionID: "temp:0", EdgeType: "start", EdgeOrder: 0},
					{SourceActionID: "temp:0", TargetActionID: "temp:1", EdgeType: "sequence", EdgeOrder: 1},
				},
			},
			GotResp: &workflowsaveapp.SaveWorkflowResponse{},
			ExpResp: &workflowsaveapp.SaveWorkflowResponse{},
			CmpFunc: func(got any, exp any) string {
				gotResp, ok := got.(*workflowsaveapp.SaveWorkflowResponse)
				if !ok {
					return "failed to cast response"
				}

				// Verify all IDs are assigned (not empty)
				if gotResp.ID == "" {
					return "rule ID should not be empty"
				}

				// Build map of action indices to their assigned UUIDs
				actionIDMap := make(map[int]string)
				for i, action := range gotResp.Actions {
					if action.ID == "" {
						return fmt.Sprintf("action[%d] ID should not be empty", i)
					}
					actionIDMap[i] = action.ID
				}

				// Verify edges have resolved UUIDs (no temp: prefix)
				for i, edge := range gotResp.Edges {
					if edge.ID == "" {
						return fmt.Sprintf("edge[%d] ID should not be empty", i)
					}
					if strings.HasPrefix(edge.TargetActionID, "temp:") {
						return fmt.Sprintf("edge[%d] target not resolved: %s", i, edge.TargetActionID)
					}
					if edge.SourceActionID != "" && strings.HasPrefix(edge.SourceActionID, "temp:") {
						return fmt.Sprintf("edge[%d] source not resolved: %s", i, edge.SourceActionID)
					}

					// Verify the resolved UUID matches an actual action
					if edge.TargetActionID != "" {
						found := false
						for _, actionID := range actionIDMap {
							if edge.TargetActionID == actionID {
								found = true
								break
							}
						}
						if !found {
							return fmt.Sprintf("edge[%d] target %s doesn't match any action", i, edge.TargetActionID)
						}
					}
				}

				return ""
			},
		},
	}
}

// create400 tests validation error responses for create endpoint.
func create400(sd SaveSeedData) []apitest.Table {
	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 {
		return nil
	}

	return []apitest.Table{
		{
			Name:       "missing-name",
			URL:        "/v1/workflow/rules/full",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPost,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          "", // Missing required field
				EntityID:      sd.Entities[0].ID.String(),
				TriggerTypeID: sd.TriggerTypes[0].ID.String(),
				Actions: []workflowsaveapp.SaveActionRequest{
					{Name: "Action", ActionType: "create_alert", IsActive: true,
						ActionConfig: json.RawMessage(`{"alert_type":"test","severity":"info","title":"T","message":"M"}`)},
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
		{
			Name:       "missing-trigger-type",
			URL:        "/v1/workflow/rules/full",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPost,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          "Test Rule",
				EntityID:      sd.Entities[0].ID.String(),
				TriggerTypeID: "", // Missing
				Actions: []workflowsaveapp.SaveActionRequest{
					{Name: "Action", ActionType: "create_alert", IsActive: true,
						ActionConfig: json.RawMessage(`{"alert_type":"test","severity":"info","title":"T","message":"M"}`)},
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
		{
			Name:       "invalid-action-type",
			URL:        "/v1/workflow/rules/full",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPost,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          "Test Rule",
				EntityID:      sd.Entities[0].ID.String(),
				TriggerTypeID: sd.TriggerTypes[0].ID.String(),
				Actions: []workflowsaveapp.SaveActionRequest{
					{Name: "Action", ActionType: "invalid_type", IsActive: true,
						ActionConfig: json.RawMessage(`{"some":"config"}`)},
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
		{
			Name:       "invalid-action-config",
			URL:        "/v1/workflow/rules/full",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPost,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          "Test Rule",
				EntityID:      sd.Entities[0].ID.String(),
				TriggerTypeID: sd.TriggerTypes[0].ID.String(),
				Actions: []workflowsaveapp.SaveActionRequest{
					{Name: "Action", ActionType: "create_alert", IsActive: true,
						ActionConfig: json.RawMessage(`{"alert_type":"test"}`)}, // Missing required fields
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
		{
			Name:       "invalid-temp-id",
			URL:        "/v1/workflow/rules/full",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPost,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          "Test Rule",
				EntityID:      sd.Entities[0].ID.String(),
				TriggerTypeID: sd.TriggerTypes[0].ID.String(),
				Actions: []workflowsaveapp.SaveActionRequest{
					{Name: "Action", ActionType: "create_alert", IsActive: true,
						ActionConfig: json.RawMessage(`{"alert_type":"test","severity":"info","title":"T","message":"M"}`)},
				},
				Edges: []workflowsaveapp.SaveEdgeRequest{
					{TargetActionID: "temp:999", EdgeType: "start"}, // Invalid temp ID
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

// create401 tests unauthorized access to create endpoint.
func create401(sd SaveSeedData) []apitest.Table {
	if len(sd.Entities) == 0 || len(sd.TriggerTypes) == 0 {
		return nil
	}

	return []apitest.Table{
		{
			Name:       "no-token",
			URL:        "/v1/workflow/rules/full",
			Token:      "",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodPost,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          "Test Workflow",
				EntityID:      sd.Entities[0].ID.String(),
				TriggerTypeID: sd.TriggerTypes[0].ID.String(),
				Actions: []workflowsaveapp.SaveActionRequest{
					{Name: "Action", ActionType: "create_alert", IsActive: true,
						ActionConfig: json.RawMessage(`{"alert_type":"test","severity":"info","title":"T","message":"M"}`)},
				},
				Edges: []workflowsaveapp.SaveEdgeRequest{{TargetActionID: "temp:0", EdgeType: "start"}},
			},
			GotResp: &map[string]any{},
			ExpResp: &map[string]any{},
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		},
	}
}

// cmpCreateResponse is the standard comparison function for successful create responses.
// It syncs server-generated fields from got to exp before comparing.
func cmpCreateResponse(got any, exp any) string {
	gotResp, ok := got.(*workflowsaveapp.SaveWorkflowResponse)
	if !ok {
		return "failed to cast got response"
	}
	expResp, ok := exp.(*workflowsaveapp.SaveWorkflowResponse)
	if !ok {
		return "failed to cast exp response"
	}

	// Sync server-generated fields from got to exp
	expResp.ID = gotResp.ID
	expResp.CreatedDate = gotResp.CreatedDate
	expResp.UpdatedDate = gotResp.UpdatedDate
	expResp.EntityID = gotResp.EntityID
	expResp.TriggerTypeID = gotResp.TriggerTypeID
	expResp.TriggerConditions = gotResp.TriggerConditions
	expResp.CanvasLayout = gotResp.CanvasLayout

	// Sync action IDs (server-generated)
	for i := range gotResp.Actions {
		if i < len(expResp.Actions) {
			expResp.Actions[i].ID = gotResp.Actions[i].ID
			expResp.Actions[i].ActionConfig = gotResp.Actions[i].ActionConfig
		}
	}

	// Sync edge IDs and resolved action references
	for i := range gotResp.Edges {
		if i < len(expResp.Edges) {
			expResp.Edges[i].ID = gotResp.Edges[i].ID
			// temp:N references are resolved to real UUIDs
			expResp.Edges[i].SourceActionID = gotResp.Edges[i].SourceActionID
			expResp.Edges[i].TargetActionID = gotResp.Edges[i].TargetActionID
		}
	}

	return cmp.Diff(gotResp, expResp)
}
