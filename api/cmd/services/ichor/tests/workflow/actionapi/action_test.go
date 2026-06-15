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

	// Execute action tests - happy path
	test.Run(t, execute200CreateAlert(sd), "execute-200-create-alert")

	// Execute action tests - transition_status
	test.Run(t, executeTransitionStatus403Denied(sd), "execute-transition-status-403")

	// Configurable action-button verbs end-to-end (registration + P4 grant + {{entity_id}}
	// resolution + real DB effect, via the production route).
	test.Run(t, executeReleaseToPicking200(sd), "execute-release-to-picking-200")
	test.Run(t, executeReleaseToPicking403(sd), "execute-release-to-picking-403")
	test.Run(t, executeClaimTransferOrder200(sd), "execute-claim-transfer-order-200")
	test.Run(t, executeExecuteTransferOrder200(sd), "execute-execute-transfer-order-200")

	// Protected-list (P3) enforcement on the manual-execute HTTP path (Path A, backend-authoritative)
	test.Run(t, executeTransitionStatusProtected400(sd), "execute-transition-status-protected-400")
	test.Run(t, executeCreateEntityProtected400(sd), "execute-create-entity-protected-400")
	test.Run(t, executeUpdateFieldNotManuallyExecutable(sd), "execute-update-field-not-manually-executable")

	// Execute action tests - error cases
	test.Run(t, execute401(sd), "execute-401")
	test.Run(t, execute403NoPermission(sd), "execute-403-no-permission")
	test.Run(t, execute404UnknownAction(sd), "execute-403-unknown-action")

	// Execution status tests
	test.Run(t, getExecutionStatus401(sd), "get-execution-status-401")
	test.Run(t, getExecutionStatus404(sd), "get-execution-status-404")
}
