package workflowsaveapi_test

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/workflow/workflowsaveapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
)

// validAlertConfig returns a valid create_alert action config for testing.
func validAlertConfig(name string) json.RawMessage {
	config := map[string]interface{}{
		"alert_type": "test",
		"severity":   "info",
		"title":      name + " Alert",
		"message":    "Test message for " + name,
	}
	data, _ := json.Marshal(config)
	return data
}

// =============================================================================
// Update Workflow Tests (PUT /v1/workflow/rules/{id}/full)
// =============================================================================

// update200RuleOnly tests updating rule metadata only (keeping actions the same).
func update200RuleOnly(sd SaveSeedData) []apitest.Table {
	if sd.ExistingRule.ID == uuid.Nil || len(sd.ExistingActions) == 0 {
		return nil
	}

	// Build action requests from existing actions with valid action configs
	actionRequests := make([]workflowsaveapp.SaveActionRequest, len(sd.ExistingActions))
	for i, action := range sd.ExistingActions {
		id := action.ID.String()
		actionRequests[i] = workflowsaveapp.SaveActionRequest{
			ID:             &id,
			Name:           action.Name,
			Description:    action.Description,
			ActionType:     "create_alert",
			ActionConfig:   validAlertConfig(action.Name), // Use valid config instead of seeded config
			ExecutionOrder: action.ExecutionOrder,
			IsActive:       action.IsActive,
		}
	}

	// Build edge requests from existing edges
	edgeRequests := make([]workflowsaveapp.SaveEdgeRequest, len(sd.ExistingEdges))
	for i, edge := range sd.ExistingEdges {
		sourceID := ""
		if edge.SourceActionID != nil {
			sourceID = edge.SourceActionID.String()
		}
		edgeRequests[i] = workflowsaveapp.SaveEdgeRequest{
			SourceActionID: sourceID,
			TargetActionID: edge.TargetActionID.String(),
			EdgeType:       edge.EdgeType,
			EdgeOrder:      edge.EdgeOrder,
		}
	}

	newName := "Updated Rule Name"
	newDescription := "Updated description for the rule"

	return []apitest.Table{
		{
			Name:       "rule-only",
			URL:        "/v1/workflow/rules/" + sd.ExistingRule.ID.String() + "/full",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPut,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          newName,
				Description:   newDescription,
				IsActive:      true,
				EntityID:      sd.ExistingRule.EntityID.String(),
				TriggerTypeID: sd.ExistingRule.TriggerTypeID.String(),
				Actions:       actionRequests,
				Edges:         edgeRequests,
			},
			GotResp: &workflowsaveapp.SaveWorkflowResponse{},
			ExpResp: &workflowsaveapp.SaveWorkflowResponse{},
			CmpFunc: func(got any, exp any) string {
				gotResp, ok := got.(*workflowsaveapp.SaveWorkflowResponse)
				if !ok {
					return "failed to cast response"
				}

				if gotResp.Name != newName {
					return fmt.Sprintf("expected name %q, got %q", newName, gotResp.Name)
				}

				if gotResp.Description != newDescription {
					return fmt.Sprintf("expected description %q, got %q", newDescription, gotResp.Description)
				}

				// Actions should be preserved
				if len(gotResp.Actions) != len(sd.ExistingActions) {
					return fmt.Sprintf("expected %d actions, got %d", len(sd.ExistingActions), len(gotResp.Actions))
				}

				return ""
			},
		},
	}
}

