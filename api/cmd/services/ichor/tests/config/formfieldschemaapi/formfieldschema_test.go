package formfieldschema_test

import (
	"testing"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func Test_FormFieldSchemaAPI(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_FormFieldSchemaAPI")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	test.Run(t, queryFieldTypes200(sd), "queryFieldTypes-200")
	test.Run(t, queryFieldTypeSchema200(sd), "queryFieldTypeSchema-200")
	test.Run(t, queryFieldTypeSchema404(sd), "queryFieldTypeSchema-404")
	test.Run(t, queryFieldTypes401(sd), "queryFieldTypes-401")
}
