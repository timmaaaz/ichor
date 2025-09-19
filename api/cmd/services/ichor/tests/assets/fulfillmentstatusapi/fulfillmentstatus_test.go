package fulfillmentstatus_test

import (
	"testing"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func Test_FulFillmentStatus(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_FulfillmentStatus")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("seeding error %s", err)
	}

	test.Run(t, query200(sd), "query-200")
	test.Run(t, queryByID200(sd), "query-by-id-200")

	test.Run(t, create200(sd), "create-200")
	test.Run(t, create400(sd), "create-400")

	test.Run(t, update200(sd), "update-200")
	test.Run(t, update400(sd), "update-400")

	test.Run(t, delete200(sd), "delete-200")
}
