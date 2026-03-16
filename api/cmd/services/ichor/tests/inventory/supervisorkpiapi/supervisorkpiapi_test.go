package supervisorkpiapi_test

import (
	"testing"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func Test_SupervisorKPIs(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_SupervisorKPIs")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	test.Run(t, query200(sd), "query-200")
	test.Run(t, query401(sd), "query-401")
}
