package scan_test

import (
	"testing"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func Test_Scan(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_Scan")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	test.Run(t, scanSerial200(sd), "scan-serial-200")
	test.Run(t, scanLot200(sd), "scan-lot-200")
	test.Run(t, scanProduct200(sd), "scan-product-200")
	test.Run(t, scanUnknown200(sd), "scan-unknown-200")

	test.Run(t, scan400(sd), "scan-400")
	test.Run(t, scan401(sd), "scan-401")
}