// update200AddAction tests adding a new action to an existing workflow.
func update200AddAction(sd SaveSeedData) []apitest.Table {
	if sd.ExistingRule.ID == uuid.Nil || len(sd.ExistingActions) == 0 {
		return nil
	}

	// Build action requests including existing and one new action
	actionRequests := make([]workflowsaveapp.SaveActionRequest, len(sd.ExistingActions)+1)
	for i, action := range sd.ExistingActions {
		id := action.ID.String()
		actionRequests[i] = workflowsaveapp.SaveActionRequest{
			ID:             &id,
			Name:           action.Name,
			Description:    action.Description,
			ActionType:     "create_alert",
			ActionConfig:   validAlertConfig(action.Name),
			ExecutionOrder: action.ExecutionOrder,
			IsActive:       action.IsActive,
		}
	}

	// Add new action (nil ID means create)
	actionRequests[len(sd.ExistingActions)] = workflowsaveapp.SaveActionRequest{
		ID:             nil,
		Name:           "New Added Action",
		Description:    "A newly added action",
		ActionType:     "create_alert",
		ExecutionOrder: len(sd.ExistingActions) + 1,
		IsActive:       true,
		ActionConfig:   json.RawMessage(`{"alert_type":"new","severity":"info","title":"New","message":"New action"}`),
	}

	// Build edges: start -> action0 -> action1 -> action2 -> new_action
	edgeRequests := []workflowsaveapp.SaveEdgeRequest{
		{TargetActionID: sd.ExistingActions[0].ID.String(), EdgeType: "start", EdgeOrder: 0},
	}
	for i := 0; i < len(sd.ExistingActions)-1; i++ {
		edgeRequests = append(edgeRequests, workflowsaveapp.SaveEdgeRequest{
			SourceActionID: sd.ExistingActions[i].ID.String(),
			TargetActionID: sd.ExistingActions[i+1].ID.String(),
			EdgeType:       "sequence",
			EdgeOrder:      i + 1,
		})
	}
	// Edge from last existing to new action using temp:N
	newActionIndex := len(sd.ExistingActions)
	edgeRequests = append(edgeRequests, workflowsaveapp.SaveEdgeRequest{
		SourceActionID: sd.ExistingActions[len(sd.ExistingActions)-1].ID.String(),
		TargetActionID: fmt.Sprintf("temp:%d", newActionIndex),
		EdgeType:       "sequence",
		EdgeOrder:      len(sd.ExistingActions),
	})

	return []apitest.Table{
		{
			Name:       "add-action",
			URL:        "/v1/workflow/rules/" + sd.ExistingRule.ID.String() + "/full",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPut,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          sd.ExistingRule.Name,
				Description:   sd.ExistingRule.Description,
				IsActive:      true,
				EntityID:      sd.ExistingRule.EntityID.String(),
				TriggerTypeID: sd.ExistingRule.TriggerTypeID.String(),
				Actions:       actionRequests,
				Edges:         edgeRequests,
			},
			GotResp: &workflowsaveapp.SaveWorkflowResponse{},
			ExpResp: &workflowsaveapp.SaveWorkflowResponse{},
			CmpFunc: func(got any, exp any) string {
				gotResp, ok := got.(*workflowsaveapp.SaveWorkflowResponse)
				if !ok {
					return "failed to cast response"
				}

				expectedActionCount := len(sd.ExistingActions) + 1
				if len(gotResp.Actions) != expectedActionCount {
					return fmt.Sprintf("expected %d actions, got %d", expectedActionCount, len(gotResp.Actions))
				}

				// Verify the new action was created with an ID
				newActionFound := false
				for _, action := range gotResp.Actions {
					if action.Name == "New Added Action" {
						newActionFound = true
						if action.ID == "" {
							return "new action should have an ID"
						}
					}
				}
				if !newActionFound {
					return "new action not found in response"
				}

				return ""
			},
		},
	}
}

// update200UpdateAction tests updating an existing action within a workflow.
func update200UpdateAction(sd SaveSeedData) []apitest.Table {
	if sd.ExistingRule.ID == uuid.Nil || len(sd.ExistingActions) == 0 {
		return nil
	}

	updatedName := "Updated Action Name"
	updatedDescription := "Updated action description"

	// Build action requests with first action updated
	actionRequests := make([]workflowsaveapp.SaveActionRequest, len(sd.ExistingActions))
	for i, action := range sd.ExistingActions {
		id := action.ID.String()
		actionRequests[i] = workflowsaveapp.SaveActionRequest{
			ID:             &id,
			Name:           action.Name,
			Description:    action.Description,
			ActionType:     "create_alert",
			ActionConfig:   validAlertConfig(action.Name),
			ExecutionOrder: action.ExecutionOrder,
			IsActive:       action.IsActive,
		}
	}

	// Update first action
	actionRequests[0].Name = updatedName
	actionRequests[0].Description = updatedDescription

	// Build edges from existing
	edgeRequests := make([]workflowsaveapp.SaveEdgeRequest, len(sd.ExistingEdges))
	for i, edge := range sd.ExistingEdges {
		sourceID := ""
		if edge.SourceActionID != nil {
			sourceID = edge.SourceActionID.String()
		}
		edgeRequests[i] = workflowsaveapp.SaveEdgeRequest{
			SourceActionID: sourceID,
			TargetActionID: edge.TargetActionID.String(),
			EdgeType:       edge.EdgeType,
			EdgeOrder:      edge.EdgeOrder,
		}
	}

	return []apitest.Table{
		{
			Name:       "update-action",
			URL:        "/v1/workflow/rules/" + sd.ExistingRule.ID.String() + "/full",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPut,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          sd.ExistingRule.Name,
				Description:   sd.ExistingRule.Description,
				IsActive:      true,
				EntityID:      sd.ExistingRule.EntityID.String(),
				TriggerTypeID: sd.ExistingRule.TriggerTypeID.String(),
				Actions:       actionRequests,
				Edges:         edgeRequests,
			},
			GotResp: &workflowsaveapp.SaveWorkflowResponse{},
			ExpResp: &workflowsaveapp.SaveWorkflowResponse{},
			CmpFunc: func(got any, exp any) string {
				gotResp, ok := got.(*workflowsaveapp.SaveWorkflowResponse)
				if !ok {
					return "failed to cast response"
				}

				// Find and verify the updated action
				var updatedAction *workflowsaveapp.SaveActionResponse
				for i := range gotResp.Actions {
					if gotResp.Actions[i].ID == sd.ExistingActions[0].ID.String() {
						updatedAction = &gotResp.Actions[i]
						break
					}
				}

				if updatedAction == nil {
					return "updated action not found in response"
				}

				if updatedAction.Name != updatedName {
					return fmt.Sprintf("expected action name %q, got %q", updatedName, updatedAction.Name)
				}

				if updatedAction.Description != updatedDescription {
					return fmt.Sprintf("expected action description %q, got %q", updatedDescription, updatedAction.Description)
				}

				return ""
			},
		},
	}
}

