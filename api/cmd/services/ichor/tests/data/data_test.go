package data_test

import (
	"testing"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func Test_Data(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_Data")

	_, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}
}
