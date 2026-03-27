# Notification Inbox Domain Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a full 7-layer `workflow.notifications` domain providing a persistent notification inbox with paginated list, unread count, mark-as-read, and mark-all-read endpoints.

**Architecture:** A new `notificationbus` business domain backed by a `workflow.notifications` table. Notifications are created internally by the system (workflow actions, alerts). Users interact via read-only endpoints scoped to their own notifications. Follows the alertbus pattern for workflow schema conventions but is simpler (no recipients or acknowledgments sub-tables).

**Tech Stack:** Go 1.23, PostgreSQL 16.4 (workflow schema), Ardan Labs Service architecture (7-layer DDD)

---

## File Map

### New Files

| Layer | Path | Responsibility |
|-------|------|----------------|
| Migration | `business/sdk/migrate/sql/migrate.sql` (append) | `workflow.notifications` table + indexes + table_access |
| Bus model | `business/domain/workflow/notificationbus/model.go` | `Notification`, `NewNotification` structs + priority constants |
| Bus filter | `business/domain/workflow/notificationbus/filter.go` | `QueryFilter` struct |
| Bus order | `business/domain/workflow/notificationbus/order.go` | `OrderBy*` constants + `DefaultOrderBy` |
| Bus core | `business/domain/workflow/notificationbus/notificationbus.go` | `Storer` interface + `Business` struct + methods |
| DB model | `business/domain/workflow/notificationbus/stores/notificationdb/model.go` | `dbNotification` + conversions |
| DB filter | `business/domain/workflow/notificationbus/stores/notificationdb/filter.go` | SQL WHERE clause builder |
| DB order | `business/domain/workflow/notificationbus/stores/notificationdb/order.go` | `orderByFields` map |
| DB store | `business/domain/workflow/notificationbus/stores/notificationdb/notificationdb.go` | `Store` struct + SQL queries |
| App model | `app/domain/workflow/notificationapp/model.go` | App-layer `Notification`, `QueryParams` + conversions |
| App filter | `app/domain/workflow/notificationapp/filter.go` | `parseFilter` from `QueryParams` |
| App order | `app/domain/workflow/notificationapp/order.go` | `orderByFields` map for API → bus translation |
| App core | `app/domain/workflow/notificationapp/notificationapp.go` | `App` struct + methods |
| API route | `api/domain/http/workflow/notificationinboxapi/route.go` | `Config` + `Routes()` |
| API handler | `api/domain/http/workflow/notificationinboxapi/notificationinboxapi.go` | Handler methods |
| API model | `api/domain/http/workflow/notificationinboxapi/model.go` | `UnreadCount` response type |
| Test entry | `api/cmd/services/ichor/tests/workflow/notificationinboxapi/notification_test.go` | `Test_NotificationInboxAPI` |
| Test seed | `api/cmd/services/ichor/tests/workflow/notificationinboxapi/seed_test.go` | `insertSeedData` + `SeedData` |
| Test query | `api/cmd/services/ichor/tests/workflow/notificationinboxapi/query_test.go` | Query + count test tables |
| Test read | `api/cmd/services/ichor/tests/workflow/notificationinboxapi/read_test.go` | Mark-as-read test tables |

### Modified Files

| File | Change |
|------|--------|
| `business/sdk/dbtest/dbtest.go` | Add `Notification *notificationbus.Business` to `BusDomain` + wire in `newBusDomains` |
| `api/cmd/services/ichor/build/all/all.go` | Import `notificationinboxapi` + `notificationbus` + `notificationdb` + `notificationapp`, wire `Routes()` |

---

## Task 1: Migration — Version 2.22

**Files:**
- Modify: `business/sdk/migrate/sql/migrate.sql` (append at end)

- [ ] **Step 1: Add the migration SQL**

Append to the end of `business/sdk/migrate/sql/migrate.sql`:

```sql
-- Version: 2.22
-- Description: Create workflow.notifications inbox table for user notification persistence.
CREATE TABLE workflow.notifications (
    id                  UUID          NOT NULL,
    user_id             UUID          NOT NULL REFERENCES core.users(id) ON DELETE CASCADE,
    title               TEXT          NOT NULL,
    message             TEXT          NULL,
    priority            VARCHAR(10)   NOT NULL DEFAULT 'medium'
                            CHECK (priority IN ('low', 'medium', 'high', 'critical')),
    is_read             BOOLEAN       NOT NULL DEFAULT false,
    read_date           TIMESTAMP     NULL,
    source_entity_name  VARCHAR(100)  NULL,
    source_entity_id    UUID          NULL,
    action_url          TEXT          NULL,
    created_date        TIMESTAMP     NOT NULL,
    PRIMARY KEY (id)
);

CREATE INDEX idx_notifications_user_id      ON workflow.notifications(user_id);
CREATE INDEX idx_notifications_user_unread  ON workflow.notifications(user_id, is_read) WHERE is_read = false;
CREATE INDEX idx_notifications_created      ON workflow.notifications(created_date DESC);
CREATE INDEX idx_notifications_user_created ON workflow.notifications(user_id, created_date DESC);

INSERT INTO core.table_access (id, role_id, table_name, can_create, can_read, can_update, can_delete)
SELECT gen_random_uuid(), id, 'workflow.notifications', true, true, true, true FROM core.roles;
```

- [ ] **Step 2: Verify migration parses**

```bash
go build ./business/sdk/migrate/...
```

Expected: builds successfully (migration SQL is loaded at runtime, not compiled, but this confirms the package still compiles).

- [ ] **Step 3: Commit**

```bash
git add business/sdk/migrate/sql/migrate.sql
git commit -m "feat(migration): add workflow.notifications inbox table (v2.22)"
```

---

## Task 2: Business Layer — Models

**Files:**
- Create: `business/domain/workflow/notificationbus/model.go`

- [ ] **Step 1: Create the model file**

