# Manitowac Inventory Schema Extensions — Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `inventory_type` to products, `stage` to zones, and a new `product_uoms` table — three additive, non-breaking schema extensions enabling manufacturing inventory tracking.

**Architecture:** Three independent changes following the standard Ardan Labs 7-layer pattern (migration → bus → db → app → api → wire). Changes A and B are column additions to existing domains. Change C is a new full domain entity. All three are nullable/additive — existing customers are unaffected.

**Tech Stack:** Go 1.23, PostgreSQL 16.4, Ardan Labs service architecture, sqlx, pgx, standard Ichor patterns.

**Design spec:** `docs/superpowers/specs/2026-03-10-manitowac-inventory-schema-design.md`
**Customer notes:** `docs/manitowac/`

---

## Scope Note

The three changes can be executed independently if needed (each has its own migration, bus, and tests). The plan executes them in a single pass for efficiency since all three share the same migration version.

---

## File Map

### Change A — `inventory_type` on products

| Action | File |
|--------|------|
| Modify | `business/sdk/migrate/sql/migrate.sql` |
| Create | `business/domain/products/productbus/inventorytype.go` |
| Modify | `business/domain/products/productbus/model.go` |
| Modify | `business/domain/products/productbus/stores/productdb/model.go` |
| Modify | `app/domain/products/productapp/model.go` |

### Change B — `stage` on zones

| Action | File |
|--------|------|
| Create | `business/domain/inventory/zonebus/stage.go` |
| Modify | `business/domain/inventory/zonebus/model.go` |
| Modify | `business/domain/inventory/zonebus/stores/zonedb/model.go` |
| Modify | `app/domain/inventory/zoneapp/model.go` |

### Change C — `product_uoms` new entity

| Action | File |
|--------|------|
| Create | `business/domain/products/productuombus/productuombus.go` |
| Create | `business/domain/products/productuombus/model.go` |
| Create | `business/domain/products/productuombus/filter.go` |
| Create | `business/domain/products/productuombus/order.go` |
| Create | `business/domain/products/productuombus/stores/productuomdb/productuomdb.go` |
| Create | `business/domain/products/productuombus/stores/productuomdb/model.go` |
| Create | `business/domain/products/productuombus/stores/productuomdb/order.go` |
| Create | `app/domain/products/productuomapp/productuomapp.go` |
| Create | `app/domain/products/productuomapp/model.go` |
| Create | `api/domain/http/products/productuomapi/productuomapi.go` |
| Create | `api/domain/http/products/productuomapi/route.go` |
| Create | `api/domain/http/products/productuomapi/filter.go` |
| Modify | `api/cmd/services/ichor/build/all/all.go` |
| Create | `api/cmd/services/ichor/tests/products/productuomapi/productuomapi_test.go` |
| Create | `api/cmd/services/ichor/tests/products/productuomapi/seed_test.go` |

---

## Chunk 1: Migration

### Task 1: Write migration SQL

**Files:**
- Modify: `business/sdk/migrate/sql/migrate.sql`

This task modifies existing `CREATE TABLE` definitions directly rather than using `ALTER TABLE`. This keeps the schema self-documenting — the full column list lives in one place. Fresh databases (tests, local dev) pick up the columns automatically. The new `product_uoms` table is added as version 2.11.

- [ ] **Step 1: Add `inventory_type` to the existing `products.products` CREATE TABLE (version 1.42, ~line 487)**

Find the block:
```sql
-- Version: 1.42
-- Description: add products
CREATE TABLE products.products (
   ...
   units_per_case INT NOT NULL,
   tracking_type VARCHAR(20) NOT NULL DEFAULT 'none' CHECK (tracking_type IN ('none', 'lot', 'serial')),
```

Add `inventory_type` and update the `units_per_case` deprecation comment so the table reads:
```sql
   units_per_case INT NOT NULL,   -- DEPRECATED: superseded by products.product_uoms
   tracking_type VARCHAR(20) NOT NULL DEFAULT 'none' CHECK (tracking_type IN ('none', 'lot', 'serial')),
   inventory_type TEXT NULL,
```

- [ ] **Step 2: Add `stage` to the existing `inventory.zones` CREATE TABLE (version 1.51, ~line 667)**

Find the block:
```sql
-- Version: 1.51
-- Description: add zones
CREATE TABLE inventory.zones (
   id UUID NOT NULL,
   warehouse_id UUID NOT NULL,
   name VARCHAR(50) NOT NULL,
   description TEXT NULL,
   created_date TIMESTAMP NOT NULL,
```

Add `stage` after `description`:
```sql
   description TEXT NULL,
   stage TEXT NULL,
   created_date TIMESTAMP NOT NULL,
```

- [ ] **Step 3: Append version 2.11 for the new `product_uoms` table only**

Add at the very end of `migrate.sql`:

```sql
-- Version: 2.11
-- Description: add product_uoms table
CREATE TABLE products.product_uoms (
    id                UUID        DEFAULT gen_random_uuid(),
    product_id        UUID        NOT NULL,
    name              TEXT        NOT NULL,
    abbreviation      TEXT,
    conversion_factor NUMERIC     NOT NULL,
    is_base           BOOLEAN     NOT NULL DEFAULT FALSE,
    is_approximate    BOOLEAN     NOT NULL DEFAULT FALSE,
    notes             TEXT,
    created_date      TIMESTAMPTZ NOT NULL,
    updated_date      TIMESTAMPTZ NOT NULL,

    PRIMARY KEY (id),
    FOREIGN KEY (product_id) REFERENCES products.products(id) ON DELETE CASCADE
);

-- Enforce one base UOM per product
CREATE UNIQUE INDEX product_uoms_base_idx
    ON products.product_uoms (product_id)
    WHERE is_base = TRUE;
```

- [ ] **Step 4: Verify the changes look correct**

```bash
grep -n "inventory_type\|stage TEXT\|product_uoms\|Version: 2.11" business/sdk/migrate/sql/migrate.sql
```

Expected: `inventory_type` near line 501, `stage TEXT` near line 671, and the full version 2.11 block at the end.

- [ ] **Step 5: Commit**

```bash
git add business/sdk/migrate/sql/migrate.sql
git commit -m "feat(migrate): add inventory_type to products, stage to zones, product_uoms table (v2.11)"
```

---

## Chunk 2: Change A — InventoryType on Products

### Task 2: Create InventoryType enum

