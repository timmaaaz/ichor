package action_test

import (
	"testing"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func Test_ActionAPI(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_ActionAPI")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// List available actions tests
	test.Run(t, listActions200Admin(sd), "list-actions-200-admin")
	test.Run(t, listActions200UserWithPermissions(sd), "list-actions-200-user-with-permissions")
	test.Run(t, listActions200UserNoPermissions(sd), "list-actions-200-user-no-permissions")
	test.Run(t, listActions401(sd), "list-actions-401")

	// Execute action tests - error cases
	test.Run(t, execute401(sd), "execute-401")
	test.Run(t, execute403NoPermission(sd), "execute-403-no-permission")
	test.Run(t, execute404UnknownAction(sd), "execute-403-unknown-action")

	// Execution status tests
	test.Run(t, getExecutionStatus401(sd), "get-execution-status-401")
	test.Run(t, getExecutionStatus404(sd), "get-execution-status-404")
}
