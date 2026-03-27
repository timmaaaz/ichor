# Pick Tasks Domain Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create a full 7-layer pick tasks domain for floor worker pick task management — the outbound counterpart to put-away tasks.

**Architecture:** Mirrors `putawaytaskbus` exactly. A pick task is a directed work instruction: pick `QuantityToPick` units of `ProductID` from `LocationID` to fulfill a sales order line item. The `complete` action performs an atomic 3-way write (update task + create PICK transaction + decrement inventory). Adds `short_picked` status for partial fulfillment.

**Tech Stack:** Go 1.23, PostgreSQL 16.4, Ardan Labs Service architecture (bus/db/app/api layers)

**Reference domain:** `putawaytaskbus` / `putawaytaskapp` / `putawaytaskapi` — follow this as the structural template for every file.

---

## Task 1: Database Migration

**Files:**
- Modify: `business/sdk/migrate/sql/migrate.sql`

**Note:** If Plan A (field-additions) has already been applied, the latest version will be 2.19. If not, it will be 2.17. Use the next available version number. This plan uses 2.20 assuming Plan A lands first. Adjust if needed.

- [ ] **Step 1: Add migration**

Append to `business/sdk/migrate/sql/migrate.sql`:

```sql
-- Version: 2.20
-- Description: Create inventory.pick_tasks table for floor worker pick task management.
CREATE TABLE inventory.pick_tasks (
    id                        UUID          NOT NULL,
    sales_order_id            UUID          NOT NULL REFERENCES sales.orders(id),
    sales_order_line_item_id  UUID          NOT NULL REFERENCES sales.order_line_items(id),
    product_id                UUID          NOT NULL REFERENCES products.products(id),
    lot_id                    UUID          NULL REFERENCES inventory.lot_trackings(id),
    serial_id                 UUID          NULL REFERENCES inventory.serial_numbers(id),
    location_id               UUID          NOT NULL REFERENCES inventory.inventory_locations(id),
    quantity_to_pick          INT           NOT NULL CHECK (quantity_to_pick > 0),
    quantity_picked           INT           NOT NULL DEFAULT 0 CHECK (quantity_picked >= 0),
    status                    VARCHAR(20)   NOT NULL DEFAULT 'pending'
                                  CHECK (status IN ('pending','in_progress','completed','short_picked','cancelled')),
    assigned_to               UUID          NULL REFERENCES core.users(id),
    assigned_at               TIMESTAMP     NULL,
    completed_by              UUID          NULL REFERENCES core.users(id),
    completed_at              TIMESTAMP     NULL,
    short_pick_reason         TEXT          NULL,
    created_by                UUID          NOT NULL REFERENCES core.users(id),
    created_date              TIMESTAMP     NOT NULL,
    updated_date              TIMESTAMP     NOT NULL,
    PRIMARY KEY (id)
);

CREATE INDEX idx_pick_tasks_status      ON inventory.pick_tasks(status);
CREATE INDEX idx_pick_tasks_order       ON inventory.pick_tasks(sales_order_id);
CREATE INDEX idx_pick_tasks_product     ON inventory.pick_tasks(product_id);
CREATE INDEX idx_pick_tasks_location    ON inventory.pick_tasks(location_id);
CREATE INDEX idx_pick_tasks_assigned    ON inventory.pick_tasks(assigned_to) WHERE assigned_to IS NOT NULL;

INSERT INTO core.table_access (id, role_id, table_name, can_create, can_read, can_update, can_delete)
SELECT gen_random_uuid(), id, 'inventory.pick_tasks', true, true, true, true FROM core.roles;
```

- [ ] **Step 2: Commit**

```bash
git add business/sdk/migrate/sql/migrate.sql
git commit -m "feat(migration): add inventory.pick_tasks table"
```

---

## Task 2: Business Layer — Model, Status, Filter, Order, Event

**Files:**
- Create: `business/domain/inventory/picktaskbus/model.go`
- Create: `business/domain/inventory/picktaskbus/status.go`
- Create: `business/domain/inventory/picktaskbus/filter.go`
- Create: `business/domain/inventory/picktaskbus/order.go`
- Create: `business/domain/inventory/picktaskbus/event.go`

- [ ] **Step 1: Create model.go**

Create `business/domain/inventory/picktaskbus/model.go`:

