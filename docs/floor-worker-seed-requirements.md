# Floor Worker Seed Data Requirements

Requirements for `make seed-frontend` to fully support the floor worker manual test plan.

**Goal:** After running `make seed-frontend`, a tester should be able to log in and run every test suite without creating data manually.

---

## 1. Products — Realistic Values & Tracking Variants

**Current state:** 20 products with generic names (`Product7342`), fake UPCs (`UpcCode7342`), all `tracking_type = "none"`, most `is_perishable = false`.

**Requirements:**

- [ ] At least 3 products must have real 12-digit UPC codes (e.g., `012345678901`, `012345678902`, `012345678903`) so they can be encoded as scannable barcodes
- [ ] At least 1 product with `tracking_type = "lot"` and `is_perishable = true` — triggers the lot capture step during receiving (T-03 steps 3.9–3.10) and lot-aware flows in picking
- [ ] At least 1 product with `tracking_type = "serial"` — triggers the serial capture step during picking (T-05 step 5.11)
- [ ] At least 1 product with `tracking_type = "none"` — baseline flow without tracking prompts
- [ ] Realistic product names help testers confirm they're looking at the right item (e.g., `"Industrial Bearing 6205"`, `"Nitrile Gloves Box/100"`, `"Hydraulic Filter HF-302"`) — nice to have, not blocking

---

## 2. Inventory Locations — Realistic Warehouse Codes

**Current state:** 25 locations with codes like `Aisle7342-Rack7342-Shelf7342-Bin7342`. Technically parses (4 dash-separated segments) but unusable for printed barcode labels and confusing to testers.

**Requirements:**

- [ ] Location codes must use short, warehouse-realistic values: aisle = `A`–`C`, rack = `01`–`05`, shelf = `01`–`04`, bin = `01`–`06`
- [ ] Resulting barcodes like `A-01-02-03`, `B-03-01-05`, `C-02-04-01` — printable, scannable, recognizable
- [ ] At least 3 distinct locations needed for transfer testing (source, destination, and one more for adjustments/counting)
- [ ] At least 1 location should be a pick location (`is_pick_location = true`)
- [ ] At least 1 location should be a reserve location (`is_reserve_location = true`)

---

## 3. Inventory Items — Stock On Hand

**Current state:** 30 inventory items seeded. Need to verify they cover the right products at the right locations with nonzero quantities.

**Requirements:**

- [ ] Each of the 3 key products (none/lot/serial tracked) must have inventory at least 1 location with `quantity > 0`
- [ ] At least 1 location must have multiple products (for cycle count testing — T-07)
- [ ] Quantities should be large enough to support picking and adjustment tests (e.g., 50+)

---

## 4. Put-Away Tasks

**Current state:** `TestSeedPutAwayTasks()` exists in `putawaytaskbus/testutil.go` but is NOT called in `seedFrontend.go`. The put-away queue page is always empty.

**Requirements:**

