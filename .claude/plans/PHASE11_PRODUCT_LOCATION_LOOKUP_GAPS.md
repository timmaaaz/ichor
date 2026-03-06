# Phase 11 Backend Gap Analysis — Product & Location Quick Lookup

**Date:** 2026-02-26
**Source:** Audit of `/ichor/` backend against `docs/usage.md` Phase 11 spec
**Status:** Pre-implementation — gaps identified, ready for worktree execution

---

## Executive Summary

All 6 endpoints required by Phase 11 exist and their core filtering works. **The backend can support Phase 11 as-is.** However, the current API design forces the frontend to make 3–5 sequential API calls per scan to render a meaningful result, and there is no way for the frontend to resolve a scanned barcode to its type without trying each endpoint one by one. The gap analysis below distinguishes what is missing from what would make Phase 11 genuinely excellent.

---

## Section 1: Required Endpoint Coverage

| Endpoint from docs/usage.md | Exists | Filter Works | Notes |
|---|---|---|---|
| `GET /v1/products/products?upc_code={upc_code}` | ✅ | ✅ exact match | Full product struct returned |
| `GET /v1/inventory/inventory-items?product_id={id}` | ✅ | ✅ UUID filter | Returns UUIDs only — no names |
| `GET /v1/inventory/inventory-items?location_id={id}` | ✅ | ✅ UUID filter | Returns UUIDs only — no names |
| `GET /v1/inventory/inventory-locations/{location_id}` | ✅ | ✅ | Has aisle/rack/shelf/bin/location_code — no warehouse/zone names |
| `GET /v1/inventory/lot-trackings?lot_number={num}` | ✅ | ✅ exact match | No `product_id` in response, only `supplier_product_id` |
| `GET /v1/inventory/serial-numbers?serial_number={sn}` | ✅ | ✅ exact match | `location_id` UUID only, no location name |

---

## Section 2: Critical Gaps

These gaps directly block the Phase 11 UX goal of "scan anything → instant contextual result."

---

### GAP 1 — No Universal Barcode Resolver *(Most Impactful)*

**The problem:** Phase 11's defining feature is "scan any barcode and the system figures out what it is." The frontend currently has no way to determine barcode type without trying each endpoint sequentially:

```
Scan barcode "ABC-12345"
→ Try GET /products/products?upc_code=ABC-12345     → empty
→ Try GET /inventory/inventory-locations?location_code=ABC-12345 → empty
→ Try GET /inventory/lot-trackings?lot_number=ABC-12345  → hit!
→ Now fetch enriched lot details (supplier_product → product → location)
```

On a warehouse device with spotty WiFi, this waterfall is 4+ round trips and potentially 1–2 seconds of spinner before anything renders. That's unacceptable for a 2-second task completion target.

**What to build:**

```
GET /v1/inventory/scan?barcode={value}
```

The handler queries all four namespaces in parallel (products, locations, lots, serials), identifies the match type, and returns a single enriched response:

```json
{
  "type": "product" | "location" | "lot" | "serial" | "purchase_order" | "unknown",
  "data": { ... type-specific enriched payload ... }
}
```

**Type-specific payloads:**

For `"type": "product"`:
```json
{
  "product_id": "uuid",
  "name": "Widget A",
  "sku": "WDG-001",
  "upc_code": "012345678901",
  "tracking_type": "lot" | "serial" | "none",
  "is_perishable": true,
  "status": "active",
  "stock_summary": [
    {
      "location_id": "uuid",
      "location_code": "A1-B2-C3",
      "aisle": "A1",
      "rack": "B2",
      "shelf": "C3",
      "bin": null,
      "quantity": 42,
      "reserved_quantity": 5
    }
  ],
  "total_quantity": 42,
  "total_reserved": 5
}
```

For `"type": "location"`:
```json
{
  "location_id": "uuid",
  "location_code": "A1-B2-C3",
  "aisle": "A1",
  "rack": "B2",
  "shelf": "C3",
  "bin": null,
  "warehouse_name": "Main DC",
  "zone_name": "Cold Storage",
  "is_pick_location": true,
  "max_capacity": 500,
  "current_utilization": 78.4,
  "items": [
    {
      "product_id": "uuid",
      "sku": "WDG-001",
      "name": "Widget A",
      "quantity": 42,
      "tracking_type": "lot"
    }
  ]
}
```

