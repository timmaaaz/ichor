package alert_test

import (
	"testing"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func Test_AlertAPI(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_AlertAPI")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// Query tests (multi-severity filter)
	test.Run(t, queryMine200(sd), "queryMine-200")
	test.Run(t, queryMineWithSeverityFilter200(sd), "queryMine-severity-filter-200")
	test.Run(t, queryMineWithInvalidSeverity400(sd), "queryMine-invalid-severity-400")

	// Bulk acknowledge tests
	test.Run(t, acknowledgeSelected200(sd), "acknowledge-selected-200")
	test.Run(t, acknowledgeSelectedPartialSkip200(sd), "acknowledge-selected-partial-skip-200")
	test.Run(t, acknowledgeSelected400(sd), "acknowledge-selected-400")
	test.Run(t, acknowledgeSelected401(sd), "acknowledge-selected-401")
	test.Run(t, acknowledgeAll200(sd), "acknowledge-all-200")
	test.Run(t, acknowledgeAll401(sd), "acknowledge-all-401")

	// Bulk dismiss tests
	test.Run(t, dismissSelected200(sd), "dismiss-selected-200")
	test.Run(t, dismissAll200(sd), "dismiss-all-200")
	test.Run(t, dismissAll401(sd), "dismiss-all-401")
}
