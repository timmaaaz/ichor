package userrole_test

import (
	"testing"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func Test_UserRole(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_UserRole")
	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("seeding error %s", err)
	}

	// Query tests
	test.Run(t, query200(sd), "query-200")
	test.Run(t, queryByID200(sd), "queryByID-200")
	test.Run(t, query401(sd), "query-401")

	// Create tests
	test.Run(t, create200(sd), "create-200")
	test.Run(t, create400(sd), "create-400")
	test.Run(t, create401(sd), "create-401")

	// Delete tests
	test.Run(t, delete200(sd), "delete-200")
	test.Run(t, delete401(sd), "delete-401")
}
