package userpreferencesapi_test

import (
	"testing"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func Test_UserPreferencesAPI(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_UserPreferencesAPI")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("seeding error: %s", err)
	}

	test.Run(t, set200(sd), "set-200")
	test.Run(t, get200(sd), "get-200")
	test.Run(t, getAll200(sd), "getall-200")
	test.Run(t, delete200(sd), "delete-200")
}