```go
package picktaskbus

import (
	"time"

	"github.com/google/uuid"
)

// JSON tags are required for workflow event serialization. The workflow system
// marshals business models to JSON for RawData in TriggerEvents.

// PickTask represents a single directed work instruction for a floor worker:
// pick QuantityToPick units of ProductID from LocationID to fulfill a sales order line item.
type PickTask struct {
	ID                   uuid.UUID  `json:"id"`
	SalesOrderID         uuid.UUID  `json:"sales_order_id"`
	SalesOrderLineItemID uuid.UUID  `json:"sales_order_line_item_id"`
	ProductID            uuid.UUID  `json:"product_id"`
	LotID                *uuid.UUID `json:"lot_id,omitempty"`
	SerialID             *uuid.UUID `json:"serial_id,omitempty"`
	LocationID           uuid.UUID  `json:"location_id"`
	QuantityToPick       int        `json:"quantity_to_pick"`
	QuantityPicked       int        `json:"quantity_picked"`
	Status               Status     `json:"status"`
	AssignedTo           uuid.UUID  `json:"assigned_to"`
	AssignedAt           time.Time  `json:"assigned_at"`
	CompletedBy          uuid.UUID  `json:"completed_by"`
	CompletedAt          time.Time  `json:"completed_at"`
	ShortPickReason      string     `json:"short_pick_reason,omitempty"`
	CreatedBy            uuid.UUID  `json:"created_by"`
	CreatedDate          time.Time  `json:"created_date"`
	UpdatedDate          time.Time  `json:"updated_date"`
}

// NewPickTask contains the information needed to create a new pick task.
// Status is always set to Statuses.Pending by the business layer.
type NewPickTask struct {
	SalesOrderID         uuid.UUID  `json:"sales_order_id"`
	SalesOrderLineItemID uuid.UUID  `json:"sales_order_line_item_id"`
	ProductID            uuid.UUID  `json:"product_id"`
	LotID                *uuid.UUID `json:"lot_id,omitempty"`
	SerialID             *uuid.UUID `json:"serial_id,omitempty"`
	LocationID           uuid.UUID  `json:"location_id"`
	QuantityToPick       int        `json:"quantity_to_pick"`
	CreatedBy            uuid.UUID  `json:"created_by"`
}

// UpdatePickTask contains the information that can be changed on a pick task.
// All fields are optional pointers; nil means "do not update this field."
type UpdatePickTask struct {
	LotID           *uuid.UUID `json:"lot_id,omitempty"`
	SerialID        *uuid.UUID `json:"serial_id,omitempty"`
	LocationID      *uuid.UUID `json:"location_id,omitempty"`
	QuantityToPick  *int       `json:"quantity_to_pick,omitempty"`
	QuantityPicked  *int       `json:"quantity_picked,omitempty"`
	Status          *Status    `json:"status,omitempty"`
	AssignedTo      *uuid.UUID `json:"assigned_to,omitempty"`
	AssignedAt      *time.Time `json:"assigned_at,omitempty"`
	CompletedBy     *uuid.UUID `json:"completed_by,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	ShortPickReason *string    `json:"short_pick_reason,omitempty"`
}
```

- [ ] **Step 2: Create status.go**

Create `business/domain/inventory/picktaskbus/status.go`:

```go
package picktaskbus

import "fmt"

type statusSet struct {
	Pending     Status
	InProgress  Status
	Completed   Status
	ShortPicked Status
	Cancelled   Status
}

// Statuses represents the set of valid pick task statuses.
var Statuses = statusSet{
	Pending:     newStatus("pending"),
	InProgress:  newStatus("in_progress"),
	Completed:   newStatus("completed"),
	ShortPicked: newStatus("short_picked"),
	Cancelled:   newStatus("cancelled"),
}

// =============================================================================

// Set of known statuses.
var statuses = make(map[string]Status)

// Status represents a pick task status in the system.
type Status struct {
	name string
}

func newStatus(s string) Status {
	st := Status{s}
	statuses[s] = st
	return st
}

// String returns the name of the status.
func (s Status) String() string {
	return s.name
}

// Equal provides support for the go-cmp package and testing.
func (s Status) Equal(s2 Status) bool {
	return s.name == s2.name
}

