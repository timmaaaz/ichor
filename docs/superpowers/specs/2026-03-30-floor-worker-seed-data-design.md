# Floor Worker Seed Data Design

**Date:** 2026-03-30
**Status:** Approved
**Goal:** After `make seed-frontend`, a tester can log in as floor_worker1 and run every manual test suite (T-03 through T-14) without creating data manually. Seed data must support physical warehouse walkthroughs with printed barcode labels.

---

## Strategy

**Unified modification** of existing `TestSeed*` functions. The debate analysis confirmed:
- 67 integration test suites use `cmp.Diff` against seed structs (self-healing on value changes)
- Zero tests filter by status and assert counts
- Mixed-value distributions already exist on filterable fields without incident
- The `n` parameter already decouples volume per caller

For 4 new seed categories (put-away tasks, pick tasks, cycle counts, workflow approvals), we wire existing `TestSeed*` functions into `seedFrontend.go` or write new ones following the standard `{entity}bus/testutil.go` pattern.

---

## Physical Warehouse Context

The office will be set up as a mini warehouse with labeled shelves:
- **~15 physical locations** across 3 aisles (A = receiving/put-away, B = main pick, C = reserve)
- **~20 products** with printed barcode labels
- **2-3 testers** running walkthroughs concurrently
- Location codes printed as barcodes: `A-01-01-01`, `B-02-01-03`, etc.

---

## PR Plan (7 PRs)

| PR | Scope | Est. Files | Tier |
|----|-------|-----------|------|
| **1a** | Products + Inspections | ~33 | Blocks suites |
| **1b** | Locations + Inventory Items | ~28 | Blocks suites |
| **1c** | Put-Away + Pick Tasks | ~5 | Blocks suites |
| **2a** | Lots + Serials | ~25 | Blocks steps |
| **2b** | POs + Transfers | ~20 | Blocks steps |
| **3a** | Adjustments + Transactions | ~11 | Nice to have |
| **3b** | Cycle Counts + Approvals + Config | ~5 | Nice to have |

**Dependency order:** 1a must land before 2a (lots/serials FK to products). 1b must land before 1c (tasks FK to locations). All tier-1 PRs before tier-2. All tier-2 before tier-3.

---

## Frontend Seed Count Overrides

`seedFrontend.go` passes its own `n` parameter, independent of integration tests. Bumped counts for physical testing:

| Entity | Integration Test `n` | Frontend Seed `n` | Rationale |
|--------|---------------------|-------------------|-----------|
| Products | 30 | 20 (unchanged) | 20 is plenty for ~20 physical SKUs |
| Locations | varies | 25 (unchanged) | 15 physical + extras for reserve |
| Put-away tasks | 2 (API tests) | **15** | 3-5 per tester, 2-3 testers |
| Pick tasks | 1 (API tests) | **15** | 3-5 per tester, 2-3 testers |
| POs | 10 | **15** | More active POs for receiving |
| PO line items | 25 | **40** | More items to receive per PO |
| Inspections | 25 (API tests) | 10 (unchanged) | 5 assigned to floor_worker1 |
| Cycle count sessions | varies | **3** | 1 per tester |
| Workflow approvals | 0 | **5** | Supervisor inbox content |
| Inventory item qty | random | **100+ per item** | Survive multiple pick/transfer sessions |

---

## PR 1a: Products + Inspections

### Products (`business/domain/products/productbus/testutil.go`)

**File:** `TestSeedProducts` / `TestSeedProductsHistorical`

