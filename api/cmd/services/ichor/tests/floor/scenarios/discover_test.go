package scenarios_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func TestDeriveFamily(t *testing.T) {
	cases := []struct {
		name string
		want family
	}{
		{"transfer-intra-zone", familyTransfer},
		{"transfer-cross-zone", familyTransfer},
		{"pick-whole-order", familyPick},
		{"pick-short-pick", familyPick},
		{"receive-lot-tracking", familyReceive},
		{"cycle-count-variance-over", familyCycleCount},
		{"profile-medical-device-rental", familyProfile},
		{"profile-strict-regulated", familyProfile},
		{"rush-receiving", familyReceive},   // override
		{"e2e-pick-strict", familyPick},     // override
		{"e2e-baseline", ""},                // unset — falls through to Custom
		{"unknown-prefix", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := deriveFamily(tc.name); got != tc.want {
				t.Errorf("deriveFamily(%q) = %q, want %q", tc.name, got, tc.want)
			}
		})
	}
}

func TestDiscoverScenarios_Smoke(t *testing.T) {
	rows, err := discoverScenarios(scenarioRoots())
	if err != nil {
		t.Fatalf("discoverScenarios: %v", err)
	}
	// We expect exactly 21 scenarios in deployments/scenarios/ as of 2026-05-20.
	// If this count drifts, either a scenario was added/removed or the
	// discovery glob broke — investigate, do not blindly update the number.
	const wantCount = 21
	if len(rows) != wantCount {
		names := make([]string, 0, len(rows))
		for _, r := range rows {
			names = append(names, r.Name)
		}
		t.Errorf("discoverScenarios returned %d rows, want %d. Got: %v", len(rows), wantCount, names)
	}
}

// discoverTransferInputs queries the DB for the single pending-or-approved
// transfer order that belongs to the given scenario, and enriches it with the
// source/destination location codes and product UPC needed by walkTransfer.
//
// The query accepts both "pending" and "approved" statuses so it works
// regardless of whether the scenario seeds the order as pending (the current
// convention) or pre-approved. walkTransfer advances the status itself.
//
// All failure paths call t.Fatalf; the returned TransferInputs is always valid.
func discoverTransferInputs(t *testing.T, h *apitest.Test, scenarioID uuid.UUID) TransferInputs {
	t.Helper()
	ctx := context.Background()
	db := h.DB.DB

	// Fetch the transfer order row.
	var row struct {
		ID             uuid.UUID `db:"id"`
		FromLocationID uuid.UUID `db:"from_location_id"`
		ToLocationID   uuid.UUID `db:"to_location_id"`
		ProductID      uuid.UUID `db:"product_id"`
		Quantity       int       `db:"quantity"`
	}
	err := db.GetContext(ctx, &row, `
		SELECT id, from_location_id, to_location_id, product_id, quantity
		FROM inventory.transfer_orders
		WHERE scenario_id = $1
		  AND status IN ('pending', 'approved')
		ORDER BY transfer_date ASC
		LIMIT 1
	`, scenarioID)
	if err != nil {
		t.Fatalf("discoverTransferInputs: query transfer_orders for scenario %s: %v", scenarioID, err)
	}

	// Resolve source location code.
	var fromCode string
	if err := db.GetContext(ctx, &fromCode, `
		SELECT location_code
		FROM inventory.inventory_locations
		WHERE id = $1
	`, row.FromLocationID); err != nil {
		t.Fatalf("discoverTransferInputs: query from_location code for %s: %v", row.FromLocationID, err)
	}

	// Resolve destination location code.
	var toCode string
	if err := db.GetContext(ctx, &toCode, `
		SELECT location_code
		FROM inventory.inventory_locations
		WHERE id = $1
	`, row.ToLocationID); err != nil {
		t.Fatalf("discoverTransferInputs: query to_location code for %s: %v", row.ToLocationID, err)
	}

	// Resolve product UPC.
	var upc string
	if err := db.GetContext(ctx, &upc, `
		SELECT upc_code
		FROM products.products
		WHERE id = $1
	`, row.ProductID); err != nil {
		t.Fatalf("discoverTransferInputs: query upc_code for product %s: %v", row.ProductID, err)
	}

	return TransferInputs{
		TransferID: row.ID,
		FromCode:   fromCode,
		ToCode:     toCode,
		ProductID:  row.ProductID,
		UPC:        upc,
		Quantity:   row.Quantity,
	}
}
