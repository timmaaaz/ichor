package rule_test

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/domain/http/workflow/ruleapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

// =============================================================================
// Create Rule Tests

func createRule201(sd RuleSeedData) []apitest.Table {
	if len(sd.Entities) == 0 || len(sd.EntityTypes) == 0 || len(sd.TriggerTypes) == 0 {
		return nil
	}

	triggerConditions := map[string]interface{}{
		"field":    "status",
		"operator": "equals",
		"value":    "active",
	}
	triggerConditionsJSON, _ := json.Marshal(triggerConditions)

	actionConfig := map[string]interface{}{
		"type":    "send_notification",
		"message": "Test notification",
	}
	actionConfigJSON, _ := json.Marshal(actionConfig)

	newRuleID := uuid.New()

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/workflow/rules",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			Input: ruleapi.CreateRuleRequest{
				Name:              "Test Rule " + newRuleID.String()[:8],
				Description:       "A test rule for API testing",
				EntityID:          sd.Entities[0].ID,
				EntityTypeID:      sd.EntityTypes[0].ID,
				TriggerTypeID:     sd.TriggerTypes[0].ID,
				TriggerConditions: triggerConditionsJSON,
				IsActive:          true,
			},
			GotResp: &ruleapi.RuleResponse{},
			ExpResp: &ruleapi.RuleResponse{},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*ruleapi.RuleResponse)
				if !exists {
					return "error getting rule response"
				}

				if gotResp.ID == uuid.Nil {
					return "rule ID should not be nil"
				}

				if gotResp.Name == "" {
					return "rule name should not be empty"
				}

				if !gotResp.IsActive {
					return "rule should be active"
				}

				return ""
			},
		},
		{
			Name:       "with-actions",
			URL:        "/v1/workflow/rules",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			Input: ruleapi.CreateRuleRequest{
				Name:              "Rule With Actions " + newRuleID.String()[:8],
				Description:       "A rule with embedded actions",
				EntityID:          sd.Entities[0].ID,
				EntityTypeID:      sd.EntityTypes[0].ID,
				TriggerTypeID:     sd.TriggerTypes[0].ID,
				TriggerConditions: triggerConditionsJSON,
				IsActive:          true,
				Actions: []ruleapi.CreateActionInput{
					{
						Name:           "Action 1",
						Description:    "First action",
						ActionConfig:   actionConfigJSON,
						IsActive:       true,
					},
				},
			},
			GotResp: &ruleapi.RuleResponse{},
			ExpResp: &ruleapi.RuleResponse{},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*ruleapi.RuleResponse)
				if !exists {
					return "error getting rule response"
				}

				if gotResp.ID == uuid.Nil {
					return "rule ID should not be nil"
				}

				if len(gotResp.Actions) == 0 {
					return "expected rule to have actions"
				}

				return ""
			},
		},
	}

	return table
}

func createRule400(sd RuleSeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "missing-name",
			URL:        "/v1/workflow/rules",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPost,
			Input: ruleapi.CreateRuleRequest{
				Name:          "", // Missing name
				EntityID:      uuid.New(),
				EntityTypeID:  uuid.New(),
				TriggerTypeID: uuid.New(),
			},
			GotResp: &map[string]any{},
			ExpResp: &map[string]any{},
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		},
		{
			Name:       "missing-entity-id",
			URL:        "/v1/workflow/rules",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPost,
			Input: ruleapi.CreateRuleRequest{
				Name:          "Test Rule",
				EntityID:      uuid.Nil, // Missing entity ID
				EntityTypeID:  uuid.New(),
				TriggerTypeID: uuid.New(),
			},
			GotResp: &map[string]any{},
			ExpResp: &map[string]any{},
			CmpFunc: func(got any, exp any) string {
				return ""
			},
		},
		{
			Name:       "missing-trigger-type-id",
			URL:        "/v1/workflow/rules",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodPost,
			Input: ruleapi.CreateRuleRequest{
				Name:          "Test Rule",
				EntityID:      uuid.New(),
				EntityTypeID:  uuid.New(),
				TriggerTypeID: uuid.Nil, // Missing trigger type ID
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

func createRule401(sd RuleSeedData) []apitest.Table {
	if len(sd.Entities) == 0 || len(sd.EntityTypes) == 0 || len(sd.TriggerTypes) == 0 {
		return nil
	}

	table := []apitest.Table{
		{
			Name:       "no-token",
			URL:        "/v1/workflow/rules",
			Token:      "",
			StatusCode: http.StatusUnauthorized,
			Method:     http.MethodPost,
			Input: ruleapi.CreateRuleRequest{
				Name:          "Test Rule",
				EntityID:      sd.Entities[0].ID,
				EntityTypeID:  sd.EntityTypes[0].ID,
				TriggerTypeID: sd.TriggerTypes[0].ID,
				IsActive:      true,
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