// MarshalText implements encoding.TextMarshaler.
func (s Status) MarshalText() ([]byte, error) {
	return []byte(s.name), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (s *Status) UnmarshalText(data []byte) error {
	st, err := ParseStatus(string(data))
	if err != nil {
		return err
	}
	*s = st
	return nil
}

// =============================================================================

// ParseStatus parses the string value and returns a status if one exists.
func ParseStatus(value string) (Status, error) {
	st, exists := statuses[value]
	if !exists {
		return Status{}, fmt.Errorf("invalid status %q", value)
	}
	return st, nil
}

// MustParseStatus parses the string value and returns a status if one exists.
// Panics if the status is invalid.
func MustParseStatus(value string) Status {
	st, err := ParseStatus(value)
	if err != nil {
		panic(err)
	}
	return st
}
```

- [ ] **Step 3: Create filter.go**

Create `business/domain/inventory/picktaskbus/filter.go`:

```go
package picktaskbus

import (
	"time"

	"github.com/google/uuid"
)

// QueryFilter holds optional filters for querying pick tasks.
type QueryFilter struct {
	ID                   *uuid.UUID
	SalesOrderID         *uuid.UUID
	SalesOrderLineItemID *uuid.UUID
	ProductID            *uuid.UUID
	LocationID           *uuid.UUID
	Status               *Status
	AssignedTo           *uuid.UUID
	CreatedBy            *uuid.UUID
	CreatedDate          *time.Time
	UpdatedDate          *time.Time
}
```

- [ ] **Step 4: Create order.go**

Create `business/domain/inventory/picktaskbus/order.go`:

```go
package picktaskbus

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy represents the default ordering for pick task queries.
var DefaultOrderBy = order.NewBy(OrderByID, order.ASC)

const (
	OrderByID             = "id"
	OrderBySalesOrderID   = "sales_order_id"
	OrderByProductID      = "product_id"
	OrderByLocationID     = "location_id"
	OrderByQuantityToPick = "quantity_to_pick"
	OrderByStatus         = "status"
	OrderByAssignedTo     = "assigned_to"
	OrderByCreatedBy      = "created_by"
	OrderByCreatedDate    = "created_date"
	OrderByUpdatedDate    = "updated_date"
)
```

- [ ] **Step 5: Create event.go**

Create `business/domain/inventory/picktaskbus/event.go`:

```go
package picktaskbus

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
)

// DomainName represents the name of this domain for delegate events.
const DomainName = "picktask"

// EntityName is the workflow entity name used for event matching.
const EntityName = "pick_tasks"

// Delegate action constants.
const (
	ActionCreated = "created"
	ActionUpdated = "updated"
	ActionDeleted = "deleted"
)

// =============================================================================
// Created Event
// =============================================================================

type ActionCreatedParms struct {
	EntityID uuid.UUID `json:"entityID"`
	UserID   uuid.UUID `json:"userID"`
	Entity   PickTask  `json:"entity"`
}

func (p *ActionCreatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

func ActionCreatedData(pt PickTask) delegate.Data {
	params := ActionCreatedParms{
		EntityID: pt.ID,
		UserID:   pt.CreatedBy,
		Entity:   pt,
	}

	rawParams, err := params.Marshal()
	if err != nil {
		panic(err)
	}

	return delegate.Data{
		Domain:    DomainName,
		Action:    ActionCreated,
		RawParams: rawParams,
	}
}

// =============================================================================
// Updated Event
// =============================================================================

type ActionUpdatedParms struct {
	EntityID     uuid.UUID `json:"entityID"`
	UserID       uuid.UUID `json:"userID"`
	Entity       PickTask  `json:"entity"`
	BeforeEntity PickTask  `json:"beforeEntity,omitempty"`
}

func (p *ActionUpdatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

func ActionUpdatedData(before, after PickTask) delegate.Data {
	params := ActionUpdatedParms{
		EntityID:     after.ID,
		UserID:       after.CreatedBy,
		Entity:       after,
		BeforeEntity: before,
	}

	rawParams, err := params.Marshal()
	if err != nil {
		panic(err)
	}

	return delegate.Data{
		Domain:    DomainName,
		Action:    ActionUpdated,
		RawParams: rawParams,
	}
}

// =============================================================================
// Deleted Event
// =============================================================================

type ActionDeletedParms struct {
	EntityID uuid.UUID `json:"entityID"`
	UserID   uuid.UUID `json:"userID"`
	Entity   PickTask  `json:"entity"`
}

func (p *ActionDeletedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

func ActionDeletedData(pt PickTask) delegate.Data {
	params := ActionDeletedParms{
		EntityID: pt.ID,
		UserID:   pt.CreatedBy,
		Entity:   pt,
	}

	rawParams, err := params.Marshal()
	if err != nil {
		panic(err)
	}

	return delegate.Data{
		Domain:    DomainName,
		Action:    ActionDeleted,
		RawParams: rawParams,
	}
}
```

- [ ] **Step 6: Build**

Run: `go build ./business/domain/inventory/picktaskbus/...`
Expected: Clean build.

- [ ] **Step 7: Commit**

```bash
git add business/domain/inventory/picktaskbus/
git commit -m "feat(picktask): add business layer models, status, filter, order, events"
```

---

## Task 3: Business Layer — Storer Interface and Business Logic

**Files:**
- Create: `business/domain/inventory/picktaskbus/picktaskbus.go`

- [ ] **Step 1: Create picktaskbus.go**

Create `business/domain/inventory/picktaskbus/picktaskbus.go`. Mirror `putawaytaskbus/putawaytaskbus.go` exactly, with these differences:
- Package name: `picktaskbus`
- Type names: `PickTask`, `NewPickTask`, `UpdatePickTask`
- Error messages: "pick task not found", etc.
- New field in Create: `SalesOrderID`, `SalesOrderLineItemID`, `LotID`, `SerialID`, `QuantityToPick` (instead of `Quantity`), `QuantityPicked: 0`
- Update method applies: `LotID`, `SerialID`, `LocationID`, `QuantityToPick`, `QuantityPicked`, `Status`, `AssignedTo`, `AssignedAt`, `CompletedBy`, `CompletedAt`, `ShortPickReason`

```go
package picktaskbus

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
	"github.com/timmaaaz/ichor/foundation/otel"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound            = errors.New("pick task not found")
	ErrUniqueEntry         = errors.New("pick task entry is not unique")
	ErrForeignKeyViolation = errors.New("foreign key violation")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, task PickTask) error
	Update(ctx context.Context, task PickTask) error
	Delete(ctx context.Context, task PickTask) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]PickTask, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, taskID uuid.UUID) (PickTask, error)
}

