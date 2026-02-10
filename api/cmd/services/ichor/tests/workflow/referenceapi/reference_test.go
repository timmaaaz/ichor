package reference_test

import (
	"testing"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func Test_ReferenceAPI(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_ReferenceAPI")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// Trigger types tests
	test.Run(t, queryTriggerTypes200(sd), "queryTriggerTypes-200")
	test.Run(t, queryTriggerTypes401(sd), "queryTriggerTypes-401")

	// Entity types tests
	test.Run(t, queryEntityTypes200(sd), "queryEntityTypes-200")
	test.Run(t, queryEntityTypes401(sd), "queryEntityTypes-401")

	// Entities tests
	test.Run(t, queryEntities200(sd), "queryEntities-200")
	test.Run(t, queryEntities401(sd), "queryEntities-401")
	test.Run(t, queryEntitiesWithFilter200(sd), "queryEntities-with-filter-200")

	// Action types tests
	test.Run(t, queryActionTypes200(sd), "queryActionTypes-200")
	test.Run(t, queryActionTypes401(sd), "queryActionTypes-401")

	// Action type schema tests
	test.Run(t, queryActionTypeSchema200(sd), "queryActionTypeSchema-200")
	test.Run(t, queryActionTypeSchema404(sd), "queryActionTypeSchema-404")
	test.Run(t, queryActionTypeSchema401(sd), "queryActionTypeSchema-401")
}