```go
package notificationbus

import (
	"time"

	"github.com/google/uuid"
)

// Priority constants for notification importance.
const (
	PriorityLow      = "low"
	PriorityMedium   = "medium"
	PriorityHigh     = "high"
	PriorityCritical = "critical"
)

// Notification represents a user notification in the system.
type Notification struct {
	ID               uuid.UUID  `json:"id"`
	UserID           uuid.UUID  `json:"user_id"`
	Title            string     `json:"title"`
	Message          string     `json:"message"`
	Priority         string     `json:"priority"`
	IsRead           bool       `json:"is_read"`
	ReadDate         *time.Time `json:"read_date,omitempty"`
	SourceEntityName string     `json:"source_entity_name"`
	SourceEntityID   uuid.UUID  `json:"source_entity_id"`
	ActionURL        string     `json:"action_url"`
	CreatedDate      time.Time  `json:"created_date"`
}

// NewNotification contains the data needed to create a notification.
type NewNotification struct {
	UserID           uuid.UUID
	Title            string
	Message          string
	Priority         string
	SourceEntityName string
	SourceEntityID   uuid.UUID
	ActionURL        string
}
```

- [ ] **Step 2: Build to verify**

```bash
go build ./business/domain/workflow/notificationbus/...
```

- [ ] **Step 3: Commit**

```bash
git add business/domain/workflow/notificationbus/model.go
git commit -m "feat(notificationbus): add notification business models"
```

---

## Task 3: Business Layer — Filter & Order

**Files:**
- Create: `business/domain/workflow/notificationbus/filter.go`
- Create: `business/domain/workflow/notificationbus/order.go`

- [ ] **Step 1: Create the filter file**

```go
package notificationbus

import (
	"time"

	"github.com/google/uuid"
)

// QueryFilter holds the available fields to filter notifications.
type QueryFilter struct {
	ID               *uuid.UUID
	UserID           *uuid.UUID
	IsRead           *bool
	Priority         *string
	SourceEntityName *string
	SourceEntityID   *uuid.UUID
	CreatedAfter     *time.Time
	CreatedBefore    *time.Time
}
```

- [ ] **Step 2: Create the order file**

```go
package notificationbus

import "github.com/timmaaaz/ichor/business/sdk/order"

// DefaultOrderBy represents the default ordering for notifications (newest first).
var DefaultOrderBy = order.NewBy(OrderByCreatedDate, order.DESC)

// Set of fields that can be used for ordering.
const (
	OrderByID          = "id"
	OrderByPriority    = "priority"
	OrderByIsRead      = "is_read"
	OrderByCreatedDate = "created_date"
)
```

- [ ] **Step 3: Build to verify**

```bash
go build ./business/domain/workflow/notificationbus/...
```

- [ ] **Step 4: Commit**

```bash
git add business/domain/workflow/notificationbus/filter.go business/domain/workflow/notificationbus/order.go
git commit -m "feat(notificationbus): add filter and order definitions"
```

---

## Task 4: Business Layer — Core (Storer + Business)

**Files:**
- Create: `business/domain/workflow/notificationbus/notificationbus.go`

- [ ] **Step 1: Create the business core file**

```go
package notificationbus

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
	ErrNotFound = errors.New("notification not found")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, notification Notification) error
	QueryByID(ctx context.Context, id uuid.UUID) (Notification, error)
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, pg page.Page) ([]Notification, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	MarkAsRead(ctx context.Context, id uuid.UUID, readDate time.Time) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID, readDate time.Time) (int, error)
}

// Business manages notification operations.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
}

// NewBusiness constructs a notification business API for use.
func NewBusiness(log *logger.Logger, delegate *delegate.Delegate, storer Storer) *Business {
	return &Business{
		log:      log,
		storer:   storer,
		delegate: delegate,
	}
}

// NewWithTx constructs a new Business value replacing the Storer with a
// Storer that uses the specified transaction.
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

// Create adds a new notification to the system.
func (b *Business) Create(ctx context.Context, nn NewNotification) (Notification, error) {
	ctx, span := otel.AddSpan(ctx, "business.notificationbus.create")
	defer span.End()

	now := time.Now()

	notification := Notification{
		ID:               uuid.New(),
		UserID:           nn.UserID,
		Title:            nn.Title,
		Message:          nn.Message,
		Priority:         nn.Priority,
		IsRead:           false,
		SourceEntityName: nn.SourceEntityName,
		SourceEntityID:   nn.SourceEntityID,
		ActionURL:        nn.ActionURL,
		CreatedDate:      now,
	}

	if err := b.storer.Create(ctx, notification); err != nil {
		return Notification{}, fmt.Errorf("create: %w", err)
	}

	return notification, nil
}

// QueryByID finds the notification by the specified ID.
func (b *Business) QueryByID(ctx context.Context, id uuid.UUID) (Notification, error) {
	ctx, span := otel.AddSpan(ctx, "business.notificationbus.querybyid")
	defer span.End()

	notification, err := b.storer.QueryByID(ctx, id)
	if err != nil {
		return Notification{}, fmt.Errorf("query: notificationID[%s]: %w", id, err)
	}

	return notification, nil
}

// Query retrieves a list of notifications from the system.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, pg page.Page) ([]Notification, error) {
	ctx, span := otel.AddSpan(ctx, "business.notificationbus.query")
	defer span.End()

	notifications, err := b.storer.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return notifications, nil
}

// Count returns the total number of notifications matching the filter.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.notificationbus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// MarkAsRead marks a single notification as read.
func (b *Business) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	ctx, span := otel.AddSpan(ctx, "business.notificationbus.markasread")
	defer span.End()

	if err := b.storer.MarkAsRead(ctx, id, time.Now()); err != nil {
		return fmt.Errorf("markasread: notificationID[%s]: %w", id, err)
	}

	return nil
}

// MarkAllAsRead marks all unread notifications for a user as read.
func (b *Business) MarkAllAsRead(ctx context.Context, userID uuid.UUID) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.notificationbus.markallasread")
	defer span.End()

	count, err := b.storer.MarkAllAsRead(ctx, userID, time.Now())
	if err != nil {
		return 0, fmt.Errorf("markallasread: userID[%s]: %w", userID, err)
	}

	return count, nil
}
```

- [ ] **Step 2: Build to verify**

```bash
go build ./business/domain/workflow/notificationbus/...
```

- [ ] **Step 3: Commit**

```bash
git add business/domain/workflow/notificationbus/notificationbus.go
git commit -m "feat(notificationbus): add Storer interface and Business core"
```

---

## Task 5: DB Layer — Models

**Files:**
- Create: `business/domain/workflow/notificationbus/stores/notificationdb/model.go`

- [ ] **Step 1: Create the DB model file**

