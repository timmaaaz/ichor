package execution_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/workflow/executionapp"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
)

// =============================================================================
// Workflow Execution Re-run Integration Tests
//
// This test drives the shared ichor mux built by
// api/cmd/services/ichor/build/all/all.go via StartTestWithTemporal, which feeds
// a real Temporal client into mux.Config. Task 7 wired all.go's WorkflowTrigger
// into executionapi.Config.Trigger (and permissionsBus into .PermissionsBus), so
// the POST /rerun route now runs with a live Reranner and a non-nil PermissionsBus.
// The happy-path 200 returns a fresh execution id (proving dispatch succeeded);
// the unknown-id 404 maps workflow.ErrNotFound; the non-admin/no-token 401 cases
// are rejected by OPA RuleAdminOnly before reaching the handler.
//
// The happy path reuses sd.Executions[0]: insertSeedData seeds it via
// TestSeedRuleActions, which attaches a start edge to rule[0]'s first action, so
// the rule has a non-empty active graph (Task 4's empty-graph guard would
// otherwise return FailedPrecondition). Its TriggerData carries entity_id + status + total.
// =============================================================================

func Test_Execution_Rerun(t *testing.T) {
	t.Parallel()

	// StartTestWithTemporal stands up a real Temporal client and wires all.go's
	// WorkflowTrigger into executionapi.Config.Trigger (Task 7). The plain StartTest
	// mux has no Temporal, so the rerun route would run with a nil Reranner and the
	// happy path could only ever surface Internal — see the package doc above.
	test := apitest.StartTestWithTemporal(t, "Test_Execution_Rerun")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// insertSeedData only seeds an admin user; mint a non-admin token here to
	// exercise the admin gate (RuleAdminOnly → 401).
	nonAdminToken := seedNonAdminToken(t, test)

	test.Run(t, rerun200(sd), "rerun-200")
	test.Run(t, rerun401(sd, nonAdminToken), "rerun-401")
	test.Run(t, rerun404(sd), "rerun-404")
}

// seedNonAdminToken creates a regular (non-admin) user and returns a bearer
// token for it.
func seedNonAdminToken(t *testing.T, test *apitest.Test) string {
	t.Helper()

	ctx := context.Background()
	users, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, test.DB.BusDomain.User)
	if err != nil {
		t.Fatalf("seeding non-admin user: %v", err)
	}
	return apitest.Token(test.DB.BusDomain.User, test.Auth, users[0].Email.Address)
}

// rerun200 is the happy path: an admin re-runs an existing (rerunnable)
// execution and receives a fresh execution id, distinct from the original and
// non-nil.
func rerun200(sd ExecutionSeedData) []apitest.Table {
	if len(sd.Executions) == 0 {
		return nil
	}

	// Executions[0] points at Rules[0], which has a start edge (non-empty graph).
	original := sd.Executions[0].ID

	return []apitest.Table{
		{
			Name:       "admin",
			URL:        fmt.Sprintf("/v1/workflow/executions/%s/rerun", original),
			Token:      sd.Users[0].Token, // admin
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			GotResp:    &executionapp.RerunResponse{},
			ExpResp:    &executionapp.RerunResponse{},
			CmpFunc: func(got any, _ any) string {
				resp, ok := got.(*executionapp.RerunResponse)
				if !ok {
					return "error getting rerun response"
				}
				if resp.OriginalExecutionID != original {
					return fmt.Sprintf("original_execution_id mismatch: got %s want %s", resp.OriginalExecutionID, original)
				}
				if resp.NewExecutionID == uuid.Nil {
					return "new_execution_id must not be nil"
				}
				if resp.NewExecutionID == original {
					return "new_execution_id must differ from the original"
				}
				return ""
			},
		},
	}
}

// rerun401 proves the admin gate: a non-admin token (and a missing token) are
// rejected before reaching the handler (OPA RuleAdminOnly → Unauthenticated/401).
func rerun401(sd ExecutionSeedData, nonAdminToken string) []apitest.Table {
	if len(sd.Executions) == 0 {
		return nil
	}

	original := sd.Executions[0].ID

	return []apitest.Table{
		{
			Name:       "non-admin",
			URL:        fmt.Sprintf("/v1/workflow/executions/%s/rerun", original),
			Token:      nonAdminToken,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &map[string]any{},
			ExpResp:    &map[string]any{},
			CmpFunc:    func(_ any, _ any) string { return "" },
		},
		{
			Name:       "no-token",
			URL:        fmt.Sprintf("/v1/workflow/executions/%s/rerun", original),
			Token:      "",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &map[string]any{},
			ExpResp:    &map[string]any{},
			CmpFunc:    func(_ any, _ any) string { return "" },
		},
	}
}

// rerun404 proves an unknown execution id surfaces as NotFound: the app layer
// maps workflow.ErrNotFound → 404.
func rerun404(sd ExecutionSeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "unknown-id",
			URL:        fmt.Sprintf("/v1/workflow/executions/%s/rerun", uuid.New()),
			Token:      sd.Users[0].Token, // admin
			Method:     http.MethodPost,
			StatusCode: http.StatusNotFound,
			GotResp:    &map[string]any{},
			ExpResp:    &map[string]any{},
			CmpFunc:    func(_ any, _ any) string { return "" },
		},
	}
}
