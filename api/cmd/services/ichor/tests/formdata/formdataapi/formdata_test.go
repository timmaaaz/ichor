package formdataapi_test

import (
	"testing"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func Test_FormData(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_FormData")

	// -------------------------------------------------------------------------

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	test.Run(t, upsertSingleEntityCreate200(sd), "upsert-single-entity-create-200")
	test.Run(t, upsertSingleEntityUpdate200(sd), "upsert-single-entity-update-200")
	test.Run(t, upsertMultiEntityWithForeignKey200(sd), "upsert-multi-entity-with-fk-200")

	test.Run(t, upsert400(sd), "upsert-400")
	test.Run(t, upsert401(sd), "upsert-401")
	test.Run(t, upsert404(sd), "upsert-404")
}
