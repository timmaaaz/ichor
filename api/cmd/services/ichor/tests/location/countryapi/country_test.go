package country_test

import (
	"testing"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func Test_Country(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_Country")

	// -------------------------------------------------------------------------

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	// test.Run(t, query200(sd), "query-200")
	test.Run(t, countryQueryByID200(sd), "query-by-id-200")
}
