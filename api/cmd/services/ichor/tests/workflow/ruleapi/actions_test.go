package rule_test

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/domain/http/workflow/ruleapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

// =============================================================================
// Query Actions Tests

func queryActions200(sd RuleSeedData) []apitest.Table {
	if len(sd.Rules) == 0 {
		return nil
	}

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/workflow/rules/%s/actions", sd.Rules[0].ID),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &[]ruleapi.ActionResponse{},
			ExpResp:    &[]ruleapi.ActionResponse{},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*[]ruleapi.ActionResponse)
				if !exists {
					return "error getting actions response"
				}
				// Response should be an array (can be empty)
				_ = gotResp
				return ""
			},
		},
	}

	return table
}

func queryActions404(sd RuleSeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "rule-not-found",
			URL:        fmt.Sprintf("/v1/workflow/rules/%s/actions", uuid.New()),
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

func queryActions401(sd RuleSeedData) []apitest.Table {
	if len(sd.Rules) == 0 {
		return nil
	}

	table := []apitest.Table{
		{
			Name:       "no-token",
			URL:        fmt.Sprintf("/v1/workflow/rules/%s/actions", sd.Rules[0].ID),
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
// Create Action Tests

func createAction201(sd RuleSeedData) []apitest.Table {
	if len(sd.Rules) == 0 {
		return nil
	}

	actionConfig := map[string]interface{}{
		"type":    "send_notification",
		"message": "Test notification from action API",
	}
	actionConfigJSON, _ := json.Marshal(actionConfig)

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/workflow/rules/%s/actions", sd.Rules[0].ID),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			Input: ruleapi.CreateActionRequest{
				Name:           "Test Action " + uuid.New().String()[:8],
				Description:    "A test action created via API",
				ActionConfig:   actionConfigJSON,
				ExecutionOrder: 10,
				IsActive:       true,
			},
			GotResp: &ruleapi.ActionResponse{},
			ExpResp: &ruleapi.ActionResponse{},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*ruleapi.ActionResponse)
				if !exists {
					return "error getting action response"
				}

				if gotResp.ID == uuid.Nil {
					return "action ID should not be nil"
				}

				if gotResp.Name == "" {
					return "action name should not be empty"
				}

				if !gotResp.IsActive {
					return "action should be active"
				}

				return ""
			},
		},
	}

	return table
}

func createAction400(sd RuleSeedData) []apitest.Table {
	if len(sd.Rules) == 0 {
		return nil
	}

	table := []apitest.Table{
		{
			Name:       "missing-name",
			URL:        fmt.Sprintf("/v1/workflow/rules/%s/actions", sd.Rules[0].ID),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPost,
			Input: ruleapi.CreateActionRequest{
				Name:           "", // Missing name
				ActionConfig:   json.RawMessage(`{"type": "test"}`),
				ExecutionOrder: 1,
				IsActive:       true,
			},
			GotResp: &map[string]any{},
			ExpResp: &map[string]any{},
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		},
		{
			Name:       "missing-action-config",
			URL:        fmt.Sprintf("/v1/workflow/rules/%s/actions", sd.Rules[0].ID),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPost,
			Input: ruleapi.CreateActionRequest{
				Name:           "Test Action",
				ActionConfig:   nil, // Missing action config
				ExecutionOrder: 1,
				IsActive:       true,
			},
			GotResp: &map[string]any{},
			ExpResp: &map[string]any{},
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		},
		// Note: Removed "invalid-action-config" test case because it's impossible to test:
		// json.RawMessage validates on marshal, so invalid JSON can't be sent in a valid request struct.
		// In practice, invalid JSON would fail at the Decode step, not validation.
	}

	return table
}