For `"type": "lot"`:
```json
{
  "lot_id": "uuid",
  "lot_number": "LOT-2025-001",
  "product_id": "uuid",
  "product_name": "Widget A",
  "sku": "WDG-001",
  "supplier_product_id": "uuid",
  "expiration_date": "2025-12-31",
  "manufacture_date": "2025-01-15",
  "quality_status": "good",
  "quantity": 240,
  "locations": [
    { "location_id": "uuid", "location_code": "A1-B2-C3", "quantity": 120 }
  ]
}
```

For `"type": "serial"`:
```json
{
  "serial_id": "uuid",
  "serial_number": "SN-9482710",
  "product_id": "uuid",
  "product_name": "Widget A",
  "sku": "WDG-001",
  "lot_id": "uuid",
  "lot_number": "LOT-2025-001",
  "status": "in_stock",
  "location_id": "uuid",
  "location_code": "A1-B2-C3",
  "aisle": "A1",
  "rack": "B2",
  "shelf": "C3"
}
```

**Why this belongs in the backend:** Type detection and join resolution require database-level queries. The backend can run all four namespace lookups in parallel with Go goroutines. The frontend cannot do this without serialized HTTP calls.

**Implementation notes:**
- New handler in `api/domain/http/inventory/` — could live in a new `scanapi/` package or as a route on an existing inventory handler
- Queries products, inventory_locations, lot_trackings, serial_numbers in parallel via goroutines
- Returns first non-empty match (or `"unknown"` if none)
- The enriched response data for each type is assembled via JOINs at the DB query level, not N+1 calls in the handler

---

### GAP 2 — LotTracking Response Missing `product_id` and Product Fields

**The problem:** `GET /v1/inventory/lot-trackings?lot_number={num}` returns `supplier_product_id` — a procurement domain FK — but not `product_id`. To display "this is a lot of Widget A (SKU: WDG-001)" the frontend needs:

```
GET /inventory/lot-trackings?lot_number=LOT-001         → supplier_product_id
GET /procurement/supplier-products/{supplier_product_id} → product_id
GET /products/products/{product_id}                      → product name, SKU
```

That is 3 sequential calls before anything useful renders.

**Relevant files:**
- `app/domain/inventory/lottrackingsapp/model.go` — LotTracking response struct (add fields here)
- `business/domain/inventory/lottrackingsbus/stores/lottrackingsdb/db.go` — base SELECT query (add JOIN here)

**What to build:** Add denormalized fields to the `LotTracking` response struct:

```go
ProductID   string `json:"product_id"`
ProductName string `json:"product_name"`
ProductSKU  string `json:"product_sku"`
```

The existing `product_id` filter already does a subquery through `procurement.supplier_products → products.products` — the JOIN path is proven. Extend the base SELECT to include these columns.

---

### GAP 3 — SerialNumber Has No Location Resolution Endpoint

**The problem:** Lots have a dedicated sub-resource:
```
GET /v1/inventory/lot-trackings/{lot_id}/locations
```
which returns enriched location data (aisle/rack/shelf/bin + quantity per location). Serials have no equivalent. The serial response returns only `location_id` (UUID). To display where a serial is, the frontend needs a second call.

**Relevant files:**
- `api/domain/http/inventory/serialnumberapi/routes.go` — add new route
- `app/domain/inventory/serialnumberapp/serialnumberapp.go` — add handler
- `app/domain/inventory/serialnumberapp/model.go` — add response type
- `business/domain/inventory/serialnumberbus/serialnumberbus.go` — add business method

**What to build:**

Option A — New sub-resource (mirrors lot pattern, preferred):
```
GET /v1/inventory/serial-numbers/{serial_id}/location
```
Returns:
```json
{
  "location_id": "uuid",
  "location_code": "A1-B2-C3",
  "aisle": "A1",
  "rack": "B2",
  "shelf": "C3",
  "bin": null,
  "warehouse_name": "Main DC",
  "zone_name": "Cold Storage"
}
```

Option B — Enrich the base serial response (simpler, less RESTful):
Add `location_code`, `aisle`, `rack`, `shelf`, `bin` to the `SerialNumber` response via a JOIN. Since a serial has exactly one location, embedding is clean.

Option A is preferred to keep the serial endpoint consistent with the lot pattern.

---

### GAP 4 — InventoryItem Response Has No Denormalized Product or Location Names

**The problem:** `GET /v1/inventory/inventory-items?product_id={id}` is the "where is this product stored?" call. The response returns:

```json
{ "id": "uuid", "product_id": "uuid", "location_id": "uuid", "quantity": "42", ... }
```

