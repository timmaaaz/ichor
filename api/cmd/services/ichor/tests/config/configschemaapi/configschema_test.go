package configschema_test

import (
	"testing"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func Test_ConfigSchemaAPI(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_ConfigSchemaAPI")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	test.Run(t, queryTableConfigSchema200(sd), "queryTableConfigSchema-200")
	test.Run(t, queryLayoutSchema200(sd), "queryLayoutSchema-200")
	test.Run(t, queryContentTypes200(sd), "queryContentTypes-200")
	test.Run(t, querySchemas401(sd), "querySchemas-401")
}
