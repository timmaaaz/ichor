package ordersapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/sales/ordersapp"
	"github.com/timmaaaz/ichor/business/domain/labels/labelbus"
	"github.com/timmaaaz/ichor/business/sdk/seedid"
)

// Test_Orders_BindContainer_AcceptsV5Label guards the FK-validator fix: foreign-key
// reference fields must accept ANY UUID version (validate:"uuid"), not version 4 only
// (validate:"uuid4"). Deterministic seed IDs are UUID v5 (seedid.Stable -> uuid.NewSHA1),
// so a uuid4 validator 400s any real seeded reference.
//
// container_label_id is the canonical case: production seeds the label catalog with v5 ids
// (business/sdk/dbtest/seed_labels.go via seedid.Stable), and the bind endpoint validates
// container_label_id. This test SeedCreates a real v5 label row (the production path, which
// bypasses request validation) so the row satisfies the FK, then POSTs it to the bind
// endpoint — where the struct-tag validator is the ONLY version gate in the path.
//
// Critically it seeds the v5 id via Label.SeedCreate, NOT busDomain.Label.Create (which mints
// uuid.New() = v4). The existing bindContainer200 uses the v4 Create path, which is exactly why
// CI stayed green while the v5 case was broken — the success fixture couldn't reproduce production.
//
//	RED  (validator = uuid4): v5 id rejected at validation -> 400.
//	GREEN (validator = uuid):  v5 id accepted -> 200.
func Test_Orders_BindContainer_AcceptsV5Label(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_Orders_BindContainer_AcceptsV5Label")

	sd, err := insertSeedData(test.DB, test.Auth)
	require.NoError(t, err, "seeding")

	ctx := context.Background()

	// Mint a deterministic v5 label id the same way production seeding does, and create a real
	// catalog row for it via the validation-bypassing SeedCreate path so the FK is satisfiable.
	v5ID := seedid.Stable("label:V5-BIND-TEST")
	require.EqualValues(t, 5, v5ID.Version(), "seedid.Stable must mint a v5 UUID")

	require.NoError(t, test.DB.BusDomain.Label.SeedCreate(ctx, labelbus.LabelCatalog{
		ID:          v5ID,
		Code:        "V5-BIND-TEST",
		Type:        labelbus.TypeContainer,
		PayloadJSON: "{}",
	}), "seedcreate v5 container label")

	// Orders[0] is free (only Orders[4] is pre-bound in the seed); the v5 container is brand new,
	// so the one_active_binding_per_container constraint is not implicated.
	body, err := json.Marshal(ordersapp.NewOrderContainerBinding{ContainerLabelID: v5ID.String()})
	require.NoError(t, err, "marshal binding")

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/v1/sales/orders/%s/bindings", sd.Orders[0].ID),
		bytes.NewReader(body))
	r.Header.Set("Authorization", "Bearer "+sd.Admins[0].Token)
	r.Header.Set("Content-Type", "application/json")
	test.ServeHTTP(w, r)

	require.Equal(t, http.StatusOK, w.Code,
		"a v5 container_label_id must be accepted — FK reference fields validate `uuid` (any version), "+
			"not `uuid4`; body=%s", w.Body.String())
}