To render "42 units at A1-B2-C3 (Cold Storage)" the frontend needs a second call to `/inventory/inventory-locations/{location_id}` for each row in the result, and a third call to `/inventory/zones/{zone_id}` to get the zone name. For a product stored across 8 locations, that's 8+ calls.

**Relevant files:**
- `api/domain/http/inventory/inventoryitemapi/filter.go` — add `include_location_details` param
- `app/domain/inventory/inventoryitemapp/model.go` — add optional location fields to response struct
- `business/domain/inventory/inventoryitembus/stores/inventoryitemdb/db.go` — add conditional JOIN

**What to build:** Add an optional `include_location_details=true` query parameter that triggers a JOIN and populates extra fields in the response:

```json
{
  "id": "uuid",
  "product_id": "uuid",
  "location_id": "uuid",
  "location_code": "A1-B2-C3",
  "aisle": "A1",
  "rack": "B2",
  "shelf": "C3",
  "bin": null,
  "zone_name": "Cold Storage",
  "warehouse_name": "Main DC",
  "quantity": "42",
  "reserved_quantity": "5"
}
```

**Note:** If GAP 1 (scan resolver) is built first, this enrichment is handled there. GAP 4 becomes lower priority once GAP 1 exists — but it is still useful for any direct inventory-items query outside the scan context.

---

## Section 3: Notable Gaps

Not blockers, but significant enough to impact Phase 11 quality.

---

### GAP 5 — Location Lookup by Code Uses ILIKE, Not Exact Match

**The situation:** `GET /v1/inventory/inventory-locations?location_code=A1-B2-C3` uses `ILIKE '%A1-B2-C3%'` — substring matching. For scan flows where the scanned code is precise, this mostly works. However:

1. If a location code `A1-B2-C3` is scanned, it will also match `ZA1-B2-C3-X` if such a record exists
2. There is no dedicated route for exact canonical lookup by code

**Relevant files:**
- `api/domain/http/inventory/inventorylocationapi/routes.go` — add new route
- `business/domain/inventory/inventorylocationbus/stores/inventorylocationdb/filter.go` line 48 — change to exact match or add second param

**What to build:** Either:
- Add a second filter param `location_code_exact=A1-B2-C3` that uses `= :value` SQL
- Or a dedicated lookup route: `GET /v1/inventory/inventory-locations/code/{code}` (404 if no match, returns single record if found)

Exact-match is the correct semantic for scan flows where the barcode encodes a verbatim location code.

---

### GAP 6 — No Serial Number Lifecycle / Assignment Tracking

**The situation:** The `SerialNumber` model has `status` and `location_id`, but no:
- `order_id` or `sales_order_id` — which order it was picked/sold on
- `customer_id` — which customer received it
- History/events table showing status transitions over time

For Phase 11 "scan a serial → see its full context," a worker currently can't answer "has this unit been sold? to whom? from which order?" with the current schema.

**What to build (schema changes required):**
- Add `order_id` (nullable FK to `sales.orders`) to `inventory.serial_numbers`
- Add `customer_id` (nullable FK to customers) to `inventory.serial_numbers`
- Or add a `serial_number_events` table: `(serial_id, event_type, from_location_id, to_location_id, order_id, user_id, created_at)` for full audit trail

This is a larger schema change. Recommend tackling in a dedicated worktree after GAPs 1–4.

---

### GAP 7 — No Product Image Support

**The situation:** The `Product` model has no `image_url` or `image_id` field. For a scan-to-lookup UI, showing a product thumbnail is a significant usability win — workers visually confirm they have the right item before taking action on it.

**What to build:** Add `image_url string` (nullable) to `products.products`. Expose it in the product response. For Phase 11 this is a minor enhancement — the core lookup works without it — but it materially improves confidence in fast-moving warehouse environments where SKUs can look similar.

---

### GAP 8 — `tracking_type` Filter Broken on Products Endpoint *(Pre-existing Bug)*

**The situation:** `GET /v1/products/products?tracking_type=lot` silently returns all products. The filter is wired in `productapp/filter.go:98` and the `QueryParams` struct but is never read from the HTTP query string.

**Relevant file:** `api/domain/http/products/productapi/filter.go`

**Fix:** Add one line:
```go
TrackingType: values.Get("tracking_type"),
```

Zero-risk trivial fix.

---

### GAP 9 — `recieved_date` Typo on Lot Trackings Filter *(Pre-existing Bug)*

**The situation:** The HTTP query parameter for received date on lot trackings is `recieved_date` (misspelled). The JSON response field is correctly `received_date`. A developer constructing a filter from the response will use the correct spelling and get no results — silent failure.

