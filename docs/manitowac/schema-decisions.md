# Schema Decisions — Manitowac

Architectural decisions for representing Manitowac's inventory in Ichor. Updated as decisions are finalized.

---

## Decision 1: Supplier Variants (Italy vs MR&P)

**Status:** ✅ Decided — one product, two supplier records, lot tracking

**The question:** Items like "Italy Frames" and "MR&P Frames" appear as separate line items in the spreadsheet. Should they be:
- One product with two supplier records (industry standard)
- Two separate products

**Industry standard:** Every major ERP (SAP, Oracle, NetSuite) uses a single material master per logical product, with multiple vendor/supplier records. Inventory is tracked in aggregate; lot tracking captures which supplier a specific batch came from. This is the `products.products` + `procurement.supplier_products` + `inventory.lot_trackings` pattern that Ichor already supports.

**Ichor's existing schema supports this without changes:**
```
products.products: id=X, name="Frames", sku="FRM-001"
  ↓
procurement.supplier_products:
  row 1: product_id=X, supplier_id=MR&P_id, supplier_part_number="119554"
  row 2: product_id=X, supplier_id=Italy_id, supplier_part_number=TBD
        ↓
inventory.lot_trackings:
  lot A: supplier_product_id → MR&P row  (MR&P stock)
  lot B: supplier_product_id → Italy row (Italy stock)
```

**Key diagnostic:** Are Italy and MR&P frames interchangeable on the production floor? If a worker can pull from either bin without any process change → same product. If job orders must specify which source → separate products.

**Decision rationale:** Italy and MR&P parts are technically interchangeable on the floor but have a significant quality difference. This is the textbook lot-tracking use case:
- Aggregate inventory counts as one product
- Quality differences handled via `lot_trackings.quality_status` (can quarantine a specific supplier's batch without affecting the other)
- Defect traceability: finished relays can be traced back to which supplier's lot was used
- Supplier quality reporting: defect rates correlated by `supplier_product_id` over time

**No schema changes needed.** Splitting into two products would lose the aggregate view and duplicate all product metadata — solving the wrong problem.

---

## Decision 2: Manufacturing Stage Field

**Status:** ✅ Decided — two orthogonal fields (industry standard A + B hybrid)

**The question:** The client wants stage tracking (Inbound → Received → Processing → Assembly → Calibration → QA → Outbound). No such field exists in the current Ichor schema.

**Decision:** Two independent, nullable fields. Companies can use either, both, or neither.

### Field 1: `products.products.inventory_type` (enum)
Answers: **"What kind of thing is this product?"** — static, set once per product definition.

Values: `raw_material` / `component` / `consumable` / `wip` / `finished_good`

- SAP equivalent: Material Type. NetSuite: Item Type. Fishbowl: Part Type.
- Frames → `component`, Steel Bars → `raw_material`, Finished Relay → `finished_good`, Unfinished Coil → `wip`

### Field 2: `inventory.zones.stage` (enum)
Answers: **"Where in the process is this batch right now?"** — dynamic, changes as stock moves between zones.

Values: `inbound` / `received` / `processing` / `assembly` / `calibration` / `qa` / `outbound`

- An item's current stage = the stage of the zone it physically lives in
- Zero manual status updates — movement IS the stage transition
- Nullable: businesses that don't use zone stages simply don't configure it

**Flexibility matrix:**

| Customer type | inventory_type | zone stage |
|--------------|----------------|-----------|
| Simple retailer | ✓ (everything is `finished_good`) | not configured |
| Distributor | optional | ✓ (receiving → outbound) |
| Manufacturer (Manitowac) | ✓ full | ✓ full |
| Basic stock tracker | not configured | not configured |

**Schema changes:** 2 new nullable enum columns — one on `products.products`, one on `inventory.zones`. No new tables, no new joins for customers who don't need it.

---

## Decision 3: WIP Items Without Part Numbers

**Status:** Decision made — assign part numbers

**The question:** ~70% of line items have no part number (unfinished coils, raw materials, WIP stages).

**Decision:** Assign internal part numbers for all WIP items. Part number will be required.
- Unfinished coils #2–#9 → assign internal SKUs (e.g., `COIL-UNF-002` through `COIL-UNF-009`)
- Raw materials → assign internal SKUs
- Assembly WIP (unfinished relays, unfinished bases) → assign internal SKUs

**Schema impact:** None — `products.products.sku` already exists and is the right field.

---

## Decision 4: Unit of Measure / Packaging

**Status:** ✅ Decided — new `products.product_uoms` table (industry standard)

**The question:** Three UOM situations to solve:
1. Packaging units — "1000/Box", "350/tray", "6000/Box, 3000/Bin"
2. Wire dual-unit — spools (physical count) AND bobbins (estimated yield, ~700/spool, approximate)
3. Raw materials by weight/length — received as bulk, consumed by piece

**Decision:** New `products.product_uoms` table. One row per UOM per product. One row marked as base (`is_base = true`, `conversion_factor = 1`); all others store their conversion factor relative to the base.

```sql
CREATE TABLE products.product_uoms (
    id                UUID PRIMARY KEY,
    product_id        UUID NOT NULL REFERENCES products.products(id),
    name              TEXT NOT NULL,          -- "spool", "bobbin", "box", "each", "lb"
    abbreviation      TEXT,                   -- "spl", "bob", "ea"
    conversion_factor NUMERIC NOT NULL,       -- relative to base UOM
    is_base           BOOLEAN NOT NULL DEFAULT false,
    is_approximate    BOOLEAN NOT NULL DEFAULT false,  -- display with ~ prefix, never use for allocation
    notes             TEXT,
    created_date      TIMESTAMP,
    updated_date      TIMESTAMP
);
```

**Examples:**

| Product | name | is_base | factor | is_approximate |
|---------|------|---------|--------|----------------|
| Wire 0.05 | spool | true | 1 | false |
| Wire 0.05 | bobbin | false | 700 | **true** |
| Covers, Black | each | true | 1 | false |
| Covers, Black | box | false | 1000 | false |
| Steel Bars | bar | true | 1 | false |
| Steel Bars | lb | false | 8.5 | false |

**Approximate flag rule:** `is_approximate = true` means the conversion is an estimate. The UI displays these with a `~` prefix. Business logic must never use approximate quantities for exact allocation, sales commitments, or ledger entries — planning/visibility only.

**Migration:** Deprecate `products.products.units_per_case` (existing integer). Data migrated to `product_uoms` as a "case" UOM entry where applicable.

---

## Schema Gaps Summary

| Gap | Current Schema | Decision |
|-----|---------------|----------|
| Supplier source variants (Italy vs MR&P) | `supplier_products` + `lot_trackings` | ✅ No change — one product, two supplier records, lot tracks source |
| WIP part numbers | `products.sku` exists | ✅ Assign internal SKUs operationally |
| Manufacturing stage — product type | None | ✅ Add `inventory_type` enum to `products.products` |
| Manufacturing stage — batch location | None | ✅ Add `stage` enum to `inventory.zones` |
| Unit of measure / packaging | `units_per_case` (single int) | ✅ New `products.product_uoms` table, deprecate `units_per_case` |
| Wire dual-unit (spool + bobbin) | Not supported | ✅ Handled by `product_uoms.is_approximate` flag |
| Coil number (#2–#9) as variant | Not modeled | ⚠ Open — separate products or product attribute? |
