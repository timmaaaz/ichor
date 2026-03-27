# Cycle Count Sessions Domain — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the `cycle_count_sessions` and `cycle_count_items` domains (full 7-layer) so floor workers can perform inventory cycle counts. The `complete` endpoint locks a session and atomically generates `inventory_adjustment` records for every approved-variance item.

**Architecture:** Two independent bus domains (`cyclecountsessionbus`, `cyclecountitembus`) with standard CRUD. The session app layer (`cyclecountsessionapp`) owns the multi-bus transaction for the `complete` flow — identical to how `picktaskapp` atomically writes pick task + inventory transaction + inventory item. A new `cycle_count` reason code is added to `inventoryadjustmentbus` so adjustments are properly categorized.

**Tech Stack:** Go 1.23, PostgreSQL 16.4, sqlx, Ardan Labs Service architecture (7-layer DDD)

**Reference patterns:**
- `putawaytaskbus` — single-table domain with status machine, `NewWithTx`, delegate events
- `picktaskbus` — atomic complete flow: app-layer `BeginTxx` + 3 `NewWithTx` calls in one transaction
- `inventoryadjustmentbus` — target for generated adjustment records

---

## File Structure

### New files to create

**Business layer — Session:**
- `business/domain/inventory/cyclecountsessionbus/model.go` — CycleCountSession, NewCycleCountSession, UpdateCycleCountSession structs
- `business/domain/inventory/cyclecountsessionbus/status.go` — Status value-object (draft, in_progress, completed, cancelled)
- `business/domain/inventory/cyclecountsessionbus/event.go` — delegate event types + factory functions
- `business/domain/inventory/cyclecountsessionbus/filter.go` — QueryFilter struct
- `business/domain/inventory/cyclecountsessionbus/order.go` — OrderBy constants
- `business/domain/inventory/cyclecountsessionbus/cyclecountsessionbus.go` — Business struct, CRUD, NewWithTx
- `business/domain/inventory/cyclecountsessionbus/testutil.go` — TestNewCycleCountSessions, TestSeedCycleCountSessions
- `business/domain/inventory/cyclecountsessionbus/stores/cyclecountsessiondb/model.go` — DB model, conversions
- `business/domain/inventory/cyclecountsessionbus/stores/cyclecountsessiondb/cyclecountsessiondb.go` — Store CRUD
- `business/domain/inventory/cyclecountsessionbus/stores/cyclecountsessiondb/filter.go` — applyFilter
- `business/domain/inventory/cyclecountsessionbus/stores/cyclecountsessiondb/order.go` — orderByFields map

**Business layer — Item:**
- `business/domain/inventory/cyclecountitembus/model.go` — CycleCountItem, NewCycleCountItem, UpdateCycleCountItem structs
- `business/domain/inventory/cyclecountitembus/status.go` — Status value-object (pending, counted, variance_approved, variance_rejected)
- `business/domain/inventory/cyclecountitembus/event.go` — delegate event types
- `business/domain/inventory/cyclecountitembus/filter.go` — QueryFilter struct
- `business/domain/inventory/cyclecountitembus/order.go` — OrderBy constants
- `business/domain/inventory/cyclecountitembus/cyclecountitembus.go` — Business struct, CRUD, NewWithTx
- `business/domain/inventory/cyclecountitembus/testutil.go` — TestNewCycleCountItems, TestSeedCycleCountItems
- `business/domain/inventory/cyclecountitembus/stores/cyclecountitemdb/model.go` — DB model, conversions
- `business/domain/inventory/cyclecountitembus/stores/cyclecountitemdb/cyclecountitemdb.go` — Store CRUD
- `business/domain/inventory/cyclecountitembus/stores/cyclecountitemdb/filter.go` — applyFilter
- `business/domain/inventory/cyclecountitembus/stores/cyclecountitemdb/order.go` — orderByFields map

**App layer:**
- `app/domain/inventory/cyclecountsessionapp/cyclecountsessionapp.go` — App struct with complete flow
- `app/domain/inventory/cyclecountsessionapp/model.go` — request/response models, conversions
- `app/domain/inventory/cyclecountitemapp/cyclecountitemapp.go` — App struct, standard CRUD
- `app/domain/inventory/cyclecountitemapp/model.go` — request/response models, conversions

**API layer:**
- `api/domain/http/inventory/cyclecountsessionapi/cyclecountsessionapi.go` — HTTP handlers
- `api/domain/http/inventory/cyclecountsessionapi/routes.go` — route registration
- `api/domain/http/inventory/cyclecountsessionapi/filter.go` — parseQueryParams
- `api/domain/http/inventory/cyclecountitemapi/cyclecountitemapi.go` — HTTP handlers
- `api/domain/http/inventory/cyclecountitemapi/routes.go` — route registration
- `api/domain/http/inventory/cyclecountitemapi/filter.go` — parseQueryParams

**Integration tests:**
- `api/cmd/services/ichor/tests/inventory/cyclecountsessionapi/cyclecountsession_test.go` — top-level test runner
- `api/cmd/services/ichor/tests/inventory/cyclecountsessionapi/seed_test.go` — seed data
- `api/cmd/services/ichor/tests/inventory/cyclecountsessionapi/create_test.go` — create tests
- `api/cmd/services/ichor/tests/inventory/cyclecountsessionapi/update_test.go` — update + complete tests
- `api/cmd/services/ichor/tests/inventory/cyclecountsessionapi/query_test.go` — query tests
- `api/cmd/services/ichor/tests/inventory/cyclecountsessionapi/delete_test.go` — delete tests
- `api/cmd/services/ichor/tests/inventory/cyclecountitemapi/cyclecountitem_test.go` — top-level test runner
- `api/cmd/services/ichor/tests/inventory/cyclecountitemapi/seed_test.go` — seed data
- `api/cmd/services/ichor/tests/inventory/cyclecountitemapi/create_test.go` — create tests
- `api/cmd/services/ichor/tests/inventory/cyclecountitemapi/update_test.go` — update tests
- `api/cmd/services/ichor/tests/inventory/cyclecountitemapi/query_test.go` — query tests
- `api/cmd/services/ichor/tests/inventory/cyclecountitemapi/delete_test.go` — delete tests

### Existing files to modify

- `business/sdk/migrate/sql/migrate.sql` — add version 2.21 (two tables + reason code constraint)
- `business/domain/inventory/inventoryadjustmentbus/inventoryadjustmentbus.go` — add `ReasonCodeCycleCount` constant + map entry
- `business/sdk/dbtest/dbtest.go` — add CycleCountSession + CycleCountItem to BusDomain
- `api/sdk/http/apitest/model.go` — add to SeedData struct
- `api/cmd/services/ichor/build/all/all.go` — wire bus instances, register delegates, mount routes

---

## Task 1: Migration SQL (version 2.21)

**Files:**
- Modify: `business/sdk/migrate/sql/migrate.sql` (append after line 2312)

- [ ] **Step 1: Append migration 2.21 to migrate.sql**

Add at the end of the file:

```sql
-- Version: 2.21
-- Description: Create cycle count sessions and items tables for inventory cycle counting.
CREATE TABLE inventory.cycle_count_sessions (
    id              UUID          NOT NULL,
    name            VARCHAR(200)  NOT NULL,
    status          VARCHAR(20)   NOT NULL DEFAULT 'draft'
                        CHECK (status IN ('draft','in_progress','completed','cancelled')),
    created_by      UUID          NOT NULL REFERENCES core.users(id),
    created_date    TIMESTAMP     NOT NULL,
    updated_date    TIMESTAMP     NOT NULL,
    completed_date  TIMESTAMP     NULL,
    PRIMARY KEY (id)
);

CREATE INDEX idx_cycle_count_sessions_status ON inventory.cycle_count_sessions(status);
CREATE INDEX idx_cycle_count_sessions_created_by ON inventory.cycle_count_sessions(created_by);

INSERT INTO core.table_access (id, role_id, table_name, can_create, can_read, can_update, can_delete)
SELECT gen_random_uuid(), id, 'inventory.cycle_count_sessions', true, true, true, true FROM core.roles;

CREATE TABLE inventory.cycle_count_items (
    id                UUID          NOT NULL,
    session_id        UUID          NOT NULL REFERENCES inventory.cycle_count_sessions(id) ON DELETE CASCADE,
    product_id        UUID          NOT NULL REFERENCES products.products(id),
    location_id       UUID          NOT NULL REFERENCES inventory.inventory_locations(id),
    system_quantity   INT           NOT NULL,
    counted_quantity  INT           NULL,
    variance          INT           NULL,
    status            VARCHAR(20)   NOT NULL DEFAULT 'pending'
                          CHECK (status IN ('pending','counted','variance_approved','variance_rejected')),
    counted_by        UUID          NULL REFERENCES core.users(id),
    counted_date      TIMESTAMP     NULL,
    created_date      TIMESTAMP     NOT NULL,
    updated_date      TIMESTAMP     NOT NULL,
    PRIMARY KEY (id)
);

CREATE INDEX idx_cycle_count_items_session ON inventory.cycle_count_items(session_id);
CREATE INDEX idx_cycle_count_items_product ON inventory.cycle_count_items(product_id);
CREATE INDEX idx_cycle_count_items_location ON inventory.cycle_count_items(location_id);
CREATE INDEX idx_cycle_count_items_status ON inventory.cycle_count_items(status);

INSERT INTO core.table_access (id, role_id, table_name, can_create, can_read, can_update, can_delete)
SELECT gen_random_uuid(), id, 'inventory.cycle_count_items', true, true, true, true FROM core.roles;

ALTER TABLE inventory.inventory_adjustments DROP CONSTRAINT inventory_adjustments_reason_code_check;
ALTER TABLE inventory.inventory_adjustments ADD CONSTRAINT inventory_adjustments_reason_code_check
    CHECK (reason_code IN ('damaged', 'theft', 'data_entry_error', 'receiving_error', 'picking_error', 'found_stock', 'other', 'cycle_count'));
```

- [ ] **Step 2: Verify the migration appends cleanly**

Run: `tail -5 business/sdk/migrate/sql/migrate.sql`
Expected: The last line should be the closing of the CHECK constraint.

- [ ] **Step 3: Commit**

```bash
git add business/sdk/migrate/sql/migrate.sql
git commit -m "feat(migration): add cycle count sessions and items tables (v2.21)"
```

---

## Task 2: Add cycle_count reason code to inventoryadjustmentbus

**Files:**
- Modify: `business/domain/inventory/inventoryadjustmentbus/inventoryadjustmentbus.go:36-55`

- [ ] **Step 1: Add the constant and map entry**

In the reason code constants block (after line 43), add:

```go
ReasonCodeCycleCount = "cycle_count"
```

In the `ValidReasonCodes` map (after the `ReasonCodeOther` entry), add:

```go
ReasonCodeCycleCount: true,
```

- [ ] **Step 2: Verify the file compiles**

Run: `go build ./business/domain/inventory/inventoryadjustmentbus/...`
Expected: Clean build, no errors.

- [ ] **Step 3: Commit**

```bash
git add business/domain/inventory/inventoryadjustmentbus/inventoryadjustmentbus.go
git commit -m "feat(inventoryadjustment): add cycle_count reason code"
```

---

## Task 3: cyclecountsessionbus — Business Layer

**Files:**
- Create: `business/domain/inventory/cyclecountsessionbus/model.go`
- Create: `business/domain/inventory/cyclecountsessionbus/status.go`
- Create: `business/domain/inventory/cyclecountsessionbus/event.go`
- Create: `business/domain/inventory/cyclecountsessionbus/filter.go`
- Create: `business/domain/inventory/cyclecountsessionbus/order.go`
- Create: `business/domain/inventory/cyclecountsessionbus/cyclecountsessionbus.go`

**Reference:** Copy patterns from `business/domain/inventory/picktaskbus/` — same struct layout, status machine, delegate events, Storer interface, NewWithTx.

- [ ] **Step 1: Create model.go**

```go
package cyclecountsessionbus

import (
	"time"

	"github.com/google/uuid"
)

// JSON tags are required for workflow event serialization.

// CycleCountSession represents a cycle count session in the system.
type CycleCountSession struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Status        Status    `json:"status"`
	CreatedBy     uuid.UUID `json:"created_by"`
	CreatedDate   time.Time `json:"created_date"`
	UpdatedDate   time.Time `json:"updated_date"`
	CompletedDate time.Time `json:"completed_date"`
}

// NewCycleCountSession contains the information needed to create a new session.
type NewCycleCountSession struct {
	Name      string    `json:"name"`
	CreatedBy uuid.UUID `json:"created_by"`
}

// UpdateCycleCountSession contains optional fields for updating a session.
type UpdateCycleCountSession struct {
	Name          *string    `json:"name,omitempty"`
	Status        *Status    `json:"status,omitempty"`
	CompletedDate *time.Time `json:"completed_date,omitempty"`
}
```

- [ ] **Step 2: Create status.go**

```go
package cyclecountsessionbus

import "fmt"

type statusSet struct {
	Draft      Status
	InProgress Status
	Completed  Status
	Cancelled  Status
}

// Statuses represents the set of valid cycle count session statuses.
var Statuses = statusSet{
	Draft:      newStatus("draft"),
	InProgress: newStatus("in_progress"),
	Completed:  newStatus("completed"),
	Cancelled:  newStatus("cancelled"),
}

// =============================================================================

// Set of known statuses.
var statuses = make(map[string]Status)

// Status represents a cycle count session status in the system.
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

- [ ] **Step 3: Create event.go**

```go
package cyclecountsessionbus

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
)

// DomainName represents the name of this domain for delegate events.
const DomainName = "cyclecountsession"

// EntityName is the workflow entity name used for event matching.
const EntityName = "cycle_count_sessions"

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
	EntityID uuid.UUID         `json:"entityID"`
	UserID   uuid.UUID         `json:"userID"`
	Entity   CycleCountSession `json:"entity"`
}