// update200DeleteAction tests removing an action from an existing workflow.
func update200DeleteAction(sd SaveSeedData) []apitest.Table {
	if sd.ExistingRule.ID == uuid.Nil || len(sd.ExistingActions) < 2 {
		return nil
	}

	// Build action requests excluding the last action (simulating deletion)
	actionRequests := make([]workflowsaveapp.SaveActionRequest, len(sd.ExistingActions)-1)
	for i := 0; i < len(sd.ExistingActions)-1; i++ {
		action := sd.ExistingActions[i]
		id := action.ID.String()
		actionRequests[i] = workflowsaveapp.SaveActionRequest{
			ID:             &id,
			Name:           action.Name,
			Description:    action.Description,
			ActionType:     "create_alert",
			ActionConfig:   validAlertConfig(action.Name),
			ExecutionOrder: action.ExecutionOrder,
			IsActive:       action.IsActive,
		}
	}

	// Build edges without the deleted action
	edgeRequests := []workflowsaveapp.SaveEdgeRequest{
		{TargetActionID: sd.ExistingActions[0].ID.String(), EdgeType: "start", EdgeOrder: 0},
	}
	for i := 0; i < len(sd.ExistingActions)-2; i++ {
		edgeRequests = append(edgeRequests, workflowsaveapp.SaveEdgeRequest{
			SourceActionID: sd.ExistingActions[i].ID.String(),
			TargetActionID: sd.ExistingActions[i+1].ID.String(),
			EdgeType:       "sequence",
			EdgeOrder:      i + 1,
		})
	}

	deletedActionID := sd.ExistingActions[len(sd.ExistingActions)-1].ID.String()

	return []apitest.Table{
		{
			Name:       "delete-action",
			URL:        "/v1/workflow/rules/" + sd.ExistingRule.ID.String() + "/full",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPut,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          sd.ExistingRule.Name,
				Description:   sd.ExistingRule.Description,
				IsActive:      true,
				EntityID:      sd.ExistingRule.EntityID.String(),
				TriggerTypeID: sd.ExistingRule.TriggerTypeID.String(),
				Actions:       actionRequests,
				Edges:         edgeRequests,
			},
			GotResp: &workflowsaveapp.SaveWorkflowResponse{},
			ExpResp: &workflowsaveapp.SaveWorkflowResponse{},
			CmpFunc: func(got any, exp any) string {
				gotResp, ok := got.(*workflowsaveapp.SaveWorkflowResponse)
				if !ok {
					return "failed to cast response"
				}

				// Verify the deleted action is NOT in the response
				for _, action := range gotResp.Actions {
					if action.ID == deletedActionID {
						return fmt.Sprintf("action %s should have been deleted but still exists", deletedActionID)
					}
				}

				// Verify remaining actions count
				expectedCount := len(sd.ExistingActions) - 1
				if len(gotResp.Actions) != expectedCount {
					return fmt.Sprintf("expected %d actions, got %d", expectedCount, len(gotResp.Actions))
				}

				return ""
			},
		},
	}
}