| Field | Current | New |
|-------|---------|-----|
| Name | `Product{idx}` | Cycle through realistic pool: `"Industrial Bearing 6205"`, `"Nitrile Gloves Box/100"`, `"Hydraulic Filter HF-302"`, `"LED Panel Light 60W"`, `"Stainless Steel Bolt M10"`, `"Thermal Paste Tube 5g"`, `"Safety Goggles Clear"`, `"Rubber Gasket Set"`, `"Wire Spool CAT6 100m"`, `"Epoxy Adhesive 2-Part"` |
| UpcCode | `UpcCode{idx}` | `fmt.Sprintf("0123456789%02d", idx%100)` — real 12-digit format |
| TrackingType | unset (defaults `"none"`) | `trackingTypes[idx%3]` cycling `"none"`, `"lot"`, `"serial"` |
| IsPerishable | `idx%2==0 && idx%5==0` (10%) | `trackingType == "lot"` — perishable when lot-tracked |

**Test impact:** Self-healing via `cmp.Diff`. Verify `productapi/create_test.go` UPC conflict detection uses `sd.Products[0].UpcCode` dynamically (confirmed by agent).

### Inspections (`business/domain/inventory/inspectionbus/testutil.go`)

**File:** `TestSeedInspections`

| Field | Current | New |
|-------|---------|-----|
| InspectorID | `inspectorIDs[idx%len(inspectorIDs)]` | First 5 → floor_worker1 UUID, rest cycle normally |
| Status | all `"pending"` | Keep all `"pending"` (requirement satisfied) |
| LotID | not set | Set on at least 1 floor_worker1 inspection. Add `lotIDs []uuid.UUID` param to `TestSeedInspections` if not already present. |
| InspectionDate | not set / random | `time.Now().AddDate(0, 0, -(idx%7))` — within last 7 days |
| NextInspectionDate | not set / random | `time.Now().AddDate(0, 0, idx+7)` — future dates |

**Test impact:** Count assertions unchanged. Inspector assignment change is self-healing.

---

## PR 1b: Locations + Inventory Items

### Locations (`business/domain/inventory/inventorylocationbus/testutil.go`)

**File:** `TestSeedInventoryLocations`

| Field | Current | New |
|-------|---------|-----|
| LocationCode | `Aisle{idx}-Rack{idx}-Shelf{idx}-Bin{idx}` | Warehouse-realistic: `fmt.Sprintf("%s-%02d-%02d-%02d", aisles[idx%3], rack, shelf, bin)` |
| IsPickLocation | `idx%2 == 0` | Keep (gives ~50% pick locations) |
| IsReserveLocation | `idx%2==0 && idx%5==0` | `idx%2 != 0` — complement of pick (every location is one or the other) |

**Location code scheme:**
```
aisles := []string{"A", "B", "C"}
rack  := (idx/12)%5 + 1   // 01-05
shelf := (idx/3)%4 + 1    // 01-04
bin   := idx%6 + 1         // 01-06
```

Produces: `A-01-01-01`, `B-01-01-02`, `C-01-01-03`, `A-01-02-04`, ...

**Physical mapping:**
- Aisle A = Receiving / put-away area
- Aisle B = Main pick area
- Aisle C = Reserve / overflow

### Inventory Items (`business/domain/inventory/inventoryitembus/testutil.go`)

**File:** `TestSeedInventoryItems`

| Field | Current | New |
|-------|---------|-----|
| Quantity | random / low | Set `Quantity: 100 + rand.Intn(50)` floor value — survives multiple pick/transfer sessions |
| Product coverage | random assignment | Ensure each tracking type (none/lot/serial) has inventory at >= 1 location |
| Multi-product locations | not guaranteed | At least 1 location with 2+ products (for cycle count testing) |

**Test impact:** `"TESTLOC-001"` hardcoded in API test is set via separate `.Update()` — safe. Count assertions unchanged.

---

## PR 1c: Put-Away + Pick Tasks

### Put-Away Tasks (`business/domain/inventory/putawaytaskbus/testutil.go`)

**File:** `TestSeedPutAwayTasks` — already exists, not called from seedFrontend.

**Changes to testutil.go:**
- Ensure at least 1 task has `assigned_to = floor_worker1` UUID
- Ensure `reference_number` is set (e.g., `fmt.Sprintf("PO-HIST-%d", idx+1)`)
- Products must have real UPCs (satisfied by PR 1a)

