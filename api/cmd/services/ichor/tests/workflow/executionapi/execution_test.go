package execution_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/domain/http/workflow/executionapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/query"
)

func Test_ExecutionAPI(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_ExecutionAPI")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// Query executions tests
	test.Run(t, queryExecutions200(sd), "queryExecutions-200")
	test.Run(t, queryExecutions401(sd), "queryExecutions-401")

	// Query execution by ID tests
	test.Run(t, queryExecutionByID200(sd), "queryExecutionByID-200")
	test.Run(t, queryExecutionByID200ExecutedByName(sd), "queryExecutionByID-200-executedByName")
	test.Run(t, queryExecutionByID404(sd), "queryExecutionByID-404")
	test.Run(t, queryExecutionByID401(sd), "queryExecutionByID-401")
}

// =============================================================================
// Query Executions Tests

func queryExecutions200(sd ExecutionSeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic-query",
			URL:        "/v1/workflow/executions",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[executionapi.ExecutionResponse]{},
			ExpResp:    &query.Result[executionapi.ExecutionResponse]{},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*query.Result[executionapi.ExecutionResponse])
				if !exists {
					return "error getting query result"
				}

				// Should have at least the seeded executions
				if gotResp.Total < 3 {
					return fmt.Sprintf("expected at least 3 executions, got %d", gotResp.Total)
				}

				return ""
			},
		},
		{
			Name:       "query-with-pagination",
			URL:        "/v1/workflow/executions?page=1&rows=1",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[executionapi.ExecutionResponse]{},
			ExpResp:    &query.Result[executionapi.ExecutionResponse]{},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*query.Result[executionapi.ExecutionResponse])
				if !exists {
					return "error getting query result"
				}

				if len(gotResp.Items) != 1 {
					return fmt.Sprintf("expected 1 item with rows=1, got %d", len(gotResp.Items))
				}

				return ""
			},
		},
		{
			Name:       "query-with-status-filter",
			URL:        "/v1/workflow/executions?status=completed",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[executionapi.ExecutionResponse]{},
			ExpResp:    &query.Result[executionapi.ExecutionResponse]{},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*query.Result[executionapi.ExecutionResponse])
				if !exists {
					return "error getting query result"
				}

				// All returned items should have status "completed"
				for _, item := range gotResp.Items {
					if item.Status != "completed" {
						return fmt.Sprintf("expected status completed, got %s", item.Status)
					}
				}

				return ""
			},
		},
		{
			Name:       "query-with-trigger-source-filter",
			URL:        "/v1/workflow/executions?trigger_source=manual",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[executionapi.ExecutionResponse]{},
			ExpResp:    &query.Result[executionapi.ExecutionResponse]{},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*query.Result[executionapi.ExecutionResponse])
				if !exists {
					return "error getting query result"
				}

				// Should have at least 1 manual execution from seed data
				if gotResp.Total < 1 {
					return "expected at least 1 manual execution"
				}

				// All returned items should have trigger_source "manual"
				for _, item := range gotResp.Items {
					if item.TriggerSource != "manual" {
						return fmt.Sprintf("expected trigger_source manual, got %s", item.TriggerSource)
					}
				}

				return ""
			},
		},
		{
			// Scope ②: list rows project the joined rule_name. Rule-backed
			// executions must expose their rule's name; manual (rule-less)
			// executions leave it empty.
			Name:       "list-exposes-rule-name",
			URL:        "/v1/workflow/executions",
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[executionapi.ExecutionResponse]{},
			ExpResp:    &query.Result[executionapi.ExecutionResponse]{},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*query.Result[executionapi.ExecutionResponse])
				if !exists {
					return "error getting query result"
				}

				ruleNames := make(map[uuid.UUID]string, len(sd.Rules))
				for _, r := range sd.Rules {
					ruleNames[r.ID] = r.Name
				}

				foundRuleName := false
				for _, item := range gotResp.Items {
					if item.AutomationRuleID == nil {
						continue
					}
					want := ruleNames[*item.AutomationRuleID]
					if item.RuleName != want {
						return fmt.Sprintf("expected rule_name %q for rule %s, got %q", want, *item.AutomationRuleID, item.RuleName)
					}
					if item.RuleName != "" {
						foundRuleName = true
					}
				}

				if !foundRuleName {
					return "expected at least one list row to expose a non-empty rule_name"
				}

				return ""
			},
		},
	}
}

