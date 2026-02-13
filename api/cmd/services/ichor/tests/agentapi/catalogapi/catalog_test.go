package catalog_test

import (
	"testing"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func Test_CatalogAPI(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_CatalogAPI")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	test.Run(t, queryCatalog200(sd), "queryCatalog-200")
	test.Run(t, queryCatalog401(sd), "queryCatalog-401")
}
