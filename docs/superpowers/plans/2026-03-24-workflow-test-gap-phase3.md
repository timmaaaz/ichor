# Phase 3: CreatePurchaseOrderHandler Tests

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add tests for `CreatePurchaseOrderHandler` — the largest untested action handler (503 lines) covering PO creation with supplier lookups, event extraction, and transactional line items.

**Architecture:** Needs real Postgres via `dbtest` for supplier product queries, PO creation, and line item creation. Follow the existing action handler test patterns (see `approve_po_test.go`, `check_inventory_test.go` for examples).

**Tech Stack:** Go testing, `dbtest`, `cmp.diff`, real bus dependencies

**Spec:** `docs/superpowers/specs/2026-03-24-workflow-test-gap-remediation-design.md` (Phase 3)

---

### Task 1: Understand Test Patterns and Dependencies

- [ ] **Step 1: Read existing handler test patterns**

Read these files to understand how action handler tests are structured:
- `business/sdk/workflow/workflowactions/procurement/approve_po_test.go`
- `business/sdk/workflow/workflowactions/inventory/check_inventory_test.go`

Key patterns to follow:
- How they get DB access and bus instances
- How they construct the handler with real dependencies
- How they seed test data (products, suppliers, etc.)
- How they build `workflow.ActionExecutionContext`

- [ ] **Step 2: Read the handler's dependencies**

Read to understand what `NewCreatePurchaseOrderHandler` needs:
- `business/domain/procurement/purchaseorderbus/` — NewBusiness constructor
- `business/domain/procurement/purchaseorderlineitembus/` — NewBusiness constructor
- `business/domain/procurement/supplierproductbus/` — NewBusiness constructor, filter, testutil

---

### Task 2: Validate Tests

**Files:**
- Create: `business/sdk/workflow/workflowactions/procurement/createpo_test.go`
- Reference: `business/sdk/workflow/workflowactions/procurement/createpo.go:100-185` (Validate method)

- [ ] **Step 1: Write Validate test cases**