// Business manages the set of APIs for pick task access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a pick task business API for use.
func NewBusiness(log *logger.Logger, delegate *delegate.Delegate, storer Storer) *Business {
	return &Business{
		log:      log,
		delegate: delegate,
		storer:   storer,
	}
}

// NewWithTx constructs a new Business value replacing the Storer
// value with a Storer value that is currently inside a transaction.
func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (*Business, error) {
	storer, err := b.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	return &Business{
		log:      b.log,
		storer:   storer,
		delegate: b.delegate,
	}, nil
}

// Create adds a new pick task to the system.
func (b *Business) Create(ctx context.Context, npt NewPickTask) (PickTask, error) {
	ctx, span := otel.AddSpan(ctx, "business.picktaskbus.create")
	defer span.End()

	now := time.Now()

	task := PickTask{
		ID:                   uuid.New(),
		SalesOrderID:         npt.SalesOrderID,
		SalesOrderLineItemID: npt.SalesOrderLineItemID,
		ProductID:            npt.ProductID,
		LotID:                npt.LotID,
		SerialID:             npt.SerialID,
		LocationID:           npt.LocationID,
		QuantityToPick:       npt.QuantityToPick,
		QuantityPicked:       0,
		Status:               Statuses.Pending,
		CreatedBy:            npt.CreatedBy,
		CreatedDate:          now,
		UpdatedDate:          now,
	}

	if err := b.storer.Create(ctx, task); err != nil {
		return PickTask{}, fmt.Errorf("create: %w", err)
	}

	if b.delegate != nil {
		b.delegate.Call(ctx, ActionCreatedData(task))
	}

	return task, nil
}

// Update modifies an existing pick task in the system.
func (b *Business) Update(ctx context.Context, task PickTask, upt UpdatePickTask) (PickTask, error) {
	ctx, span := otel.AddSpan(ctx, "business.picktaskbus.update")
	defer span.End()

	before := task

	if upt.LotID != nil {
		task.LotID = upt.LotID
	}
	if upt.SerialID != nil {
		task.SerialID = upt.SerialID
	}
	if upt.LocationID != nil {
		task.LocationID = *upt.LocationID
	}
	if upt.QuantityToPick != nil {
		task.QuantityToPick = *upt.QuantityToPick
	}
	if upt.QuantityPicked != nil {
		task.QuantityPicked = *upt.QuantityPicked
	}
	if upt.Status != nil {
		task.Status = *upt.Status
	}
	if upt.AssignedTo != nil {
		task.AssignedTo = *upt.AssignedTo
	}
	if upt.AssignedAt != nil {
		task.AssignedAt = *upt.AssignedAt
	}
	if upt.CompletedBy != nil {
		task.CompletedBy = *upt.CompletedBy
	}
	if upt.CompletedAt != nil {
		task.CompletedAt = *upt.CompletedAt
	}
	if upt.ShortPickReason != nil {
		task.ShortPickReason = *upt.ShortPickReason
	}

	task.UpdatedDate = time.Now()

	if err := b.storer.Update(ctx, task); err != nil {
		return PickTask{}, fmt.Errorf("update: %w", err)
	}

	if b.delegate != nil {
		b.delegate.Call(ctx, ActionUpdatedData(before, task))
	}

	return task, nil
}