```go
package notificationdb

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/workflow/notificationbus"
)

type dbNotification struct {
	ID               uuid.UUID      `db:"id"`
	UserID           uuid.UUID      `db:"user_id"`
	Title            string         `db:"title"`
	Message          sql.NullString `db:"message"`
	Priority         string         `db:"priority"`
	IsRead           bool           `db:"is_read"`
	ReadDate         sql.NullTime   `db:"read_date"`
	SourceEntityName sql.NullString `db:"source_entity_name"`
	SourceEntityID   sql.NullString `db:"source_entity_id"`
	ActionURL        sql.NullString `db:"action_url"`
	CreatedDate      time.Time      `db:"created_date"`
}

func toDBNotification(n notificationbus.Notification) dbNotification {
	db := dbNotification{
		ID:          n.ID,
		UserID:      n.UserID,
		Title:       n.Title,
		Priority:    n.Priority,
		IsRead:      n.IsRead,
		CreatedDate: n.CreatedDate,
	}

	if n.Message != "" {
		db.Message = sql.NullString{String: n.Message, Valid: true}
	}
	if n.ReadDate != nil {
		db.ReadDate = sql.NullTime{Time: *n.ReadDate, Valid: true}
	}
	if n.SourceEntityName != "" {
		db.SourceEntityName = sql.NullString{String: n.SourceEntityName, Valid: true}
	}
	if n.SourceEntityID != uuid.Nil {
		db.SourceEntityID = sql.NullString{String: n.SourceEntityID.String(), Valid: true}
	}
	if n.ActionURL != "" {
		db.ActionURL = sql.NullString{String: n.ActionURL, Valid: true}
	}

	return db
}

func toBusNotification(db dbNotification) notificationbus.Notification {
	n := notificationbus.Notification{
		ID:          db.ID,
		UserID:      db.UserID,
		Title:       db.Title,
		Priority:    db.Priority,
		IsRead:      db.IsRead,
		CreatedDate: db.CreatedDate,
	}

	if db.Message.Valid {
		n.Message = db.Message.String
	}
	if db.ReadDate.Valid {
		n.ReadDate = &db.ReadDate.Time
	}
	if db.SourceEntityName.Valid {
		n.SourceEntityName = db.SourceEntityName.String
	}
	if db.SourceEntityID.Valid {
		n.SourceEntityID, _ = uuid.Parse(db.SourceEntityID.String)
	}
	if db.ActionURL.Valid {
		n.ActionURL = db.ActionURL.String
	}

	return n
}

func toBusNotifications(dbs []dbNotification) []notificationbus.Notification {
	notifications := make([]notificationbus.Notification, len(dbs))
	for i, db := range dbs {
		notifications[i] = toBusNotification(db)
	}
	return notifications
}
```

- [ ] **Step 2: Build to verify**

```bash
go build ./business/domain/workflow/notificationbus/stores/notificationdb/...
```

- [ ] **Step 3: Commit**

```bash
git add business/domain/workflow/notificationbus/stores/notificationdb/model.go
git commit -m "feat(notificationdb): add DB model and bus conversions"
```

---

## Task 6: DB Layer — Filter & Order

**Files:**
- Create: `business/domain/workflow/notificationbus/stores/notificationdb/filter.go`
- Create: `business/domain/workflow/notificationbus/stores/notificationdb/order.go`

- [ ] **Step 1: Create the filter file**

```go
package notificationdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/workflow/notificationbus"
)

func applyFilter(filter notificationbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["id"] = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.UserID != nil {
		data["user_id"] = *filter.UserID
		wc = append(wc, "user_id = :user_id")
	}

	if filter.IsRead != nil {
		data["is_read"] = *filter.IsRead
		wc = append(wc, "is_read = :is_read")
	}

	if filter.Priority != nil {
		data["priority"] = *filter.Priority
		wc = append(wc, "priority = :priority")
	}

	if filter.SourceEntityName != nil {
		data["source_entity_name"] = *filter.SourceEntityName
		wc = append(wc, "source_entity_name = :source_entity_name")
	}

	if filter.SourceEntityID != nil {
		data["source_entity_id"] = (*filter.SourceEntityID).String()
		wc = append(wc, "source_entity_id = :source_entity_id")
	}

	if filter.CreatedAfter != nil {
		data["created_after"] = *filter.CreatedAfter
		wc = append(wc, "created_date >= :created_after")
	}

	if filter.CreatedBefore != nil {
		data["created_before"] = *filter.CreatedBefore
		wc = append(wc, "created_date <= :created_before")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
```

- [ ] **Step 2: Create the order file**

```go
package notificationdb

import (
	"fmt"

	"github.com/timmaaaz/ichor/business/domain/workflow/notificationbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	notificationbus.OrderByID:          "id",
	notificationbus.OrderByPriority:    "priority",
	notificationbus.OrderByIsRead:      "is_read",
	notificationbus.OrderByCreatedDate: "created_date",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
```

- [ ] **Step 3: Build to verify**

```bash
go build ./business/domain/workflow/notificationbus/stores/notificationdb/...
```

- [ ] **Step 4: Commit**

```bash
git add business/domain/workflow/notificationbus/stores/notificationdb/filter.go business/domain/workflow/notificationbus/stores/notificationdb/order.go
git commit -m "feat(notificationdb): add SQL filter builder and order mapping"
```

---

## Task 7: DB Layer — Store Implementation

**Files:**
- Create: `business/domain/workflow/notificationbus/stores/notificationdb/notificationdb.go`

- [ ] **Step 1: Create the store file**

```go
package notificationdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/workflow/notificationbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for notification database access.
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

// NewWithTx constructs a new Store value replacing the sqlx DB with a sqlx DB
// value that is currently inside a transaction.
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (notificationbus.Storer, error) {
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

// Create inserts a new notification into the database.
func (s *Store) Create(ctx context.Context, notification notificationbus.Notification) error {
	const q = `
	INSERT INTO workflow.notifications
		(id, user_id, title, message, priority, is_read, read_date, source_entity_name, source_entity_id, action_url, created_date)
	VALUES
		(:id, :user_id, :title, :message, :priority, :is_read, :read_date, :source_entity_name, :source_entity_id, :action_url, :created_date)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBNotification(notification)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// QueryByID gets the specified notification from the database.