**Orchestration:** Add to `seedFrontend.go` dependency chain after `seedProcurement`:
```
seedProcurement(...)
    ↓
seedPutAwayTasks(ctx, busDomain, foundation, products, inventory)   ← NEW
    ↓
seedPickTasks(ctx, busDomain, foundation, products, inventory, sales) ← NEW
```

New file: `business/sdk/dbtest/seed_tasks.go`

### Pick Tasks (`business/domain/inventory/picktaskbus/testutil.go`)

**File:** `TestSeedPickTasks` — already exists, not called from seedFrontend.

**Changes to testutil.go:**
- Tasks reference sales order line items, products, and pick locations
- At least 1 task for a lot-tracked product
- At least 1 task for a serial-tracked product
- Pick locations must have sufficient inventory (qty >= pick qty)

**Test impact:** Isolated — only 2-3 API test files reference these. No count assertions change.

---

## PR 2a: Lots + Serials

### Lot Trackings (`business/domain/inventory/lottrackingsbus/testutil.go`)

**File:** `TestSeedLotTrackings`

| Field | Current | New |
|-------|---------|-----|
| LotNumber | `LotNumber{idx}` | `fmt.Sprintf("LOT-2026-%03d", idx+1)` |
| ManufactureDate | `RandomDate()` (2020-2030) | `time.Now().AddDate(0, -3, 0)` — 3 months ago |
| ExpirationDate | `RandomDate()` (2020-2030) | Distributed: idx 0 = now+5d (urgent), idx 1 = now+20d (warning), idx 2 = now+60d (monitor), rest = now+30-180d |
| ReceivedDate | `RandomDate()` (2020-2030) | `time.Now().AddDate(0, 0, -rand.Intn(30))` — within last 30 days |
| QualityStatus | cycles all 5 | idx 0-2 = `"good"` (dashboard visible), ensure at least 1 `"quarantined"` |

**Dependency:** Lots must link to lot-tracked products from PR 1a.

### Serial Numbers (`business/domain/inventory/serialnumberbus/testutil.go`)

**File:** `TestSeedSerialNumbers`

| Field | Current | New |
|-------|---------|-----|
| SerialNumber | `SN-{idx}` | `fmt.Sprintf("SN-2026-%04d", idx+1)` |
| Status | `Status-{idx%2}` | `statuses[idx%4]`: `"available"`, `"reserved"`, `"quarantined"`, `"shipped"` |
| LotID | not set | Set on at least 1 serial (lot link navigation) |

**Dependency:** Serials must link to serial-tracked products from PR 1a.

**Test impact:** Count assertions unchanged (`Total: 15`, `Total: 20`). Scan API tests reference values dynamically — self-healing.

---

## PR 2b: POs + Transfers

### Purchase Orders (`business/domain/procurement/purchaseorderbus/testutil.go`)

**File:** `TestSeedPurchaseOrdersHistorical`

| Field | Current | New |
|-------|---------|-----|
| ExpectedDeliveryDate | Always `orderDate + 14 days` | Varied: at least 1 = today, 1 = past (overdue), 1 = within 7 days, rest spread |
| Status | Cycles DRAFT→CLOSED | Ensure at least 1 `SENT` or `APPROVED` (active for receiving) |
| ActualDeliveryDate | not checked | Leave null on overdue PO (triggers overdue display) |
| PO line items | 25 total | Ensure at least 1 has `quantity_received < quantity_ordered` |

**Frontend seed override:** Bump to `n=15` POs, `n=40` line items.