// Delete removes a pick task from the system.
func (b *Business) Delete(ctx context.Context, task PickTask) error {
	ctx, span := otel.AddSpan(ctx, "business.picktaskbus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, task); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	if b.delegate != nil {
		b.delegate.Call(ctx, ActionDeletedData(task))
	}

	return nil
}

// Query retrieves a list of pick tasks from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]PickTask, error) {
	ctx, span := otel.AddSpan(ctx, "business.picktaskbus.query")
	defer span.End()

	tasks, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return tasks, nil
}

// Count returns the total number of pick tasks matching the filter.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.picktaskbus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID retrieves a single pick task by its ID.
func (b *Business) QueryByID(ctx context.Context, taskID uuid.UUID) (PickTask, error) {
	ctx, span := otel.AddSpan(ctx, "business.picktaskbus.querybyid")
	defer span.End()

	task, err := b.storer.QueryByID(ctx, taskID)
	if err != nil {
		return PickTask{}, fmt.Errorf("query: %w", err)
	}

	return task, nil
}
```

- [ ] **Step 2: Build**

Run: `go build ./business/domain/inventory/picktaskbus/...`
Expected: Clean build.

- [ ] **Step 3: Commit**

```bash
git add business/domain/inventory/picktaskbus/picktaskbus.go
git commit -m "feat(picktask): add business layer with Storer interface and CRUD"
```

---

## Task 4: Database Store Layer

**Files:**
- Create: `business/domain/inventory/picktaskbus/stores/picktaskdb/model.go`
- Create: `business/domain/inventory/picktaskbus/stores/picktaskdb/filter.go`
- Create: `business/domain/inventory/picktaskbus/stores/picktaskdb/order.go`
- Create: `business/domain/inventory/picktaskbus/stores/picktaskdb/picktaskdb.go`

**Pattern reference:** Mirror `putawaytaskdb` exactly. Key differences:
- Extra columns: `sales_order_id`, `sales_order_line_item_id`, `lot_id` (NullUUID), `serial_id` (NullUUID), `quantity_to_pick`, `quantity_picked`, `short_pick_reason` (NullString)
- Nullable UUIDs (`lot_id`, `serial_id`, `assigned_to`, `completed_by`) use `uuid.NullUUID` in DB struct, `*uuid.UUID` in bus struct
- Nullable timestamps (`assigned_at`, `completed_at`) use `sql.NullTime`
- Nullable string (`short_pick_reason`) uses `sql.NullString`

- [ ] **Step 1: Create model.go**

Create `business/domain/inventory/picktaskbus/stores/picktaskdb/model.go`. The DB struct must have `db:` tags matching the SQL column names exactly. The `toBusPickTask` and `toDBPickTask` functions handle nullable field conversions.

For nullable UUID fields, use this pattern (from `putawaytaskdb`):
```go
// DB → Bus (in toBusPickTask):
var lotID *uuid.UUID
if db.LotID.Valid {
    id := db.LotID.UUID
    lotID = &id
}

