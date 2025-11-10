package pageconfigapi_test

import (
	"testing"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func Test_PageConfig(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_PageConfig")

	// -------------------------------------------------------------------------

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	test.Run(t, queryByID200(sd), "query-by-id-200")
	test.Run(t, queryByName200(sd), "query-by-name-200")

	test.Run(t, create200(sd), "create-200")
	test.Run(t, create400(sd), "create-400")
	test.Run(t, create401(sd), "create-401")

	test.Run(t, update200(sd), "update-200")
	test.Run(t, update400(sd), "update-400")
	test.Run(t, update401(sd), "update-401")
	test.Run(t, update404(sd), "update-404")

	test.Run(t, delete200(sd), "delete-200")
	test.Run(t, delete400(sd), "delete-400")
	test.Run(t, delete401(sd), "delete-401")
	test.Run(t, delete404(sd), "delete-404")
}