**Test impact:** HIGHEST RISK in this PR — `purchaseorderapi/query_test.go` has 8+ Total assertions. Must audit each for status filtering. If all are unfiltered queries, counts only change if `n` changes in tests (it won't).

### Transfer Orders (`business/domain/inventory/transferorderbus/testutil.go`)

**File:** `TestSeedTransferOrders`

| Field | Current | New |
|-------|---------|-----|
| Status | all `StatusPending` | `statuses[idx%5]`: pending, pending, approved, approved, completed (~40/40/20 split) |

**Test impact:** `Total: 10` assertion is unfiltered — safe.

---

## PR 3a: Adjustments + Transactions

### Inventory Adjustments (`business/domain/inventory/inventoryadjustmentbus/testutil.go`)

| Field | Current | New |
|-------|---------|-----|
| ApprovalStatus | all `"pending"` (bus default) | Keep 5+ pending, set at least 1 `"approved"` |
| ReasonCode | all `"other"` | Cycle: `"damaged"`, `"expired"`, `"found_stock"`, `"theft"`, `"data_entry_error"` |
| CreatedBy | random | At least 2 set to floor_worker1 UUID |

### Inventory Transactions (`business/domain/inventory/inventorytransactionbus/testutil.go`)

| Field | Current | New |
|-------|---------|-----|
| TransactionType | all `"Movement"` | Cycle: `"receive"`, `"pick"`, `"putaway"`, `"transfer"`, `"adjustment"`, `"count"` |
| TransactionDate | `time.Now()` | Already good — keep |
| CreatedBy | random | Set some to floor_worker1 UUID |

**Test impact:** Low — 6 and 5 files respectively, no filtered count assertions.

---

## PR 3b: Cycle Counts + Approvals + Config

### Cycle Count Sessions

**File:** `business/domain/inventory/cyclecountsessionbus/testutil.go` — `TestSeedCycleCountSessions` already exists.

**Orchestration:** Wire into `seedFrontend.go` after `seedInventory`. New call in `seed_tasks.go` or `seed_inventory.go`.
- Seed 3 sessions referencing locations with inventory items
- Status allowing worker participation

### Workflow Approval Instances

**File:** `business/domain/workflow/approvalrequestbus/testutil.go` — does NOT exist.

**New work:**
1. Add `ApprovalRequest *approvalrequestbus.Business` to BusDomain struct in `dbtest.go`
2. Instantiate in `newBusDomains()`
3. Write `TestSeedApprovalRequests` in `approvalrequestbus/testutil.go`
4. Call from `seedFrontend.go` after `seedWorkflow`
5. Seed 5 instances: all `status = "pending"`, with `task_name`, `description`, `requester_id`, `source_entity_name`, `source_entity_id`

### Config Settings

**File:** Seed via `busDomain.ConfigStore` in `seedFrontend.go` or new `seed_config.go`.

| Key | Value | Purpose |
|-----|-------|---------|
| `inventory.expiry_warning_1_days` | `7` | Urgent bucket threshold |
| `inventory.expiry_warning_2_days` | `30` | Warning bucket threshold |
| `inventory.expiry_warning_3_days` | `90` | Monitor bucket threshold |

---

## Fixed UUIDs Reference

Used across multiple PRs:

```
floor_worker1   id=c0000000-0000-4000-8000-000000000001
FLOOR_WORKER    role_id=b0000000-0000-4000-8000-000000000001
admin_gopher    id=5cf37266-3473-4006-984f-9325122678b7
```

---

## Test Verification Strategy

For each PR:
1. Run integration tests for all modified domains: `go test ./business/domain/{area}/{entity}bus/...`
2. Run API tests for all affected seed_test.go consumers: `go test ./api/cmd/services/ichor/tests/{area}/{entity}api/...`
3. Run `make seed-frontend` on a fresh DB and verify data via psql spot-checks
4. `go build ./...` to catch any compilation issues

---

## Risk Mitigation

- **Self-healing tests:** `cmp.Diff` pattern means value changes propagate through `sd.*` structs automatically
- **Count stability:** `n` parameters unchanged for integration test callers
- **Compile-time safety:** Any missing fields in `Create()` calls caught by compiler
- **PR isolation:** Each PR is independently shippable and testable
- **Dependency order:** PRs respect FK chains — no PR depends on an unmerged PR except by tier ordering