**Files:**
- Create: `business/domain/products/productbus/inventorytype.go`

- [ ] **Step 1: Create the file**

```go
package productbus

import "fmt"

// InventoryType classifies a product's role in the supply chain.
// Nullable — customers without manufacturing do not need to set this.
type inventoryTypeSet struct {
	RawMaterial  InventoryType
	Component    InventoryType
	Consumable   InventoryType
	WIP          InventoryType
	FinishedGood InventoryType
}

// InventoryTypes is the set of allowed inventory type values.
var InventoryTypes = inventoryTypeSet{
	RawMaterial:  newInventoryType("raw_material"),
	Component:    newInventoryType("component"),
	Consumable:   newInventoryType("consumable"),
	WIP:          newInventoryType("wip"),
	FinishedGood: newInventoryType("finished_good"),
}

var inventoryTypeMap = make(map[string]InventoryType)

// InventoryType represents the product's role in the supply chain.
type InventoryType struct {
	name string
}

func newInventoryType(s string) InventoryType {
	it := InventoryType{s}
	inventoryTypeMap[s] = it
	return it
}

// String returns the string representation of the InventoryType.
func (it InventoryType) String() string {
	return it.name
}

// Equal returns true if the two InventoryTypes are equal.
func (it InventoryType) Equal(it2 InventoryType) bool {
	return it.name == it2.name
}

// MarshalText implements encoding.TextMarshaler.
func (it InventoryType) MarshalText() ([]byte, error) {
	return []byte(it.name), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (it *InventoryType) UnmarshalText(data []byte) error {
	parsed, err := ParseInventoryType(string(data))
	if err != nil {
		return err
	}
	*it = parsed
	return nil
}

// ParseInventoryType parses a string into an InventoryType.
func ParseInventoryType(value string) (InventoryType, error) {
	it, exists := inventoryTypeMap[value]
	if !exists {
		return InventoryType{}, fmt.Errorf("invalid inventory type %q", value)
	}
	return it, nil
}
```

- [ ] **Step 2: Build to verify it compiles**

```bash
go build ./business/domain/products/productbus/...
```

Expected: no errors.

### Task 3: Update Product business model

**Files:**
- Modify: `business/domain/products/productbus/model.go`

- [ ] **Step 1: Add InventoryType field to NewProduct, Product, and UpdateProduct**

In `Product` (the response struct), add after the last existing field:

```go
InventoryType *InventoryType `json:"inventory_type"`
```

In `NewProduct` (the create struct), add:

```go
InventoryType *InventoryType `json:"inventory_type"`
```

In `UpdateProduct` (the update struct), add:

```go
InventoryType *InventoryType `json:"inventory_type"`
```

> Single pointer — nil means "don't change". This is the consistent pattern across all Update structs in this codebase.

- [ ] **Step 2: Build**

```bash
go build ./business/domain/products/productbus/...
```

### Task 4: Update product DB model

**Files:**
- Modify: `business/domain/products/productbus/stores/productdb/model.go`

- [ ] **Step 1: Add InventoryType to the db struct**

In the unexported `product` db struct, add:

```go
InventoryType sql.NullString `db:"inventory_type"`
```

- [ ] **Step 2: Update toDBProduct conversion**

In `toDBProduct`, add:

```go
if bus.InventoryType != nil {
    dest.InventoryType = sql.NullString{String: bus.InventoryType.String(), Valid: true}
}
```

- [ ] **Step 3: Update toBusProduct conversion**

In `toBusProduct`, add:

```go
if db.InventoryType.Valid {
    it, err := productbus.ParseInventoryType(db.InventoryType.String)
    if err == nil {
        dest.InventoryType = &it
    }
}
```

- [ ] **Step 4: Build**

```bash
go build ./business/domain/products/productbus/...
```

### Task 5: Update product app model

**Files:**
- Modify: `app/domain/products/productapp/model.go`

- [ ] **Step 1: Add inventory_type to the app Product struct**

In the `Product` app struct (JSON response), add:

```go
InventoryType string `json:"inventory_type"`
```

- [ ] **Step 2: Update ToAppProduct**

In `ToAppProduct` (bus → app conversion), add:

```go
if bus.InventoryType != nil {
    app.InventoryType = bus.InventoryType.String()
}
```

- [ ] **Step 3: Add inventory_type to QueryParams**

In `QueryParams`, add:

```go
InventoryType string
```

- [ ] **Step 4: Add inventory_type to NewProduct and UpdateProduct app structs**

In `NewProduct`:
```go
InventoryType string `json:"inventory_type"`
```

In `UpdateProduct`:
```go
InventoryType *string `json:"inventory_type"`
```

- [ ] **Step 5: Update toBusNewProduct to parse InventoryType**

```go
if app.InventoryType != "" {
    it, err := productbus.ParseInventoryType(app.InventoryType)
    if err != nil {
        return productbus.NewProduct{}, errs.Newf(errs.InvalidArgument, "invalid inventory_type: %s", err)
    }
    dest.InventoryType = &it
}
```

- [ ] **Step 6: Update toBusUpdateProduct similarly**

```go
if app.InventoryType != nil {
    it, err := productbus.ParseInventoryType(*app.InventoryType)
    if err != nil {
        return productbus.UpdateProduct{}, errs.Newf(errs.InvalidArgument, "invalid inventory_type: %s", err)
    }
    dest.InventoryType = &it
}
```

- [ ] **Step 7: Build entire products stack**

```bash
go build ./business/domain/products/... ./app/domain/products/... ./api/domain/http/products/...
```

Expected: no errors.

- [ ] **Step 8: Commit**

```bash
git add business/domain/products/productbus/inventorytype.go \
        business/domain/products/productbus/model.go \
        business/domain/products/productbus/stores/productdb/model.go \
        app/domain/products/productapp/model.go
git commit -m "feat(products): add inventory_type enum field to product domain"
```

---

## Chunk 3: Change B — Stage on Zones

### Task 6: Create Stage enum

**Files:**
- Create: `business/domain/inventory/zonebus/stage.go`

- [ ] **Step 1: Create the file**

