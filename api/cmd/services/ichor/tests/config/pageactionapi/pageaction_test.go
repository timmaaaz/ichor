package pageaction_test

import (
	"testing"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func Test_PageAction(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_PageAction")
	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("seeding error %s", err)
	}

	// Query tests
	test.Run(t, query200(sd), "query-200")
	test.Run(t, queryByID200(sd), "queryByID-200")
	test.Run(t, queryByPageConfigID200(sd), "queryByPageConfigID-200")
	test.Run(t, query401(sd), "query-401")

	// Button create tests
	test.Run(t, createButton200(sd), "createButton-200")
	test.Run(t, createButton400(sd), "createButton-400")
	test.Run(t, createButton401(sd), "createButton-401")

	// Dropdown create tests
	test.Run(t, createDropdown200(sd), "createDropdown-200")
	test.Run(t, createDropdown400(sd), "createDropdown-400")
	test.Run(t, createDropdown401(sd), "createDropdown-401")

	// Separator create tests
	test.Run(t, createSeparator200(sd), "createSeparator-200")
	test.Run(t, createSeparator400(sd), "createSeparator-400")
	test.Run(t, createSeparator401(sd), "createSeparator-401")

	// Button update tests
	test.Run(t, updateButton200(sd), "updateButton-200")
	test.Run(t, updateButton400(sd), "updateButton-400")
	test.Run(t, updateButton401(sd), "updateButton-401")

	// Dropdown update tests
	test.Run(t, updateDropdown200(sd), "updateDropdown-200")
	test.Run(t, updateDropdown400(sd), "updateDropdown-400")
	test.Run(t, updateDropdown401(sd), "updateDropdown-401")

	// Separator update tests
	test.Run(t, updateSeparator200(sd), "updateSeparator-200")
	test.Run(t, updateSeparator400(sd), "updateSeparator-400")
	test.Run(t, updateSeparator401(sd), "updateSeparator-401")

	// Delete tests
	test.Run(t, delete401(sd), "delete-401")
	test.Run(t, delete200(sd), "delete-200")

	// Batch tests
	test.Run(t, batchCreate200(sd), "batchCreate-200")
	test.Run(t, batchCreate400(sd), "batchCreate-400")
	test.Run(t, batchCreate401(sd), "batchCreate-401")
}