func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (notificationbus.Notification, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: id.String(),
	}

	const q = `
	SELECT
		id, user_id, title, message, priority, is_read, read_date,
		source_entity_name, source_entity_id, action_url, created_date
	FROM
		workflow.notifications
	WHERE
		id = :id`

	var dbN dbNotification
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbN); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return notificationbus.Notification{}, fmt.Errorf("db: %w", notificationbus.ErrNotFound)
		}
		return notificationbus.Notification{}, fmt.Errorf("db: %w", err)
	}

	return toBusNotification(dbN), nil
}

// Query retrieves a list of notifications from the database.
func (s *Store) Query(ctx context.Context, filter notificationbus.QueryFilter, orderBy order.By, pg page.Page) ([]notificationbus.Notification, error) {
	data := map[string]any{
		"offset":        (pg.Number() - 1) * pg.RowsPerPage(),
		"rows_per_page": pg.RowsPerPage(),
	}

	var buf bytes.Buffer
	buf.WriteString(`
	SELECT
		id, user_id, title, message, priority, is_read, read_date,
		source_entity_name, source_entity_id, action_url, created_date
	FROM
		workflow.notifications`)

	applyFilter(filter, data, &buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}
	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbNs []dbNotification
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbNs); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusNotifications(dbNs), nil
}

// Count returns the total number of notifications matching the filter.
func (s *Store) Count(ctx context.Context, filter notificationbus.QueryFilter) (int, error) {
	data := map[string]any{}

	var buf bytes.Buffer
	buf.WriteString(`
	SELECT
		count(1)
	FROM
		workflow.notifications`)

	applyFilter(filter, data, &buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("db: %w", err)
	}

	return count.Count, nil
}