```go
package zonebus

import "fmt"

// stageSet holds all valid stage values for a warehouse zone.
type stageSet struct {
	Inbound     Stage
	Received    Stage
	Processing  Stage
	Assembly    Stage
	Calibration Stage
	QA          Stage
	Outbound    Stage
}

// Stages is the set of allowed zone stage values.
var Stages = stageSet{
	Inbound:     newStage("inbound"),
	Received:    newStage("received"),
	Processing:  newStage("processing"),
	Assembly:    newStage("assembly"),
	Calibration: newStage("calibration"),
	QA:          newStage("qa"),
	Outbound:    newStage("outbound"),
}

var stageMap = make(map[string]Stage)

// Stage represents the manufacturing lifecycle stage associated with a zone.
// Nullable — businesses without stage-based tracking leave this unset.
type Stage struct {
	name string
}

func newStage(s string) Stage {
	st := Stage{s}
	stageMap[s] = st
	return st
}

// String returns the string representation of the Stage.
func (s Stage) String() string {
	return s.name
}

// Equal returns true if the two Stages are equal.
func (s Stage) Equal(s2 Stage) bool {
	return s.name == s2.name
}

// MarshalText implements encoding.TextMarshaler.
func (s Stage) MarshalText() ([]byte, error) {
	return []byte(s.name), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (s *Stage) UnmarshalText(data []byte) error {
	parsed, err := ParseStage(string(data))
	if err != nil {
		return err
	}
	*s = parsed
	return nil
}

// ParseStage parses a string into a Stage.
func ParseStage(value string) (Stage, error) {
	st, exists := stageMap[value]
	if !exists {
		return Stage{}, fmt.Errorf("invalid stage %q", value)
	}
	return st, nil
}
```

- [ ] **Step 2: Build**

```bash
go build ./business/domain/inventory/zonebus/...
```

### Task 7: Update Zone business model

**Files:**
- Modify: `business/domain/inventory/zonebus/model.go`

- [ ] **Step 1: Add Stage to NewZone, Zone, and UpdateZone**

In the `Zone` response struct, add:

```go
Stage *Stage `json:"stage"`
```

In `NewZone` (the create struct), add:

```go
Stage *Stage `json:"stage"`
```

In `UpdateZone`, add:

```go
Stage *Stage `json:"stage"`
```

- [ ] **Step 2: Build**

```bash
go build ./business/domain/inventory/zonebus/...
```

### Task 8: Update zone DB model

**Files:**
- Modify: `business/domain/inventory/zonebus/stores/zonedb/model.go`

- [ ] **Step 1: Add Stage to the db struct**

```go
Stage sql.NullString `db:"stage"`
```

- [ ] **Step 2: Update toDBZone**

```go
if bus.Stage != nil {
    dest.Stage = sql.NullString{String: bus.Stage.String(), Valid: true}
}
```

- [ ] **Step 3: Update toBusZone**

```go
if db.Stage.Valid {
    st, err := zonebus.ParseStage(db.Stage.String)
    if err == nil {
        dest.Stage = &st
    }
}
```

- [ ] **Step 4: Build**

```bash
go build ./business/domain/inventory/zonebus/...
```

### Task 9: Update zone app model

**Files:**
- Modify: `app/domain/inventory/zoneapp/model.go`

- [ ] **Step 1: Add stage to the app Zone struct**

```go
Stage string `json:"stage"`
```

- [ ] **Step 2: Update ToAppZone**

```go
if bus.Stage != nil {
    app.Stage = bus.Stage.String()
}
```

- [ ] **Step 3: Add stage to NewZone, UpdateZone app structs and conversion functions**

In `NewZone`:
```go
Stage string `json:"stage"`
```

In `UpdateZone`:
```go
Stage *string `json:"stage"`
```

In `toBusNewZone`:
```go
if app.Stage != "" {
    st, err := zonebus.ParseStage(app.Stage)
    if err != nil {
        return zonebus.NewZone{}, errs.Newf(errs.InvalidArgument, "invalid stage: %s", err)
    }
    dest.Stage = &st
}
```

In `toBusUpdateZone`:
```go
if app.Stage != nil {
    st, err := zonebus.ParseStage(*app.Stage)
    if err != nil {
        return zonebus.UpdateZone{}, errs.Newf(errs.InvalidArgument, "invalid stage: %s", err)
    }
    dest.Stage = &st
}
```

- [ ] **Step 4: Add stage to QueryParams**

```go
Stage string
```

- [ ] **Step 5: Build entire inventory zones stack**

```bash
go build ./business/domain/inventory/zonebus/... ./app/domain/inventory/zoneapp/... ./api/domain/http/inventory/zoneapi/...
```

- [ ] **Step 6: Commit**

```bash
git add business/domain/inventory/zonebus/stage.go \
        business/domain/inventory/zonebus/model.go \
        business/domain/inventory/zonebus/stores/zonedb/model.go \
        app/domain/inventory/zoneapp/model.go
git commit -m "feat(zones): add stage enum field to zone domain"
```

---

## Chunk 4: Change C — ProductUOMs Business + DB Layer

### Task 10: Create productuombus model

**Files:**
- Create: `business/domain/products/productuombus/model.go`

- [ ] **Step 1: Create model.go**

```go
package productuombus

import (
	"time"

	"github.com/google/uuid"
)

// ProductUOM represents a unit of measure for a product.
type ProductUOM struct {
	ID               uuid.UUID `json:"id"`
	ProductID        uuid.UUID `json:"product_id"`
	Name             string    `json:"name"`
	Abbreviation     string    `json:"abbreviation"`
	ConversionFactor float64   `json:"conversion_factor"`
	IsBase           bool      `json:"is_base"`
	IsApproximate    bool      `json:"is_approximate"`
	Notes            string    `json:"notes"`
	CreatedDate      time.Time `json:"created_date"`
	UpdatedDate      time.Time `json:"updated_date"`
}

// NewProductUOM contains the data required to create a new UOM.
type NewProductUOM struct {
	ProductID        uuid.UUID `json:"product_id"`
	Name             string    `json:"name"`
	Abbreviation     string    `json:"abbreviation"`
	ConversionFactor float64   `json:"conversion_factor"`
	IsBase           bool      `json:"is_base"`
	IsApproximate    bool      `json:"is_approximate"`
	Notes            string    `json:"notes"`
}

// UpdateProductUOM contains the fields that can be updated on a UOM.
// Pointer fields are optional — nil means do not change.
type UpdateProductUOM struct {
	Name             *string  `json:"name"`
	Abbreviation     *string  `json:"abbreviation"`
	ConversionFactor *float64 `json:"conversion_factor"`
	IsBase           *bool    `json:"is_base"`
	IsApproximate    *bool    `json:"is_approximate"`
	Notes            *string  `json:"notes"`
}
```

