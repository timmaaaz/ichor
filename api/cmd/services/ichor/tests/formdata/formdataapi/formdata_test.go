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

	test.Run(t, validate200_ValidForm(sd), "validate-200-valid-form")
	test.Run(t, validate200_MultiEntityForm(sd), "validate-200-multi-entity")
	test.Run(t, validate200_UnregisteredEntity(sd), "validate-200-unregistered-entity")

	test.Run(t, validate400(sd), "validate-400")
	test.Run(t, validate401(sd), "validate-401")
	test.Run(t, validate404(sd), "validate-404")
}