func createAction404(sd RuleSeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "rule-not-found",
			URL:        fmt.Sprintf("/v1/workflow/rules/%s/actions", uuid.New()),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusNotFound,
			Method:     http.MethodPost,
			Input: ruleapi.CreateActionRequest{
				Name:           "Test Action",
				ActionConfig:   json.RawMessage(`{"type": "test"}`),
				ExecutionOrder: 1,
				IsActive:       true,
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

func createAction401(sd RuleSeedData) []apitest.Table {
	if len(sd.Rules) == 0 {
		return nil
	}

	table := []apitest.Table{
		{
			Name:       "no-token",
			URL:        fmt.Sprintf("/v1/workflow/rules/%s/actions", sd.Rules[0].ID),
			Token:      "",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodPost,
			Input: ruleapi.CreateActionRequest{
				Name:           "Test Action",
				ActionConfig:   json.RawMessage(`{"type": "test"}`),
				ExecutionOrder: 1,
				IsActive:       true,
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

// =============================================================================
// Update Action Tests

func updateAction200(sd RuleSeedData) []apitest.Table {
	if len(sd.Rules) == 0 || len(sd.Actions) == 0 {
		return nil
	}

	// Find an action that belongs to the first rule
	var actionForRule *ruleapi.ActionResponse
	for _, action := range sd.Actions {
		if action.AutomationRuleID == sd.Rules[0].ID {
			actionForRule = &ruleapi.ActionResponse{
				ID:     action.ID,
				RuleID: action.AutomationRuleID,
			}
			break
		}
	}
	if actionForRule == nil {
		return nil
	}

	newName := "Updated Action Name"
	newOrder := 99

	table := []apitest.Table{
		{
			Name:       "update-name",
			URL:        fmt.Sprintf("/v1/workflow/rules/%s/actions/%s", actionForRule.RuleID, actionForRule.ID),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPut,
			Input: ruleapi.UpdateActionRequest{
				Name: &newName,
			},
			GotResp: &ruleapi.ActionResponse{},
			ExpResp: &ruleapi.ActionResponse{},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*ruleapi.ActionResponse)
				if !exists {
					return "error getting action response"
				}

				if gotResp.Name != newName {
					return fmt.Sprintf("expected name %s, got %s", newName, gotResp.Name)
				}

				return ""
			},
		},
		{
			Name:       "update-execution-order",
			URL:        fmt.Sprintf("/v1/workflow/rules/%s/actions/%s", actionForRule.RuleID, actionForRule.ID),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPut,
			Input: ruleapi.UpdateActionRequest{
				ExecutionOrder: &newOrder,
			},
			GotResp: &ruleapi.ActionResponse{},
			ExpResp: &ruleapi.ActionResponse{},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*ruleapi.ActionResponse)
				if !exists {
					return "error getting action response"
				}

				if gotResp.ExecutionOrder != newOrder {
					return fmt.Sprintf("expected execution order %d, got %d", newOrder, gotResp.ExecutionOrder)
				}

				return ""
			},
		},
	}

	return table
}

func updateAction404(sd RuleSeedData) []apitest.Table {
	if len(sd.Rules) == 0 {
		return nil
	}

	// Find an action that belongs to a different rule (if we have multiple rules)
	var actionFromDifferentRule uuid.UUID
	if len(sd.Rules) > 1 && len(sd.Actions) > 0 {
		for _, action := range sd.Actions {
			if action.AutomationRuleID != sd.Rules[0].ID {
				actionFromDifferentRule = action.ID
				break
			}
		}
	}

	newName := "Updated Name"

	table := []apitest.Table{
		{
			Name:       "action-not-found",
			URL:        fmt.Sprintf("/v1/workflow/rules/%s/actions/%s", sd.Rules[0].ID, uuid.New()),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusNotFound,
			Method:     http.MethodPut,
			Input: ruleapi.UpdateActionRequest{
				Name: &newName,
			},
			GotResp: &map[string]any{},
			ExpResp: &map[string]any{},
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		},
		{
			Name:       "rule-not-found",
			URL:        fmt.Sprintf("/v1/workflow/rules/%s/actions/%s", uuid.New(), uuid.New()),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusNotFound,
			Method:     http.MethodPut,
			Input: ruleapi.UpdateActionRequest{
				Name: &newName,
			},
			GotResp: &map[string]any{},
			ExpResp: &map[string]any{},
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		},
	}

	// Add test for action belonging to different rule (if available)
	if actionFromDifferentRule != uuid.Nil {
		table = append(table, apitest.Table{
			Name:       "action-belongs-to-different-rule",
			URL:        fmt.Sprintf("/v1/workflow/rules/%s/actions/%s", sd.Rules[0].ID, actionFromDifferentRule),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusNotFound,
			Method:     http.MethodPut,
			Input: ruleapi.UpdateActionRequest{
				Name: &newName,
			},
			GotResp: &map[string]any{},
			ExpResp: &map[string]any{},
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		})
	}

	return table
}

func updateAction401(sd RuleSeedData) []apitest.Table {
	if len(sd.Rules) == 0 || len(sd.Actions) == 0 {
		return nil
	}

	newName := "Updated Name"

	table := []apitest.Table{
		{
			Name:       "no-token",
			URL:        fmt.Sprintf("/v1/workflow/rules/%s/actions/%s", sd.Rules[0].ID, sd.Actions[0].ID),
			Token:      "",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodPut,
			Input: ruleapi.UpdateActionRequest{
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

// =============================================================================
// Delete Action Tests

func deleteAction200(sd RuleSeedData) []apitest.Table {
	if len(sd.Rules) == 0 || len(sd.Actions) == 0 {
		return nil
	}

	// Find an action that belongs to the first rule
	var actionForRule *ruleapi.ActionResponse
	for _, action := range sd.Actions {
		if action.AutomationRuleID == sd.Rules[0].ID {
			actionForRule = &ruleapi.ActionResponse{
				ID:     action.ID,
				RuleID: action.AutomationRuleID,
			}
			break
		}
	}
	if actionForRule == nil {
		return nil
	}

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/workflow/rules/%s/actions/%s", actionForRule.RuleID, actionForRule.ID),
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

func deleteAction404(sd RuleSeedData) []apitest.Table {
	if len(sd.Rules) == 0 {
		return nil
	}

	// Find an action that belongs to a different rule (if we have multiple rules)
	var actionFromDifferentRule uuid.UUID
	if len(sd.Rules) > 1 && len(sd.Actions) > 0 {
		for _, action := range sd.Actions {
			if action.AutomationRuleID != sd.Rules[0].ID {
				actionFromDifferentRule = action.ID
				break
			}
		}
	}

	table := []apitest.Table{
		{
			Name:       "action-not-found",
			URL:        fmt.Sprintf("/v1/workflow/rules/%s/actions/%s", sd.Rules[0].ID, uuid.New()),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusNotFound,
			Method:     http.MethodDelete,
			GotResp:    &map[string]any{},
			ExpResp:    &map[string]any{},
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		},
		{
			Name:       "rule-not-found",
			URL:        fmt.Sprintf("/v1/workflow/rules/%s/actions/%s", uuid.New(), uuid.New()),
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

	// Add test for action belonging to different rule (if available)
	if actionFromDifferentRule != uuid.Nil {
		table = append(table, apitest.Table{
			Name:       "action-belongs-to-different-rule",
			URL:        fmt.Sprintf("/v1/workflow/rules/%s/actions/%s", sd.Rules[0].ID, actionFromDifferentRule),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusNotFound,
			Method:     http.MethodDelete,
			GotResp:    &map[string]any{},
			ExpResp:    &map[string]any{},
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		})
	}

	return table
}

func deleteAction401(sd RuleSeedData) []apitest.Table {
	if len(sd.Rules) == 0 || len(sd.Actions) == 0 {
		return nil
	}

	table := []apitest.Table{
		{
			Name:       "no-token",
			URL:        fmt.Sprintf("/v1/workflow/rules/%s/actions/%s", sd.Rules[0].ID, sd.Actions[0].ID),
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