### Task 11: Create productuombus filter and order

**Files:**
- Create: `business/domain/products/productuombus/filter.go`
- Create: `business/domain/products/productuombus/order.go`

- [ ] **Step 1: Create filter.go**

```go
package productuombus

import "github.com/google/uuid"

// QueryFilter holds the available filters for querying product UOMs.
type QueryFilter struct {
	ID        *uuid.UUID
	ProductID *uuid.UUID
	IsBase    *bool
	Name      *string
}
```

- [ ] **Step 2: Create order.go**

Look at an existing order.go in the products domain (e.g., `business/domain/products/productbus/order.go`) and follow the same pattern. Define `DefaultOrderBy` and the exported order-by constant names (`OrderByID`, `OrderByProductID`, `OrderByName`, `OrderByCreatedDate`) that the app layer will reference.

```go
package productuombus

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.NewBy(OrderByName, order.ASC)

// Set of fields that the results can be ordered by.
const (
    OrderByID          = "id"
    OrderByProductID   = "product_id"
    OrderByName        = "name"
    OrderByCreatedDate = "created_date"
)
```

### Task 11.5: Create productuombus errors

**Files:**
- Create: `business/domain/products/productuombus/errors.go`

- [ ] **Step 1: Create errors.go**

```go
package productuombus

import "errors"

// Set of error variables for CRUD operations.
var (
    ErrNotFound    = errors.New("product uom not found")
    ErrUniqueEntry = errors.New("product uom entry is not unique")
)
```

- [ ] **Step 2: Build**

```bash
go build ./business/domain/products/productuombus/...
```

### Task 12: Create productuombus business layer

**Files:**
- Create: `business/domain/products/productuombus/productuombus.go`

- [ ] **Step 1: Create productuombus.go**

```go
package productuombus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Storer defines the required persistence interface for ProductUOMs.
type Storer interface {
	Create(ctx context.Context, uom ProductUOM) error
	Update(ctx context.Context, uom ProductUOM) error
	Delete(ctx context.Context, uomID uuid.UUID) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]ProductUOM, error)
	QueryByID(ctx context.Context, uomID uuid.UUID) (ProductUOM, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
}

// Business manages the set of APIs for product UOM access.
type Business struct {
	log      *logger.Logger
	delegate *delegate.Delegate
	storer   Storer
}

// NewBusiness constructs a Business for product UOM API access.
func NewBusiness(log *logger.Logger, delegate *delegate.Delegate, storer Storer) *Business {
	return &Business{
		log:      log,
		delegate: delegate,
		storer:   storer,
	}
}

// Create adds a new product UOM to the system.
func (b *Business) Create(ctx context.Context, npu NewProductUOM) (ProductUOM, error) {
	now := time.Now()

	uom := ProductUOM{
		ID:               uuid.New(),
		ProductID:        npu.ProductID,
		Name:             npu.Name,
		Abbreviation:     npu.Abbreviation,
		ConversionFactor: npu.ConversionFactor,
		IsBase:           npu.IsBase,
		IsApproximate:    npu.IsApproximate,
		Notes:            npu.Notes,
		CreatedDate:      now,
		UpdatedDate:      now,
	}

	if err := b.storer.Create(ctx, uom); err != nil {
		return ProductUOM{}, fmt.Errorf("create: %w", err)
	}

	return uom, nil
}

// Update modifies information about an existing product UOM.
func (b *Business) Update(ctx context.Context, uom ProductUOM, upu UpdateProductUOM) (ProductUOM, error) {
	if upu.Name != nil {
		uom.Name = *upu.Name
	}
	if upu.Abbreviation != nil {
		uom.Abbreviation = *upu.Abbreviation
	}
	if upu.ConversionFactor != nil {
		uom.ConversionFactor = *upu.ConversionFactor
	}
	if upu.IsBase != nil {
		uom.IsBase = *upu.IsBase
	}
	if upu.IsApproximate != nil {
		uom.IsApproximate = *upu.IsApproximate
	}
	if upu.Notes != nil {
		uom.Notes = *upu.Notes
	}

	uom.UpdatedDate = time.Now()

	if err := b.storer.Update(ctx, uom); err != nil {
		return ProductUOM{}, fmt.Errorf("update: %w", err)
	}

	return uom, nil
}

// Delete removes a product UOM from the system.
func (b *Business) Delete(ctx context.Context, uom ProductUOM) error {
	if err := b.storer.Delete(ctx, uom.ID); err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	return nil
}

// Query retrieves a list of product UOMs from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]ProductUOM, error) {
	uoms, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	return uoms, nil
}

// QueryByID finds the product UOM by the specified ID.
func (b *Business) QueryByID(ctx context.Context, uomID uuid.UUID) (ProductUOM, error) {
	uom, err := b.storer.QueryByID(ctx, uomID)
	if err != nil {
		return ProductUOM{}, fmt.Errorf("querybyid: %w", err)
	}
	return uom, nil
}

// Count returns the total number of product UOMs.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	return b.storer.Count(ctx, filter)
}
```

- [ ] **Step 2: Build**

```bash
go build ./business/domain/products/productuombus/...
```

### Task 13: Create productuomdb store

**Files:**
- Create: `business/domain/products/productuombus/stores/productuomdb/model.go`
- Create: `business/domain/products/productuombus/stores/productuomdb/productuomdb.go`
- Create: `business/domain/products/productuombus/stores/productuomdb/order.go`

- [ ] **Step 1: Create model.go**