// update200ReplaceEdges tests that all edges are replaced when updating.
func update200ReplaceEdges(sd SaveSeedData) []apitest.Table {
	if sd.ExistingRule.ID == uuid.Nil || len(sd.ExistingActions) < 2 {
		return nil
	}

	// Keep actions the same
	actionRequests := make([]workflowsaveapp.SaveActionRequest, len(sd.ExistingActions))
	for i, action := range sd.ExistingActions {
		id := action.ID.String()
		actionRequests[i] = workflowsaveapp.SaveActionRequest{
			ID:             &id,
			Name:           action.Name,
			Description:    action.Description,
			ActionType:     "create_alert",
			ActionConfig:   validAlertConfig(action.Name),
			ExecutionOrder: action.ExecutionOrder,
			IsActive:       action.IsActive,
		}
	}

	// Create different edge structure (reverse order)
	lastActionIdx := len(sd.ExistingActions) - 1
	edgeRequests := []workflowsaveapp.SaveEdgeRequest{
		// Start now points to last action instead of first
		{TargetActionID: sd.ExistingActions[lastActionIdx].ID.String(), EdgeType: "start", EdgeOrder: 0},
	}
	// Reverse sequence
	for i := lastActionIdx; i > 0; i-- {
		edgeRequests = append(edgeRequests, workflowsaveapp.SaveEdgeRequest{
			SourceActionID: sd.ExistingActions[i].ID.String(),
			TargetActionID: sd.ExistingActions[i-1].ID.String(),
			EdgeType:       "sequence",
			EdgeOrder:      lastActionIdx - i + 1,
		})
	}

	return []apitest.Table{
		{
			Name:       "replace-edges",
			URL:        "/v1/workflow/rules/" + sd.ExistingRule.ID.String() + "/full",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPut,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          sd.ExistingRule.Name,
				Description:   sd.ExistingRule.Description,
				IsActive:      true,
				EntityID:      sd.ExistingRule.EntityID.String(),
				TriggerTypeID: sd.ExistingRule.TriggerTypeID.String(),
				Actions:       actionRequests,
				Edges:         edgeRequests,
			},
			GotResp: &workflowsaveapp.SaveWorkflowResponse{},
			ExpResp: &workflowsaveapp.SaveWorkflowResponse{},
			CmpFunc: func(got any, exp any) string {
				gotResp, ok := got.(*workflowsaveapp.SaveWorkflowResponse)
				if !ok {
					return "failed to cast response"
				}

				// Verify edge count matches request
				if len(gotResp.Edges) != len(edgeRequests) {
					return fmt.Sprintf("expected %d edges, got %d", len(edgeRequests), len(gotResp.Edges))
				}

				// Find the start edge and verify it points to the last action
				var startEdge *workflowsaveapp.SaveEdgeResponse
				for i := range gotResp.Edges {
					if gotResp.Edges[i].EdgeType == "start" {
						startEdge = &gotResp.Edges[i]
						break
					}
				}

				if startEdge == nil {
					return "start edge not found"
				}

				lastActionID := sd.ExistingActions[lastActionIdx].ID.String()
				if startEdge.TargetActionID != lastActionID {
					return fmt.Sprintf("expected start edge to point to %s, got %s", lastActionID, startEdge.TargetActionID)
				}

				return ""
			},
		},
	}
}

// update200CanvasLayout tests updating the canvas layout.
func update200CanvasLayout(sd SaveSeedData) []apitest.Table {
	if sd.ExistingRule.ID == uuid.Nil || len(sd.ExistingActions) == 0 {
		return nil
	}

	newCanvasLayout := json.RawMessage(`{"viewport":{"x":500,"y":500,"zoom":2},"node_positions":{"updated":true}}`)

	// Build action requests from existing
	actionRequests := make([]workflowsaveapp.SaveActionRequest, len(sd.ExistingActions))
	for i, action := range sd.ExistingActions {
		id := action.ID.String()
		actionRequests[i] = workflowsaveapp.SaveActionRequest{
			ID:             &id,
			Name:           action.Name,
			Description:    action.Description,
			ActionType:     "create_alert",
			ActionConfig:   validAlertConfig(action.Name),
			ExecutionOrder: action.ExecutionOrder,
			IsActive:       action.IsActive,
		}
	}

	// Build edge requests from existing
	edgeRequests := make([]workflowsaveapp.SaveEdgeRequest, len(sd.ExistingEdges))
	for i, edge := range sd.ExistingEdges {
		sourceID := ""
		if edge.SourceActionID != nil {
			sourceID = edge.SourceActionID.String()
		}
		edgeRequests[i] = workflowsaveapp.SaveEdgeRequest{
			SourceActionID: sourceID,
			TargetActionID: edge.TargetActionID.String(),
			EdgeType:       edge.EdgeType,
			EdgeOrder:      edge.EdgeOrder,
		}
	}

	return []apitest.Table{
		{
			Name:       "canvas-layout",
			URL:        "/v1/workflow/rules/" + sd.ExistingRule.ID.String() + "/full",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPut,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          sd.ExistingRule.Name,
				Description:   sd.ExistingRule.Description,
				IsActive:      true,
				EntityID:      sd.ExistingRule.EntityID.String(),
				TriggerTypeID: sd.ExistingRule.TriggerTypeID.String(),
				CanvasLayout:  newCanvasLayout,
				Actions:       actionRequests,
				Edges:         edgeRequests,
			},
			GotResp: &workflowsaveapp.SaveWorkflowResponse{},
			ExpResp: &workflowsaveapp.SaveWorkflowResponse{},
			CmpFunc: func(got any, exp any) string {
				gotResp, ok := got.(*workflowsaveapp.SaveWorkflowResponse)
				if !ok {
					return "failed to cast response"
				}

				// Verify canvas layout is updated
				if len(gotResp.CanvasLayout) == 0 {
					return "canvas_layout should not be empty"
				}

				var layout map[string]interface{}
				if err := json.Unmarshal(gotResp.CanvasLayout, &layout); err != nil {
					return fmt.Sprintf("failed to parse canvas_layout: %v", err)
				}

				viewport, ok := layout["viewport"].(map[string]interface{})
				if !ok {
					return "viewport not found in canvas_layout"
				}

				zoom, ok := viewport["zoom"].(float64)
				if !ok || zoom != 2 {
					return fmt.Sprintf("expected zoom 2, got %v", viewport["zoom"])
				}

				return ""
			},
		},
	}
}