```go
package procurement_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/procurement"
)

func TestCreatePurchaseOrder_Validate(t *testing.T) {
	handler := procurement.NewCreatePurchaseOrderHandler(nil, nil, nil, nil, nil)

	validStatusID := uuid.New().String()
	validWarehouseID := uuid.New().String()
	validLocationID := uuid.New().String()
	validCurrencyID := uuid.New().String()
	validProductID := uuid.New().String()
	validLineItemStatusID := uuid.New().String()

	validConfig := procurement.CreatePurchaseOrderConfig{
		PurchaseOrderStatusID: validStatusID,
		DeliveryWarehouseID:   validWarehouseID,
		DeliveryLocationID:    validLocationID,
		CurrencyID:            validCurrencyID,
		LineItems: []procurement.CreatePOLineItemConfig{
			{
				ProductID:        validProductID,
				QuantityOrdered:  10,
				LineItemStatusID: validLineItemStatusID,
			},
		},
	}

	tests := []struct {
		name      string
		modify    func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "valid config",
			modify:  func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig { return c },
			wantErr: false,
		},
		{
			name: "missing purchase_order_status_id",
			modify: func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
				c.PurchaseOrderStatusID = ""
				return c
			},
			wantErr:   true,
			errSubstr: "purchase_order_status_id is required",
		},
		{
			name: "invalid purchase_order_status_id",
			modify: func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
				c.PurchaseOrderStatusID = "not-a-uuid"
				return c
			},
			wantErr:   true,
			errSubstr: "invalid purchase_order_status_id",
		},
		{
			name: "missing delivery_warehouse_id",
			modify: func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
				c.DeliveryWarehouseID = ""
				return c
			},
			wantErr:   true,
			errSubstr: "delivery_warehouse_id is required",
		},
		{
			name: "missing delivery_location_id",
			modify: func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
				c.DeliveryLocationID = ""
				return c
			},
			wantErr:   true,
			errSubstr: "delivery_location_id is required",
		},
		{
			name: "missing currency_id",
			modify: func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
				c.CurrencyID = ""
				return c
			},
			wantErr:   true,
			errSubstr: "currency_id is required",
		},
		{
			name: "no line items when source_from_event is false",
			modify: func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
				c.LineItems = nil
				return c
			},
			wantErr:   true,
			errSubstr: "at least one line item is required",
		},
		{
			name: "source_from_event without default_line_item_status_id",
			modify: func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
				c.SourceFromEvent = true
				c.LineItems = nil
				return c
			},
			wantErr:   true,
			errSubstr: "default_line_item_status_id is required",
		},
		{
			name: "source_from_event with valid default_line_item_status_id",
			modify: func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
				c.SourceFromEvent = true
				c.DefaultLineItemStatusID = validLineItemStatusID
				c.LineItems = nil
				return c
			},
			wantErr: false,
		},
		{
			name: "line item missing product_id",
			modify: func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
				c.LineItems[0].ProductID = ""
				return c
			},
			wantErr:   true,
			errSubstr: "product_id is required",
		},
		{
			name: "line item zero quantity",
			modify: func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
				c.LineItems[0].QuantityOrdered = 0
				return c
			},
			wantErr:   true,
			errSubstr: "quantity_ordered must be greater than 0",
		},
		{
			name: "line item negative quantity",
			modify: func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
				c.LineItems[0].QuantityOrdered = -5
				return c
			},
			wantErr:   true,
			errSubstr: "quantity_ordered must be greater than 0",
		},
		{
			name: "line item missing line_item_status_id",
			modify: func(c procurement.CreatePurchaseOrderConfig) procurement.CreatePurchaseOrderConfig {
				c.LineItems[0].LineItemStatusID = ""
				return c
			},
			wantErr:   true,
			errSubstr: "line_item_status_id is required",
		},
		{
			name:      "invalid json",
			modify:    nil, // special case
			wantErr:   true,
			errSubstr: "invalid configuration format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var configBytes json.RawMessage
			if tt.modify == nil {
				configBytes = json.RawMessage(`{invalid`)
			} else {
				cfg := tt.modify(validConfig)
				data, _ := json.Marshal(cfg)
				configBytes = data
			}

			err := handler.Validate(configBytes)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.errSubstr)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if tt.wantErr && err != nil && tt.errSubstr != "" {
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Fatalf("error %q does not contain %q", err.Error(), tt.errSubstr)
				}
			}
		})
	}
}
```

Add `"strings"` to the import block.

- [ ] **Step 2: Run Validate tests**

Run: `go test ./business/sdk/workflow/workflowactions/procurement/... -run TestCreatePurchaseOrder_Validate -v -count=1`
Expected: All PASS.

- [ ] **Step 3: Commit**

```
git add business/sdk/workflow/workflowactions/procurement/createpo_test.go
git commit -m "test(createpo): add Validate method tests"
```

---

### Task 3: Metadata Method Tests

- [ ] **Step 1: Add metadata tests to the same file**

```go
func TestCreatePurchaseOrder_Metadata(t *testing.T) {
	handler := procurement.NewCreatePurchaseOrderHandler(nil, nil, nil, nil, nil)

	t.Run("GetType", func(t *testing.T) {
		if got := handler.GetType(); got != "create_purchase_order" {
			t.Fatalf("expected create_purchase_order, got %s", got)
		}
	})

	t.Run("SupportsManualExecution", func(t *testing.T) {
		if !handler.SupportsManualExecution() {
			t.Fatal("expected true")
		}
	})

	t.Run("IsAsync", func(t *testing.T) {
		if handler.IsAsync() {
			t.Fatal("expected false")
		}
	})

	t.Run("GetDescription", func(t *testing.T) {
		desc := handler.GetDescription()
		if desc == "" {
			t.Fatal("expected non-empty description")
		}
	})

	t.Run("GetOutputPorts", func(t *testing.T) {
		ports := handler.GetOutputPorts()
		if len(ports) != 3 {
			t.Fatalf("expected 3 output ports, got %d", len(ports))
		}
		portNames := make(map[string]bool)
		for _, p := range ports {
			portNames[p.Name] = true
		}
		for _, expected := range []string{"created", "no_supplier_found", "failure"} {
			if !portNames[expected] {
				t.Fatalf("missing output port: %s", expected)
			}
		}
	})

	t.Run("GetEntityModifications", func(t *testing.T) {
		mods := handler.GetEntityModifications(nil)
		if len(mods) != 2 {
			t.Fatalf("expected 2 entity modifications, got %d", len(mods))
		}
	})
}
```

- [ ] **Step 2: Run and commit**

Run: `go test ./business/sdk/workflow/workflowactions/procurement/... -run TestCreatePurchaseOrder_Metadata -v -count=1`

```
git add business/sdk/workflow/workflowactions/procurement/createpo_test.go
git commit -m "test(createpo): add metadata method tests"
```

---

### Task 4: Execute Tests (DB-backed)

This requires real Postgres with seeded suppliers, products, PO statuses, warehouses, etc.

- [ ] **Step 1: Study the seed data patterns**

Read `business/sdk/workflow/workflowactions/procurement/approve_po_test.go` to see how it seeds:
- Purchase order statuses
- Suppliers and supplier products
- Warehouses, locations
- Currencies

You'll need to replicate this seed pattern for the Execute tests.

- [ ] **Step 2: Write Execute happy path and error tests**

The Execute test requires a full `dbtest.NewDatabase` setup with seeded procurement data. Follow the pattern from `approve_po_test.go` exactly. Key test cases:

1. **Happy path**: seed a supplier product with known cost, create PO with one line item, verify PO created with correct totals
2. **No supplier found**: use a product ID with no supplier product mapping, expect `no_supplier_found` output
3. **extractFromEvent**: set `SourceFromEvent=true` and put `product_id`/`quantity` in `RawData`

Since this is DB-heavy and depends on the specific seed data available, the test structure should:
- Create a single `Test_CreatePurchaseOrder_Execute(t *testing.T)` that creates a DB and seeds procurement data
- Run subtests for each scenario

```go
func Test_CreatePurchaseOrder_Execute(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_CreatePO_Execute")

	// Seed the required procurement data.
	// Follow the pattern from approve_po_test.go for seeding:
	// - users, suppliers, supplier products, PO statuses,
	//   line item statuses, warehouses, locations, currencies
	// ... (read approve_po_test.go for exact seed pattern)

	// Then construct the handler with real bus instances and test.
}
```

The exact seed code depends on what helpers exist in the procurement bus packages. Read them first, then write the seed function.

- [ ] **Step 3: Run Execute tests**

Run: `go test ./business/sdk/workflow/workflowactions/procurement/... -run Test_CreatePurchaseOrder_Execute -v -count=1`

- [ ] **Step 4: Commit**

```
git add business/sdk/workflow/workflowactions/procurement/createpo_test.go
git commit -m "test(createpo): add Execute tests with real DB"
```

---

### Task 5: extractFromEvent Tests

- [ ] **Step 1: Add extractFromEvent test cases**

`extractFromEvent` is an unexported method, but its behavior is exercised through `Execute` with `SourceFromEvent=true`. Add test cases to the Execute test that cover:

1. Event has `product_id` and `quantity` → line items extracted correctly
2. Event has `product_id` and `reorder_quantity` (fallback field) → extracted correctly
3. Event missing `product_id` → error returned (output="failure")
4. Event missing both `quantity` and `reorder_quantity` → error returned
5. Event has `quantity=0` → error returned

These are tested through the `Execute` method by setting `SourceFromEvent=true` on the config and varying `execCtx.RawData`.

- [ ] **Step 2: Run all tests**

Run: `go test ./business/sdk/workflow/workflowactions/procurement/... -v -count=1`
Expected: All PASS.

- [ ] **Step 3: Final commit**

```
git add business/sdk/workflow/workflowactions/procurement/createpo_test.go
git commit -m "test(createpo): add extractFromEvent integration tests"
```