**Relevant file:** `api/domain/http/inventory/lottrackingsapi/filter.go`

**Fix:** Rename query param from `recieved_date` to `received_date`. Verify no existing callers depend on the misspelled form before deploying.

---

## Section 4: Prioritized Build List

| # | Gap | Priority | Effort | Impact |
|---|---|---|---|---|
| 1 | Unified scan resolver `GET /v1/inventory/scan?barcode={val}` | **Critical** | High | Eliminates waterfall; enables Phase 11's core UX |
| 2 | Add `product_id`, `product_name`, `product_sku` to LotTracking response | **High** | Low | Eliminates 2 extra API calls per lot scan |
| 3 | Add location details to SerialNumber (sub-endpoint or enriched response) | **High** | Low | Eliminates extra call per serial scan |
| 4 | Add `include_location_details=true` to inventory-items list | **High** | Medium | Eliminates N+1 location lookups for product stock view |
| 5 | Exact-match location code lookup endpoint | **Medium** | Low | Correct scan-flow semantics; prevents false matches |
| 6 | Serial number lifecycle fields (`order_id`, `customer_id`, events table) | **Medium** | High | Makes serial traceability genuinely useful |
| 7 | Product image URL field | **Low** | Medium | Visual confirmation UX enhancement |
| 8 | Fix `tracking_type` filter bug on products | **Low** | Trivial | Silent bug; one-line fix |
| 9 | Fix `recieved_date` typo on lot-trackings filter | **Low** | Trivial | API correctness; one-line fix |

---

## Section 5: Recommended Worktree Phasing

### Worktree 1 — Scan Resolver (GAP 1)
The architectural backbone. Build first; everything else builds on top of it or becomes less critical once it exists.

- New package: `api/domain/http/inventory/scanapi/`
- New business method: parallel namespace lookup via goroutines
- Enriched response types per scan result type
- Route: `GET /v1/inventory/scan?barcode={value}`

### Worktree 2 — Response Enrichment (GAPs 2, 3, 4)
Low-effort, high-return JOIN additions to existing endpoints.

- Add `product_id`, `product_name`, `product_sku` to LotTracking response
- Add `/serial-numbers/{id}/location` sub-endpoint
- Add `include_location_details=true` to inventory-items list

### Worktree 3 — Location Code Exact Match + Bug Fixes (GAPs 5, 8, 9)
Quick cleanup and correctness fixes.

- Add exact location code lookup
- Fix `tracking_type` filter
- Fix `recieved_date` typo

### Worktree 4 — Serial Lifecycle (GAP 6)
Schema change — requires migration, separate planning.

- Schema design for serial events / assignment tracking
- Migration
- New endpoints for serial history

### Worktree 5 — Product Images (GAP 7)
Schema change + media storage decision (S3/CDN vs. URL reference).

- Requires infrastructure decision before implementation

---

## Key File Reference

| File | Relevance |
|---|---|
| `api/domain/http/inventory/inventoryitemapi/routes.go` | Inventory items routes |
| `api/domain/http/inventory/inventoryitemapi/filter.go` | HTTP query param parsing for items |
| `app/domain/inventory/inventoryitemapp/model.go` | InventoryItem JSON response struct |
| `api/domain/http/inventory/inventorylocationapi/routes.go` | Location routes |
| `api/domain/http/inventory/inventorylocationapi/filter.go` | Location HTTP filter (location_code ILIKE at line 23) |
| `app/domain/inventory/inventorylocationapp/model.go` | InventoryLocation response struct |
| `business/domain/inventory/inventorylocationbus/stores/inventorylocationdb/filter.go` | Location SQL WHERE generation |
| `api/domain/http/inventory/lottrackingsapi/routes.go` | Lot tracking routes (incl. /{lot_id}/locations) |
| `app/domain/inventory/lottrackingsapp/model.go` | LotTracking response struct + location sub-response |
| `business/domain/inventory/lottrackingsbus/stores/lottrackingsdb/filter.go` | Lot SQL WHERE (product_id subquery pattern) |
| `api/domain/http/inventory/serialnumberapi/routes.go` | Serial number routes |
| `app/domain/inventory/serialnumberapp/model.go` | SerialNumber response struct |
| `api/domain/http/products/productapi/filter.go` | Products HTTP filter (tracking_type bug, line ~25) |
| `app/domain/products/productapp/model.go` | Product response struct + QueryParams |