// Bus → DB (in toDBPickTask):
var lotID uuid.NullUUID
if bus.LotID != nil {
    lotID = uuid.NullUUID{UUID: *bus.LotID, Valid: true}
}
```

For nullable `uuid.UUID` fields that use zero-value (like `AssignedTo`, `CompletedBy`), use the same `uuid.NullUUID` pattern but check `!= uuid.Nil` in `toDBPickTask`:
```go
var assignedTo uuid.NullUUID
if bus.AssignedTo != (uuid.UUID{}) {
    assignedTo = uuid.NullUUID{UUID: bus.AssignedTo, Valid: true}
}
```

- [ ] **Step 2: Create filter.go**

Create `business/domain/inventory/picktaskbus/stores/picktaskdb/filter.go`. Apply filters for: `ID`, `SalesOrderID`, `SalesOrderLineItemID`, `ProductID`, `LocationID`, `Status`, `AssignedTo`, `CreatedBy`, `CreatedDate`, `UpdatedDate`.

- [ ] **Step 3: Create order.go**

Create `business/domain/inventory/picktaskbus/stores/picktaskdb/order.go`. Map the bus-layer `OrderBy*` constants to SQL column names.

- [ ] **Step 4: Create picktaskdb.go**

Create `business/domain/inventory/picktaskbus/stores/picktaskdb/picktaskdb.go` with `Store`, `NewStore`, `NewWithTx`, `Create`, `Update`, `Delete`, `Query`, `Count`, `QueryByID`.

The SQL column list for INSERT/SELECT:
```sql
id, sales_order_id, sales_order_line_item_id, product_id, lot_id, serial_id,
location_id, quantity_to_pick, quantity_picked, status, assigned_to, assigned_at,
completed_by, completed_at, short_pick_reason, created_by, created_date, updated_date
```

- [ ] **Step 5: Build**

Run: `go build ./business/domain/inventory/picktaskbus/...`
Expected: Clean build.

- [ ] **Step 6: Commit**

```bash
git add business/domain/inventory/picktaskbus/stores/
git commit -m "feat(picktask): add database store layer"
```

---

## Task 5: App Layer — Models, Filter, Order

**Files:**
- Create: `app/domain/inventory/picktaskapp/model.go`
- Create: `app/domain/inventory/picktaskapp/filter.go`
- Create: `app/domain/inventory/picktaskapp/order.go`

**Pattern reference:** Mirror `putawaytaskapp` exactly. App models use `string` for all fields (UUIDs are stringified, ints are formatted, timestamps use `timeutil.FORMAT`). Nullable bus fields (`*uuid.UUID`) map to empty string `""` in app layer.

- [ ] **Step 1: Create model.go**

Create `app/domain/inventory/picktaskapp/model.go` with:
- `PickTask` struct (all string fields, JSON tags, `Encode()` method)
- `ToAppPickTask(bus) PickTask` conversion
- `ToAppPickTasks([]bus) []PickTask` conversion
- `NewPickTask` struct with validation tags
- `toBusNewPickTask` conversion
- `UpdatePickTask` struct with optional pointer fields
- `toBusUpdatePickTask` conversion

- [ ] **Step 2: Create filter.go**

Create `app/domain/inventory/picktaskapp/filter.go` with `QueryParams` and `parseFilter` function.

- [ ] **Step 3: Create order.go**

Create `app/domain/inventory/picktaskapp/order.go` with `defaultOrderBy`, `orderByFields` map, and `parseOrder` function.

- [ ] **Step 4: Build**

Run: `go build ./app/domain/inventory/picktaskapp/...`
Expected: Clean build.

- [ ] **Step 5: Commit**

```bash
git add app/domain/inventory/picktaskapp/
git commit -m "feat(picktask): add app layer models, filter, order"
```

---

## Task 6: App Layer — Business Logic with Complete Flow

**Files:**
- Create: `app/domain/inventory/picktaskapp/picktaskapp.go`

- [ ] **Step 1: Create picktaskapp.go**

Create `app/domain/inventory/picktaskapp/picktaskapp.go` with `App` struct holding 4 dependencies:

```go
type App struct {
	pickTaskBus       *picktaskbus.Business
	invTransactionBus *inventorytransactionbus.Business
	invItemBus        *inventoryitembus.Business
	db                *sqlx.DB
}
```

Methods: `Create`, `Update`, `Query`, `QueryByID`, `Delete`, and `complete`.

The `Update` method must detect when status is being set to `completed` or `short_picked` and route to the `complete` method (same pattern as `putawaytaskapp`).

The `complete` method performs the atomic 3-way write:

```go
func (a *App) complete(ctx context.Context, task picktaskbus.PickTask, upt picktaskbus.UpdatePickTask) (PickTask, error) {
    userID, err := mid.GetUserID(ctx)
    if err != nil {
        return PickTask{}, errs.Newf(errs.Unauthenticated, "get user id: %s", err)
    }

    now := time.Now()
    upt.CompletedBy = &userID
    upt.CompletedAt = &now

    // Determine actual quantity picked
    quantityPicked := task.QuantityToPick
    if upt.QuantityPicked != nil {
        quantityPicked = *upt.QuantityPicked
    }
    upt.QuantityPicked = &quantityPicked

    tx, err := a.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
    if err != nil {
        return PickTask{}, fmt.Errorf("begin transaction: %w", err)
    }
    defer tx.Rollback()

    // 1. Update pick task status inside the transaction.
    ptBusTx, err := a.pickTaskBus.NewWithTx(tx)
    if err != nil {
        return PickTask{}, fmt.Errorf("new picktask tx: %w", err)
    }

    updated, err := ptBusTx.Update(ctx, task, upt)
    if err != nil {
        return PickTask{}, fmt.Errorf("update task: %w", err)
    }

    // 2. Create PICK inventory transaction (ledger record).
    txBusTx, err := a.invTransactionBus.NewWithTx(tx)
    if err != nil {
        return PickTask{}, fmt.Errorf("new invtransaction tx: %w", err)
    }

    _, err = txBusTx.Create(ctx, inventorytransactionbus.NewInventoryTransaction{
        ProductID:       task.ProductID,
        LocationID:      task.LocationID,
        UserID:          userID,
        LotID:           task.LotID,
        Quantity:        quantityPicked,
        TransactionType: "PICK",
        ReferenceNumber: task.SalesOrderID.String(),
        TransactionDate: now,
    })
    if err != nil {
        return PickTask{}, fmt.Errorf("create inventory transaction: %w", err)
    }

    // 3. Decrement inventory_item quantity at the source location.
    itemBusTx, err := a.invItemBus.NewWithTx(tx)
    if err != nil {
        return PickTask{}, fmt.Errorf("new invitem tx: %w", err)
    }

    if err := itemBusTx.DecrementQuantity(ctx, task.ProductID, task.LocationID, quantityPicked); err != nil {
        return PickTask{}, fmt.Errorf("decrement inventory quantity: %w", err)
    }

    if err := tx.Commit(); err != nil {
        return PickTask{}, fmt.Errorf("commit transaction: %w", err)
    }

    return ToAppPickTask(updated), nil
}
```

Key differences from put-away complete:
- Uses `DecrementQuantity` instead of `UpsertQuantity` (outbound, not inbound)
- Transaction type is `"PICK"` instead of `"PUT_AWAY"`
- Reference number is the sales order ID
- Passes `LotID` to the transaction (put-away doesn't)
- Handles `QuantityPicked` which may differ from `QuantityToPick` (short pick)

- [ ] **Step 2: Build**

Run: `go build ./app/domain/inventory/picktaskapp/...`
Expected: Clean build.

- [ ] **Step 3: Commit**

```bash
git add app/domain/inventory/picktaskapp/picktaskapp.go
git commit -m "feat(picktask): add app layer with atomic complete flow"
```

---

## Task 7: API Layer — Routes and Handlers

**Files:**
- Create: `api/domain/http/inventory/picktaskapi/routes.go`
- Create: `api/domain/http/inventory/picktaskapi/picktaskapi.go`
- Create: `api/domain/http/inventory/picktaskapi/filter.go`

- [ ] **Step 1: Create routes.go**

Create `api/domain/http/inventory/picktaskapi/routes.go`:

```go
package picktaskapi

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/inventory/picktaskapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/core/permissionsbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorytransactionbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/picktaskbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