```go
package productuomdb

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/products/productuombus"
)

type productUOM struct {
	ID               uuid.UUID      `db:"id"`
	ProductID        uuid.UUID      `db:"product_id"`
	Name             string         `db:"name"`
	Abbreviation     sql.NullString `db:"abbreviation"`
	ConversionFactor float64        `db:"conversion_factor"`
	IsBase           bool           `db:"is_base"`
	IsApproximate    bool           `db:"is_approximate"`
	Notes            sql.NullString `db:"notes"`
	CreatedDate      time.Time      `db:"created_date"`
	UpdatedDate      time.Time      `db:"updated_date"`
}

func toDBProductUOM(bus productuombus.ProductUOM) productUOM {
	return productUOM{
		ID:               bus.ID,
		ProductID:        bus.ProductID,
		Name:             bus.Name,
		Abbreviation:     sql.NullString{String: bus.Abbreviation, Valid: bus.Abbreviation != ""},
		ConversionFactor: bus.ConversionFactor,
		IsBase:           bus.IsBase,
		IsApproximate:    bus.IsApproximate,
		Notes:            sql.NullString{String: bus.Notes, Valid: bus.Notes != ""},
		CreatedDate:      bus.CreatedDate.UTC(),
		UpdatedDate:      bus.UpdatedDate.UTC(),
	}
}

func toBusProductUOM(db productUOM) productuombus.ProductUOM {
	return productuombus.ProductUOM{
		ID:               db.ID,
		ProductID:        db.ProductID,
		Name:             db.Name,
		Abbreviation:     db.Abbreviation.String,
		ConversionFactor: db.ConversionFactor,
		IsBase:           db.IsBase,
		IsApproximate:    db.IsApproximate,
		Notes:            db.Notes.String,
		CreatedDate:      db.CreatedDate.Local(),
		UpdatedDate:      db.UpdatedDate.Local(),
	}
}

func toBusProductUOMs(dbs []productUOM) []productuombus.ProductUOM {
	bus := make([]productuombus.ProductUOM, len(dbs))
	for i, db := range dbs {
		bus[i] = toBusProductUOM(db)
	}
	return bus
}
```

- [ ] **Step 2: Create order.go**

Follow the exact pattern of any existing `*db/order.go` file (e.g., `productdb/order.go`). Map the allowed `order.By` field names (`id`, `product_id`, `name`, `created_date`) to their SQL column names.

- [ ] **Step 3: Create productuomdb.go**

```go
package productuomdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/products/productuombus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/jmoiern/sqlx"
)

// Store manages the set of APIs for product_uoms database access.
type Store struct {
	log *logger.Logger
	db  sqlx.ExtContext
}

// NewStore constructs the api for data access.
func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// Create inserts a new product UOM into the database.
func (s *Store) Create(ctx context.Context, uom productuombus.ProductUOM) error {
	const q = `
    INSERT INTO products.product_uoms
        (id, product_id, name, abbreviation, conversion_factor, is_base, is_approximate, notes, created_date, updated_date)
    VALUES
        (:id, :product_id, :name, :abbreviation, :conversion_factor, :is_base, :is_approximate, :notes, :created_date, :updated_date)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBProductUOM(uom)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Update modifies a product UOM in the database.
