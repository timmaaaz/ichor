package vproduct_test

import (
	"testing"

	"bitbucket.org/superiortechnologies/ichor/api/sdk/http/apitest"
)

func Test_VProduct(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_VProduct")

	// -------------------------------------------------------------------------

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	test.Run(t, query200(sd), "query-200")
}
