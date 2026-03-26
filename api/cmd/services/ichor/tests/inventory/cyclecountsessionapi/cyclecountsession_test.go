package cyclecountsessionapi_test

import (
	"testing"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func Test_CycleCountSession(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_CycleCountSession")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// Query
	test.Run(t, query200(sd), "query-200")
	test.Run(t, queryByID200(sd), "query-by-id-200")
	test.Run(t, queryByID404(sd), "query-by-id-404")

	// Create
	test.Run(t, create200(sd), "create-200")
	test.Run(t, create400(sd), "create-400")
	test.Run(t, create401(sd), "create-401")

	// Update
	test.Run(t, update200(sd), "update-200")
	test.Run(t, update400(sd), "update-400")
	test.Run(t, update401(sd), "update-401")
	test.Run(t, update404(sd), "update-404")

	// Delete
	test.Run(t, delete200(sd), "delete-200")
	test.Run(t, delete401(sd), "delete-401")
	test.Run(t, delete404(sd), "delete-404")
}