func queryExecutions401(sd ExecutionSeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "no-token",
			URL:        "/v1/workflow/executions",
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
}

// =============================================================================
// Query By ID Tests

func queryExecutionByID200(sd ExecutionSeedData) []apitest.Table {
	if len(sd.Executions) == 0 {
		return nil
	}

	return []apitest.Table{
		{
			Name:       "get-execution-detail",
			URL:        fmt.Sprintf("/v1/workflow/executions/%s", sd.Executions[0].ID),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &executionapi.ExecutionDetail{},
			ExpResp:    &executionapi.ExecutionDetail{},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*executionapi.ExecutionDetail)
				if !exists {
					return "error getting execution detail"
				}

				// Verify the ID matches
				if gotResp.ID != sd.Executions[0].ID {
					return fmt.Sprintf("expected ID %s, got %s", sd.Executions[0].ID, gotResp.ID)
				}

				// Verify trigger_data is present
				if len(gotResp.TriggerData) == 0 {
					return "expected trigger_data to be populated"
				}

				return ""
			},
		},
	}
}

// queryExecutionByID200ExecutedByName verifies scope ①: the detail endpoint
// projects the joined executor display name (executed_by_name) for a
// user-attributed manual execution, and leaves it empty for automation
// executions (NULL executed_by → LEFT JOIN yields no user).
func queryExecutionByID200ExecutedByName(sd ExecutionSeedData) []apitest.Table {
	// Pick a manual execution attributed to a user and an automation execution
	// with no executor, by property rather than fragile seed ordering.
	var manualID, automationID, executorID uuid.UUID
	for _, e := range sd.Executions {
		if e.ExecutedBy != nil && manualID == uuid.Nil {
			manualID = e.ID
			executorID = *e.ExecutedBy
		}
		if e.AutomationRuleID != nil && e.ExecutedBy == nil && automationID == uuid.Nil {
			automationID = e.ID
		}
	}
	if manualID == uuid.Nil || automationID == uuid.Nil {
		return nil
	}

	// Resolve the expected display name from the user the manual execution is
	// actually attributed to, rather than assuming a position in sd.Users.
	var expectedName string
	for _, u := range sd.Users {
		if u.User.ID == executorID {
			expectedName = u.User.FirstName.String() + " " + u.User.LastName.String()
			break
		}
	}
	if expectedName == "" {
		return nil
	}

	return []apitest.Table{
		{
			Name:       "executed-by-name-populated",
			URL:        fmt.Sprintf("/v1/workflow/executions/%s", manualID),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &executionapi.ExecutionDetail{},
			ExpResp:    &executionapi.ExecutionDetail{},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*executionapi.ExecutionDetail)
				if !exists {
					return "error getting execution detail"
				}
				if gotResp.ExecutedByName != expectedName {
					return fmt.Sprintf("expected executed_by_name %q, got %q", expectedName, gotResp.ExecutedByName)
				}
				return ""
			},
		},
		{
			Name:       "executed-by-name-empty-for-automation",
			URL:        fmt.Sprintf("/v1/workflow/executions/%s", automationID),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &executionapi.ExecutionDetail{},
			ExpResp:    &executionapi.ExecutionDetail{},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*executionapi.ExecutionDetail)
				if !exists {
					return "error getting execution detail"
				}
				if gotResp.ExecutedByName != "" {
					return fmt.Sprintf("expected empty executed_by_name for automation execution, got %q", gotResp.ExecutedByName)
				}
				return ""
			},
		},
	}
}

func queryExecutionByID404(sd ExecutionSeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "execution-not-found",
			URL:        fmt.Sprintf("/v1/workflow/executions/%s", uuid.New()),
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
}

func queryExecutionByID401(sd ExecutionSeedData) []apitest.Table {
	if len(sd.Executions) == 0 {
		return nil
	}

	return []apitest.Table{
		{
			Name:       "no-token",
			URL:        fmt.Sprintf("/v1/workflow/executions/%s", sd.Executions[0].ID),
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
}