type Config struct {
	Log               *logger.Logger
	PickTaskBus       *picktaskbus.Business
	InvTransactionBus *inventorytransactionbus.Business
	InvItemBus        *inventoryitembus.Business
	DB                *sqlx.DB
	AuthClient        *authclient.Client
	PermissionsBus    *permissionsbus.Business
}

const (
	RouteTable = "inventory.pick_tasks"
)

func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	api := newAPI(picktaskapp.NewApp(cfg.PickTaskBus, cfg.InvTransactionBus, cfg.InvItemBus, cfg.DB))

	app.HandlerFunc(http.MethodGet, version, "/inventory/pick-tasks", api.query, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))

	app.HandlerFunc(http.MethodGet, version, "/inventory/pick-tasks/{task_id}", api.queryByID, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))

	app.HandlerFunc(http.MethodPost, version, "/inventory/pick-tasks", api.create, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Create, auth.RuleAny))

	app.HandlerFunc(http.MethodPut, version, "/inventory/pick-tasks/{task_id}", api.update, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Update, auth.RuleAny))

	app.HandlerFunc(http.MethodDelete, version, "/inventory/pick-tasks/{task_id}", api.delete, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Delete, auth.RuleAny))
}
```

- [ ] **Step 2: Create picktaskapi.go**

Create `api/domain/http/inventory/picktaskapi/picktaskapi.go` with handlers: `create`, `update`, `delete`, `query`, `queryByID`. Mirror `putawaytaskapi/putawaytaskapi.go` exactly.

- [ ] **Step 3: Create filter.go**

Create `api/domain/http/inventory/picktaskapi/filter.go` with `parseQueryParams` that extracts query string parameters and builds `picktaskapp.QueryParams`.

- [ ] **Step 4: Build**

Run: `go build ./api/domain/http/inventory/picktaskapi/...`
Expected: Clean build.

- [ ] **Step 5: Commit**

```bash
git add api/domain/http/inventory/picktaskapi/
git commit -m "feat(picktask): add API layer with routes and handlers"
```

---

## Task 8: Wiring — all.go, dbtest.go, apitest model

**Files:**
- Modify: `api/cmd/services/ichor/build/all/all.go`
- Modify: `business/sdk/dbtest/dbtest.go`
- Modify: `api/sdk/http/apitest/model.go`

**Important:** Read each file before editing. These are large files with many insertion points.

- [ ] **Step 1: Wire into all.go**

In `api/cmd/services/ichor/build/all/all.go`:

1. **Imports** — add:
   ```go
   "github.com/timmaaaz/ichor/api/domain/http/inventory/picktaskapi"
   "github.com/timmaaaz/ichor/business/domain/inventory/picktaskbus"
   "github.com/timmaaaz/ichor/business/domain/inventory/picktaskbus/stores/picktaskdb"
   ```

2. **Business instantiation** — add after `putAwayTaskBus`:
   ```go
   pickTaskBus := picktaskbus.NewBusiness(cfg.Log, delegate, picktaskdb.NewStore(cfg.Log, cfg.DB))
   ```

3. **BusDomain struct** — add `PickTask: pickTaskBus,` after the `PutAwayTask` field.

4. **Delegate registration** — add after putawaytaskbus registration:
   ```go
   delegateHandler.RegisterDomain(delegate, picktaskbus.DomainName, picktaskbus.EntityName)
   ```

5. **Route registration** — add after the putawaytaskapi.Routes block:
   ```go
   picktaskapi.Routes(app, picktaskapi.Config{
       Log:               cfg.Log,
       PickTaskBus:       pickTaskBus,
       InvTransactionBus: inventoryTransactionBus,
       InvItemBus:        inventoryItemBus,
       DB:                cfg.DB,
       AuthClient:        cfg.AuthClient,
       PermissionsBus:    permissionsBus,
   })
   ```

- [ ] **Step 2: Wire into dbtest.go**

In `business/sdk/dbtest/dbtest.go`:

1. Add import for `picktaskbus` and `picktaskdb`.
2. Add `PickTask *picktaskbus.Business` field to `BusDomain` struct (after `PutAwayTask`).
3. In `newBusDomains`, add instantiation and assignment:
   ```go
   pickTaskBus := picktaskbus.NewBusiness(log, delegate, picktaskdb.NewStore(log, db))
   ```
   And `PickTask: pickTaskBus,` in the return struct.

- [ ] **Step 3: Wire into apitest model**

In `api/sdk/http/apitest/model.go`:

Add `PickTasks []picktaskapp.PickTask` to the `SeedData` struct (after `PutAwayTasks`). Add the import for `picktaskapp`.

- [ ] **Step 4: Build full service**

Run: `go build ./api/cmd/...`
Expected: Clean build.

- [ ] **Step 5: Commit**

```bash
git add api/cmd/services/ichor/build/all/all.go business/sdk/dbtest/dbtest.go api/sdk/http/apitest/model.go
git commit -m "feat(picktask): wire pick tasks into service, dbtest, and apitest"
```

---

## Task 9: Test Utilities

**Files:**
- Create: `business/domain/inventory/picktaskbus/testutil.go`

- [ ] **Step 1: Create testutil.go**

Create `business/domain/inventory/picktaskbus/testutil.go` following the `putawaytaskbus/testutil.go` pattern. Include:
- `TestNewPickTasks(n int, ...)` — generates `n` `NewPickTask` values with the provided FK IDs
- `TestSeedPickTasks(ctx, n, api)` — seeds `n` pick tasks using the business layer

The seed function requires `SalesOrderID`, `SalesOrderLineItemID`, `ProductID`, `LocationID`, and `CreatedBy` UUIDs as parameters. These must come from already-seeded data.

- [ ] **Step 2: Build**

Run: `go build ./business/domain/inventory/picktaskbus/...`
Expected: Clean build.

- [ ] **Step 3: Commit**

```bash
git add business/domain/inventory/picktaskbus/testutil.go
git commit -m "feat(picktask): add test utilities for seeding pick tasks"
```

---

## Task 10: Final Verification

- [ ] **Step 1: Full service build**

Run: `go build ./api/cmd/...`
Expected: Clean build, no errors.

- [ ] **Step 2: Verify migration**

Run: `grep "pick_tasks" business/sdk/migrate/sql/migrate.sql`
Expected: Shows CREATE TABLE and indexes.

- [ ] **Step 3: Verify wiring**

Run: `grep -r "picktask" api/cmd/services/ichor/build/all/all.go`
Expected: Shows import, bus instantiation, delegate registration, and route registration.

---

## Future Work (Not In This Plan)

These items are tracked but intentionally excluded from this plan:

1. **Integration tests** — Create `api/cmd/services/ichor/tests/inventory/picktaskapi/` with full test suite. This requires a running database and extensive seed data (users → geo → warehouses → zones → locations → products → orders → order_line_items → pick tasks). Do this as a follow-up.

2. **Order line item fulfillment update** — When a pick task completes, the corresponding `order_line_items.picked_quantity` should be updated. This creates a dependency on `orderlineitembus` which should be wired as a follow-up.

3. **Statuses endpoint** — Add `GET /v1/inventory/pick-tasks/statuses` following the pattern from Group B (transfer order statuses). Trivial follow-up.
