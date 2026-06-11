package data_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/protected"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/data"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

func testLog() *logger.Logger {
	var buf bytes.Buffer
	return logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})
}

// Blocked paths return before touching the DB, so a nil *sqlx.DB is never used.

func Test_UpdateField_BlocksProtectedField(t *testing.T) {
	t.Parallel()

	reg := protected.New()
	reg.ProtectField("procurement.purchase_orders", "purchase_order_status_id", "approve_purchase_order")

	h := data.NewUpdateFieldHandler(testLog(), nil, data.WithProtectedRegistry(reg))

	cfg := json.RawMessage(`{
		"target_entity": "procurement.purchase_orders",
		"target_field": "purchase_order_status_id",
		"new_value": "approved"
	}`)

	res, err := h.Execute(context.Background(), cfg, workflow.ActionExecutionContext{})
	if !errors.Is(err, protected.ErrProtectedField) {
		t.Fatalf("err = %v, want ErrProtectedField", err)
	}
	// update_field returns its failed result map alongside the error.
	m, ok := res.(map[string]any)
	if !ok || m["status"] != "failed" {
		t.Fatalf("result = %#v, want status=failed map", res)
	}
}

func Test_UpdateField_AllowsUnprotectedFieldThroughGate(t *testing.T) {
	t.Parallel()

	// Registry protects a DIFFERENT field; the gate must let an unprotected field pass.
	reg := protected.New()
	reg.ProtectField("procurement.purchase_orders", "purchase_order_status_id", "approve_purchase_order")

	h := data.NewUpdateFieldHandler(testLog(), nil, data.WithProtectedRegistry(reg))

	cfg := json.RawMessage(`{
		"target_entity": "procurement.purchase_orders",
		"target_field": "purchase_order_status_id_NOT",
		"new_value": "x",
		"conditions": [{"field_name": "id", "operator": "equals", "value": "00000000-0000-0000-0000-000000000000"}]
	}`)

	// We do not run the DB here; we only assert the protection gate did not fire.
	// Recover from the nil-DB panic that happens *after* the gate, proving the gate passed.
	defer func() { _ = recover() }()
	_, err := h.Execute(context.Background(), cfg, workflow.ActionExecutionContext{})
	if errors.Is(err, protected.ErrProtectedField) {
		t.Fatalf("unprotected field was blocked: %v", err)
	}
}

func Test_TransitionStatus_BlocksInvariantStatus(t *testing.T) {
	t.Parallel()

	reg := protected.New()
	reg.ProtectField("procurement.purchase_orders", "purchase_order_status_id", "approve_purchase_order")

	h := data.NewTransitionStatusHandler(testLog(), nil, data.WithProtectedRegistry(reg))

	cfg := json.RawMessage(`{
		"target_entity": "procurement.purchase_orders",
		"target_id": "00000000-0000-0000-0000-000000000001",
		"status_field": "purchase_order_status_id",
		"to_status": "approved",
		"valid_from_statuses": ["draft"]
	}`)

	_, err := h.Execute(context.Background(), cfg, workflow.ActionExecutionContext{})
	if !errors.Is(err, protected.ErrProtectedField) {
		t.Fatalf("err = %v, want ErrProtectedField", err)
	}
}

func Test_TransitionStatus_AllowsPlainStatusThroughGate(t *testing.T) {
	t.Parallel()

	reg := protected.New() // protects nothing
	h := data.NewTransitionStatusHandler(testLog(), nil, data.WithProtectedRegistry(reg))

	cfg := json.RawMessage(`{
		"target_entity": "workflow.automation_rules",
		"target_id": "00000000-0000-0000-0000-000000000001",
		"status_field": "some_status",
		"to_status": "done",
		"valid_from_statuses": ["pending"]
	}`)

	defer func() { _ = recover() }()
	_, err := h.Execute(context.Background(), cfg, workflow.ActionExecutionContext{})
	if errors.Is(err, protected.ErrProtectedField) {
		t.Fatalf("plain status field was blocked: %v", err)
	}
}

func Test_CreateEntity_BlocksWholeTableEntity(t *testing.T) {
	t.Parallel()

	reg := protected.New()
	reg.ProtectEntity("inventory.inventory_transactions", "")

	h := data.NewCreateEntityHandler(testLog(), nil, data.WithProtectedRegistry(reg))

	cfg := json.RawMessage(`{
		"target_entity": "inventory.inventory_transactions",
		"fields": {"quantity": 5}
	}`)

	_, err := h.Execute(context.Background(), cfg, workflow.ActionExecutionContext{})
	if !errors.Is(err, protected.ErrProtectedField) {
		t.Fatalf("err = %v, want ErrProtectedField", err)
	}
}

func Test_CreateEntity_BlocksProtectedFieldInPayload(t *testing.T) {
	t.Parallel()

	reg := protected.New()
	reg.ProtectField("sales.order_line_items", "picked_quantity", "")

	h := data.NewCreateEntityHandler(testLog(), nil, data.WithProtectedRegistry(reg))

	cfg := json.RawMessage(`{
		"target_entity": "sales.order_line_items",
		"fields": {"quantity": 1, "picked_quantity": 9}
	}`)

	_, err := h.Execute(context.Background(), cfg, workflow.ActionExecutionContext{})
	if !errors.Is(err, protected.ErrProtectedField) {
		t.Fatalf("err = %v, want ErrProtectedField", err)
	}
}

// DB-backed: a populated registry must not blanket-block; an unprotected field still writes,
// and the protected field on the same entity is rejected end-to-end.
func Test_UpdateField_PopulatedRegistry_AllowsAndBlocks(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_UpdateField_PopulatedRegistry")
	sd, err := insertUpdateFieldSeedData(t, db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	reg := protected.New()
	// Protect an unrelated real column on customers; "name" stays writable.
	reg.ProtectField("sales.customers", "delivery_address_id", "some_action")

	h := data.NewUpdateFieldHandler(testLog(), db.DB, data.WithProtectedRegistry(reg))

	// ALLOWED: update the (unprotected) name of the seeded customer.
	allowCfg := mustJSON(map[string]any{
		"target_entity": "sales.customers",
		"target_field":  "name",
		"new_value":     "Renamed Co",
		"conditions": []map[string]any{
			{"field_name": "id", "operator": "equals", "value": sd.Customer.ID.String()},
		},
	})
	res, err := h.Execute(context.Background(), allowCfg, workflow.ActionExecutionContext{})
	if err != nil {
		t.Fatalf("allowed update returned error: %v", err)
	}
	if m, ok := res.(map[string]any); !ok || m["status"] != "success" {
		t.Fatalf("allowed update result = %#v, want status=success", res)
	}

	// BLOCKED: the protected column is rejected with the sentinel.
	blockCfg := mustJSON(map[string]any{
		"target_entity": "sales.customers",
		"target_field":  "delivery_address_id",
		"new_value":     "00000000-0000-0000-0000-000000000000",
		"conditions": []map[string]any{
			{"field_name": "id", "operator": "equals", "value": sd.Customer.ID.String()},
		},
	})
	if _, err := h.Execute(context.Background(), blockCfg, workflow.ActionExecutionContext{}); !errors.Is(err, protected.ErrProtectedField) {
		t.Fatalf("blocked update err = %v, want ErrProtectedField", err)
	}
}

func mustJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