func (p *ActionCreatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

func ActionCreatedData(s CycleCountSession) delegate.Data {
	params := ActionCreatedParms{
		EntityID: s.ID,
		UserID:   s.CreatedBy,
		Entity:   s,
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
	EntityID     uuid.UUID         `json:"entityID"`
	UserID       uuid.UUID         `json:"userID"`
	Entity       CycleCountSession `json:"entity"`
	BeforeEntity CycleCountSession `json:"beforeEntity,omitempty"`
}

func (p *ActionUpdatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

func ActionUpdatedData(before, after CycleCountSession) delegate.Data {
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
	EntityID uuid.UUID         `json:"entityID"`
	UserID   uuid.UUID         `json:"userID"`
	Entity   CycleCountSession `json:"entity"`
}

func (p *ActionDeletedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

func ActionDeletedData(s CycleCountSession) delegate.Data {
	params := ActionDeletedParms{
		EntityID: s.ID,
		UserID:   s.CreatedBy,
		Entity:   s,
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

- [ ] **Step 4: Create filter.go**

```go
package cyclecountsessionbus

import (
	"time"

	"github.com/google/uuid"
)

// QueryFilter holds optional filters for querying cycle count sessions.
type QueryFilter struct {
	ID          *uuid.UUID
	Name        *string
	Status      *Status
	CreatedBy   *uuid.UUID
	CreatedDate *time.Time
}
```

- [ ] **Step 5: Create order.go**

```go
package cyclecountsessionbus

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy represents the default ordering for session queries.
var DefaultOrderBy = order.NewBy(OrderByID, order.ASC)

const (
	OrderByID          = "id"
	OrderByName        = "name"
	OrderByStatus      = "status"
	OrderByCreatedBy   = "created_by"
	OrderByCreatedDate = "created_date"
)
```

- [ ] **Step 6: Create cyclecountsessionbus.go**

```go
package cyclecountsessionbus

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
	ErrNotFound            = errors.New("cycleCountSession not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry         = errors.New("cycleCountSession entry is not unique")
	ErrForeignKeyViolation = errors.New("foreign key violation")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, session CycleCountSession) error
	Update(ctx context.Context, session CycleCountSession) error
	Delete(ctx context.Context, session CycleCountSession) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]CycleCountSession, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, id uuid.UUID) (CycleCountSession, error)
}

// Business manages the set of APIs for cycle count session access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a cycle count session business API for use.
func NewBusiness(log *logger.Logger, delegate *delegate.Delegate, storer Storer) *Business {
	return &Business{
		log:      log,
		storer:   storer,
		delegate: delegate,
	}
}

// NewWithTx constructs a new Business value replacing the Storer with
// a Storer that uses the specified transaction.
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

// Create adds a new cycle count session to the system.
func (b *Business) Create(ctx context.Context, ncs NewCycleCountSession) (CycleCountSession, error) {
	ctx, span := otel.AddSpan(ctx, "business.cyclecountsessionbus.Create")
	defer span.End()

	now := time.Now()

	session := CycleCountSession{
		ID:          uuid.New(),
		Name:        ncs.Name,
		Status:      Statuses.Draft,
		CreatedBy:   ncs.CreatedBy,
		CreatedDate: now,
		UpdatedDate: now,
	}

	if err := b.storer.Create(ctx, session); err != nil {
		return CycleCountSession{}, fmt.Errorf("create: %w", err)
	}

	if err := b.delegate.Call(ctx, ActionCreatedData(session)); err != nil {
		b.log.Info(ctx, "WARNING: delegate.Call", "err", err)
	}

	return session, nil
}

// Update modifies data about a cycle count session.
func (b *Business) Update(ctx context.Context, session CycleCountSession, ucs UpdateCycleCountSession) (CycleCountSession, error) {
	ctx, span := otel.AddSpan(ctx, "business.cyclecountsessionbus.Update")
	defer span.End()

	before := session

	if ucs.Name != nil {
		session.Name = *ucs.Name
	}

	if ucs.Status != nil {
		session.Status = *ucs.Status
	}

	if ucs.CompletedDate != nil {
		session.CompletedDate = *ucs.CompletedDate
	}

	session.UpdatedDate = time.Now()

	if err := b.storer.Update(ctx, session); err != nil {
		return CycleCountSession{}, fmt.Errorf("update: %w", err)
	}

	if err := b.delegate.Call(ctx, ActionUpdatedData(before, session)); err != nil {
		b.log.Info(ctx, "WARNING: delegate.Call", "err", err)
	}

	return session, nil
}

// Delete removes the specified cycle count session.
func (b *Business) Delete(ctx context.Context, session CycleCountSession) error {
	ctx, span := otel.AddSpan(ctx, "business.cyclecountsessionbus.Delete")
	defer span.End()

	if err := b.storer.Delete(ctx, session); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	if err := b.delegate.Call(ctx, ActionDeletedData(session)); err != nil {
		b.log.Info(ctx, "WARNING: delegate.Call", "err", err)
	}

	return nil
}

// Query retrieves a list of cycle count sessions from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]CycleCountSession, error) {
	ctx, span := otel.AddSpan(ctx, "business.cyclecountsessionbus.Query")
	defer span.End()

	sessions, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return sessions, nil
}

// Count returns the total number of sessions.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.cyclecountsessionbus.Count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the session by the specified ID.
func (b *Business) QueryByID(ctx context.Context, id uuid.UUID) (CycleCountSession, error) {
	ctx, span := otel.AddSpan(ctx, "business.cyclecountsessionbus.QueryByID")
	defer span.End()

	session, err := b.storer.QueryByID(ctx, id)
	if err != nil {
		return CycleCountSession{}, fmt.Errorf("query: sessionID[%s]: %w", id, err)
	}

	return session, nil
}
```

- [ ] **Step 7: Verify the package compiles (will fail until DB store exists — that's expected)**

Run: `go vet ./business/domain/inventory/cyclecountsessionbus/...`
Expected: Compiles (no store dependency yet, but the package itself should parse cleanly).

- [ ] **Step 8: Commit**

```bash
git add business/domain/inventory/cyclecountsessionbus/
git commit -m "feat(cyclecountsession): add business layer for cycle count sessions"
```

---

## Task 4: cyclecountsessiondb — DB Store

**Files:**
- Create: `business/domain/inventory/cyclecountsessionbus/stores/cyclecountsessiondb/model.go`
- Create: `business/domain/inventory/cyclecountsessionbus/stores/cyclecountsessiondb/cyclecountsessiondb.go`
- Create: `business/domain/inventory/cyclecountsessionbus/stores/cyclecountsessiondb/filter.go`
- Create: `business/domain/inventory/cyclecountsessionbus/stores/cyclecountsessiondb/order.go`

**Reference:** `business/domain/inventory/picktaskbus/stores/picktaskdb/` — same Store struct, NamedQuery patterns, error mapping, filter builder.

- [ ] **Step 1: Create model.go**

```go
package cyclecountsessiondb

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountsessionbus"
)

type cycleCountSession struct {
	ID            uuid.UUID    `db:"id"`
	Name          string       `db:"name"`
	Status        string       `db:"status"`
	CreatedBy     uuid.UUID    `db:"created_by"`
	CreatedDate   time.Time    `db:"created_date"`
	UpdatedDate   time.Time    `db:"updated_date"`
	CompletedDate sql.NullTime `db:"completed_date"`
}

func toBusCycleCountSession(db cycleCountSession) cyclecountsessionbus.CycleCountSession {
	s := cyclecountsessionbus.CycleCountSession{
		ID:          db.ID,
		Name:        db.Name,
		Status:      cyclecountsessionbus.MustParseStatus(db.Status),
		CreatedBy:   db.CreatedBy,
		CreatedDate: db.CreatedDate,
		UpdatedDate: db.UpdatedDate,
	}

	if db.CompletedDate.Valid {
		s.CompletedDate = db.CompletedDate.Time
	}

	return s
}

func toBusCycleCountSessions(dbs []cycleCountSession) []cyclecountsessionbus.CycleCountSession {
	sessions := make([]cyclecountsessionbus.CycleCountSession, len(dbs))
	for i, db := range dbs {
		sessions[i] = toBusCycleCountSession(db)
	}
	return sessions
}

func toDBCycleCountSession(bus cyclecountsessionbus.CycleCountSession) cycleCountSession {
	db := cycleCountSession{
		ID:          bus.ID,
		Name:        bus.Name,
		Status:      bus.Status.String(),
		CreatedBy:   bus.CreatedBy,
		CreatedDate: bus.CreatedDate,
		UpdatedDate: bus.UpdatedDate,
	}

	if !bus.CompletedDate.IsZero() {
		db.CompletedDate = sql.NullTime{Time: bus.CompletedDate, Valid: true}
	}

	return db
}
```

- [ ] **Step 2: Create cyclecountsessiondb.go**

```go
package cyclecountsessiondb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountsessionbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for cycle count session database access.
type Store struct {
	log *logger.Logger
	db  sqlx.ExtContext
}

// NewStore constructs the API for data access.
func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// NewWithTx constructs a new Store value replacing the sqlx DB value with a
// sqlx DB value that is currently inside a transaction.
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (cyclecountsessionbus.Storer, error) {
	ec, err := sqldb.GetExtContext(tx)
	if err != nil {
		return nil, err
	}

	store := Store{
		log: s.log,
		db:  ec,
	}

	return &store, nil
}

// Create inserts a new cycle count session into the database.
func (s *Store) Create(ctx context.Context, session cyclecountsessionbus.CycleCountSession) error {
	const q = `
    INSERT INTO inventory.cycle_count_sessions
        (id, name, status, created_by, created_date, updated_date, completed_date)
    VALUES
        (:id, :name, :status, :created_by, :created_date, :updated_date, :completed_date)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBCycleCountSession(session)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return cyclecountsessionbus.ErrForeignKeyViolation
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return cyclecountsessionbus.ErrUniqueEntry
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update replaces a cycle count session document in the database.
func (s *Store) Update(ctx context.Context, session cyclecountsessionbus.CycleCountSession) error {
	const q = `
    UPDATE
        inventory.cycle_count_sessions
    SET
        name = :name,
        status = :status,
        updated_date = :updated_date,
        completed_date = :completed_date
    WHERE
        id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBCycleCountSession(session)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return cyclecountsessionbus.ErrForeignKeyViolation
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return cyclecountsessionbus.ErrUniqueEntry
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a cycle count session from the database.
func (s *Store) Delete(ctx context.Context, session cyclecountsessionbus.CycleCountSession) error {
	const q = `
    DELETE FROM
        inventory.cycle_count_sessions
    WHERE
        id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBCycleCountSession(session)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of cycle count sessions from the database.
func (s *Store) Query(ctx context.Context, filter cyclecountsessionbus.QueryFilter, orderBy order.By, p page.Page) ([]cyclecountsessionbus.CycleCountSession, error) {
	data := map[string]any{
		"offset":        (p.Number() - 1) * p.RowsPerPage(),
		"rows_per_page": p.RowsPerPage(),
	}

	const q = `SELECT id, name, status, created_by, created_date, updated_date, completed_date FROM inventory.cycle_count_sessions`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbs []cycleCountSession
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbs); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusCycleCountSessions(dbs), nil
}

// Count returns the total number of sessions in the DB.
func (s *Store) Count(ctx context.Context, filter cyclecountsessionbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `SELECT count(1) AS count FROM inventory.cycle_count_sessions`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("namedquerystruct: %w", err)
	}

	return count.Count, nil
}

// QueryByID retrieves a single session by its id.
func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (cyclecountsessionbus.CycleCountSession, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: id.String(),
	}

	const q = `SELECT id, name, status, created_by, created_date, updated_date, completed_date FROM inventory.cycle_count_sessions WHERE id = :id`

	var db cycleCountSession
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &db); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return cyclecountsessionbus.CycleCountSession{}, cyclecountsessionbus.ErrNotFound
		}
		return cyclecountsessionbus.CycleCountSession{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusCycleCountSession(db), nil
}
```

- [ ] **Step 3: Create filter.go**

```go
package cyclecountsessiondb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountsessionbus"
)

func applyFilter(filter cyclecountsessionbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.Name != nil {
		data["name"] = "%" + *filter.Name + "%"
		wc = append(wc, "name ILIKE :name")
	}

	if filter.Status != nil {
		data["status"] = filter.Status.String()
		wc = append(wc, "status = :status")
	}

	if filter.CreatedBy != nil {
		data["created_by"] = *filter.CreatedBy
		wc = append(wc, "created_by = :created_by")
	}

	if filter.CreatedDate != nil {
		data["created_date"] = *filter.CreatedDate
		wc = append(wc, "created_date >= :created_date")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
```

- [ ] **Step 4: Create order.go**

```go
package cyclecountsessiondb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountsessionbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	cyclecountsessionbus.OrderByID:          "id",
	cyclecountsessionbus.OrderByName:        "name",
	cyclecountsessionbus.OrderByStatus:      "status",
	cyclecountsessionbus.OrderByCreatedBy:   "created_by",
	cyclecountsessionbus.OrderByCreatedDate: "created_date",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
```

- [ ] **Step 5: Verify full package compiles**

Run: `go build ./business/domain/inventory/cyclecountsessionbus/...`
Expected: Clean build, no errors.

- [ ] **Step 6: Commit**

```bash
git add business/domain/inventory/cyclecountsessionbus/stores/
git commit -m "feat(cyclecountsession): add database store layer"
```

---

## Task 5: cyclecountitembus — Business Layer

**Files:**
- Create: `business/domain/inventory/cyclecountitembus/model.go`
- Create: `business/domain/inventory/cyclecountitembus/status.go`
- Create: `business/domain/inventory/cyclecountitembus/event.go`
- Create: `business/domain/inventory/cyclecountitembus/filter.go`
- Create: `business/domain/inventory/cyclecountitembus/order.go`
- Create: `business/domain/inventory/cyclecountitembus/cyclecountitembus.go`

**Reference:** Same patterns as Task 3, adapted for cycle count items.

- [ ] **Step 1: Create model.go**

```go
package cyclecountitembus

import (
	"time"

	"github.com/google/uuid"
)

// JSON tags are required for workflow event serialization.

// CycleCountItem represents a single line item within a cycle count session.
type CycleCountItem struct {
	ID              uuid.UUID `json:"id"`
	SessionID       uuid.UUID `json:"session_id"`
	ProductID       uuid.UUID `json:"product_id"`
	LocationID      uuid.UUID `json:"location_id"`
	SystemQuantity  int       `json:"system_quantity"`
	CountedQuantity *int      `json:"counted_quantity,omitempty"`
	Variance        *int      `json:"variance,omitempty"`
	Status          Status    `json:"status"`
	CountedBy       uuid.UUID `json:"counted_by"`
	CountedDate     time.Time `json:"counted_date"`
	CreatedDate     time.Time `json:"created_date"`
	UpdatedDate     time.Time `json:"updated_date"`
}

// NewCycleCountItem contains the information needed to create a new item.
type NewCycleCountItem struct {
	SessionID      uuid.UUID `json:"session_id"`
	ProductID      uuid.UUID `json:"product_id"`
	LocationID     uuid.UUID `json:"location_id"`
	SystemQuantity int       `json:"system_quantity"`
}

// UpdateCycleCountItem contains optional fields for updating an item.
type UpdateCycleCountItem struct {
	CountedQuantity *int       `json:"counted_quantity,omitempty"`
	Status          *Status    `json:"status,omitempty"`
	CountedBy       *uuid.UUID `json:"counted_by,omitempty"`
	CountedDate     *time.Time `json:"counted_date,omitempty"`
}
```

- [ ] **Step 2: Create status.go**

```go
package cyclecountitembus

import "fmt"

type statusSet struct {
	Pending          Status
	Counted          Status
	VarianceApproved Status
	VarianceRejected Status
}

// Statuses represents the set of valid cycle count item statuses.
var Statuses = statusSet{
	Pending:          newStatus("pending"),
	Counted:          newStatus("counted"),
	VarianceApproved: newStatus("variance_approved"),
	VarianceRejected: newStatus("variance_rejected"),
}

// =============================================================================

// Set of known statuses.
var statuses = make(map[string]Status)

// Status represents a cycle count item status in the system.
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

- [ ] **Step 3: Create event.go**

```go
package cyclecountitembus

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
)

// DomainName represents the name of this domain for delegate events.
const DomainName = "cyclecountitem"

// EntityName is the workflow entity name used for event matching.
const EntityName = "cycle_count_items"

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
	EntityID uuid.UUID      `json:"entityID"`
	UserID   uuid.UUID      `json:"userID"`
	Entity   CycleCountItem `json:"entity"`
}

func (p *ActionCreatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

func ActionCreatedData(item CycleCountItem) delegate.Data {
	params := ActionCreatedParms{
		EntityID: item.ID,
		Entity:   item,
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
	EntityID     uuid.UUID      `json:"entityID"`
	UserID       uuid.UUID      `json:"userID"`
	Entity       CycleCountItem `json:"entity"`
	BeforeEntity CycleCountItem `json:"beforeEntity,omitempty"`
}

func (p *ActionUpdatedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

func ActionUpdatedData(before, after CycleCountItem) delegate.Data {
	params := ActionUpdatedParms{
		EntityID:     after.ID,
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
	EntityID uuid.UUID      `json:"entityID"`
	UserID   uuid.UUID      `json:"userID"`
	Entity   CycleCountItem `json:"entity"`
}

func (p *ActionDeletedParms) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

func ActionDeletedData(item CycleCountItem) delegate.Data {
	params := ActionDeletedParms{
		EntityID: item.ID,
		Entity:   item,
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

- [ ] **Step 4: Create filter.go**

```go
package cyclecountitembus

import (
	"github.com/google/uuid"
)

// QueryFilter holds optional filters for querying cycle count items.
type QueryFilter struct {
	ID         *uuid.UUID
	SessionID  *uuid.UUID
	ProductID  *uuid.UUID
	LocationID *uuid.UUID
	Status     *Status
	CountedBy  *uuid.UUID
}
```

- [ ] **Step 5: Create order.go**

```go
package cyclecountitembus

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy represents the default ordering for item queries.
var DefaultOrderBy = order.NewBy(OrderByID, order.ASC)

const (
	OrderByID             = "id"
	OrderBySessionID      = "session_id"
	OrderByProductID      = "product_id"
	OrderByLocationID     = "location_id"
	OrderBySystemQuantity = "system_quantity"
	OrderByStatus         = "status"
	OrderByCreatedDate    = "created_date"
)
```

- [ ] **Step 6: Create cyclecountitembus.go**

```go
package cyclecountitembus

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
	ErrNotFound            = errors.New("cycleCountItem not found")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrUniqueEntry         = errors.New("cycleCountItem entry is not unique")
	ErrForeignKeyViolation = errors.New("foreign key violation")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, item CycleCountItem) error
	Update(ctx context.Context, item CycleCountItem) error
	Delete(ctx context.Context, item CycleCountItem) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]CycleCountItem, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, id uuid.UUID) (CycleCountItem, error)
}

// Business manages the set of APIs for cycle count item access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a cycle count item business API for use.
func NewBusiness(log *logger.Logger, delegate *delegate.Delegate, storer Storer) *Business {
	return &Business{
		log:      log,
		storer:   storer,
		delegate: delegate,
	}
}

// NewWithTx constructs a new Business value replacing the Storer with
// a Storer that uses the specified transaction.
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

// Create adds a new cycle count item to the system.
func (b *Business) Create(ctx context.Context, nci NewCycleCountItem) (CycleCountItem, error) {
	ctx, span := otel.AddSpan(ctx, "business.cyclecountitembus.Create")
	defer span.End()

	now := time.Now()

	item := CycleCountItem{
		ID:             uuid.New(),
		SessionID:      nci.SessionID,
		ProductID:      nci.ProductID,
		LocationID:     nci.LocationID,
		SystemQuantity: nci.SystemQuantity,
		Status:         Statuses.Pending,
		CreatedDate:    now,
		UpdatedDate:    now,
	}

	if err := b.storer.Create(ctx, item); err != nil {
		return CycleCountItem{}, fmt.Errorf("create: %w", err)
	}

	if err := b.delegate.Call(ctx, ActionCreatedData(item)); err != nil {
		b.log.Info(ctx, "WARNING: delegate.Call", "err", err)
	}

	return item, nil
}

// Update modifies data about a cycle count item.
func (b *Business) Update(ctx context.Context, item CycleCountItem, uci UpdateCycleCountItem) (CycleCountItem, error) {
	ctx, span := otel.AddSpan(ctx, "business.cyclecountitembus.Update")
	defer span.End()

	before := item

	if uci.CountedQuantity != nil {
		item.CountedQuantity = uci.CountedQuantity
		variance := *uci.CountedQuantity - item.SystemQuantity
		item.Variance = &variance
	}

	if uci.Status != nil {
		item.Status = *uci.Status
	}

	if uci.CountedBy != nil {
		item.CountedBy = *uci.CountedBy
	}

	if uci.CountedDate != nil {
		item.CountedDate = *uci.CountedDate
	}

	item.UpdatedDate = time.Now()

	if err := b.storer.Update(ctx, item); err != nil {
		return CycleCountItem{}, fmt.Errorf("update: %w", err)
	}

	if err := b.delegate.Call(ctx, ActionUpdatedData(before, item)); err != nil {
		b.log.Info(ctx, "WARNING: delegate.Call", "err", err)
	}

	return item, nil
}

// Delete removes the specified cycle count item.
func (b *Business) Delete(ctx context.Context, item CycleCountItem) error {
	ctx, span := otel.AddSpan(ctx, "business.cyclecountitembus.Delete")
	defer span.End()

	if err := b.storer.Delete(ctx, item); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	if err := b.delegate.Call(ctx, ActionDeletedData(item)); err != nil {
		b.log.Info(ctx, "WARNING: delegate.Call", "err", err)
	}

	return nil
}

// Query retrieves a list of cycle count items from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]CycleCountItem, error) {
	ctx, span := otel.AddSpan(ctx, "business.cyclecountitembus.Query")
	defer span.End()

	items, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return items, nil
}

