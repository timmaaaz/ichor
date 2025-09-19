package role_test

import (
	"testing"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func Test_Role(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_Role")
	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("seeding error %s", err)
	}

	test.Run(t, query200(sd), "query-200")
	test.Run(t, queryByID200(sd), "queryByID-200")
	test.Run(t, query401(sd), "query-401")

	test.Run(t, create200(sd), "create-200")
	test.Run(t, create400(sd), "create-400")
	test.Run(t, create401(sd), "create-401")

	test.Run(t, update200(sd), "update-200")
	test.Run(t, update400(sd), "update-400")
	test.Run(t, update401(sd), "update-401")

	test.Run(t, delete401(sd), "delete-401")
	test.Run(t, delete200(sd), "delete-200")

}