- [ ] Call `TestSeedPutAwayTasks()` in `seedFrontend.go` (or equivalent)
- [ ] At least 3 put-away tasks with `status = "pending"`
- [ ] At least 1 task should have `assigned_to` set to `floor_worker1`'s UUID (`c0000000-0000-4000-8000-000000000001`) to test the "assigned to me" filter
- [ ] Each task needs: `product_id` (a product with a UPC), `location_id` (a realistic location), `quantity`, `reference_number` (e.g., `"PO-HIST-1"`)
- [ ] Products on put-away tasks must have real UPC codes (from requirement #1) so the item scan step works

---

## 5. Pick Tasks

**Current state:** `TestSeedPickTasks()` exists in `picktaskbus/testutil.go` but is NOT called in `seedFrontend.go`.

**Requirements:**

- [ ] Call `TestSeedPickTasks()` in `seedFrontend.go` (or equivalent)
- [ ] At least 3 pick tasks with `status = "pending"`
- [ ] Tasks should reference sales order line items, products, and pick locations
- [ ] At least 1 pick task for a lot-tracked product
- [ ] At least 1 pick task for a serial-tracked product
- [ ] Pick locations must have sufficient inventory (`inventory_items.quantity >= pick quantity`)

---

## 6. Lot Trackings — Expiry Spread & Quality Statuses

**Current state:** 15 lots with random dates between 2020–2030 and quality statuses cycling through `good`, `on_hold`, `quarantined`, `released`, `expired`.

**Requirements:**

- [ ] `lot_number` values should be human-readable (e.g., `LOT-2026-001` instead of `LotNumber7342`)
- [ ] At least 1 lot with `quality_status = "good"` and `expiration_date` within the next 7 days — shows in the "urgent" bucket on the expiry dashboard (T-11)
- [ ] At least 1 lot with `quality_status = "good"` and `expiration_date` within 30 days — shows in the "warning" bucket
- [ ] At least 1 lot with `quality_status = "good"` and `expiration_date` within 90 days — shows in the "monitor" bucket
- [ ] At least 1 lot with `quality_status = "quarantined"` — tests the "Contact supervisor to release" display (T-11 step 11.8)
- [ ] Lots must be linked to lot-tracked products (from requirement #1)
- [ ] Lot `expiration_date` values should be relative to seed time (e.g., `now + 5 days`, `now + 20 days`, `now + 60 days`) so the expiry dashboard always has relevant data regardless of when the seed runs

---

## 7. Serial Numbers — Real Statuses

**Current state:** 50 serial numbers with `status` alternating between `Status-0` and `Status-1`. These aren't real status values.

**Requirements:**

- [ ] `serial_number` values should be human-readable (e.g., `SN-2026-0001` instead of `SN-7342`)
- [ ] Status values must be real strings the frontend uses for badge display:
  - At least 1 with `status = "available"`
  - At least 1 with `status = "reserved"`
  - At least 1 with `status = "quarantined"` — shows lock icon (T-12 step 12.2)
  - At least 1 with `status = "shipped"` — shows check icon (T-12 step 12.3)
- [ ] Serials must be linked to the serial-tracked product (from requirement #1)
- [ ] At least 1 serial should have a `lot_id` set — tests the lot link navigation (T-12 step 12.4)

---

## 8. Quality Inspections — Assigned to Test User

**Current state:** 10 inspections with `status = "pending"` but `inspector_id` is randomly assigned from seeded users. `floor_worker1` likely has none.

**Requirements:**

- [ ] At least 2 inspections with `inspector_id = c0000000-0000-4000-8000-000000000001` (floor_worker1's UUID)
- [ ] Both should have `status = "pending"`
- [ ] At least 1 should have a `lot_id` set — enables the quarantine-on-fail flow (T-09 steps 9.7–9.8)
- [ ] `inspection_date` should be recent (within last 7 days)
- [ ] `next_inspection_date` should be in the future

---

## 9. Workflow Approval Instances

**Current state:** Workflow action templates and automation rules are seeded, but no actual pending approval instances exist. The supervisor inbox workflow section is empty.

**Requirements:**

- [ ] At least 2 approval instances with `status = "pending"`
- [ ] Each needs: `task_name` (display title), `description`, `requester_id`, `created_date`
- [ ] At least 1 should have a `source_entity_name` and `source_entity_id` so the "view details" link works
- [ ] These appear in the Supervisor Dashboard inbox (T-14 steps 14.2–14.5)

---

## 10. Purchase Orders — Date Spread

**Current state:** 10 POs (`PO-HIST-1` through `PO-HIST-10`) with statuses cycling through DRAFT → CLOSED. Dates are distributed across a 120-day historical window. `expected_delivery_date` is always 14 days after order date.

**Requirements:**

- [ ] At least 1 PO with `expected_delivery_date = today` (or within today's date range) — shows in the "Today" window on PO Visibility (T-13 step 13.1)
- [ ] At least 1 PO with `expected_delivery_date` in the past and no `actual_delivery_date` — shows as "Overdue" (T-13 step 13.3)
- [ ] At least 1 PO with `expected_delivery_date` within the next 7 days — shows in "7 Days" window (T-13 step 13.2)
- [ ] At least 1 PO with status = `PENDING_APPROVAL` or `APPROVED` or `SENT` (an active, non-received PO) — needed for the receiving flow (T-03)
- [ ] PO line items should reference products with real UPC codes (from requirement #1) so scanning during receiving works
- [ ] At least 1 PO line item with `quantity_received < quantity_ordered` — so there's something left to receive

---

## 11. Transfer Orders — Mixed Statuses

**Current state:** 20 transfers, ALL with `status = "pending"`. No approved or completed transfers exist.

**Requirements:**

- [ ] At least 2 transfers with `status = "pending"` — appear in supervisor inbox for approval (T-14)
- [ ] At least 2 transfers with `status = "approved"` — appear in the worker's execute queue (T-08 steps 8.5–8.10)
- [ ] At least 1 transfer with `status = "completed"` — tests the read-only completed view (T-08 step 8.12) and shows in History tab (T-08 step 8.2)
- [ ] Source and destination locations should use the realistic location codes (from requirement #2)

---

## 12. Inventory Adjustments — Mixed Statuses

**Current state:** 20 adjustments, all with `approval_status = "pending"` and `reason_code = "other"`.

**Requirements:**

- [ ] Keep at least 5 with `approval_status = "pending"` — supervisor inbox (T-14)
- [ ] At least 1 with `approval_status = "approved"` — shows in list as resolved
- [ ] Use varied `reason_code` values: `"damaged"`, `"expired"`, `"found_stock"`, `"theft"`, `"data_entry_error"` — not all `"other"`
- [ ] `created_by` on at least 2 should be `floor_worker1`'s UUID — so the worker's own adjustments appear in their list

---

## 13. Inventory Transactions — Activity Feed

**Current state:** 40 transactions seeded. Need variety for the supervisor Activity tab.

**Requirements:**

- [ ] Use varied `transaction_type` values: `"receive"`, `"pick"`, `"putaway"`, `"transfer"`, `"adjustment"`, `"count"`
- [ ] Recent timestamps (within last 7 days) so the Activity tab isn't showing only old data
- [ ] `created_by` should include `floor_worker1`'s UUID on some transactions

---

## 14. Cycle Count Sessions

**Current state:** Not seeded at all. Schema exists.

**Requirements:**

- [ ] At least 1 cycle count session with a status that allows the worker to participate
- [ ] Session should reference locations that have inventory items (from requirement #3)
- [ ] This unblocks T-07 scheduled count flow

---

## 15. Config Settings — Expiry Thresholds (Optional)

**Current state:** Not seeded. Frontend falls back to hardcoded defaults.

**Requirements (nice to have):**

- [ ] `inventory.expiry_warning_1_days` = `7`
- [ ] `inventory.expiry_warning_2_days` = `30`
- [ ] `inventory.expiry_warning_3_days` = `90`
- [ ] These control the color-coded buckets on the lot expiry dashboard (T-11 step 11.2)

---

## Summary — Priority Order

### Must fix (blocks entire test suites)

| # | Requirement | Test Suites Blocked |
|---|---|---|
| 1 | Products with real UPCs + tracking types | T-03, T-05, T-10 (scan steps) |
| 2 | Realistic location codes | T-04, T-05, T-06, T-07, T-08 (all location scanning) |
| 4 | Seed put-away tasks | T-04 (entire suite) |
| 5 | Seed pick tasks | T-05 (task assignment) |
| 8 | Inspections assigned to floor_worker1 | T-09 (entire suite) |

### Must fix (blocks individual steps)

| # | Requirement | Steps Blocked |
|---|---|---|
| 6 | Lot expiry date spread | T-11 expiry dashboard |
| 7 | Real serial statuses | T-12 status icons |
| 9 | Workflow approval instances | T-14 workflow section |
| 10 | PO date spread | T-13 Today/Overdue/7 Days |
| 11 | Mixed transfer statuses | T-08 approved/completed views |

### Nice to have (improves realism)

| # | Requirement | Impact |
|---|---|---|
| 3 | Inventory item coverage | Better cycle count / picking tests |
| 12 | Mixed adjustment statuses + reasons | More realistic supervisor inbox |
| 13 | Transaction type variety | Better activity feed |
| 14 | Cycle count sessions | T-07 scheduled flow |
| 15 | Config settings for expiry thresholds | Expiry dashboard colors |