// Count returns the total number of items.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.cyclecountitembus.Count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the item by the specified ID.
func (b *Business) QueryByID(ctx context.Context, id uuid.UUID) (CycleCountItem, error) {
	ctx, span := otel.AddSpan(ctx, "business.cyclecountitembus.QueryByID")
	defer span.End()

	item, err := b.storer.QueryByID(ctx, id)
	if err != nil {
		return CycleCountItem{}, fmt.Errorf("query: itemID[%s]: %w", id, err)
	}

	return item, nil
}
```

- [ ] **Step 7: Verify the package compiles**

Run: `go vet ./business/domain/inventory/cyclecountitembus/...`
Expected: Clean build.

- [ ] **Step 8: Commit**

```bash
git add business/domain/inventory/cyclecountitembus/
git commit -m "feat(cyclecountitem): add business layer for cycle count items"
```

---

## Task 6: cyclecountitemdb — DB Store

**Files:**
- Create: `business/domain/inventory/cyclecountitembus/stores/cyclecountitemdb/model.go`
- Create: `business/domain/inventory/cyclecountitembus/stores/cyclecountitemdb/cyclecountitemdb.go`
- Create: `business/domain/inventory/cyclecountitembus/stores/cyclecountitemdb/filter.go`
- Create: `business/domain/inventory/cyclecountitembus/stores/cyclecountitemdb/order.go`

- [ ] **Step 1: Create model.go**

```go
package cyclecountitemdb

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountitembus"
	"github.com/timmaaaz/ichor/business/sdk/nulltypes"
)

type cycleCountItem struct {
	ID              uuid.UUID      `db:"id"`
	SessionID       uuid.UUID      `db:"session_id"`
	ProductID       uuid.UUID      `db:"product_id"`
	LocationID      uuid.UUID      `db:"location_id"`
	SystemQuantity  int            `db:"system_quantity"`
	CountedQuantity sql.NullInt64  `db:"counted_quantity"`
	Variance        sql.NullInt64  `db:"variance"`
	Status          string         `db:"status"`
	CountedBy       sql.NullString `db:"counted_by"`
	CountedDate     sql.NullTime   `db:"counted_date"`
	CreatedDate     time.Time      `db:"created_date"`
	UpdatedDate     time.Time      `db:"updated_date"`
}

func toBusCycleCountItem(db cycleCountItem) cyclecountitembus.CycleCountItem {
	item := cyclecountitembus.CycleCountItem{
		ID:             db.ID,
		SessionID:      db.SessionID,
		ProductID:      db.ProductID,
		LocationID:     db.LocationID,
		SystemQuantity: db.SystemQuantity,
		Status:         cyclecountitembus.MustParseStatus(db.Status),
		CountedBy:      nulltypes.FromNullableUUID(db.CountedBy),
		CreatedDate:    db.CreatedDate,
		UpdatedDate:    db.UpdatedDate,
	}

	if db.CountedQuantity.Valid {
		v := int(db.CountedQuantity.Int64)
		item.CountedQuantity = &v
	}

	if db.Variance.Valid {
		v := int(db.Variance.Int64)
		item.Variance = &v
	}

	if db.CountedDate.Valid {
		item.CountedDate = db.CountedDate.Time
	}

	return item
}

func toBusCycleCountItems(dbs []cycleCountItem) []cyclecountitembus.CycleCountItem {
	items := make([]cyclecountitembus.CycleCountItem, len(dbs))
	for i, db := range dbs {
		items[i] = toBusCycleCountItem(db)
	}
	return items
}

func toDBCycleCountItem(bus cyclecountitembus.CycleCountItem) cycleCountItem {
	db := cycleCountItem{
		ID:             bus.ID,
		SessionID:      bus.SessionID,
		ProductID:      bus.ProductID,
		LocationID:     bus.LocationID,
		SystemQuantity: bus.SystemQuantity,
		Status:         bus.Status.String(),
		CountedBy:      nulltypes.ToNullableUUID(bus.CountedBy),
		CreatedDate:    bus.CreatedDate,
		UpdatedDate:    bus.UpdatedDate,
	}

	if bus.CountedQuantity != nil {
		db.CountedQuantity = sql.NullInt64{Int64: int64(*bus.CountedQuantity), Valid: true}
	}

	if bus.Variance != nil {
		db.Variance = sql.NullInt64{Int64: int64(*bus.Variance), Valid: true}
	}

	if !bus.CountedDate.IsZero() {
		db.CountedDate = sql.NullTime{Time: bus.CountedDate, Valid: true}
	}

	return db
}
```

- [ ] **Step 2: Create cyclecountitemdb.go**

```go
package cyclecountitemdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountitembus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for cycle count item database access.
type Store struct {
	log *logger.Logger
	db  sqlx.ExtContext
}

// NewStore constructs the API for data access.
func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// NewWithTx constructs a new Store value replacing the sqlx DB value with a
// sqlx DB value that is currently inside a transaction.
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (cyclecountitembus.Storer, error) {
	ec, err := sqldb.GetExtContext(tx)
	if err != nil {
		return nil, err
	}

	store := Store{
		log: s.log,
		db:  ec,
	}

	return &store, nil
}