// update400 tests validation error responses for update endpoint.
func update400(sd SaveSeedData) []apitest.Table {
	if sd.ExistingRule.ID == uuid.Nil {
		return nil
	}

	return []apitest.Table{
		{
			Name:       "missing-name",
			URL:        "/v1/workflow/rules/" + sd.ExistingRule.ID.String() + "/full",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPut,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          "", // Missing
				EntityID:      sd.ExistingRule.EntityID.String(),
				TriggerTypeID: sd.ExistingRule.TriggerTypeID.String(),
				Actions: []workflowsaveapp.SaveActionRequest{
					{Name: "Action", ActionType: "create_alert", ExecutionOrder: 1, IsActive: true,
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
			Name:       "invalid-action-id",
			URL:        "/v1/workflow/rules/" + sd.ExistingRule.ID.String() + "/full",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPut,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          "Test Rule",
				EntityID:      sd.ExistingRule.EntityID.String(),
				TriggerTypeID: sd.ExistingRule.TriggerTypeID.String(),
				Actions: []workflowsaveapp.SaveActionRequest{
					{
						ID:             strPtr(uuid.New().String()), // ID that doesn't belong to this rule
						Name:           "Action",
						ActionType:     "create_alert",
						ExecutionOrder: 1,
						IsActive:       true,
						ActionConfig:   json.RawMessage(`{"alert_type":"test","severity":"info","title":"T","message":"M"}`),
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

// update401 tests unauthorized access to update endpoint.
func update401(sd SaveSeedData) []apitest.Table {
	if sd.ExistingRule.ID == uuid.Nil {
		return nil
	}

	return []apitest.Table{
		{
			Name:       "no-token",
			URL:        "/v1/workflow/rules/" + sd.ExistingRule.ID.String() + "/full",
			Token:      "",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodPut,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          "Test Workflow",
				EntityID:      sd.ExistingRule.EntityID.String(),
				TriggerTypeID: sd.ExistingRule.TriggerTypeID.String(),
				Actions: []workflowsaveapp.SaveActionRequest{
					{Name: "Action", ActionType: "create_alert", ExecutionOrder: 1, IsActive: true,
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

// update404 tests not found error for update endpoint.
func update404(sd SaveSeedData) []apitest.Table {
	nonExistentID := uuid.New().String()

	return []apitest.Table{
		{
			Name:       "not-found",
			URL:        "/v1/workflow/rules/" + nonExistentID + "/full",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusNotFound,
			Method:     http.MethodPut,
			Input: workflowsaveapp.SaveWorkflowRequest{
				Name:          "Test Workflow",
				EntityID:      sd.Entities[0].ID.String(),
				TriggerTypeID: sd.TriggerTypes[0].ID.String(),
				Actions: []workflowsaveapp.SaveActionRequest{
					{Name: "Action", ActionType: "create_alert", ExecutionOrder: 1, IsActive: true,
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
				if gotErr.Code != errs.NotFound {
					return fmt.Sprintf("expected NotFound, got %v", gotErr.Code)
				}
				return ""
			},
		},
	}
}

// strPtr is a helper to get a pointer to a string.
func strPtr(s string) *string {
	return &s
}