func (s *Store) Update(ctx context.Context, uom productuombus.ProductUOM) error {
	const q = `
    UPDATE products.product_uoms
    SET
        name             = :name,
        abbreviation     = :abbreviation,
        conversion_factor = :conversion_factor,
        is_base          = :is_base,
        is_approximate   = :is_approximate,
        notes            = :notes,
        updated_date     = :updated_date
    WHERE
        id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBProductUOM(uom)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Delete removes a product UOM from the database.
func (s *Store) Delete(ctx context.Context, uomID uuid.UUID) error {
	const q = `DELETE FROM products.product_uoms WHERE id = :id`

	data := struct {
		ID uuid.UUID `db:"id"`
	}{ID: uomID}

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Count returns the number of product UOMs in the store.
func (s *Store) Count(ctx context.Context, filter productuombus.QueryFilter) (int, error) {
	var count struct {
		Count int `db:"count"`
	}

	const q = `SELECT count(1) AS count FROM products.product_uoms`

	var buf bytes.Buffer
	buf.WriteString(q)
	applyFilter(filter, &buf)

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), filter, &count); err != nil {
		return 0, fmt.Errorf("namedquerystruct: %w", err)
	}
	return count.Count, nil
}

// Query retrieves a list of product UOMs from the database.
func (s *Store) Query(ctx context.Context, filter productuombus.QueryFilter, orderBy order.By, page page.Page) ([]productuombus.ProductUOM, error) {
	const q = `SELECT * FROM products.product_uoms`

	var buf bytes.Buffer
	buf.WriteString(q)
	applyFilter(filter, &buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}
	buf.WriteString(orderByClause)
	buf.WriteString(page.InSQL())

	var dbs []productUOM
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), filter, &dbs); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}
	return toBusProductUOMs(dbs), nil
}

// QueryByID retrieves a single product UOM by its ID.
func (s *Store) QueryByID(ctx context.Context, uomID uuid.UUID) (productuombus.ProductUOM, error) {
	data := struct {
		ID uuid.UUID `db:"id"`
	}{ID: uomID}

	const q = `SELECT * FROM products.product_uoms WHERE id = :id`

	var db productUOM
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &db); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return productuombus.ProductUOM{}, fmt.Errorf("namedquerystruct: %w", productuombus.ErrNotFound)
		}
		return productuombus.ProductUOM{}, fmt.Errorf("namedquerystruct: %w", err)
	}
	return toBusProductUOM(db), nil
}

// applyFilter adds WHERE clauses for the provided filter.
func applyFilter(filter productuombus.QueryFilter, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		wc = append(wc, "id = :id")
	}
	if filter.ProductID != nil {
		wc = append(wc, "product_id = :product_id")
	}
	if filter.IsBase != nil {
		wc = append(wc, "is_base = :is_base")
	}
	if filter.Name != nil {
		wc = append(wc, "name = :name")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
```

> Note: Add `"strings"` to the imports in `productuomdb.go`. `ErrNotFound` is defined in `productuombus/errors.go` (Task 11.5).

- [ ] **Step 4: Build**

```bash
go build ./business/domain/products/productuombus/...
```

- [ ] **Step 5: Run tests for the products domain**

```bash
go test ./business/domain/products/productuombus/...
```

- [ ] **Step 6: Commit**

```bash
git add business/domain/products/productuombus/
git commit -m "feat(productuombus): add product UOM business and db layers"
```

---

## Chunk 5: Change C — ProductUOMs App + API + Wiring

### Task 14: Create productuomapp layer

**Files:**
- Create: `app/domain/products/productuomapp/order.go`
- Create: `app/domain/products/productuomapp/model.go`
- Create: `app/domain/products/productuomapp/productuomapp.go`

- [ ] **Step 1: Create order.go**

```go
package productuomapp

import (
	"github.com/timmaaaz/ichor/business/domain/products/productuombus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	"id":           productuombus.OrderByID,
	"product_id":   productuombus.OrderByProductID,
	"name":         productuombus.OrderByName,
	"created_date": productuombus.OrderByCreatedDate,
}
```

- [ ] **Step 2: Create model.go**

```go
package productuomapp

import (
	"encoding/json"
	"strconv"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/products/productuombus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// QueryParams holds the raw query string values for filtering.
type QueryParams struct {
	Page      string
	Rows      string
	OrderBy   string
	ID        string
	ProductID string
	IsBase    string
	Name      string
}

// ProductUOM is the app-layer representation of a product UOM.
type ProductUOM struct {
	ID               string `json:"id"`
	ProductID        string `json:"product_id"`
	Name             string `json:"name"`
	Abbreviation     string `json:"abbreviation"`
	ConversionFactor string `json:"conversion_factor"`
	IsBase           bool   `json:"is_base"`
	IsApproximate    bool   `json:"is_approximate"`
	Notes            string `json:"notes"`
	CreatedDate      string `json:"created_date"`
	UpdatedDate      string `json:"updated_date"`
}

// Encode implements web.Encoder.
func (app ProductUOM) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// ToAppProductUOM converts a bus ProductUOM to an app ProductUOM.
func ToAppProductUOM(bus productuombus.ProductUOM) ProductUOM {
	return ProductUOM{
		ID:               bus.ID.String(),
		ProductID:        bus.ProductID.String(),
		Name:             bus.Name,
		Abbreviation:     bus.Abbreviation,
		ConversionFactor: strconv.FormatFloat(bus.ConversionFactor, 'f', -1, 64),
		IsBase:           bus.IsBase,
		IsApproximate:    bus.IsApproximate,
		Notes:            bus.Notes,
		CreatedDate:      bus.CreatedDate.Format("2006-01-02T15:04:05Z"),
		UpdatedDate:      bus.UpdatedDate.Format("2006-01-02T15:04:05Z"),
	}
}

// ToAppProductUOMs converts a slice of bus ProductUOMs.
func ToAppProductUOMs(uoms []productuombus.ProductUOM) []ProductUOM {
	app := make([]ProductUOM, len(uoms))
	for i, u := range uoms {
		app[i] = ToAppProductUOM(u)
	}
	return app
}

// =============================================================================
// Create

// NewProductUOM is the app-layer create request.
type NewProductUOM struct {
	ProductID        string  `json:"product_id" validate:"required,min=36,max=36"`
	Name             string  `json:"name" validate:"required"`
	Abbreviation     string  `json:"abbreviation"`
	ConversionFactor float64 `json:"conversion_factor" validate:"required"`
	IsBase           bool    `json:"is_base"`
	IsApproximate    bool    `json:"is_approximate"`
	Notes            string  `json:"notes"`
}

// Decode implements web.Decoder.
func (app *NewProductUOM) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

// Validate checks the data in the model is considered clean.
func (app NewProductUOM) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewProductUOM(app NewProductUOM) (productuombus.NewProductUOM, error) {
	productID, err := uuid.Parse(app.ProductID)
	if err != nil {
		return productuombus.NewProductUOM{}, errs.Newf(errs.InvalidArgument, "invalid product_id: %s", err)
	}

	return productuombus.NewProductUOM{
		ProductID:        productID,
		Name:             app.Name,
		Abbreviation:     app.Abbreviation,
		ConversionFactor: app.ConversionFactor,
		IsBase:           app.IsBase,
		IsApproximate:    app.IsApproximate,
		Notes:            app.Notes,
	}, nil
}

// =============================================================================
// Update

// UpdateProductUOM is the app-layer update request.
type UpdateProductUOM struct {
	Name             *string  `json:"name"`
	Abbreviation     *string  `json:"abbreviation"`
	ConversionFactor *float64 `json:"conversion_factor"`
	IsBase           *bool    `json:"is_base"`
	IsApproximate    *bool    `json:"is_approximate"`
	Notes            *string  `json:"notes"`
}

// Decode implements web.Decoder.
func (app *UpdateProductUOM) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

// Validate checks the data in the model is considered clean.
func (app UpdateProductUOM) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusUpdateProductUOM(app UpdateProductUOM) productuombus.UpdateProductUOM {
	return productuombus.UpdateProductUOM{
		Name:             app.Name,
		Abbreviation:     app.Abbreviation,
		ConversionFactor: app.ConversionFactor,
		IsBase:           app.IsBase,
		IsApproximate:    app.IsApproximate,
		Notes:            app.Notes,
	}
}

// =============================================================================
// Query helpers

func parseQueryParams(qp QueryParams) (productuombus.QueryFilter, order.By, page.Page, error) {
	// Follow the exact pattern from any existing *app/model.go parseQueryParams.
	// Parse page/rows into page.Page, parse OrderBy into order.By using the bus
	// DefaultOrderBy, parse filter fields.
	filter := productuombus.QueryFilter{}

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return productuombus.QueryFilter{}, order.By{}, page.Page{}, errs.Newf(errs.InvalidArgument, "invalid id: %s", err)
		}
		filter.ID = &id
	}

	if qp.ProductID != "" {
		pid, err := uuid.Parse(qp.ProductID)
		if err != nil {
			return productuombus.QueryFilter{}, order.By{}, page.Page{}, errs.Newf(errs.InvalidArgument, "invalid product_id: %s", err)
		}
		filter.ProductID = &pid
	}

	if qp.Name != "" {
		filter.Name = &qp.Name
	}

	// Parse page/rows/orderBy following the same pattern as other *app packages.
	// See app/domain/products/productapp/model.go for the exact parseQueryParams implementation.
	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return productuombus.QueryFilter{}, order.By{}, page.Page{}, errs.Newf(errs.InvalidArgument, "page: %s", err)
	}

	ob, err := order.Parse(orderByFields, qp.OrderBy, productuombus.DefaultOrderBy)
	if err != nil {
		return productuombus.QueryFilter{}, order.By{}, page.Page{}, errs.Newf(errs.InvalidArgument, "orderby: %s", err)
	}

	return filter, ob, pg, nil
}
```

- [ ] **Step 2: Create productuomapp.go**

```go
package productuomapp

import (
	"context"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/products/productuombus"
	"github.com/timmaaaz/ichor/business/sdk/query"
)

// App manages the set of app layer API functions for the product UOM domain.
type App struct {
	productuombus *productuombus.Business
}

// NewApp constructs a product UOM App.
func NewApp(productuombus *productuombus.Business) *App {
	return &App{productuombus: productuombus}
}

// Create adds a new product UOM.
func (a *App) Create(ctx context.Context, app NewProductUOM) (ProductUOM, error) {
	np, err := toBusNewProductUOM(app)
	if err != nil {
		return ProductUOM{}, err
	}

	uom, err := a.productuombus.Create(ctx, np)
	if err != nil {
		return ProductUOM{}, err
	}

	return ToAppProductUOM(uom), nil
}

// Update modifies an existing product UOM.
func (a *App) Update(ctx context.Context, app UpdateProductUOM, uomID uuid.UUID) (ProductUOM, error) {
	uom, err := a.productuombus.QueryByID(ctx, uomID)
	if err != nil {
		return ProductUOM{}, err
	}

	updated, err := a.productuombus.Update(ctx, uom, toBusUpdateProductUOM(app))
	if err != nil {
		return ProductUOM{}, err
	}

	return ToAppProductUOM(updated), nil
}

// Delete removes a product UOM.
func (a *App) Delete(ctx context.Context, uomID uuid.UUID) error {
	uom, err := a.productuombus.QueryByID(ctx, uomID)
	if err != nil {
		return err
	}

	return a.productuombus.Delete(ctx, uom)
}

// Query returns a list of product UOMs.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[ProductUOM], error) {
	filter, ob, pg, err := parseQueryParams(qp)
	if err != nil {
		return query.Result[ProductUOM]{}, err
	}

	uoms, err := a.productuombus.Query(ctx, filter, ob, pg)
	if err != nil {
		return query.Result[ProductUOM]{}, err
	}

	total, err := a.productuombus.Count(ctx, filter)
	if err != nil {
		return query.Result[ProductUOM]{}, err
	}

	return query.NewResult(ToAppProductUOMs(uoms), total, pg), nil
}

// QueryByID returns a single product UOM by ID.
func (a *App) QueryByID(ctx context.Context, uomID uuid.UUID) (ProductUOM, error) {
	uom, err := a.productuombus.QueryByID(ctx, uomID)
	if err != nil {
		return ProductUOM{}, err
	}

	return ToAppProductUOM(uom), nil
}
```

- [ ] **Step 3: Build**

```bash
go build ./app/domain/products/productuomapp/...
```

### Task 15: Create productuomapi layer

**Files:**
- Create: `api/domain/http/products/productuomapi/productuomapi.go`
- Create: `api/domain/http/products/productuomapi/route.go`
- Create: `api/domain/http/products/productuomapi/filter.go`

- [ ] **Step 1: Create filter.go**

```go
package productuomapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/products/productuomapp"
	"github.com/timmaaaz/ichor/foundation/web"
)

func parseQueryParams(r *http.Request) (productuomapp.QueryParams, error) {
	values := r.URL.Query()

	return productuomapp.QueryParams{
		Page:      web.GetParam(values, "page"),
		Rows:      web.GetParam(values, "rows"),
		OrderBy:   web.GetParam(values, "orderBy"),
		ID:        web.GetParam(values, "id"),
		ProductID: web.GetParam(values, "product_id"),
		IsBase:    web.GetParam(values, "is_base"),
		Name:      web.GetParam(values, "name"),
	}, nil
}
```

- [ ] **Step 2: Create productuomapi.go**

```go
package productuomapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/products/productuomapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	productuomapp *productuomapp.App
}

func newAPI(app *productuomapp.App) *api {
	return &api{productuomapp: app}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app productuomapp.NewProductUOM
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	uom, err := api.productuomapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return uom
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app productuomapp.UpdateProductUOM
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	uomID, err := uuid.Parse(web.Param(r, "product_uom_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	uom, err := api.productuomapp.Update(ctx, app, uomID)
	if err != nil {
		return errs.NewError(err)
	}

	return uom
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	uomID, err := uuid.Parse(web.Param(r, "product_uom_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := api.productuomapp.Delete(ctx, uomID); err != nil {
		return errs.NewError(err)
	}

	return nil
}

func (api *api) query(ctx context.Context, r *http.Request) web.Encoder {
	qp, err := parseQueryParams(r)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	uoms, err := api.productuomapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return uoms
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	uomID, err := uuid.Parse(web.Param(r, "product_uom_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	uom, err := api.productuomapp.QueryByID(ctx, uomID)
	if err != nil {
		return errs.NewError(err)
	}

	return uom
}
```

- [ ] **Step 3: Create route.go**

```go
package productuomapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/products/productuomapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/core/permissionsbus"
	"github.com/timmaaaz/ichor/business/domain/products/productuombus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log            *logger.Logger
	ProductUOMBus  *productuombus.Business
	AuthClient     *authclient.Client
	PermissionsBus *permissionsbus.Business
}

const RouteTable = "products.product_uoms"

// Routes adds specific routes for this group of handlers.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)

	api := newAPI(productuomapp.NewApp(cfg.ProductUOMBus))

	app.HandlerFunc(http.MethodGet, version, "/products/product-uoms", api.query, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))
	app.HandlerFunc(http.MethodGet, version, "/products/product-uoms/{product_uom_id}", api.queryByID, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))
	app.HandlerFunc(http.MethodPost, version, "/products/product-uoms", api.create, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Create, auth.RuleAny))
	app.HandlerFunc(http.MethodPut, version, "/products/product-uoms/{product_uom_id}", api.update, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Update, auth.RuleAny))
	app.HandlerFunc(http.MethodDelete, version, "/products/product-uoms/{product_uom_id}", api.delete, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Delete, auth.RuleAny))
}
```

### Task 16: Wire ProductUOMs into all.go

**Files:**
- Modify: `api/cmd/services/ichor/build/all/all.go`

- [ ] **Step 1: Add imports**

Add to the import block:
```go
"github.com/timmaaaz/ichor/api/domain/http/products/productuomapi"
"github.com/timmaaaz/ichor/business/domain/products/productuombus"
"github.com/timmaaaz/ichor/business/domain/products/productuombus/stores/productuomdb"
```

- [ ] **Step 2: Instantiate the business layer**

Find where other product buses are instantiated (search for `productbus.NewBusiness`). Add nearby:

```go
productUOMBus := productuombus.NewBusiness(cfg.Log, delegate, productuomdb.NewStore(cfg.Log, cfg.DB))
```

- [ ] **Step 3: Register routes**

Find where other product routes are registered (search for `productapi.Routes`). Add:

```go
productuomapi.Routes(app, productuomapi.Config{
    Log:            cfg.Log,
    ProductUOMBus:  productUOMBus,
    AuthClient:     cfg.AuthClient,
    PermissionsBus: permissionsBus,
})
```

- [ ] **Step 4: Build the entire service**

```bash
go build ./api/cmd/services/ichor/...
```

Expected: no errors.

- [ ] **Step 5: Commit**

```bash
git add app/domain/products/productuomapp/ \
        api/domain/http/products/productuomapi/ \
        api/cmd/services/ichor/build/all/all.go
git commit -m "feat(productuomapi): wire product UOM app, api, and routes"
```

---

## Chunk 6: Integration Tests

### Task 17: Write integration tests for product_uoms

**Files:**
- Create: `api/cmd/services/ichor/tests/products/productuomapi/seed_test.go`
- Create: `api/cmd/services/ichor/tests/products/productuomapi/productuomapi_test.go`

- [ ] **Step 1: Create seed_test.go**

Read `api/cmd/services/ichor/tests/products/productapi/seed_test.go` for the exact pattern. Your `insertSeedData` must:
- Accept `(db *dbtest.Database, ath *auth.Auth)` (same signature as all other seed functions)
- Return a local struct (not `apitest.SeedData`) containing the seeded UOMs so tests can reference IDs

```go
package productuomapi_test

import (
    "context"
    "testing"

    "github.com/timmaaaz/ichor/api/sdk/http/apitest"
    "github.com/timmaaaz/ichor/app/sdk/auth"
    "github.com/timmaaaz/ichor/business/domain/products/productuombus"
    "github.com/timmaaaz/ichor/business/sdk/dbtest"
)

type seedData struct {
    UOMs     []productuombus.ProductUOM
    Products []apitest.SeedProduct // use whatever SeedProduct type the existing tests use
}

func insertSeedData(db *dbtest.Database, ath *auth.Auth) (seedData, error) {
    ctx := context.Background()

    // Get a seeded product to attach UOMs to — follow existing pattern for
    // accessing seeded data (see productapi/seed_test.go for the exact call)
    sd, err := apitest.Seed(ctx, db, ath)
    if err != nil {
        return seedData{}, err
    }

    bus := productuombus.NewBusiness(db.Log, db.Delegate, /* store */)

    uom1, err := bus.Create(ctx, productuombus.NewProductUOM{
        ProductID:        sd.Products[0].ProductID,
        Name:             "each",
        ConversionFactor: 1,
        IsBase:           true,
        IsApproximate:    false,
    })
    if err != nil {
        return seedData{}, err
    }

    uom2, err := bus.Create(ctx, productuombus.NewProductUOM{
        ProductID:        sd.Products[0].ProductID,
        Name:             "box",
        Abbreviation:     "bx",
        ConversionFactor: 1000,
        IsBase:           false,
        IsApproximate:    false,
    })
    if err != nil {
        return seedData{}, err
    }

    return seedData{
        UOMs:     []productuombus.ProductUOM{uom1, uom2},
        Products: sd.Products,
    }, nil
}
```

> **Important:** The exact seed infrastructure (how to access `dbtest.Database`, `auth.Auth`, and seeded products) is defined in the existing test packages. Read `tests/products/productapi/seed_test.go` and `tests/products/productapi/productapi_test.go` before writing these files — copy the startup pattern exactly.

- [ ] **Step 2: Create productuomapi_test.go**

Follow the exact `apitest.StartTest` + `apitest.Table` + `test.Run` pattern from existing test files. The test entry point, authentication, and table runner all come from that infrastructure.

```go
package productuomapi_test

import (
    "net/http"
    "testing"

    "github.com/timmaaaz/ichor/api/sdk/http/apitest"
    "github.com/timmaaaz/ichor/app/domain/products/productuomapp"
)

func Test_ProductUOMAPI(t *testing.T) {
    t.Parallel()

    test := apitest.StartTest(t, "Test_ProductUOMAPI")
    sd, err := insertSeedData(test.DB, test.Auth)
    if err != nil {
        t.Fatalf("insertSeedData: %s", err)
    }

    // Follow the exact table structure from tests/products/productapi/*.
    // Each table entry maps to one HTTP call + expected status.

    test.Run(t, query200(sd), "query-200")
    test.Run(t, queryByID200(sd), "queryByID-200")
    test.Run(t, create201(sd), "create-201")
    test.Run(t, update200(sd), "update-200")
    test.Run(t, delete204(sd), "delete-204")
}

func query200(sd seedData) []apitest.Table {
    return []apitest.Table{
        {
            Name:       "basic-query",
            URL:        "/v1/products/product-uoms",
            Token:      "", // use test.Token() — follow existing pattern
            Method:     http.MethodGet,
            StatusCode: http.StatusOK,
        },
    }
}

// Implement queryByID200, create201, update200, delete204 following the same
// pattern — see tests/products/productapi/ for complete examples of each.
```

> The exact `apitest.Table` fields, `test.Token()` usage, and input encoding follow the pattern in the existing test packages. Read them before implementing. Do not invent new helpers.

- [ ] **Step 3: Run the integration tests**

```bash
go test ./api/cmd/services/ichor/tests/products/productuomapi/... -v -count=1
```

Expected: all 5 sub-tests pass.

- [ ] **Step 4: Also verify Change A and B didn't break existing tests**

```bash
go test ./api/cmd/services/ichor/tests/products/productapi/... -v -count=1
go test ./api/cmd/services/ichor/tests/inventory/zoneapi/... -v -count=1
```

Expected: all pass.

- [ ] **Step 5: Final build check**

```bash
go build ./...
```

- [ ] **Step 6: Final commit**

```bash
git add api/cmd/services/ichor/tests/products/productuomapi/
git commit -m "test(productuomapi): integration tests for product UOM CRUD endpoints"
```

---

## Completion Checklist

- [ ] Migration 2.11 applied (`inventory_type` column, `stage` column, `product_uoms` table)
- [ ] `InventoryType` enum in `productbus` with 5 values
- [ ] `Stage` enum in `zonebus` with 7 values
- [ ] Products domain: model, db, and app layers accept/return `inventory_type`
- [ ] Zones domain: model, db, and app layers accept/return `stage`
- [ ] `productuombus` package: Business, Storer, models, filter, order
- [ ] `productuomdb` package: Store with CRUD + filter SQL
- [ ] `productuomapp` package: App + models + conversion
- [ ] `productuomapi` package: 5 handlers + routes
- [ ] Wired in `all.go`
- [ ] Integration tests passing for all three changes
- [ ] `go build ./...` clean
