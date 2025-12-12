package data_test

import (
	"testing"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

// NOTE: Before writing validate tests, the stored config needs to have all its
// validation tags filled out properly.

func Test_Data(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_Data")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	test.Run(t, queryByID200(sd), "querybyid-200")
	test.Run(t, queryByName200(sd), "querybyname-200")
	test.Run(t, queryByUser200(sd), "querybyuser-200")
	test.Run(t, queryAll200(sd), "queryall-200")
	test.Run(t, queryAll401(sd), "queryall-401")

	test.Run(t, create200(sd), "create-200")

	test.Run(t, execute200(sd), "execute-200")
	test.Run(t, executeCountByID200(sd), "executecountbyid-200")
	test.Run(t, executeByName200(sd), "executebyname-200")
	test.Run(t, executeCountByName200(sd), "executecountbyname-200")

	test.Run(t, update200(sd), "update-200")

	test.Run(t, delete200(sd), "delete-200")

	// Chart tests
	test.Run(t, executeChartByID200(sd), "executechartbyid-200")
	test.Run(t, executeChartByID400(sd), "executechartbyid-400")
	test.Run(t, executeChartByID401(sd), "executechartbyid-401")
	test.Run(t, executeChartByID404(sd), "executechartbyid-404")
	test.Run(t, executeChartByName200(sd), "executechartbyname-200")
	test.Run(t, executeChartByName404(sd), "executechartbyname-404")
	test.Run(t, previewChartData200(sd), "previewchartdata-200")
	test.Run(t, previewChartData400(sd), "previewchartdata-400")
	test.Run(t, previewChartData401(sd), "previewchartdata-401")
}
