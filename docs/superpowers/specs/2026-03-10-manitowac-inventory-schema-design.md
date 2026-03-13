# Design Spec: Manitowac Inventory Schema Extensions

**Date:** 2026-03-10
**Customer:** Manitowac (test customer — relay manufacturer)
**Scope:** Schema additions to support manufacturing inventory tracking

---

## Background

Manitowac manufactures electromechanical relays for engines. Their December 2024 inventory snapshot (87 distinct items, 6 manufacturing stages) revealed four gaps between Ichor's current schema and what's needed to model a manufacturing business. These changes are designed to be **general-purpose** — useful for all Ichor customers, not just Manitowac.

Full customer notes: `docs/manitowac/`

---

## Decisions Made

### 1. Supplier Source Variants (Italy vs MR&P)
No schema change needed. The existing `supplier_products` + `lot_trackings` model handles this. One product per logical item, two supplier records, lot tracking captures source. Quality differences handled via `lot_trackings.quality_status`.

### 2. WIP Part Numbers
No schema change needed. Assign internal SKUs operationally (e.g., `COIL-UNF-002`). `products.products.sku` already exists.

### 3. Bill of Materials / Production Orders
Out of scope for this phase. BOM and manufacturing execution (MRP) are a future system. Current work covers inventory tracking only.

---

## Schema Changes

### Change A: `products.products.inventory_type` (enum)

**Answers:** "What kind of thing is this product?"

Static classification per product definition. Set once, doesn't change.

```sql
-- Add to products.products
ALTER TABLE products.products
  ADD COLUMN inventory_type TEXT;

-- Allowed values (enforced at application layer):
-- raw_material  — Steel bars, wire, resins, solder
-- component     — Frames, cores, pins, bobbins, springs
-- consumable    — Mylar tape, solder (used in process, not in finished product)
-- wip           — Unfinished coils, unfinished relays, unfinished bases
-- finished_good — Finished relays, finished bases
```

**Nullable:** Yes — existing customers without manufacturing don't need to configure this.

**Manitowac examples:**

| Product | inventory_type |
|---------|---------------|
| Steel Bars | raw_material |
| Wire 0.05 gauge | raw_material |
| Frames | component |
| Bobbins (plastic) | component |
| Springs #30 | component |
| Mylar Tape | consumable |
| Unfinished Coil #4 | wip |
| Unfinished Relay | wip |
| Finished Relay | finished_good |

---

### Change B: `inventory.zones.stage` (enum)

**Answers:** "Where in the process is this batch right now?"

Dynamic — an item's current stage is derived from the zone it physically lives in. Moving inventory between zones is the stage transition. No manual status updates required.

```sql
-- Add to inventory.zones
ALTER TABLE inventory.zones
  ADD COLUMN stage TEXT;

-- Allowed values:
-- inbound      — Receiving dock, incoming goods area
-- received     — Checked-in stock awaiting put-away
-- processing   — Winding floor, riveting stations
-- assembly     — Assembly floor
-- calibration  — Calibration area (sub-stage of assembly)
-- qa           — Quality hold / inspection area
-- outbound     — Finished goods, shipping staging
```

**Nullable:** Yes — businesses that don't use zone-based stage tracking simply don't configure this. No errors, no empty data, feature is transparently absent.

**Manitowac zone setup example:**

| Zone | stage |
|------|-------|
| Receiving Dock | inbound |
| Component Storage | received |
| Winding Floor | processing |
| Coil Riveting | processing |
| Contact Riveting | processing |
| Assembly Floor | assembly |
| Calibration Bench | calibration |
| QA Hold | qa |
| Finished Goods Bay | outbound |

**Customer flexibility:**

| Customer type | inventory_type | zone stage |
|--------------|----------------|-----------|
| Simple retailer | ✓ (finished_good) | not configured |
| Distributor | optional | ✓ (inbound / outbound) |
| Manufacturer | ✓ full | ✓ full |
| Basic tracker | not configured | not configured |

---

### Change C: New `products.product_uoms` table

**Answers:** "What units is this product measured in, and how do they convert?"

Replaces the existing `products.products.units_per_case` integer column. Supports unlimited UOMs per product, each with a conversion factor relative to the base unit.

```sql
CREATE TABLE products.product_uoms (
    id                UUID         PRIMARY KEY,
    product_id        UUID         NOT NULL REFERENCES products.products(id) ON DELETE CASCADE,
    name              TEXT         NOT NULL,   -- "spool", "bobbin", "box", "each", "lb", "tray"
    abbreviation      TEXT,                    -- "spl", "bob", "ea"
    conversion_factor NUMERIC      NOT NULL,   -- relative to base UOM (base = 1)
    is_base           BOOLEAN      NOT NULL DEFAULT false,
    is_approximate    BOOLEAN      NOT NULL DEFAULT false,
    notes             TEXT,
    created_date      TIMESTAMPTZ  NOT NULL,
    updated_date      TIMESTAMPTZ  NOT NULL
);

-- One base UOM per product
CREATE UNIQUE INDEX product_uoms_base_idx
  ON products.product_uoms(product_id)
  WHERE is_base = true;
```

**The `is_approximate` flag:** When true, the conversion is an estimate. UI displays with `~` prefix. Business logic must never use approximate quantities for exact allocation, sales commitments, or ledger entries — planning and visibility only. The wire bobbin count (~700/spool) is the canonical example.

**Migration:** `products.products.units_per_case` deprecated and migrated to `product_uoms` as a packaging UOM entry where set.

**Manitowac examples:**

| Product | UOM name | is_base | factor | is_approximate |
|---------|----------|---------|--------|----------------|
| Wire 0.05 | spool | true | 1 | false |
| Wire 0.05 | bobbin | false | 700 | **true** |
| Covers, Black | each | true | 1 | false |
| Covers, Black | box | false | 1000 | false |
| Bobbins (plastic) | each | true | 1 | false |
| Bobbins (plastic) | box | false | 6000 | false |
| Bobbins (plastic) | bin | false | 3000 | false |
| Steel Bars | bar | true | 1 | false |
| Finished Relay | each | true | 1 | false |
| Finished Relay | box | false | 100 | false |

---

## Migration Path

All three changes are additive and non-breaking:

1. Add `inventory_type` column to `products.products` (nullable, no existing data affected)
2. Add `stage` column to `inventory.zones` (nullable, no existing data affected)
3. Create `products.product_uoms` table
4. Migrate existing `units_per_case` values to `product_uoms` rows
5. Deprecate `units_per_case` (keep column for one release cycle, then drop)

---

## Out of Scope

- **Bill of Materials (BOM)** — consumption relationships between raw materials and WIP/finished goods
- **Production Orders / MRP** — manufacturing execution, planned vs actual output
- **Coil #2–#9 / Spring #29–#35 product definitions** — pending client answers on whether these are separate products or variants
- **Armature bent vs unbent modeling** — pending client answer
- **KS suffix classification** — pending client answer

---

## Open Client Questions

See `docs/manitowac/client-qa.md` for the full list. Questions that could affect this design:

- Coil numbers — separate products or variants? (affects product data setup, not schema)
- Armatures bent vs unbent — WIP stage or separate product? (affects zone config or product data)
- Calibration — separate physical zone? (affects zone setup)
- Mylar Tape — component or consumable? (affects `inventory_type` value, not schema)
