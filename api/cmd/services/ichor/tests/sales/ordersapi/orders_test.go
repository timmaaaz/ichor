package ordersapi_test

import (
	"testing"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func Test_Order(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_Order")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	test.Run(t, query200(sd), "query-200")
	test.Run(t, queryByID200(sd), "query-by-id-200")

	test.Run(t, create200(sd), "create-200")
	test.Run(t, create400(sd), "create-400")
	test.Run(t, create401(sd), "create-401")

	test.Run(t, update200(sd), "update-200")
	test.Run(t, update400(sd), "update-400")
	test.Run(t, update401(sd), "update-401")

	test.Run(t, delete200(sd), "delete-200")
	test.Run(t, delete401(sd), "delete-401")

	// =========================================================================
	// Order container bindings (Phase 0g.B7)
	// =========================================================================
	test.Run(t, bindContainer200(sd), "bind-container-200")
	test.Run(t, bindContainer409(sd), "bind-container-409")
	test.Run(t, bindContainer400(sd), "bind-container-400")
	test.Run(t, bindContainer404(sd), "bind-container-404")
	test.Run(t, bindContainer401(sd), "bind-container-401")

	test.Run(t, unbindContainer200(sd), "unbind-container-200")
	test.Run(t, unbindContainerIdempotent(sd), "unbind-container-idempotent")
	test.Run(t, unbindContainer404(sd), "unbind-container-404")
	test.Run(t, unbindContainer401(sd), "unbind-container-401")

	test.Run(t, queryBindings200(sd), "query-bindings-200")
	test.Run(t, queryBindingsEmpty(sd), "query-bindings-empty")
	test.Run(t, queryBindings401(sd), "query-bindings-401")
}