// Create inserts a new cycle count item into the database.
func (s *Store) Create(ctx context.Context, item cyclecountitembus.CycleCountItem) error {
	const q = `
    INSERT INTO inventory.cycle_count_items
        (id, session_id, product_id, location_id, system_quantity, counted_quantity, variance, status, counted_by, counted_date, created_date, updated_date)
    VALUES
        (:id, :session_id, :product_id, :location_id, :system_quantity, :counted_quantity, :variance, :status, :counted_by, :counted_date, :created_date, :updated_date)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBCycleCountItem(item)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return cyclecountitembus.ErrForeignKeyViolation
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return cyclecountitembus.ErrUniqueEntry
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update replaces a cycle count item document in the database.
func (s *Store) Update(ctx context.Context, item cyclecountitembus.CycleCountItem) error {
	const q = `
    UPDATE
        inventory.cycle_count_items
    SET
        counted_quantity = :counted_quantity,
        variance = :variance,
        status = :status,
        counted_by = :counted_by,
        counted_date = :counted_date,
        updated_date = :updated_date
    WHERE
        id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBCycleCountItem(item)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return cyclecountitembus.ErrForeignKeyViolation
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return cyclecountitembus.ErrUniqueEntry
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a cycle count item from the database.
func (s *Store) Delete(ctx context.Context, item cyclecountitembus.CycleCountItem) error {
	const q = `
    DELETE FROM
        inventory.cycle_count_items
    WHERE
        id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBCycleCountItem(item)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of cycle count items from the database.
func (s *Store) Query(ctx context.Context, filter cyclecountitembus.QueryFilter, orderBy order.By, p page.Page) ([]cyclecountitembus.CycleCountItem, error) {
	data := map[string]any{
		"offset":        (p.Number() - 1) * p.RowsPerPage(),
		"rows_per_page": p.RowsPerPage(),
	}

	const q = `SELECT id, session_id, product_id, location_id, system_quantity, counted_quantity, variance, status, counted_by, counted_date, created_date, updated_date FROM inventory.cycle_count_items`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbs []cycleCountItem
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbs); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusCycleCountItems(dbs), nil
}

// Count returns the total number of items in the DB.
func (s *Store) Count(ctx context.Context, filter cyclecountitembus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `SELECT count(1) AS count FROM inventory.cycle_count_items`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("namedquerystruct: %w", err)
	}

	return count.Count, nil
}

// QueryByID retrieves a single item by its id.
func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (cyclecountitembus.CycleCountItem, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: id.String(),
	}

	const q = `SELECT id, session_id, product_id, location_id, system_quantity, counted_quantity, variance, status, counted_by, counted_date, created_date, updated_date FROM inventory.cycle_count_items WHERE id = :id`

	var db cycleCountItem
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &db); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return cyclecountitembus.CycleCountItem{}, cyclecountitembus.ErrNotFound
		}
		return cyclecountitembus.CycleCountItem{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusCycleCountItem(db), nil
}
```

- [ ] **Step 3: Create filter.go**

```go
package cyclecountitemdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountitembus"
)

func applyFilter(filter cyclecountitembus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.SessionID != nil {
		data["session_id"] = *filter.SessionID
		wc = append(wc, "session_id = :session_id")
	}

	if filter.ProductID != nil {
		data["product_id"] = *filter.ProductID
		wc = append(wc, "product_id = :product_id")
	}

	if filter.LocationID != nil {
		data["location_id"] = *filter.LocationID
		wc = append(wc, "location_id = :location_id")
	}

	if filter.Status != nil {
		data["status"] = filter.Status.String()
		wc = append(wc, "status = :status")
	}

	if filter.CountedBy != nil {
		data["counted_by"] = *filter.CountedBy
		wc = append(wc, "counted_by = :counted_by")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
```

- [ ] **Step 4: Create order.go**

```go
package cyclecountitemdb

import (
	"fmt"

	"github.com/timmaaez/ichor/business/domain/inventory/cyclecountitembus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	cyclecountitembus.OrderByID:             "id",
	cyclecountitembus.OrderBySessionID:      "session_id",
	cyclecountitembus.OrderByProductID:      "product_id",
	cyclecountitembus.OrderByLocationID:     "location_id",
	cyclecountitembus.OrderBySystemQuantity: "system_quantity",
	cyclecountitembus.OrderByStatus:         "status",
	cyclecountitembus.OrderByCreatedDate:    "created_date",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
```

- [ ] **Step 5: Verify full package compiles**

Run: `go build ./business/domain/inventory/cyclecountitembus/...`
Expected: Clean build.

- [ ] **Step 6: Commit**

```bash
git add business/domain/inventory/cyclecountitembus/stores/
git commit -m "feat(cyclecountitem): add database store layer"
```

---

## Task 7: cyclecountsessionapp — App Layer with Complete Flow

**Files:**
- Create: `app/domain/inventory/cyclecountsessionapp/model.go`
- Create: `app/domain/inventory/cyclecountsessionapp/cyclecountsessionapp.go`

**Reference:** `app/domain/inventory/picktaskapp/picktaskapp.go` for the atomic `complete()` transaction pattern.

This is the most critical task — it contains the `complete` flow that locks the session and generates inventory adjustment records.

- [ ] **Step 1: Create model.go**

```go
package cyclecountsessionapp

import (
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountsessionbus"
)

// QueryParams represents the set of possible query strings.
type QueryParams struct {
	Page             string
	Rows             string
	OrderBy          string
	ID               string
	Name             string
	Status           string
	CreatedBy        string
	CreatedDate      string
}

// =============================================================================

// CycleCountSession represents information about an individual session.
type CycleCountSession struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Status        string `json:"status"`
	CreatedBy     string `json:"createdBy"`
	CreatedDate   string `json:"createdDate"`
	UpdatedDate   string `json:"updatedDate"`
	CompletedDate string `json:"completedDate"`
}

// Encode implements the web.Encoder interface.
func (app CycleCountSession) Encode() ([]byte, string, error) {
	data, err := errs.MarshalIndentJSON(app)
	return data, "application/json", err
}

func ToAppCycleCountSession(bus cyclecountsessionbus.CycleCountSession) CycleCountSession {
	completedDate := ""
	if !bus.CompletedDate.IsZero() {
		completedDate = bus.CompletedDate.Format(time.RFC3339)
	}

	return CycleCountSession{
		ID:            bus.ID.String(),
		Name:          bus.Name,
		Status:        bus.Status.String(),
		CreatedBy:     bus.CreatedBy.String(),
		CreatedDate:   bus.CreatedDate.Format(time.RFC3339),
		UpdatedDate:   bus.UpdatedDate.Format(time.RFC3339),
		CompletedDate: completedDate,
	}
}

func ToAppCycleCountSessions(sessions []cyclecountsessionbus.CycleCountSession) []CycleCountSession {
	app := make([]CycleCountSession, len(sessions))
	for i, s := range sessions {
		app[i] = ToAppCycleCountSession(s)
	}
	return app
}

// =============================================================================

// NewCycleCountSession defines the data needed to create a new session.
type NewCycleCountSession struct {
	Name string `json:"name" validate:"required"`
}

// Decode implements the web.Decoder interface.
func (app *NewCycleCountSession) Decode(data []byte) error {
	return errs.UnmarshalJSON(data, app)
}

// Validate checks the data in the model is considered clean.
func (app NewCycleCountSession) Validate() error {
	if err := errs.Check(app); err != nil {
		return err
	}
	return nil
}

func toBusNewCycleCountSession(app NewCycleCountSession, createdBy uuid.UUID) cyclecountsessionbus.NewCycleCountSession {
	return cyclecountsessionbus.NewCycleCountSession{
		Name:      app.Name,
		CreatedBy: createdBy,
	}
}

// =============================================================================

// UpdateCycleCountSession defines the data needed to update a session.
type UpdateCycleCountSession struct {
	Name   *string `json:"name" validate:"omitempty"`
	Status *string `json:"status" validate:"omitempty"`
}

// Decode implements the web.Decoder interface.
func (app *UpdateCycleCountSession) Decode(data []byte) error {
	return errs.UnmarshalJSON(data, app)
}

// Validate checks the data in the model is considered clean.
func (app UpdateCycleCountSession) Validate() error {
	if err := errs.Check(app); err != nil {
		return err
	}
	return nil
}

func toBusUpdateCycleCountSession(app UpdateCycleCountSession) (cyclecountsessionbus.UpdateCycleCountSession, error) {
	var ucs cyclecountsessionbus.UpdateCycleCountSession

	if app.Name != nil {
		ucs.Name = app.Name
	}

	if app.Status != nil {
		status, err := cyclecountsessionbus.ParseStatus(*app.Status)
		if err != nil {
			return cyclecountsessionbus.UpdateCycleCountSession{}, err
		}
		ucs.Status = &status
	}

	return ucs, nil
}

// =============================================================================

// parseFilter builds a session bus QueryFilter from query params.
func parseFilter(qp QueryParams) (cyclecountsessionbus.QueryFilter, error) {
	var filter cyclecountsessionbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return cyclecountsessionbus.QueryFilter{}, err
		}
		filter.ID = &id
	}

	if qp.Name != "" {
		filter.Name = &qp.Name
	}

	if qp.Status != "" {
		status, err := cyclecountsessionbus.ParseStatus(qp.Status)
		if err != nil {
			return cyclecountsessionbus.QueryFilter{}, err
		}
		filter.Status = &status
	}

	if qp.CreatedBy != "" {
		id, err := uuid.Parse(qp.CreatedBy)
		if err != nil {
			return cyclecountsessionbus.QueryFilter{}, err
		}
		filter.CreatedBy = &id
	}

	return filter, nil
}

// parseOrder builds the order from query params.
func parseOrder(qp QueryParams) (string, error) {
	if qp.OrderBy == "" {
		return "", nil
	}
	return qp.OrderBy, nil
}

// parsePage builds page from query params.
func parsePage(qp QueryParams) (int, int) {
	pg := 1
	rows := 10

	if qp.Page != "" {
		if v, err := strconv.Atoi(qp.Page); err == nil {
			pg = v
		}
	}

	if qp.Rows != "" {
		if v, err := strconv.Atoi(qp.Rows); err == nil {
			rows = v
		}
	}

	return pg, rows
}
```

- [ ] **Step 2: Create cyclecountsessionapp.go**

This is the core file — contains the `complete` method that atomically locks the session and generates inventory adjustments for approved variances.

```go
package cyclecountsessionapp

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/mid"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaez/ichor/business/domain/inventory/cyclecountitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountsessionbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryadjustmentbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaez/ichor/business/sdk/page"
)

// App manages the set of app layer API functions for cycle count sessions.
type App struct {
	cycleCountSessionBus *cyclecountsessionbus.Business
	cycleCountItemBus    *cyclecountitembus.Business
	invAdjustmentBus     *inventoryadjustmentbus.Business
	db                   *sqlx.DB
}

// NewApp constructs a cycle count session app API for use.
func NewApp(sessionBus *cyclecountsessionbus.Business, itemBus *cyclecountitembus.Business, adjBus *inventoryadjustmentbus.Business, db *sqlx.DB) *App {
	return &App{
		cycleCountSessionBus: sessionBus,
		cycleCountItemBus:    itemBus,
		invAdjustmentBus:     adjBus,
		db:                   db,
	}
}

// Create adds a new cycle count session to the system.
func (a *App) Create(ctx context.Context, app NewCycleCountSession) (CycleCountSession, error) {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return CycleCountSession{}, errs.New(errs.Unauthenticated, err)
	}

	ncs := toBusNewCycleCountSession(app, userID)

	session, err := a.cycleCountSessionBus.Create(ctx, ncs)
	if err != nil {
		if errors.Is(err, cyclecountsessionbus.ErrUniqueEntry) {
			return CycleCountSession{}, errs.New(errs.Aborted, cyclecountsessionbus.ErrUniqueEntry)
		}
		return CycleCountSession{}, errs.Newf(errs.Internal, "create: session[%+v]: %s", app, err)
	}

	return ToAppCycleCountSession(session), nil
}

// Update modifies data about a cycle count session. If the new status is
// "completed", the update is routed through the complete flow which locks
// the session and generates inventory adjustments for approved variances.
func (a *App) Update(ctx context.Context, session cyclecountsessionbus.CycleCountSession, app UpdateCycleCountSession) (CycleCountSession, error) {
	ucs, err := toBusUpdateCycleCountSession(app)
	if err != nil {
		return CycleCountSession{}, errs.New(errs.FailedPrecondition, err)
	}

	// Terminal state guard: completed and cancelled sessions cannot be modified.
	if session.Status.Equal(cyclecountsessionbus.Statuses.Completed) || session.Status.Equal(cyclecountsessionbus.Statuses.Cancelled) {
		return CycleCountSession{}, errs.New(errs.FailedPrecondition, fmt.Errorf("session is in terminal state %q", session.Status))
	}

	// Route to the atomic complete flow when status is being set to completed.
	if ucs.Status != nil && ucs.Status.Equal(cyclecountsessionbus.Statuses.Completed) {
		// Only in_progress sessions can be completed.
		if !session.Status.Equal(cyclecountsessionbus.Statuses.InProgress) {
			return CycleCountSession{}, errs.New(errs.FailedPrecondition, fmt.Errorf("session must be in_progress to complete, current status: %q", session.Status))
		}

		return a.complete(ctx, session)
	}

	updated, err := a.cycleCountSessionBus.Update(ctx, session, ucs)
	if err != nil {
		return CycleCountSession{}, errs.Newf(errs.Internal, "update: sessionID[%s]: %s", session.ID, err)
	}

	return ToAppCycleCountSession(updated), nil
}

// complete locks a session and generates inventory_adjustment records for all
// items with variance_approved status and non-zero variance. This runs as an
// atomic transaction.
func (a *App) complete(ctx context.Context, session cyclecountsessionbus.CycleCountSession) (CycleCountSession, error) {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return CycleCountSession{}, errs.New(errs.Unauthenticated, err)
	}

	now := time.Now()

	tx, err := a.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return CycleCountSession{}, errs.Newf(errs.Internal, "begin tx: %s", err)
	}
	defer tx.Rollback()

	// 1. Lock session — set status to completed.
	sessionBusTx, err := a.cycleCountSessionBus.NewWithTx(tx)
	if err != nil {
		return CycleCountSession{}, errs.Newf(errs.Internal, "session newwithtx: %s", err)
	}

	completedStatus := cyclecountsessionbus.Statuses.Completed
	updatedSession, err := sessionBusTx.Update(ctx, session, cyclecountsessionbus.UpdateCycleCountSession{
		Status:        &completedStatus,
		CompletedDate: &now,
	})
	if err != nil {
		return CycleCountSession{}, errs.Newf(errs.Internal, "complete: lock session[%s]: %s", session.ID, err)
	}

	// 2. Query all items for this session with variance_approved status.
	itemBusTx, err := a.cycleCountItemBus.NewWithTx(tx)
	if err != nil {
		return CycleCountSession{}, errs.Newf(errs.Internal, "item newwithtx: %s", err)
	}

	approvedStatus := cyclecountitembus.Statuses.VarianceApproved
	items, err := itemBusTx.Query(ctx, cyclecountitembus.QueryFilter{
		SessionID: &session.ID,
		Status:    &approvedStatus,
	}, cyclecountitembus.DefaultOrderBy, page.MustParse("1", "1000"))
	if err != nil {
		return CycleCountSession{}, errs.Newf(errs.Internal, "complete: query items[%s]: %s", session.ID, err)
	}

	// 3. Generate inventory adjustments for items with non-zero variance.
	if len(items) > 0 {
		adjBusTx, err := a.invAdjustmentBus.NewWithTx(tx)
		if err != nil {
			return CycleCountSession{}, errs.Newf(errs.Internal, "adj newwithtx: %s", err)
		}

		for _, item := range items {
			if item.Variance == nil || *item.Variance == 0 {
				continue
			}

			_, err := adjBusTx.Create(ctx, inventoryadjustmentbus.NewInventoryAdjustment{
				ProductID:      item.ProductID,
				LocationID:     item.LocationID,
				AdjustedBy:     userID,
				ApprovedBy:     &userID,
				ApprovalStatus: inventoryadjustmentbus.ApprovalStatusApproved,
				QuantityChange: *item.Variance,
				ReasonCode:     inventoryadjustmentbus.ReasonCodeCycleCount,
				Notes:          fmt.Sprintf("Cycle count session %s: system_qty=%d, counted_qty=%d", session.Name, item.SystemQuantity, *item.CountedQuantity),
				AdjustmentDate: now,
			})
			if err != nil {
				return CycleCountSession{}, errs.Newf(errs.Internal, "complete: create adjustment for item[%s]: %s", item.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return CycleCountSession{}, errs.Newf(errs.Internal, "commit: %s", err)
	}

	return ToAppCycleCountSession(updatedSession), nil
}

// Delete removes the specified cycle count session.
func (a *App) Delete(ctx context.Context, session cyclecountsessionbus.CycleCountSession) error {
	if err := a.cycleCountSessionBus.Delete(ctx, session); err != nil {
		return errs.Newf(errs.Internal, "delete: sessionID[%s]: %s", session.ID, err)
	}

	return nil
}

// Query returns a list of sessions with paging.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[CycleCountSession], error) {
	pg, rows := parsePage(qp)
	p, err := page.Parse(pg, rows)
	if err != nil {
		return query.Result[CycleCountSession]{}, errs.New(errs.FailedPrecondition, err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[CycleCountSession]{}, errs.New(errs.FailedPrecondition, err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, cyclecountsessionbus.DefaultOrderBy)
	if err != nil {
		return query.Result[CycleCountSession]{}, errs.New(errs.FailedPrecondition, err)
	}

	sessions, err := a.cycleCountSessionBus.Query(ctx, filter, orderBy, p)
	if err != nil {
		return query.Result[CycleCountSession]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.cycleCountSessionBus.Count(ctx, filter)
	if err != nil {
		return query.Result[CycleCountSession]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppCycleCountSessions(sessions), total, p), nil
}

// QueryByID finds the session by the specified ID.
func (a *App) QueryByID(ctx context.Context, session cyclecountsessionbus.CycleCountSession) (CycleCountSession, error) {
	return ToAppCycleCountSession(session), nil
}

// =============================================================================

// orderByFields maps app-level field names to bus-level field names.
var orderByFields = map[string]string{
	"id":          cyclecountsessionbus.OrderByID,
	"name":        cyclecountsessionbus.OrderByName,
	"status":      cyclecountsessionbus.OrderByStatus,
	"createdBy":   cyclecountsessionbus.OrderByCreatedBy,
	"createdDate": cyclecountsessionbus.OrderByCreatedDate,
}
```

- [ ] **Step 3: Verify the package compiles**

Run: `go build ./app/domain/inventory/cyclecountsessionapp/...`
Expected: Clean build. (Assumes bus layers from Tasks 3-6 are committed.)

- [ ] **Step 4: Commit**

```bash
git add app/domain/inventory/cyclecountsessionapp/
git commit -m "feat(cyclecountsession): add app layer with atomic complete flow"
```

---

## Task 8: cyclecountitemapp — App Layer

**Files:**
- Create: `app/domain/inventory/cyclecountitemapp/model.go`
- Create: `app/domain/inventory/cyclecountitemapp/cyclecountitemapp.go`

- [ ] **Step 1: Create model.go**

```go
package cyclecountitemapp

import (
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountitembus"
)

// QueryParams represents the set of possible query strings.
type QueryParams struct {
	Page       string
	Rows       string
	OrderBy    string
	ID         string
	SessionID  string
	ProductID  string
	LocationID string
	Status     string
	CountedBy  string
}

// =============================================================================

// CycleCountItem represents information about an individual item.
type CycleCountItem struct {
	ID              string `json:"id"`
	SessionID       string `json:"sessionID"`
	ProductID       string `json:"productID"`
	LocationID      string `json:"locationID"`
	SystemQuantity  string `json:"systemQuantity"`
	CountedQuantity string `json:"countedQuantity"`
	Variance        string `json:"variance"`
	Status          string `json:"status"`
	CountedBy       string `json:"countedBy"`
	CountedDate     string `json:"countedDate"`
	CreatedDate     string `json:"createdDate"`
	UpdatedDate     string `json:"updatedDate"`
}

// Encode implements the web.Encoder interface.
func (app CycleCountItem) Encode() ([]byte, string, error) {
	data, err := errs.MarshalIndentJSON(app)
	return data, "application/json", err
}

func ToAppCycleCountItem(bus cyclecountitembus.CycleCountItem) CycleCountItem {
	countedQuantity := ""
	if bus.CountedQuantity != nil {
		countedQuantity = strconv.Itoa(*bus.CountedQuantity)
	}

	variance := ""
	if bus.Variance != nil {
		variance = strconv.Itoa(*bus.Variance)
	}

	countedBy := ""
	if bus.CountedBy != (uuid.UUID{}) {
		countedBy = bus.CountedBy.String()
	}

	countedDate := ""
	if !bus.CountedDate.IsZero() {
		countedDate = bus.CountedDate.Format(time.RFC3339)
	}

	return CycleCountItem{
		ID:              bus.ID.String(),
		SessionID:       bus.SessionID.String(),
		ProductID:       bus.ProductID.String(),
		LocationID:      bus.LocationID.String(),
		SystemQuantity:  strconv.Itoa(bus.SystemQuantity),
		CountedQuantity: countedQuantity,
		Variance:        variance,
		Status:          bus.Status.String(),
		CountedBy:       countedBy,
		CountedDate:     countedDate,
		CreatedDate:     bus.CreatedDate.Format(time.RFC3339),
		UpdatedDate:     bus.UpdatedDate.Format(time.RFC3339),
	}
}

func ToAppCycleCountItems(items []cyclecountitembus.CycleCountItem) []CycleCountItem {
	app := make([]CycleCountItem, len(items))
	for i, item := range items {
		app[i] = ToAppCycleCountItem(item)
	}
	return app
}

// =============================================================================

// NewCycleCountItem defines the data needed to create a new item.
type NewCycleCountItem struct {
	SessionID      string `json:"sessionID" validate:"required,min=36,max=36"`
	ProductID      string `json:"productID" validate:"required,min=36,max=36"`
	LocationID     string `json:"locationID" validate:"required,min=36,max=36"`
	SystemQuantity string `json:"systemQuantity" validate:"required"`
}

// Decode implements the web.Decoder interface.
func (app *NewCycleCountItem) Decode(data []byte) error {
	return errs.UnmarshalJSON(data, app)
}

// Validate checks the data in the model is considered clean.
func (app NewCycleCountItem) Validate() error {
	if err := errs.Check(app); err != nil {
		return err
	}
	return nil
}

func toBusNewCycleCountItem(app NewCycleCountItem) (cyclecountitembus.NewCycleCountItem, error) {
	sessionID, err := uuid.Parse(app.SessionID)
	if err != nil {
		return cyclecountitembus.NewCycleCountItem{}, fmt.Errorf("parsing sessionID: %w", err)
	}

	productID, err := uuid.Parse(app.ProductID)
	if err != nil {
		return cyclecountitembus.NewCycleCountItem{}, fmt.Errorf("parsing productID: %w", err)
	}

	locationID, err := uuid.Parse(app.LocationID)
	if err != nil {
		return cyclecountitembus.NewCycleCountItem{}, fmt.Errorf("parsing locationID: %w", err)
	}

	sysQty, err := strconv.Atoi(app.SystemQuantity)
	if err != nil {
		return cyclecountitembus.NewCycleCountItem{}, fmt.Errorf("parsing systemQuantity: %w", err)
	}

	return cyclecountitembus.NewCycleCountItem{
		SessionID:      sessionID,
		ProductID:      productID,
		LocationID:     locationID,
		SystemQuantity: sysQty,
	}, nil
}

// =============================================================================

// UpdateCycleCountItem defines the data needed to update an item.
type UpdateCycleCountItem struct {
	CountedQuantity *string `json:"countedQuantity" validate:"omitempty"`
	Status          *string `json:"status" validate:"omitempty"`
}

// Decode implements the web.Decoder interface.
func (app *UpdateCycleCountItem) Decode(data []byte) error {
	return errs.UnmarshalJSON(data, app)
}

// Validate checks the data in the model is considered clean.
func (app UpdateCycleCountItem) Validate() error {
	if err := errs.Check(app); err != nil {
		return err
	}
	return nil
}

func toBusUpdateCycleCountItem(app UpdateCycleCountItem, userID uuid.UUID) (cyclecountitembus.UpdateCycleCountItem, error) {
	var uci cyclecountitembus.UpdateCycleCountItem

	if app.CountedQuantity != nil {
		qty, err := strconv.Atoi(*app.CountedQuantity)
		if err != nil {
			return cyclecountitembus.UpdateCycleCountItem{}, fmt.Errorf("parsing countedQuantity: %w", err)
		}
		uci.CountedQuantity = &qty

		now := time.Now()
		uci.CountedBy = &userID
		uci.CountedDate = &now
	}

	if app.Status != nil {
		status, err := cyclecountitembus.ParseStatus(*app.Status)
		if err != nil {
			return cyclecountitembus.UpdateCycleCountItem{}, err
		}
		uci.Status = &status
	}

	return uci, nil
}

// =============================================================================

func parseFilter(qp QueryParams) (cyclecountitembus.QueryFilter, error) {
	var filter cyclecountitembus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return cyclecountitembus.QueryFilter{}, err
		}
		filter.ID = &id
	}

	if qp.SessionID != "" {
		id, err := uuid.Parse(qp.SessionID)
		if err != nil {
			return cyclecountitembus.QueryFilter{}, err
		}
		filter.SessionID = &id
	}

	if qp.ProductID != "" {
		id, err := uuid.Parse(qp.ProductID)
		if err != nil {
			return cyclecountitembus.QueryFilter{}, err
		}
		filter.ProductID = &id
	}

	if qp.LocationID != "" {
		id, err := uuid.Parse(qp.LocationID)
		if err != nil {
			return cyclecountitembus.QueryFilter{}, err
		}
		filter.LocationID = &id
	}

	if qp.Status != "" {
		status, err := cyclecountitembus.ParseStatus(qp.Status)
		if err != nil {
			return cyclecountitembus.QueryFilter{}, err
		}
		filter.Status = &status
	}

	if qp.CountedBy != "" {
		id, err := uuid.Parse(qp.CountedBy)
		if err != nil {
			return cyclecountitembus.QueryFilter{}, err
		}
		filter.CountedBy = &id
	}

	return filter, nil
}

func parsePage(qp QueryParams) (int, int) {
	pg := 1
	rows := 10

	if qp.Page != "" {
		if v, err := strconv.Atoi(qp.Page); err == nil {
			pg = v
		}
	}

	if qp.Rows != "" {
		if v, err := strconv.Atoi(qp.Rows); err == nil {
			rows = v
		}
	}

	return pg, rows
}
```

Note: This file needs a `"fmt"` import — ensure it's present.

- [ ] **Step 2: Create cyclecountitemapp.go**

```go
package cyclecountitemapp

import (
	"context"
	"errors"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaez/ichor/app/sdk/mid"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountitembus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer API functions for cycle count items.
type App struct {
	cycleCountItemBus *cyclecountitembus.Business
}

// NewApp constructs a cycle count item app API for use.
func NewApp(itemBus *cyclecountitembus.Business) *App {
	return &App{
		cycleCountItemBus: itemBus,
	}
}

// Create adds a new cycle count item to the system.
func (a *App) Create(ctx context.Context, app NewCycleCountItem) (CycleCountItem, error) {
	nci, err := toBusNewCycleCountItem(app)
	if err != nil {
		return CycleCountItem{}, errs.New(errs.FailedPrecondition, err)
	}

	item, err := a.cycleCountItemBus.Create(ctx, nci)
	if err != nil {
		if errors.Is(err, cyclecountitembus.ErrUniqueEntry) {
			return CycleCountItem{}, errs.New(errs.Aborted, cyclecountitembus.ErrUniqueEntry)
		}
		if errors.Is(err, cyclecountitembus.ErrForeignKeyViolation) {
			return CycleCountItem{}, errs.New(errs.Aborted, cyclecountitembus.ErrForeignKeyViolation)
		}
		return CycleCountItem{}, errs.Newf(errs.Internal, "create: item[%+v]: %s", app, err)
	}

	return ToAppCycleCountItem(item), nil
}

// Update modifies data about a cycle count item. When a counted_quantity is
// provided, the app auto-injects counted_by and counted_date from the JWT user.
func (a *App) Update(ctx context.Context, item cyclecountitembus.CycleCountItem, app UpdateCycleCountItem) (CycleCountItem, error) {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return CycleCountItem{}, errs.New(errs.Unauthenticated, err)
	}

	uci, err := toBusUpdateCycleCountItem(app, userID)
	if err != nil {
		return CycleCountItem{}, errs.New(errs.FailedPrecondition, err)
	}

	updated, err := a.cycleCountItemBus.Update(ctx, item, uci)
	if err != nil {
		return CycleCountItem{}, errs.Newf(errs.Internal, "update: itemID[%s]: %s", item.ID, err)
	}

	return ToAppCycleCountItem(updated), nil
}

// Delete removes the specified cycle count item.
func (a *App) Delete(ctx context.Context, item cyclecountitembus.CycleCountItem) error {
	if err := a.cycleCountItemBus.Delete(ctx, item); err != nil {
		return errs.Newf(errs.Internal, "delete: itemID[%s]: %s", item.ID, err)
	}

	return nil
}

// Query returns a list of items with paging.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[CycleCountItem], error) {
	pg, rows := parsePage(qp)
	p, err := page.Parse(pg, rows)
	if err != nil {
		return query.Result[CycleCountItem]{}, errs.New(errs.FailedPrecondition, err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[CycleCountItem]{}, errs.New(errs.FailedPrecondition, err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, cyclecountitembus.DefaultOrderBy)
	if err != nil {
		return query.Result[CycleCountItem]{}, errs.New(errs.FailedPrecondition, err)
	}

	items, err := a.cycleCountItemBus.Query(ctx, filter, orderBy, p)
	if err != nil {
		return query.Result[CycleCountItem]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.cycleCountItemBus.Count(ctx, filter)
	if err != nil {
		return query.Result[CycleCountItem]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppCycleCountItems(items), total, p), nil
}

// QueryByID finds the item by the specified ID.
func (a *App) QueryByID(ctx context.Context, item cyclecountitembus.CycleCountItem) (CycleCountItem, error) {
	return ToAppCycleCountItem(item), nil
}

// =============================================================================

var orderByFields = map[string]string{
	"id":             cyclecountitembus.OrderByID,
	"sessionID":      cyclecountitembus.OrderBySessionID,
	"productID":      cyclecountitembus.OrderByProductID,
	"locationID":     cyclecountitembus.OrderByLocationID,
	"systemQuantity": cyclecountitembus.OrderBySystemQuantity,
	"status":         cyclecountitembus.OrderByStatus,
	"createdDate":    cyclecountitembus.OrderByCreatedDate,
}
```

- [ ] **Step 3: Verify the package compiles**

Run: `go build ./app/domain/inventory/cyclecountitemapp/...`
Expected: Clean build.

- [ ] **Step 4: Commit**

```bash
git add app/domain/inventory/cyclecountitemapp/
git commit -m "feat(cyclecountitem): add app layer for cycle count items"
```

---

## Task 9: API Layer — Both Domains

**Files:**
- Create: `api/domain/http/inventory/cyclecountsessionapi/cyclecountsessionapi.go`
- Create: `api/domain/http/inventory/cyclecountsessionapi/routes.go`
- Create: `api/domain/http/inventory/cyclecountsessionapi/filter.go`
- Create: `api/domain/http/inventory/cyclecountitemapi/cyclecountitemapi.go`
- Create: `api/domain/http/inventory/cyclecountitemapi/routes.go`
- Create: `api/domain/http/inventory/cyclecountitemapi/filter.go`

**Reference:** `api/domain/http/inventory/picktaskapi/` — same Config struct pattern, RouteTable constant, handler functions.

- [ ] **Step 1: Create cyclecountsessionapi/routes.go**

```go
package cyclecountsessionapi

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/inventory/cyclecountsessionapp"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountsessionbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryadjustmentbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/permissionsbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
	"github.com/timmaaez/ichor/sdk/auth/authclient"
)

// RouteTable is the table name used for RBAC permission lookups.
const RouteTable = "inventory.cycle_count_sessions"

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log                  *logger.Logger
	CycleCountSessionBus *cyclecountsessionbus.Business
	CycleCountItemBus    *cyclecountitembus.Business
	InvAdjustmentBus     *inventoryadjustmentbus.Business
	DB                   *sqlx.DB
	AuthClient           *authclient.Client
	PermissionsBus       *permissionsbus.Business
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	sessionApp := cyclecountsessionapp.NewApp(cfg.CycleCountSessionBus, cfg.CycleCountItemBus, cfg.InvAdjustmentBus, cfg.DB)

	authen := mid.Authenticate(cfg.AuthClient)
	ruleRead := mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read)
	ruleCreate := mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Create)
	ruleUpdate := mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Update)
	ruleDelete := mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Delete)

	api := newAPI(sessionApp, cfg.CycleCountSessionBus)

	app.HandlerFunc(http.MethodGet, version, "/inventory/cycle-count-sessions", api.query, authen, ruleRead)
	app.HandlerFunc(http.MethodGet, version, "/inventory/cycle-count-sessions/{session_id}", api.queryByID, authen, ruleRead)
	app.HandlerFunc(http.MethodPost, version, "/inventory/cycle-count-sessions", api.create, authen, ruleCreate)
	app.HandlerFunc(http.MethodPut, version, "/inventory/cycle-count-sessions/{session_id}", api.update, authen, ruleUpdate)
	app.HandlerFunc(http.MethodDelete, version, "/inventory/cycle-count-sessions/{session_id}", api.delete, authen, ruleDelete)
}
```

- [ ] **Step 2: Create cyclecountsessionapi/cyclecountsessionapi.go**

```go
package cyclecountsessionapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/inventory/cyclecountsessionapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaez/ichor/business/domain/inventory/cyclecountsessionbus"
	"github.com/timmaaez/ichor/foundation/web"
)

type api struct {
	cycleCountSessionApp *cyclecountsessionapp.App
	cycleCountSessionBus *cyclecountsessionbus.Business
}

func newAPI(app *cyclecountsessionapp.App, bus *cyclecountsessionbus.Business) *api {
	return &api{
		cycleCountSessionApp: app,
		cycleCountSessionBus: bus,
	}
}

func (a *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app cyclecountsessionapp.NewCycleCountSession
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.FailedPrecondition, err)
	}

	session, err := a.cycleCountSessionApp.Create(ctx, app)
	if err != nil {
		return err
	}

	return session
}

func (a *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app cyclecountsessionapp.UpdateCycleCountSession
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.FailedPrecondition, err)
	}

	sessionID, err := uuid.Parse(web.Param(r, "session_id"))
	if err != nil {
		return errs.New(errs.FailedPrecondition, err)
	}

	session, err := a.cycleCountSessionBus.QueryByID(ctx, sessionID)
	if err != nil {
		return errs.New(errs.NotFound, err)
	}

	updated, err := a.cycleCountSessionApp.Update(ctx, session, app)
	if err != nil {
		return err
	}

	return updated
}

func (a *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	sessionID, err := uuid.Parse(web.Param(r, "session_id"))
	if err != nil {
		return errs.New(errs.FailedPrecondition, err)
	}

	session, err := a.cycleCountSessionBus.QueryByID(ctx, sessionID)
	if err != nil {
		return errs.New(errs.NotFound, err)
	}

	if err := a.cycleCountSessionApp.Delete(ctx, session); err != nil {
		return err
	}

	return nil
}

func (a *api) query(ctx context.Context, r *http.Request) web.Encoder {
	qp := parseQueryParams(r)

	result, err := a.cycleCountSessionApp.Query(ctx, qp)
	if err != nil {
		return err
	}

	return result
}

func (a *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	sessionID, err := uuid.Parse(web.Param(r, "session_id"))
	if err != nil {
		return errs.New(errs.FailedPrecondition, err)
	}

	session, err := a.cycleCountSessionBus.QueryByID(ctx, sessionID)
	if err != nil {
		return errs.New(errs.NotFound, err)
	}

	result, err := a.cycleCountSessionApp.QueryByID(ctx, session)
	if err != nil {
		return err
	}

	return result
}
```

- [ ] **Step 3: Create cyclecountsessionapi/filter.go**

```go
package cyclecountsessionapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/inventory/cyclecountsessionapp"
)

func parseQueryParams(r *http.Request) cyclecountsessionapp.QueryParams {
	return cyclecountsessionapp.QueryParams{
		Page:        r.URL.Query().Get("page"),
		Rows:        r.URL.Query().Get("rows"),
		OrderBy:     r.URL.Query().Get("orderBy"),
		ID:          r.URL.Query().Get("id"),
		Name:        r.URL.Query().Get("name"),
		Status:      r.URL.Query().Get("status"),
		CreatedBy:   r.URL.Query().Get("createdBy"),
		CreatedDate: r.URL.Query().Get("createdDate"),
	}
}
```

- [ ] **Step 4: Create cyclecountitemapi/routes.go**

```go
package cyclecountitemapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/inventory/cyclecountitemapp"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountitembus"
	"github.com/timmaaaz/ichor/business/domain/permissions/permissionsbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
	"github.com/timmaaaz/ichor/sdk/auth/authclient"
)

// RouteTable is the table name used for RBAC permission lookups.
const RouteTable = "inventory.cycle_count_items"

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log               *logger.Logger
	CycleCountItemBus *cyclecountitembus.Business
	AuthClient        *authclient.Client
	PermissionsBus    *permissionsbus.Business
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	itemApp := cyclecountitemapp.NewApp(cfg.CycleCountItemBus)

	authen := mid.Authenticate(cfg.AuthClient)
	ruleRead := mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read)
	ruleCreate := mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Create)
	ruleUpdate := mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Update)
	ruleDelete := mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Delete)

	api := newAPI(itemApp, cfg.CycleCountItemBus)

	app.HandlerFunc(http.MethodGet, version, "/inventory/cycle-count-items", api.query, authen, ruleRead)
	app.HandlerFunc(http.MethodGet, version, "/inventory/cycle-count-items/{item_id}", api.queryByID, authen, ruleRead)
	app.HandlerFunc(http.MethodPost, version, "/inventory/cycle-count-items", api.create, authen, ruleCreate)
	app.HandlerFunc(http.MethodPut, version, "/inventory/cycle-count-items/{item_id}", api.update, authen, ruleUpdate)
	app.HandlerFunc(http.MethodDelete, version, "/inventory/cycle-count-items/{item_id}", api.delete, authen, ruleDelete)
}
```

- [ ] **Step 5: Create cyclecountitemapi/cyclecountitemapi.go**

```go
package cyclecountitemapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/inventory/cyclecountitemapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaez/ichor/business/domain/inventory/cyclecountitembus"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	cycleCountItemApp *cyclecountitemapp.App
	cycleCountItemBus *cyclecountitembus.Business
}

func newAPI(app *cyclecountitemapp.App, bus *cyclecountitembus.Business) *api {
	return &api{
		cycleCountItemApp: app,
		cycleCountItemBus: bus,
	}
}

func (a *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app cyclecountitemapp.NewCycleCountItem
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.FailedPrecondition, err)
	}

	item, err := a.cycleCountItemApp.Create(ctx, app)
	if err != nil {
		return err
	}

	return item
}

func (a *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app cyclecountitemapp.UpdateCycleCountItem
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.FailedPrecondition, err)
	}

	itemID, err := uuid.Parse(web.Param(r, "item_id"))
	if err != nil {
		return errs.New(errs.FailedPrecondition, err)
	}

	item, err := a.cycleCountItemBus.QueryByID(ctx, itemID)
	if err != nil {
		return errs.New(errs.NotFound, err)
	}

	updated, err := a.cycleCountItemApp.Update(ctx, item, app)
	if err != nil {
		return err
	}

	return updated
}

func (a *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	itemID, err := uuid.Parse(web.Param(r, "item_id"))
	if err != nil {
		return errs.New(errs.FailedPrecondition, err)
	}

	item, err := a.cycleCountItemBus.QueryByID(ctx, itemID)
	if err != nil {
		return errs.New(errs.NotFound, err)
	}

	if err := a.cycleCountItemApp.Delete(ctx, item); err != nil {
		return err
	}

	return nil
}

func (a *api) query(ctx context.Context, r *http.Request) web.Encoder {
	qp := parseQueryParams(r)

	result, err := a.cycleCountItemApp.Query(ctx, qp)
	if err != nil {
		return err
	}

	return result
}

func (a *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	itemID, err := uuid.Parse(web.Param(r, "item_id"))
	if err != nil {
		return errs.New(errs.FailedPrecondition, err)
	}

	item, err := a.cycleCountItemBus.QueryByID(ctx, itemID)
	if err != nil {
		return errs.New(errs.NotFound, err)
	}

	result, err := a.cycleCountItemApp.QueryByID(ctx, item)
	if err != nil {
		return err
	}

	return result
}
```

- [ ] **Step 6: Create cyclecountitemapi/filter.go**

```go
package cyclecountitemapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/inventory/cyclecountitemapp"
)

func parseQueryParams(r *http.Request) cyclecountitemapp.QueryParams {
	return cyclecountitemapp.QueryParams{
		Page:       r.URL.Query().Get("page"),
		Rows:       r.URL.Query().Get("rows"),
		OrderBy:    r.URL.Query().Get("orderBy"),
		ID:         r.URL.Query().Get("id"),
		SessionID:  r.URL.Query().Get("sessionID"),
		ProductID:  r.URL.Query().Get("productID"),
		LocationID: r.URL.Query().Get("locationID"),
		Status:     r.URL.Query().Get("status"),
		CountedBy:  r.URL.Query().Get("countedBy"),
	}
}
```

- [ ] **Step 7: Verify both API packages compile**

Run: `go build ./api/domain/http/inventory/cyclecountsessionapi/... ./api/domain/http/inventory/cyclecountitemapi/...`
Expected: Clean build.

- [ ] **Step 8: Commit**

```bash
git add api/domain/http/inventory/cyclecountsessionapi/ api/domain/http/inventory/cyclecountitemapi/
git commit -m "feat(cyclecountapi): add HTTP API layer for cycle count sessions and items"
```

---

## Task 10: Wiring — all.go, dbtest.go, apitest/model.go

**Files:**
- Modify: `api/cmd/services/ichor/build/all/all.go`
- Modify: `business/sdk/dbtest/dbtest.go`
- Modify: `api/sdk/http/apitest/model.go`

- [ ] **Step 1: Add bus imports and construction to all.go**

In the imports section, add:
```go
"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountsessionbus"
"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountsessionbus/stores/cyclecountsessiondb"
"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountitembus"
"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountitembus/stores/cyclecountitemdb"
"github.com/timmaaaz/ichor/api/domain/http/inventory/cyclecountsessionapi"
"github.com/timmaaaz/ichor/api/domain/http/inventory/cyclecountitemapi"
```

In the bus construction section (near pickTaskBus), add:
```go
cycleCountSessionBus := cyclecountsessionbus.NewBusiness(cfg.Log, delegate, cyclecountsessiondb.NewStore(cfg.Log, cfg.DB))
cycleCountItemBus := cyclecountitembus.NewBusiness(cfg.Log, delegate, cyclecountitemdb.NewStore(cfg.Log, cfg.DB))
```

In the delegate registration block (inside the Temporal guard), add:
```go
delegateHandler.RegisterDomain(delegate, cyclecountsessionbus.DomainName, cyclecountsessionbus.EntityName)
delegateHandler.RegisterDomain(delegate, cyclecountitembus.DomainName, cyclecountitembus.EntityName)
```

In the route registration section (after picktaskapi.Routes), add:
```go
cyclecountsessionapi.Routes(app, cyclecountsessionapi.Config{
    Log:                  cfg.Log,
    CycleCountSessionBus: cycleCountSessionBus,
    CycleCountItemBus:    cycleCountItemBus,
    InvAdjustmentBus:     inventoryAdjustmentBus,
    DB:                   cfg.DB,
    AuthClient:           cfg.AuthClient,
    PermissionsBus:       permissionsBus,
})

cyclecountitemapi.Routes(app, cyclecountitemapi.Config{
    Log:               cfg.Log,
    CycleCountItemBus: cycleCountItemBus,
    AuthClient:        cfg.AuthClient,
    PermissionsBus:    permissionsBus,
})
```

- [ ] **Step 2: Add to BusDomain in dbtest.go**

In the `BusDomain` struct (after `PickTask`), add:
```go
CycleCountSession *cyclecountsessionbus.Business
CycleCountItem    *cyclecountitembus.Business
```

In the `newBusDomains` function (after pickTaskBus construction), add:
```go
cycleCountSessionBus := cyclecountsessionbus.NewBusiness(log, delegate, cyclecountsessiondb.NewStore(log, db))
cycleCountItemBus := cyclecountitembus.NewBusiness(log, delegate, cyclecountitemdb.NewStore(log, db))
```

In the return struct, add:
```go
CycleCountSession: cycleCountSessionBus,
CycleCountItem:    cycleCountItemBus,
```

Add the required imports:
```go
"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountsessionbus"
"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountsessionbus/stores/cyclecountsessiondb"
"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountitembus"
"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountitembus/stores/cyclecountitemdb"
```

- [ ] **Step 3: Add to SeedData in apitest/model.go**

In the `SeedData` struct (after `PickTasks`), add:
```go
CycleCountSessions []cyclecountsessionapp.CycleCountSession
CycleCountItems    []cyclecountitemapp.CycleCountItem
```

Add the required imports:
```go
"github.com/timmaaaz/ichor/app/domain/inventory/cyclecountsessionapp"
"github.com/timmaaez/ichor/app/domain/inventory/cyclecountitemapp"
```

- [ ] **Step 4: Verify the full build**

Run: `go build ./api/cmd/services/ichor/...`
Expected: Clean build.

- [ ] **Step 5: Commit**

```bash
git add api/cmd/services/ichor/build/all/all.go business/sdk/dbtest/dbtest.go api/sdk/http/apitest/model.go
git commit -m "feat(wiring): wire cycle count sessions and items into service"
```

---

## Task 11: testutil.go for Both Bus Domains

**Files:**
- Create: `business/domain/inventory/cyclecountsessionbus/testutil.go`
- Create: `business/domain/inventory/cyclecountitembus/testutil.go`

- [ ] **Step 1: Create cyclecountsessionbus/testutil.go**

```go
package cyclecountsessionbus

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"
)

// TestNewCycleCountSessions generates n new cycle count sessions for testing.
func TestNewCycleCountSessions(n int, createdByIDs []uuid.UUID) []NewCycleCountSession {
	sessions := make([]NewCycleCountSession, n)

	for i := range n {
		sessions[i] = NewCycleCountSession{
			Name:      fmt.Sprintf("Cycle Count Session %d", i+1),
			CreatedBy: createdByIDs[i%len(createdByIDs)],
		}
	}

	return sessions
}

// TestSeedCycleCountSessions creates n cycle count sessions in the database for testing.
func TestSeedCycleCountSessions(ctx context.Context, n int, createdByIDs []uuid.UUID, api *Business) ([]CycleCountSession, error) {
	newSessions := TestNewCycleCountSessions(n, createdByIDs)

	sessions := make([]CycleCountSession, len(newSessions))
	for i, ncs := range newSessions {
		session, err := api.Create(ctx, ncs)
		if err != nil {
			return nil, err
		}
		sessions[i] = session
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].ID.String() < sessions[j].ID.String()
	})

	return sessions, nil
}
```

- [ ] **Step 2: Create cyclecountitembus/testutil.go**

```go
package cyclecountitembus

import (
	"context"
	"sort"

	"github.com/google/uuid"
)

// TestNewCycleCountItems generates n new cycle count items for testing.
func TestNewCycleCountItems(n int, sessionIDs, productIDs, locationIDs []uuid.UUID) []NewCycleCountItem {
	items := make([]NewCycleCountItem, n)

	for i := range n {
		items[i] = NewCycleCountItem{
			SessionID:      sessionIDs[i%len(sessionIDs)],
			ProductID:      productIDs[i%len(productIDs)],
			LocationID:     locationIDs[i%len(locationIDs)],
			SystemQuantity: (i + 1) * 10,
		}
	}

	return items
}

// TestSeedCycleCountItems creates n cycle count items in the database for testing.
func TestSeedCycleCountItems(ctx context.Context, n int, sessionIDs, productIDs, locationIDs []uuid.UUID, api *Business) ([]CycleCountItem, error) {
	newItems := TestNewCycleCountItems(n, sessionIDs, productIDs, locationIDs)

	items := make([]CycleCountItem, len(newItems))
	for i, nci := range newItems {
		item, err := api.Create(ctx, nci)
		if err != nil {
			return nil, err
		}
		items[i] = item
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].ID.String() < items[j].ID.String()
	})

	return items, nil
}
```

- [ ] **Step 3: Verify compilation**

Run: `go build ./business/domain/inventory/cyclecountsessionbus/... ./business/domain/inventory/cyclecountitembus/...`
Expected: Clean build.

- [ ] **Step 4: Commit**

```bash
git add business/domain/inventory/cyclecountsessionbus/testutil.go business/domain/inventory/cyclecountitembus/testutil.go
git commit -m "feat(cyclecount): add test utilities for seeding cycle count data"
```

---

## Task 12: Session Integration Tests — CRUD

**Files:**
- Create: `api/cmd/services/ichor/tests/inventory/cyclecountsessionapi/cyclecountsession_test.go`
- Create: `api/cmd/services/ichor/tests/inventory/cyclecountsessionapi/seed_test.go`
- Create: `api/cmd/services/ichor/tests/inventory/cyclecountsessionapi/create_test.go`
- Create: `api/cmd/services/ichor/tests/inventory/cyclecountsessionapi/query_test.go`
- Create: `api/cmd/services/ichor/tests/inventory/cyclecountsessionapi/delete_test.go`

**Reference:** `api/cmd/services/ichor/tests/inventory/putawaytaskapi/` for seed chain and test patterns.

- [ ] **Step 1: Create cyclecountsession_test.go (top-level runner)**

```go
package cyclecountsessionapi_test

import (
	"testing"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func Test_CycleCountSession(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_CycleCountSession")

	sd := insertSeedData(test.DB, test.Auth)

	t.Run("query200", query200(test, sd))
	t.Run("create200", create200(test, sd))
	t.Run("create401", create401(test, sd))
	t.Run("delete200", delete200(test, sd))
}
```

- [ ] **Step 2: Create seed_test.go**

```go
package cyclecountsessionapi_test

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/inventory/cyclecountsessionapp"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountsessionbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/sdk/auth/authclient"
)

func insertSeedData(db *sqlx.DB, ath *authclient.Client) apitest.SeedData {
	ctx := context.Background()
	busDomain := dbtest.NewBusDomain(db)

	usrs, err := dbtest.SeedUsers(ctx, 1, busDomain.User, busDomain.UserApprovalStatus)
	if err != nil {
		panic(fmt.Sprintf("seeding users: %s", err))
	}

	admins, err := dbtest.SeedUsers(ctx, 1, busDomain.User, busDomain.UserApprovalStatus)
	if err != nil {
		panic(fmt.Sprintf("seeding admins: %s", err))
	}

	createdByIDs := make([]uuid.UUID, len(admins))
	for i, a := range admins {
		createdByIDs[i] = a.ID
	}

	sessions, err := cyclecountsessionbus.TestSeedCycleCountSessions(ctx, 4, createdByIDs, busDomain.CycleCountSession)
	if err != nil {
		panic(fmt.Sprintf("seeding cycle count sessions: %s", err))
	}

	// Transition first 2 sessions to in_progress for update/complete tests.
	inProgress := cyclecountsessionbus.Statuses.InProgress
	for i := 0; i < 2 && i < len(sessions); i++ {
		sessions[i], err = busDomain.CycleCountSession.Update(ctx, sessions[i], cyclecountsessionbus.UpdateCycleCountSession{
			Status: &inProgress,
		})
		if err != nil {
			panic(fmt.Sprintf("transitioning session to in_progress: %s", err))
		}
	}

	appSessions := make([]cyclecountsessionapp.CycleCountSession, len(sessions))
	for i, s := range sessions {
		appSessions[i] = cyclecountsessionapp.ToAppCycleCountSession(s)
	}

	// Set up roles and table access for RBAC.
	roles, err := dbtest.SeedRoles(ctx, 2, busDomain.Role)
	if err != nil {
		panic(fmt.Sprintf("seeding roles: %s", err))
	}

	// User gets read-only, admin gets all.
	dbtest.SeedUserRoles(ctx, usrs[0].ID, []uuid.UUID{roles[0].ID}, busDomain.UserRole)
	dbtest.SeedUserRoles(ctx, admins[0].ID, []uuid.UUID{roles[1].ID}, busDomain.UserRole)

	dbtest.SeedTableAccess(ctx, roles[0].ID, "inventory.cycle_count_sessions", true, false, false, false, busDomain.TableAccess)
	dbtest.SeedTableAccess(ctx, roles[1].ID, "inventory.cycle_count_sessions", true, true, true, true, busDomain.TableAccess)

	sd := apitest.SeedData{
		Admins:             apitest.ToAppUsers(admins),
		Users:              apitest.ToAppUsers(usrs),
		CycleCountSessions: appSessions,
	}

	return sd
}
```

**Important note to implementer:** The exact seed helper signatures may differ from what's shown. Check `dbtest.SeedUsers`, `dbtest.SeedRoles`, `dbtest.SeedUserRoles`, and `dbtest.SeedTableAccess` signatures in the actual codebase. The patterns above follow `putawaytaskapi/seed_test.go` — adapt the user/role/table_access seeding to match the current helpers. The critical parts are:
- Seed 4 sessions (2 in `draft`, 2 transitioned to `in_progress`)
- Set up RBAC with one read-only user and one admin user

- [ ] **Step 3: Create create_test.go**

```go
package cyclecountsessionapi_test

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/inventory/cyclecountsessionapp"
)

func create200(test *apitest.Test, sd apitest.SeedData) func(t *testing.T) {
	return func(t *testing.T) {
		table := []apitest.Table{
			{
				Name:       "basic",
				URL:        "/v1/inventory/cycle-count-sessions",
				Token:      sd.Admins[0].Token,
				Method:     http.MethodPost,
				StatusCode: http.StatusOK,
				Input: &cyclecountsessionapp.NewCycleCountSession{
					Name: "Test Session Create",
				},
				GotResp: &cyclecountsessionapp.CycleCountSession{},
				ExpResp: &cyclecountsessionapp.CycleCountSession{
					Name:   "Test Session Create",
					Status: "draft",
				},
				CmpFunc: func(got any, exp any) string {
					gotResp := got.(*cyclecountsessionapp.CycleCountSession)
					expResp := exp.(*cyclecountsessionapp.CycleCountSession)

					expResp.ID = gotResp.ID
					expResp.CreatedBy = gotResp.CreatedBy
					expResp.CreatedDate = gotResp.CreatedDate
					expResp.UpdatedDate = gotResp.UpdatedDate

					return apitest.CmpJSON(gotResp, expResp)
				},
			},
		}

		test.Run(t, table, "create200")
	}
}

func create401(test *apitest.Test, sd apitest.SeedData) func(t *testing.T) {
	return func(t *testing.T) {
		table := []apitest.Table{
			{
				Name:       "no-auth",
				URL:        "/v1/inventory/cycle-count-sessions",
				Token:      "",
				Method:     http.MethodPost,
				StatusCode: http.StatusUnauthorized,
				Input: &cyclecountsessionapp.NewCycleCountSession{
					Name: "Should Fail",
				},
			},
		}

		test.Run(t, table, "create401")
	}
}
```

- [ ] **Step 4: Create query_test.go**

```go
package cyclecountsessionapi_test

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/inventory/cyclecountsessionapp"
)

func query200(test *apitest.Test, sd apitest.SeedData) func(t *testing.T) {
	return func(t *testing.T) {
		table := []apitest.Table{
			{
				Name:       "basic",
				URL:        "/v1/inventory/cycle-count-sessions?page=1&rows=10",
				Token:      sd.Users[0].Token,
				Method:     http.MethodGet,
				StatusCode: http.StatusOK,
				GotResp:    &[]cyclecountsessionapp.CycleCountSession{},
				ExpResp:    &[]cyclecountsessionapp.CycleCountSession{},
				CmpFunc: func(got any, exp any) string {
					gotResp := got.(*[]cyclecountsessionapp.CycleCountSession)
					if len(*gotResp) == 0 {
						return "expected sessions, got none"
					}
					return ""
				},
			},
		}

		test.Run(t, table, "query200")
	}
}
```

- [ ] **Step 5: Create delete_test.go**

```go
package cyclecountsessionapi_test

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func delete200(test *apitest.Test, sd apitest.SeedData) func(t *testing.T) {
	return func(t *testing.T) {
		// Delete the last session (draft status, safe to delete).
		url := "/v1/inventory/cycle-count-sessions/" + sd.CycleCountSessions[len(sd.CycleCountSessions)-1].ID

		table := []apitest.Table{
			{
				Name:       "basic",
				URL:        url,
				Token:      sd.Admins[0].Token,
				Method:     http.MethodDelete,
				StatusCode: http.StatusNoContent,
			},
		}

		test.Run(t, table, "delete200")
	}
}
```

- [ ] **Step 6: Verify test compilation**

Run: `go vet ./api/cmd/services/ichor/tests/inventory/cyclecountsessionapi/...`
Expected: Compiles (tests won't run without DB, but should compile).

- [ ] **Step 7: Commit**

```bash
git add api/cmd/services/ichor/tests/inventory/cyclecountsessionapi/
git commit -m "test(cyclecountsession): add CRUD integration tests"
```

---

## Task 13: Item Integration Tests — CRUD

**Files:**
- Create: `api/cmd/services/ichor/tests/inventory/cyclecountitemapi/cyclecountitem_test.go`
- Create: `api/cmd/services/ichor/tests/inventory/cyclecountitemapi/seed_test.go`
- Create: `api/cmd/services/ichor/tests/inventory/cyclecountitemapi/create_test.go`
- Create: `api/cmd/services/ichor/tests/inventory/cyclecountitemapi/update_test.go`
- Create: `api/cmd/services/ichor/tests/inventory/cyclecountitemapi/query_test.go`
- Create: `api/cmd/services/ichor/tests/inventory/cyclecountitemapi/delete_test.go`

- [ ] **Step 1: Create cyclecountitem_test.go**

```go
package cyclecountitemapi_test

import (
	"testing"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func Test_CycleCountItem(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_CycleCountItem")

	sd := insertSeedData(test.DB, test.Auth)

	t.Run("query200", query200(test, sd))
	t.Run("create200", create200(test, sd))
	t.Run("create401", create401(test, sd))
	t.Run("update200", update200(test, sd))
	t.Run("delete200", delete200(test, sd))
}
```

- [ ] **Step 2: Create seed_test.go**

The seed must build the full dependency chain: users → regions → cities → streets → warehouses → zones → inventory locations → products → cycle count sessions → cycle count items.

```go
package cyclecountitemapi_test

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/inventory/cyclecountitemapp"
	"github.com/timmaaez/ichor/app/domain/inventory/cyclecountsessionapp"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountsessionbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/sdk/auth/authclient"
)

func insertSeedData(db *sqlx.DB, ath *authclient.Client) apitest.SeedData {
	ctx := context.Background()
	busDomain := dbtest.NewBusDomain(db)

	// Users
	usrs, err := dbtest.SeedUsers(ctx, 1, busDomain.User, busDomain.UserApprovalStatus)
	if err != nil {
		panic(fmt.Sprintf("seeding users: %s", err))
	}
	admins, err := dbtest.SeedUsers(ctx, 1, busDomain.User, busDomain.UserApprovalStatus)
	if err != nil {
		panic(fmt.Sprintf("seeding admins: %s", err))
	}

	// Geography chain for inventory locations
	regions, err := dbtest.SeedRegions(ctx, busDomain.Region)
	if err != nil {
		panic(fmt.Sprintf("seeding regions: %s", err))
	}
	cities, err := dbtest.SeedCities(ctx, 1, regions, busDomain.City)
	if err != nil {
		panic(fmt.Sprintf("seeding cities: %s", err))
	}
	streets, err := dbtest.SeedStreets(ctx, 1, cities, busDomain.Street)
	if err != nil {
		panic(fmt.Sprintf("seeding streets: %s", err))
	}
	warehouses, err := dbtest.SeedWarehouses(ctx, 1, streets, busDomain.Warehouse)
	if err != nil {
		panic(fmt.Sprintf("seeding warehouses: %s", err))
	}

	warehouseIDs := make([]uuid.UUID, len(warehouses))
	for i, w := range warehouses {
		warehouseIDs[i] = w.ID
	}

	zones, err := dbtest.SeedZones(ctx, 1, warehouseIDs, busDomain.Zone)
	if err != nil {
		panic(fmt.Sprintf("seeding zones: %s", err))
	}

	zoneIDs := make([]uuid.UUID, len(zones))
	for i, z := range zones {
		zoneIDs[i] = z.ID
	}

	locations, err := dbtest.SeedInventoryLocations(ctx, 3, zoneIDs, busDomain.InventoryLocation)
	if err != nil {
		panic(fmt.Sprintf("seeding inventory locations: %s", err))
	}

	locationIDs := make([]uuid.UUID, len(locations))
	for i, l := range locations {
		locationIDs[i] = l.ID
	}

	// Products
	products, err := dbtest.SeedProducts(ctx, 3, busDomain)
	if err != nil {
		panic(fmt.Sprintf("seeding products: %s", err))
	}

	productIDs := make([]uuid.UUID, len(products))
	for i, p := range products {
		productIDs[i] = p.ID
	}

	// Sessions
	createdByIDs := []uuid.UUID{admins[0].ID}
	sessions, err := cyclecountsessionbus.TestSeedCycleCountSessions(ctx, 2, createdByIDs, busDomain.CycleCountSession)
	if err != nil {
		panic(fmt.Sprintf("seeding sessions: %s", err))
	}

	// Transition sessions to in_progress.
	inProgress := cyclecountsessionbus.Statuses.InProgress
	for i := range sessions {
		sessions[i], err = busDomain.CycleCountSession.Update(ctx, sessions[i], cyclecountsessionbus.UpdateCycleCountSession{
			Status: &inProgress,
		})
		if err != nil {
			panic(fmt.Sprintf("transitioning session: %s", err))
		}
	}

	sessionIDs := make([]uuid.UUID, len(sessions))
	for i, s := range sessions {
		sessionIDs[i] = s.ID
	}

	// Items (4 items across sessions)
	items, err := cyclecountitembus.TestSeedCycleCountItems(ctx, 4, sessionIDs, productIDs, locationIDs, busDomain.CycleCountItem)
	if err != nil {
		panic(fmt.Sprintf("seeding items: %s", err))
	}

	appItems := make([]cyclecountitemapp.CycleCountItem, len(items))
	for i, item := range items {
		appItems[i] = cyclecountitemapp.ToAppCycleCountItem(item)
	}

	appSessions := make([]cyclecountsessionapp.CycleCountSession, len(sessions))
	for i, s := range sessions {
		appSessions[i] = cyclecountsessionapp.ToAppCycleCountSession(s)
	}

	// RBAC
	roles, err := dbtest.SeedRoles(ctx, 2, busDomain.Role)
	if err != nil {
		panic(fmt.Sprintf("seeding roles: %s", err))
	}

	dbtest.SeedUserRoles(ctx, usrs[0].ID, []uuid.UUID{roles[0].ID}, busDomain.UserRole)
	dbtest.SeedUserRoles(ctx, admins[0].ID, []uuid.UUID{roles[1].ID}, busDomain.UserRole)

	dbtest.SeedTableAccess(ctx, roles[0].ID, "inventory.cycle_count_items", true, false, false, false, busDomain.TableAccess)
	dbtest.SeedTableAccess(ctx, roles[1].ID, "inventory.cycle_count_items", true, true, true, true, busDomain.TableAccess)

	return apitest.SeedData{
		Admins:             apitest.ToAppUsers(admins),
		Users:              apitest.ToAppUsers(usrs),
		CycleCountSessions: appSessions,
		CycleCountItems:    appItems,
	}
}
```

**Important note to implementer:** The seed helper signatures (`SeedRegions`, `SeedCities`, `SeedProducts`, etc.) must match the actual helpers in `dbtest`. The pattern above follows `picktaskapi/seed_test.go`. Adapt argument lists and return types as needed.

- [ ] **Step 3: Create create_test.go, update_test.go, query_test.go, delete_test.go**

Follow the same table-driven patterns as `putawaytaskapi` tests. Key test cases:

**create_test.go:**
- `create200`: create item with valid session/product/location IDs, assert status=pending
- `create401`: create with no auth token → 401

**update_test.go:**
- `update200`: update `counted_quantity` → assert `counted_by` and `counted_date` are auto-set, `variance` is computed

**query_test.go:**
- `query200`: basic paginated query returns seeded items

**delete_test.go:**
- `delete200`: delete last seeded item → 204

Each test file follows the same pattern as Task 12's test files, adapted for `cyclecountitemapp` types.

- [ ] **Step 4: Verify test compilation**

Run: `go vet ./api/cmd/services/ichor/tests/inventory/cyclecountitemapi/...`
Expected: Compiles.

- [ ] **Step 5: Commit**

```bash
git add api/cmd/services/ichor/tests/inventory/cyclecountitemapi/
git commit -m "test(cyclecountitem): add CRUD integration tests"
```

---

## Task 14: Complete Flow Integration Test

**Files:**
- Create: `api/cmd/services/ichor/tests/inventory/cyclecountsessionapi/update_test.go`

This is the most important integration test — it verifies the atomic complete flow end-to-end.

- [ ] **Step 1: Create update_test.go with complete flow test**

```go
package cyclecountsessionapi_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/inventory/cyclecountsessionapp"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryadjustmentbus"
	"github.com/timmaaez/ichor/business/sdk/dbtest"
	"github.com/timmaaez/ichor/business/sdk/page"
)

func update200(test *apitest.Test, sd apitest.SeedData) func(t *testing.T) {
	return func(t *testing.T) {
		table := []apitest.Table{
			{
				Name:       "draft-to-in_progress",
				URL:        "/v1/inventory/cycle-count-sessions/" + sd.CycleCountSessions[2].ID,
				Token:      sd.Admins[0].Token,
				Method:     http.MethodPut,
				StatusCode: http.StatusOK,
				Input: &cyclecountsessionapp.UpdateCycleCountSession{
					Status: dbtest.StringPointer("in_progress"),
				},
				GotResp: &cyclecountsessionapp.CycleCountSession{},
				ExpResp: &cyclecountsessionapp.CycleCountSession{
					Status: "in_progress",
				},
				CmpFunc: func(got any, exp any) string {
					gotResp := got.(*cyclecountsessionapp.CycleCountSession)
					expResp := exp.(*cyclecountsessionapp.CycleCountSession)

					expResp.ID = gotResp.ID
					expResp.Name = gotResp.Name
					expResp.CreatedBy = gotResp.CreatedBy
					expResp.CreatedDate = gotResp.CreatedDate
					expResp.UpdatedDate = gotResp.UpdatedDate

					return apitest.CmpJSON(gotResp, expResp)
				},
			},
		}

		test.Run(t, table, "update200")
	}
}

// TestUpdate200Complete is a standalone sequential test that verifies the full
// complete flow: create session → add items → count items → approve variances
// → complete session → verify adjustment records were created.
func TestUpdate200Complete(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "TestUpdate200Complete")
	busDomain := dbtest.NewBusDomain(test.DB)
	ctx := context.Background()

	// 1. Seed dependencies (users, locations, products).
	sd := insertCompleteFlowSeedData(test.DB, test.Auth, busDomain)

	// 2. Create a session.
	var createdSession cyclecountsessionapp.CycleCountSession
	test.DoRequest(t, http.MethodPost, "/v1/inventory/cycle-count-sessions", sd.Admins[0].Token,
		&cyclecountsessionapp.NewCycleCountSession{Name: "Complete Flow Test"}, &createdSession, http.StatusOK)

	// 3. Transition to in_progress.
	var inProgressSession cyclecountsessionapp.CycleCountSession
	test.DoRequest(t, http.MethodPut, "/v1/inventory/cycle-count-sessions/"+createdSession.ID, sd.Admins[0].Token,
		&cyclecountsessionapp.UpdateCycleCountSession{Status: dbtest.StringPointer("in_progress")}, &inProgressSession, http.StatusOK)

	// 4. Add items with known system quantities.
	// Item 1: system=100, will count 95 (variance=-5, approved).
	// Item 2: system=50, will count 50 (variance=0, approved — no adjustment expected).
	// Item 3: system=75, will count 80 (variance=+5, rejected — no adjustment expected).

	// ... create items via POST, then update counted_quantity via PUT,
	// then update status to variance_approved / variance_rejected ...

	// 5. Complete the session.
	var completedSession cyclecountsessionapp.CycleCountSession
	test.DoRequest(t, http.MethodPut, "/v1/inventory/cycle-count-sessions/"+createdSession.ID, sd.Admins[0].Token,
		&cyclecountsessionapp.UpdateCycleCountSession{Status: dbtest.StringPointer("completed")}, &completedSession, http.StatusOK)

	if completedSession.Status != "completed" {
		t.Fatalf("expected status completed, got %s", completedSession.Status)
	}

	if completedSession.CompletedDate == "" {
		t.Fatal("expected completed_date to be set")
	}

	// 6. Verify adjustment records were created.
	// Query inventoryadjustmentbus directly for cycle_count reason code.
	adjustments, err := busDomain.InventoryAdjustment.Query(ctx, inventoryadjustmentbus.QueryFilter{
		// Filter by reason code
	}, inventoryadjustmentbus.DefaultOrderBy, page.MustParse("1", "100"))
	if err != nil {
		t.Fatalf("querying adjustments: %s", err)
	}

	// Only item 1 should generate an adjustment (variance=-5, approved).
	// Item 2 has 0 variance (skipped), item 3 was rejected (skipped).
	cycleCountAdj := 0
	for _, adj := range adjustments {
		if adj.ReasonCode == inventoryadjustmentbus.ReasonCodeCycleCount {
			cycleCountAdj++
			if adj.QuantityChange != -5 {
				t.Errorf("expected adjustment quantity -5, got %d", adj.QuantityChange)
			}
		}
	}

	if cycleCountAdj != 1 {
		t.Errorf("expected 1 cycle_count adjustment, got %d", cycleCountAdj)
	}
}

// TestUpdate400TerminalState verifies that completed/cancelled sessions cannot
// be further modified.
func TestUpdate400TerminalState(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "TestUpdate400TerminalState")

	sd := insertSeedData(test.DB, test.Auth)

	// The first session is in_progress. Complete it.
	var completedSession cyclecountsessionapp.CycleCountSession
	test.DoRequest(t, http.MethodPut, "/v1/inventory/cycle-count-sessions/"+sd.CycleCountSessions[0].ID, sd.Admins[0].Token,
		&cyclecountsessionapp.UpdateCycleCountSession{Status: dbtest.StringPointer("completed")}, &completedSession, http.StatusOK)

	// Try to update the completed session.
	test.DoRequest(t, http.MethodPut, "/v1/inventory/cycle-count-sessions/"+sd.CycleCountSessions[0].ID, sd.Admins[0].Token,
		&cyclecountsessionapp.UpdateCycleCountSession{Name: dbtest.StringPointer("Should Fail")}, nil, http.StatusBadRequest)
}

// TestUpdate400DraftCannotComplete verifies that a draft session cannot be
// directly completed — it must first be transitioned to in_progress.
func TestUpdate400DraftCannotComplete(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "TestUpdate400DraftCannotComplete")

	sd := insertSeedData(test.DB, test.Auth)

	// Session[2] is in draft status. Try to complete directly.
	test.DoRequest(t, http.MethodPut, "/v1/inventory/cycle-count-sessions/"+sd.CycleCountSessions[2].ID, sd.Admins[0].Token,
		&cyclecountsessionapp.UpdateCycleCountSession{Status: dbtest.StringPointer("completed")}, nil, http.StatusBadRequest)
}
```

**Important note to implementer:** The `insertCompleteFlowSeedData` helper and `test.DoRequest` patterns shown above are illustrative. The actual helpers available in `apitest` may differ. Check how `putawaytaskapi/update_test.go:TestUpdate200Complete` performs sequential HTTP requests and direct bus queries. The key verification points are:
1. Session transitions draft → in_progress → completed
2. `completed_date` is set
3. Only items with `variance_approved` status AND non-zero variance generate adjustments
4. Adjustment records have `reason_code = "cycle_count"` and correct `quantity_change`
5. Terminal state guard prevents re-modification

- [ ] **Step 2: Run the complete test suite**

Run: `go test ./api/cmd/services/ichor/tests/inventory/cyclecountsessionapi/... -v -count=1`
Expected: All tests PASS.

- [ ] **Step 3: Run the item test suite**

Run: `go test ./api/cmd/services/ichor/tests/inventory/cyclecountitemapi/... -v -count=1`
Expected: All tests PASS.

- [ ] **Step 4: Commit**

```bash
git add api/cmd/services/ichor/tests/inventory/cyclecountsessionapi/update_test.go
git commit -m "test(cyclecountsession): add complete flow and terminal state integration tests"
```

---

## Task 15: Final Build Verification

- [ ] **Step 1: Full build**

Run: `go build ./...`
Expected: Clean build across the entire project.

- [ ] **Step 2: Run both test suites**

Run:
```bash
go test ./api/cmd/services/ichor/tests/inventory/cyclecountsessionapi/... -v -count=1
go test ./api/cmd/services/ichor/tests/inventory/cyclecountitemapi/... -v -count=1
```
Expected: All tests PASS.

- [ ] **Step 3: Commit any remaining fixes and push**

```bash
git push
```

---

## Dependency Graph

```
Task 1 (migration) ─────────────────────────────────────────┐
Task 2 (reason code) ───────────────────────────────────────┐│
                                                             ││
Task 3 (session bus) ──── Task 4 (session db) ──┐           ││
Task 5 (item bus) ─────── Task 6 (item db) ────┐│           ││
                                                ││           ││
Task 11 (testutil) ◄────────────────────────────┤│           ││
                                                ││           ││
Task 7 (session app) ◄─────────────────────────┘│◄──────────┘│
Task 8 (item app) ◄─────────────────────────────┘            │
                                                              │
Task 9 (APIs) ◄─── Task 7 + 8                                │
Task 10 (wiring) ◄─── Task 9                                 │
                                                              │
Task 12 (session tests) ◄─── Task 10 + 11                    │
Task 13 (item tests) ◄─── Task 10 + 11                       │
Task 14 (complete test) ◄─── Task 10 + 11 ◄──────────────────┘
Task 15 (final verify) ◄─── ALL
```

**Parallelizable:**
- Tasks 3+4 and 5+6 can run in parallel
- Tasks 7 and 8 can run in parallel
- Tasks 12 and 13 can run in parallel
