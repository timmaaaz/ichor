package scenarioapi_test

import (
	"testing"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func Test_Scenarios(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_Scenarios")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	test.Run(t, query200(sd), "query-200")
	test.Run(t, query401(sd), "query-401")
	test.Run(t, queryByID200(sd), "query-by-id-200")
	test.Run(t, queryByID401(sd), "query-by-id-401")
	test.Run(t, queryByID404(sd), "query-by-id-404")
	test.Run(t, activeNone(sd), "active-none")
	test.Run(t, active401(sd), "active-401")
	test.Run(t, fixtures200(sd), "fixtures-200")
	test.Run(t, fixtures401(sd), "fixtures-401")
	test.Run(t, fixtures404(sd), "fixtures-404")
}