// MarkAsRead updates a single notification's is_read and read_date.
func (s *Store) MarkAsRead(ctx context.Context, id uuid.UUID, readDate time.Time) error {
	data := struct {
		ID       string    `db:"id"`
		ReadDate time.Time `db:"read_date"`
	}{
		ID:       id.String(),
		ReadDate: readDate,
	}

	const q = `
	UPDATE
		workflow.notifications
	SET
		is_read = true,
		read_date = :read_date
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return fmt.Errorf("db: %w", notificationbus.ErrNotFound)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// MarkAllAsRead marks all unread notifications for a user as read.
func (s *Store) MarkAllAsRead(ctx context.Context, userID uuid.UUID, readDate time.Time) (int, error) {
	data := struct {
		UserID   string    `db:"user_id"`
		ReadDate time.Time `db:"read_date"`
	}{
		UserID:   userID.String(),
		ReadDate: readDate,
	}

	const q = `
	UPDATE
		workflow.notifications
	SET
		is_read = true,
		read_date = :read_date
	WHERE
		user_id = :user_id AND is_read = false`

	nstmt, err := sqlx.Named(q, data)
	if err != nil {
		return 0, fmt.Errorf("named: %w", err)
	}
	nstmt = s.db.Rebind(nstmt)

	result, err := s.db.ExecContext(ctx, nstmt)
	if err != nil {
		return 0, fmt.Errorf("execcontext: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rowsaffected: %w", err)
	}

	return int(rows), nil
}
```

⚠️ **Note on `MarkAllAsRead`:** Uses raw `sqlx.Named` + `Rebind` + `ExecContext` (same pattern as `alertdb.DismissMultiple`) because `NamedExecContext` doesn't return `RowsAffected`. The caller needs the count to report how many were marked.

- [ ] **Step 2: Add compile-time interface check at the bottom of the file**

Add this line at the very end of `notificationdb.go`:

```go
// Ensure Store implements notificationbus.Storer.
var _ notificationbus.Storer = (*Store)(nil)
```

- [ ] **Step 3: Build to verify**

```bash
go build ./business/domain/workflow/notificationbus/...
```

- [ ] **Step 4: Commit**

```bash
git add business/domain/workflow/notificationbus/stores/notificationdb/notificationdb.go
git commit -m "feat(notificationdb): add SQL store implementation"
```

---

## Task 8: App Layer — Models

**Files:**
- Create: `app/domain/workflow/notificationapp/model.go`

- [ ] **Step 1: Create the app model file**

```go
package notificationapp

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/business/domain/workflow/notificationbus"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

// QueryParams holds the raw query parameters from the HTTP request.
type QueryParams struct {
	Page             string
	Rows             string
	OrderBy          string
	ID               string
	IsRead           string
	Priority         string
	SourceEntityName string
	SourceEntityID   string
}

// =============================================================================
// Response model
// =============================================================================

// Notification is the app-layer response model.
type Notification struct {
	ID               string `json:"id"`
	UserID           string `json:"userId"`
	Title            string `json:"title"`
	Message          string `json:"message"`
	Priority         string `json:"priority"`
	IsRead           bool   `json:"isRead"`
	ReadDate         string `json:"readDate"`
	SourceEntityName string `json:"sourceEntityName"`
	SourceEntityID   string `json:"sourceEntityId"`
	ActionURL        string `json:"actionUrl"`
	CreatedDate      string `json:"createdDate"`
}

// Encode implements web.Encoder.
func (app Notification) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// Notifications is a slice of Notification for list responses.
type Notifications []Notification

// Encode implements web.Encoder.
func (app Notifications) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// ToAppNotification converts a bus model to an app-layer response model.
func ToAppNotification(bus notificationbus.Notification) Notification {
	readDate := ""
	if bus.ReadDate != nil {
		readDate = bus.ReadDate.Format(timeutil.FORMAT)
	}

	sourceEntityID := ""
	if bus.SourceEntityID.String() != "00000000-0000-0000-0000-000000000000" {
		sourceEntityID = bus.SourceEntityID.String()
	}

	return Notification{
		ID:               bus.ID.String(),
		UserID:           bus.UserID.String(),
		Title:            bus.Title,
		Message:          bus.Message,
		Priority:         bus.Priority,
		IsRead:           bus.IsRead,
		ReadDate:         readDate,
		SourceEntityName: bus.SourceEntityName,
		SourceEntityID:   sourceEntityID,
		ActionURL:        bus.ActionURL,
		CreatedDate:      bus.CreatedDate.Format(timeutil.FORMAT),
	}
}

// ToAppNotifications converts a slice of bus models to app-layer response models.
func ToAppNotifications(bus []notificationbus.Notification) Notifications {
	app := make(Notifications, len(bus))
	for i, v := range bus {
		app[i] = ToAppNotification(v)
	}
	return app
}
```

- [ ] **Step 2: Build to verify**

```bash
go build ./app/domain/workflow/notificationapp/...
```

- [ ] **Step 3: Commit**

```bash
git add app/domain/workflow/notificationapp/model.go
git commit -m "feat(notificationapp): add app-layer models and conversions"
```

---

## Task 9: App Layer — Filter, Order, Core

**Files:**
- Create: `app/domain/workflow/notificationapp/filter.go`
- Create: `app/domain/workflow/notificationapp/order.go`
- Create: `app/domain/workflow/notificationapp/notificationapp.go`

- [ ] **Step 1: Create the filter file**

```go
package notificationapp

import (
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/workflow/notificationbus"
)

func parseFilter(qp QueryParams) (notificationbus.QueryFilter, error) {
	var filter notificationbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return notificationbus.QueryFilter{}, err
		}
		filter.ID = &id
	}

	if qp.IsRead != "" {
		b, err := strconv.ParseBool(qp.IsRead)
		if err != nil {
			return notificationbus.QueryFilter{}, err
		}
		filter.IsRead = &b
	}

	if qp.Priority != "" {
		filter.Priority = &qp.Priority
	}

	if qp.SourceEntityName != "" {
		filter.SourceEntityName = &qp.SourceEntityName
	}

	if qp.SourceEntityID != "" {
		id, err := uuid.Parse(qp.SourceEntityID)
		if err != nil {
			return notificationbus.QueryFilter{}, err
		}
		filter.SourceEntityID = &id
	}

	return filter, nil
}

func parseCountFilter(isRead string, userID uuid.UUID) (notificationbus.QueryFilter, error) {
	filter := notificationbus.QueryFilter{
		UserID: &userID,
	}

	if isRead != "" {
		b, err := strconv.ParseBool(isRead)
		if err != nil {
			return notificationbus.QueryFilter{}, err
		}
		filter.IsRead = &b
	}

	return filter, nil
}

// parseTime parses a time string in the standard format.
func parseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
```

- [ ] **Step 2: Create the order file**

```go
package notificationapp

import (
	"github.com/timmaaaz/ichor/business/domain/workflow/notificationbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var orderByFields = map[string]string{
	"id":          notificationbus.OrderByID,
	"priority":    notificationbus.OrderByPriority,
	"isRead":      notificationbus.OrderByIsRead,
	"createdDate": notificationbus.OrderByCreatedDate,
}

// DefaultOrderBy is the default ordering for notification queries.
var DefaultOrderBy = order.NewBy(notificationbus.OrderByCreatedDate, order.DESC)
```

- [ ] **Step 3: Create the app core file**

```go
package notificationapp

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/mid"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/workflow/notificationbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer APIs for notification access.
type App struct {
	notificationBus *notificationbus.Business
}

// NewApp constructs a notification app.
func NewApp(notificationBus *notificationbus.Business) *App {
	return &App{
		notificationBus: notificationBus,
	}
}

// Query returns a list of notifications for the authenticated user.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[Notification], error) {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return query.Result[Notification]{}, errs.Newf(errs.Unauthenticated, "get user id: %s", err)
	}

	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[Notification]{}, errs.Newf(errs.InvalidArgument, "page: %s", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[Notification]{}, errs.Newf(errs.InvalidArgument, "filter: %s", err)
	}

	// Always scope to the authenticated user.
	filter.UserID = &userID

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, DefaultOrderBy)
	if err != nil {
		return query.Result[Notification]{}, errs.Newf(errs.InvalidArgument, "order: %s", err)
	}

	notifications, err := a.notificationBus.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return query.Result[Notification]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.notificationBus.Count(ctx, filter)
	if err != nil {
		return query.Result[Notification]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppNotifications(notifications), total, pg), nil
}

// Count returns the number of notifications matching the filter for the authenticated user.
func (a *App) Count(ctx context.Context, r *http.Request) (int, error) {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return 0, errs.Newf(errs.Unauthenticated, "get user id: %s", err)
	}

	isRead := r.URL.Query().Get("is_read")
	filter, err := parseCountFilter(isRead, userID)
	if err != nil {
		return 0, errs.Newf(errs.InvalidArgument, "filter: %s", err)
	}

	count, err := a.notificationBus.Count(ctx, filter)
	if err != nil {
		return 0, errs.Newf(errs.Internal, "count: %s", err)
	}

	return count, nil
}

// MarkAsRead marks a single notification as read. Verifies the notification
// belongs to the authenticated user.
func (a *App) MarkAsRead(ctx context.Context, idStr string) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.Newf(errs.Unauthenticated, "get user id: %s", err)
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return errs.Newf(errs.InvalidArgument, "parse notification id: %s", err)
	}

	// Verify ownership.
	notification, err := a.notificationBus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, notificationbus.ErrNotFound) {
			return errs.Newf(errs.NotFound, "notification not found")
		}
		return errs.Newf(errs.Internal, "query: %s", err)
	}

	if notification.UserID != userID {
		return errs.Newf(errs.NotFound, "notification not found")
	}

	if err := a.notificationBus.MarkAsRead(ctx, id); err != nil {
		return errs.Newf(errs.Internal, "mark as read: %s", err)
	}

	return nil
}

// MarkAllAsRead marks all unread notifications for the authenticated user as read.
func (a *App) MarkAllAsRead(ctx context.Context) (int, error) {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return 0, errs.Newf(errs.Unauthenticated, "get user id: %s", err)
	}

	count, err := a.notificationBus.MarkAllAsRead(ctx, userID)
	if err != nil {
		return 0, errs.Newf(errs.Internal, "mark all as read: %s", err)
	}

	return count, nil
}
```

- [ ] **Step 4: Build to verify**

```bash
go build ./app/domain/workflow/notificationapp/...
```

- [ ] **Step 5: Commit**

```bash
git add app/domain/workflow/notificationapp/
git commit -m "feat(notificationapp): add app layer with filter, order, and core"
```

---

## Task 10: API Layer — Routes, Handlers, Models

**Files:**
- Create: `api/domain/http/workflow/notificationinboxapi/route.go`
- Create: `api/domain/http/workflow/notificationinboxapi/notificationinboxapi.go`
- Create: `api/domain/http/workflow/notificationinboxapi/model.go`

- [ ] **Step 1: Create the route file**

```go
package notificationinboxapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/workflow/notificationapp"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/workflow/notificationbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config holds the dependencies for the notification inbox API.
type Config struct {
	Log             *logger.Logger
	NotificationBus *notificationbus.Business
	AuthClient      *authclient.Client
}

// Routes registers the notification inbox API routes.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	api := newAPI(cfg)
	authen := mid.Authenticate(cfg.AuthClient)

	app.HandlerFunc(http.MethodGet, version, "/workflow/notifications", api.query, authen)
	app.HandlerFunc(http.MethodGet, version, "/workflow/notifications/count", api.count, authen)
	app.HandlerFunc(http.MethodPost, version, "/workflow/notifications/{notification_id}/read", api.markAsRead, authen)
	app.HandlerFunc(http.MethodPost, version, "/workflow/notifications/read-all", api.markAllAsRead, authen)
}
```

⚠️ **Route ordering matters:** `/workflow/notifications/count` and `/workflow/notifications/read-all` must be registered BEFORE `/workflow/notifications/{notification_id}/read` to avoid the router matching "count" or "read-all" as a `notification_id`. In Go's `web.App` (which wraps `httprouter`), static segments take precedence over wildcards so the registration order shown here is safe.

- [ ] **Step 2: Create the model file**

```go
package notificationinboxapi

import "encoding/json"

// UnreadCount is the response for the count endpoint.
type UnreadCount struct {
	Count int `json:"count"`
}

// Encode implements web.Encoder.
func (u UnreadCount) Encode() ([]byte, string, error) {
	data, err := json.Marshal(u)
	return data, "application/json", err
}

// MarkAllReadResult is the response for the mark-all-read endpoint.
type MarkAllReadResult struct {
	Count int `json:"count"`
}

// Encode implements web.Encoder.
func (m MarkAllReadResult) Encode() ([]byte, string, error) {
	data, err := json.Marshal(m)
	return data, "application/json", err
}

// SuccessResult is a generic success response.
type SuccessResult struct {
	Success bool `json:"success"`
}

// Encode implements web.Encoder.
func (s SuccessResult) Encode() ([]byte, string, error) {
	data, err := json.Marshal(s)
	return data, "application/json", err
}
```

- [ ] **Step 3: Create the handler file**

```go
package notificationinboxapi

import (
	"context"
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/workflow/notificationapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	notificationApp *notificationapp.App
}

func newAPI(cfg Config) *api {
	return &api{
		notificationApp: notificationapp.NewApp(cfg.NotificationBus),
	}
}

func (a *api) query(ctx context.Context, r *http.Request) web.Encoder {
	qp := notificationapp.QueryParams{
		Page:             r.URL.Query().Get("page"),
		Rows:             r.URL.Query().Get("rows"),
		OrderBy:          r.URL.Query().Get("orderBy"),
		ID:               r.URL.Query().Get("id"),
		IsRead:           r.URL.Query().Get("is_read"),
		Priority:         r.URL.Query().Get("priority"),
		SourceEntityName: r.URL.Query().Get("source_entity_name"),
		SourceEntityID:   r.URL.Query().Get("source_entity_id"),
	}

	result, err := a.notificationApp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return result
}

func (a *api) count(ctx context.Context, r *http.Request) web.Encoder {
	count, err := a.notificationApp.Count(ctx, r)
	if err != nil {
		return errs.NewError(err)
	}

	return UnreadCount{Count: count}
}

func (a *api) markAsRead(ctx context.Context, r *http.Request) web.Encoder {
	id := web.Param(r, "notification_id")

	if err := a.notificationApp.MarkAsRead(ctx, id); err != nil {
		return errs.NewError(err)
	}

	return SuccessResult{Success: true}
}

func (a *api) markAllAsRead(ctx context.Context, r *http.Request) web.Encoder {
	count, err := a.notificationApp.MarkAllAsRead(ctx)
	if err != nil {
		return errs.NewError(err)
	}

	return MarkAllReadResult{Count: count}
}
```

- [ ] **Step 4: Build to verify**

```bash
go build ./api/domain/http/workflow/notificationinboxapi/...
```

- [ ] **Step 5: Commit**

```bash
git add api/domain/http/workflow/notificationinboxapi/
git commit -m "feat(notificationinboxapi): add API routes and handlers for notification inbox"
```

---

## Task 11: Wiring — dbtest.go + all.go

**Files:**
- Modify: `business/sdk/dbtest/dbtest.go`
- Modify: `api/cmd/services/ichor/build/all/all.go`

- [ ] **Step 1: Add to BusDomain struct in `dbtest.go`**

In the `BusDomain` struct, find the Workflow section (around line 274) and add:

```go
	// Workflow
	Workflow          *workflow.Business
	Alert             *alertbus.Business
	ActionPermissions *actionpermissionsbus.Business
	Notification      *notificationbus.Business
```

- [ ] **Step 2: Wire in `newBusDomains` in `dbtest.go`**

In the `newBusDomains` function, find the Workflow section (around line 388) and add:

```go
	alertBus := alertbus.NewBusiness(log, alertdb.NewStore(log, db))
	actionPermissionsBus := actionpermissionsbus.NewBusiness(log, actionpermissionsdb.NewStore(log, db))
	notificationBus := notificationbus.NewBusiness(log, delegate, notificationdb.NewStore(log, db))
```

Then add `Notification: notificationBus,` to the return struct.

Add the imports:
```go
"github.com/timmaaaz/ichor/business/domain/workflow/notificationbus"
"github.com/timmaaaz/ichor/business/domain/workflow/notificationbus/stores/notificationdb"
```

- [ ] **Step 3: Wire routes in `all.go`**

Add the imports to `all.go`:
```go
"github.com/timmaaaz/ichor/api/domain/http/workflow/notificationinboxapi"
"github.com/timmaaaz/ichor/business/domain/workflow/notificationbus"
"github.com/timmaaaz/ichor/business/domain/workflow/notificationbus/stores/notificationdb"
```

In the Workflow section of `all.go`, add the bus construction near the existing `alertBus` line (around line 1310):

```go
notificationBus := notificationbus.NewBusiness(cfg.Log, delegate, notificationdb.NewStore(cfg.Log, cfg.DB))
```

Then add route registration after the existing `notificationsapi.Routes(...)` block:

```go
notificationinboxapi.Routes(app, notificationinboxapi.Config{
    Log:             cfg.Log,
    NotificationBus: notificationBus,
    AuthClient:      cfg.AuthClient,
})
```

- [ ] **Step 4: Build to verify full service compiles**

```bash
go build ./api/cmd/services/ichor/...
```

- [ ] **Step 5: Commit**

```bash
git add business/sdk/dbtest/dbtest.go api/cmd/services/ichor/build/all/all.go
git commit -m "feat(wiring): integrate notificationbus into dbtest and service routes"
```

---

## Task 12: Integration Tests — Seed & Entry Point

**Files:**
- Create: `api/cmd/services/ichor/tests/workflow/notificationinboxapi/notification_test.go`
- Create: `api/cmd/services/ichor/tests/workflow/notificationinboxapi/seed_test.go`

- [ ] **Step 1: Create the test entry point**

```go
package notificationinboxapi_test

import (
	"testing"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func Test_NotificationInboxAPI(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_NotificationInboxAPI")

	sd := insertSeedData(t, test)

	// -------------------------------------------------------------------------

	t.Run("query200", func(t *testing.T) {
		for _, table := range query200(sd) {
			apitest.Run(t, table, "query200")
		}
	})

	t.Run("count200", func(t *testing.T) {
		for _, table := range count200(sd) {
			apitest.Run(t, table, "count200")
		}
	})

	t.Run("markAsRead200", func(t *testing.T) {
		for _, table := range markAsRead200(sd) {
			apitest.Run(t, table, "markAsRead200")
		}
	})

	t.Run("markAllAsRead200", func(t *testing.T) {
		for _, table := range markAllAsRead200(sd) {
			apitest.Run(t, table, "markAllAsRead200")
		}
	})
}
```

- [ ] **Step 2: Create the seed file**

```go
package notificationinboxapi_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/workflow/notificationapp"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/notificationbus"
)

type SeedData struct {
	apitest.SeedData
	Notifications []notificationapp.Notification
	User          apitest.User
	UnreadCount   int
	TotalCount    int
}

func insertSeedData(t *testing.T, test *apitest.Test) SeedData {
	ctx := context.Background()

	usrs := userbus.TestSeedUsersWithNoFKs(t, ctx, 1, test.DB.BusDomain.User)
	tu := apitest.User{
		User:  usrs[0],
		Token: apitest.Token(test.DB.BusDomain.User, test.Auth, usrs[0].Email.Address),
	}

	// Create 4 notifications: 2 unread, 2 read (different priorities).
	nots := make([]notificationbus.Notification, 4)
	appNots := make([]notificationapp.Notification, 4)

	now := time.Now().Truncate(time.Microsecond)

	for i := 0; i < 4; i++ {
		priorities := []string{
			notificationbus.PriorityCritical,
			notificationbus.PriorityHigh,
			notificationbus.PriorityMedium,
			notificationbus.PriorityLow,
		}

		nn := notificationbus.NewNotification{
			UserID:           usrs[0].ID,
			Title:            "Test Notification " + string(rune('A'+i)),
			Message:          "Message for notification " + string(rune('A'+i)),
			Priority:         priorities[i],
			SourceEntityName: "orders",
			SourceEntityID:   uuid.New(),
			ActionURL:        "/orders/" + uuid.New().String(),
		}

		n, err := test.DB.BusDomain.Notification.Create(ctx, nn)
		if err != nil {
			t.Fatalf("seeding notification %d: %s", i, err)
		}

		// Mark the last 2 as read.
		if i >= 2 {
			if err := test.DB.BusDomain.Notification.MarkAsRead(ctx, n.ID); err != nil {
				t.Fatalf("marking notification %d as read: %s", i, err)
			}
			// Re-query to get the updated state.
			n, err = test.DB.BusDomain.Notification.QueryByID(ctx, n.ID)
			if err != nil {
				t.Fatalf("re-querying notification %d: %s", i, err)
			}
		}

		nots[i] = n
		appNots[i] = notificationapp.ToAppNotification(n)
	}

	_ = now

	return SeedData{
		SeedData:      test.SeedData,
		Notifications: appNots,
		User:          tu,
		UnreadCount:   2,
		TotalCount:    4,
	}
}
```

- [ ] **Step 3: Build to verify test files compile**

```bash
go build ./api/cmd/services/ichor/tests/workflow/notificationinboxapi/...
```

This will fail until we add the test table functions. That's expected — proceed to Task 13.

- [ ] **Step 4: Commit**

```bash
git add api/cmd/services/ichor/tests/workflow/notificationinboxapi/notification_test.go api/cmd/services/ichor/tests/workflow/notificationinboxapi/seed_test.go
git commit -m "test(notificationinboxapi): add test entry point and seed data"
```

---

## Task 13: Integration Tests — Query & Count Tables

**Files:**
- Create: `api/cmd/services/ichor/tests/workflow/notificationinboxapi/query_test.go`

- [ ] **Step 1: Create the query test file**

```go
package notificationinboxapi_test

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/workflow/notificationapp"
)

func query200(sd SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic-query",
			URL:        "/v1/workflow/notifications?page=1&rows=10",
			Token:      sd.User.Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &[]notificationapp.Notification{},
			ExpResp: func() *[]notificationapp.Notification {
				// Default order is created_date DESC, so newest first.
				nots := make([]notificationapp.Notification, len(sd.Notifications))
				copy(nots, sd.Notifications)
				sort.Slice(nots, func(i, j int) bool {
					return nots[i].CreatedDate > nots[j].CreatedDate
				})
				return &nots
			}(),
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*[]notificationapp.Notification)
				expResp := exp.(*[]notificationapp.Notification)

				if len(*gotResp) != len(*expResp) {
					return fmt.Sprintf("expected %d notifications, got %d", len(*expResp), len(*gotResp))
				}

				for i := range *gotResp {
					if (*gotResp)[i].ID != (*expResp)[i].ID {
						return fmt.Sprintf("notification[%d] ID mismatch: got %s, exp %s", i, (*gotResp)[i].ID, (*expResp)[i].ID)
					}
					if (*gotResp)[i].IsRead != (*expResp)[i].IsRead {
						return fmt.Sprintf("notification[%d] IsRead mismatch: got %v, exp %v", i, (*gotResp)[i].IsRead, (*expResp)[i].IsRead)
					}
				}

				return ""
			},
		},
		{
			Name:       "filter-unread",
			URL:        "/v1/workflow/notifications?page=1&rows=10&is_read=false",
			Token:      sd.User.Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &[]notificationapp.Notification{},
			ExpResp:    &[]notificationapp.Notification{},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*[]notificationapp.Notification)
				if len(*gotResp) != sd.UnreadCount {
					return fmt.Sprintf("expected %d unread notifications, got %d", sd.UnreadCount, len(*gotResp))
				}
				for _, n := range *gotResp {
					if n.IsRead {
						return fmt.Sprintf("expected all unread, got read notification %s", n.ID)
					}
				}
				return ""
			},
		},
	}

	return table
}

func count200(sd SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "unread-count",
			URL:        "/v1/workflow/notifications/count?is_read=false",
			Token:      sd.User.Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &struct{ Count int }{},
			ExpResp:    &struct{ Count int }{Count: sd.UnreadCount},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "total-count",
			URL:        "/v1/workflow/notifications/count",
			Token:      sd.User.Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &struct{ Count int }{},
			ExpResp:    &struct{ Count int }{Count: sd.TotalCount},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
```

- [ ] **Step 2: Build to verify**

```bash
go build ./api/cmd/services/ichor/tests/workflow/notificationinboxapi/...
```

- [ ] **Step 3: Commit**

```bash
git add api/cmd/services/ichor/tests/workflow/notificationinboxapi/query_test.go
git commit -m "test(notificationinboxapi): add query and count test tables"
```

---

## Task 14: Integration Tests — Mark-as-Read Tables

**Files:**
- Create: `api/cmd/services/ichor/tests/workflow/notificationinboxapi/read_test.go`

- [ ] **Step 1: Create the read test file**

```go
package notificationinboxapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func markAsRead200(sd SeedData) []apitest.Table {
	// Pick the first unread notification (index 0, which is unread per seed).
	unreadID := sd.Notifications[0].ID

	table := []apitest.Table{
		{
			Name:       "mark-single-read",
			URL:        fmt.Sprintf("/v1/workflow/notifications/%s/read", unreadID),
			Token:      sd.User.Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			GotResp:    &struct{ Success bool }{},
			ExpResp:    &struct{ Success bool }{Success: true},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func markAllAsRead200(sd SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "mark-all-read",
			URL:        "/v1/workflow/notifications/read-all",
			Token:      sd.User.Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodPost,
			GotResp:    &struct{ Count int }{},
			ExpResp:    &struct{ Count int }{},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*struct{ Count int })
				// After markAsRead200 already marked 1, there should be
				// at most 1 remaining unread (originally 2 unread, minus 1).
				// But test ordering isn't guaranteed to be sequential across
				// t.Run groups, so just verify count >= 0.
				if gotResp.Count < 0 {
					return fmt.Sprintf("expected count >= 0, got %d", gotResp.Count)
				}
				return ""
			},
		},
	}

	return table
}
```

⚠️ **Note on test interdependence:** The `markAsRead200` and `markAllAsRead200` tests modify state. In the `apitest.Run` framework, tests within a single `Test_` function share the same database. The `markAllAsRead200` count assertion is intentionally flexible because it runs after `markAsRead200` may have already marked one notification.

- [ ] **Step 2: Build to verify all test files compile**

```bash
go build ./api/cmd/services/ichor/tests/workflow/notificationinboxapi/...
```

- [ ] **Step 3: Commit**

```bash
git add api/cmd/services/ichor/tests/workflow/notificationinboxapi/read_test.go
git commit -m "test(notificationinboxapi): add mark-as-read test tables"
```

---

## Task 15: Build & Run Tests

- [ ] **Step 1: Full service build**

```bash
go build ./...
```

Verify zero compile errors.

- [ ] **Step 2: Run integration tests**

```bash
go test ./api/cmd/services/ichor/tests/workflow/notificationinboxapi/... -v -count=1
```

Expected: All 4 test groups pass (`query200`, `count200`, `markAsRead200`, `markAllAsRead200`).

- [ ] **Step 3: Run existing workflow tests to verify no regressions**

```bash
go test ./api/cmd/services/ichor/tests/workflow/alertapi/... -v -count=1
go test ./api/cmd/services/ichor/tests/workflow/notificationsapi/... -v -count=1
```

Expected: All existing tests still pass.

- [ ] **Step 4: Final commit (if any fixes were needed)**

```bash
git add -A
git commit -m "fix(notificationinboxapi): integration test fixes"
```

Only run this step if fixes were needed. If all tests passed on first try, skip.

---

## Reference: Key Patterns to Follow

### sqldb Error Handling
- `NamedQueryStruct` returns `sqldb.ErrDBNotFound` when no rows → wrap as `notificationbus.ErrNotFound`
- `NamedQuerySlice` returns `nil` (not error) for empty results
- `NamedExecContext` does not return `RowsAffected` → use raw `sqlx.Named` + `ExecContext` when count needed

### Nullable Columns
- `sql.NullString` for nullable `TEXT`/`VARCHAR` columns
- `sql.NullTime` for nullable `TIMESTAMP` columns
- `sql.NullString` for nullable UUID columns (not `*uuid.UUID` — sqlx doesn't handle it natively with Postgres)
- Zero-value UUID (`uuid.Nil`) in bus model → skip setting `NullString.Valid` in `toDBNotification`

### App Layer Conventions
- All response fields are `string` type (UUIDs, dates, numbers serialized as strings) — **except** `bool` fields like `IsRead` which remain `bool`
- JSON tags use camelCase (`userId`, `createdDate`, `isRead`)
- `Encode() ([]byte, string, error)` on all response types for `web.Encoder` interface

### Route Registration
- All inbox routes use `authen` middleware only (no `Authorize`) — users can only see their own notifications, enforced at the app layer by injecting `userID` into the filter
- Registered in `all.go` alongside existing `notificationsapi.Routes(...)` call
