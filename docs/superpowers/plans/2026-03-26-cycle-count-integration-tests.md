# Cycle Count Integration Tests Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add integration tests for cycle count sessions and cycle count items domains — the 7th layer for PR #107.

**Architecture:** Two new test packages under `api/cmd/services/ichor/tests/inventory/`, following the exact pattern established by `picktaskapi/`. Each package has an entry-point test, seed file, and per-verb test files. A shared `insertSeedData` function builds the full dependency chain (users → geography → warehouse → products → inventory → cycle counts → permissions).

**Tech Stack:** Go 1.23, `apitest` framework, `dbtest` Docker PostgreSQL, `google/go-cmp`

---

## File Structure

### Cycle Count Session Tests
```
api/cmd/services/ichor/tests/inventory/cyclecountsessionapi/
├── cyclecountsession_test.go   — entry point, test orchestration
├── seed_test.go                — insertSeedData + shared seed state
├── create_test.go              — create 200/400/401/409
├── query_test.go               — query 200, queryByID 200/404
├── update_test.go              — update 200/400/401/404 + standalone multi-step tests
└── delete_test.go              — delete 200/401/404
```

### Cycle Count Item Tests
```
api/cmd/services/ichor/tests/inventory/cyclecountitemapi/
├── cyclecountitem_test.go      — entry point, test orchestration
├── seed_test.go                — insertSeedData + shared seed state
├── create_test.go              — create 200/400/401/409
├── query_test.go               — query 200, queryByID 200/404
├── update_test.go              — update 200/400/401/404
└── delete_test.go              — delete 200/401/404
```

---

## Reference Information

### Key URLs
| Domain | Base URL |
|--------|----------|
| Sessions | `/v1/inventory/cycle-count-sessions` |
| Items | `/v1/inventory/cycle-count-items` |

### URL Params
| Domain | Path param |
|--------|------------|
| Sessions | `{session_id}` |
| Items | `{item_id}` |

### Route Table Constants
| Domain | RouteTable |
|--------|------------|
| Sessions | `"inventory.cycle_count_sessions"` |
| Items | `"inventory.cycle_count_items"` |

### Status Enums
| Domain | Statuses |
|--------|----------|
| Sessions | `draft`, `in_progress`, `completed`, `cancelled` |
| Items | `pending`, `counted`, `variance_approved`, `variance_rejected` |

### Error Messages (exact strings)
| Scenario | Error |
|----------|-------|
| Session not found | `errs.Newf(errs.NotFound, "cycle count session not found")` |
| Item not found | `errs.Newf(errs.NotFound, "cycle count item not found")` |
| Session terminal state | `errs.Newf(errs.FailedPrecondition, "session is already %s and cannot be transitioned", status)` |
| Session complete from wrong status | `errs.Newf(errs.FailedPrecondition, "session must be in_progress to complete, current status: %s", status)` |
| Invalid session status | `errs.Newf(errs.InvalidArgument, "parse status: invalid status %q", value)` |
| Invalid item status | `errs.Newf(errs.InvalidArgument, "parse status: invalid status %q", value)` |
| Missing name (session create) | `errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"name\",\"error\":\"name is a required field\"}]")` |
| Missing sessionId (item create) | `errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"sessionId\",\"error\":\"sessionId is a required field\"}]")` |
| Missing productId (item create) | `errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"productId\",\"error\":\"productId is a required field\"}]")` |
| Missing locationId (item create) | `errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"locationId\",\"error\":\"locationId is a required field\"}]")` |
| Missing systemQuantity (item create) | `errs.Newf(errs.InvalidArgument, "validate: [{\"field\":\"systemQuantity\",\"error\":\"systemQuantity is a required field\"}]")` |
| FK violation (session create) | `errs.Newf(errs.Aborted, "create: namedexeccontext: foreign key violation")` — unlikely since only FK is created_by (from auth), but test with bad UUID path param on update |
| FK violation (item create) | `errs.Newf(errs.Aborted, "create: namedexeccontext: foreign key violation")` |
| No permission | `errs.Newf(errs.Unauthenticated, "user does not have permission %s for table: %s", action, routeTable)` |
| Bad token format | `errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments")` |
| Bad token sig | `errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]")` |

### App-Layer Models (all fields are strings)

**CycleCountSession response:**
```go
type CycleCountSession struct {
    ID            string `json:"id"`
    Name          string `json:"name"`
    Status        string `json:"status"`
    CreatedBy     string `json:"createdBy"`
    CreatedDate   string `json:"createdDate"`
    UpdatedDate   string `json:"updatedDate"`
    CompletedDate string `json:"completedDate"`
}
```

**NewCycleCountSession request:**
```go
type NewCycleCountSession struct {
    Name string `json:"name" validate:"required"`
}
```

**UpdateCycleCountSession request:**
```go
type UpdateCycleCountSession struct {
    Name   *string `json:"name" validate:"omitempty"`
    Status *string `json:"status" validate:"omitempty"`
}
```

**CycleCountItem response:**
```go
type CycleCountItem struct {
    ID              string `json:"id"`
    SessionID       string `json:"sessionId"`
    ProductID       string `json:"productId"`
    LocationID      string `json:"locationId"`
    SystemQuantity  string `json:"systemQuantity"`
    CountedQuantity string `json:"countedQuantity"`
    Variance        string `json:"variance"`
    Status          string `json:"status"`
    CountedBy       string `json:"countedBy"`
    CountedDate     string `json:"countedDate"`
    CreatedDate     string `json:"createdDate"`
    UpdatedDate     string `json:"updatedDate"`
}
```

**NewCycleCountItem request:**
```go
type NewCycleCountItem struct {
    SessionID      string `json:"sessionId" validate:"required,min=36,max=36"`
    ProductID      string `json:"productId" validate:"required,min=36,max=36"`
    LocationID     string `json:"locationId" validate:"required,min=36,max=36"`
    SystemQuantity string `json:"systemQuantity" validate:"required"`
}
```

**UpdateCycleCountItem request:**
```go
type UpdateCycleCountItem struct {
    CountedQuantity *string `json:"countedQuantity" validate:"omitempty"`
    Status          *string `json:"status" validate:"omitempty"`
}
```

### Key Behavior Notes

1. **Session `CreatedBy`** is injected server-side from the authenticated user — NOT sent by the client.
2. **Item `CountedBy` and `CountedDate`** are auto-injected by the app layer when `CountedQuantity` is set. Client never sends these.
3. **Item `Variance`** is auto-computed by the bus layer: `variance = countedQuantity - systemQuantity`.
4. **Session completion** is a transactional operation: re-queries session inside TX (TOCTOU guard), pages through `variance_approved` items, creates inventory adjustments with `ReasonCodeCycleCount`, immediately approves them.
5. **Seed data sorting:** `TestSeed*` functions sort by `ID.String()` (UUID lexicographic), so `sd.CycleCountSessions[0]` is the one with the lowest UUID string.
6. **`CompletedDate`** is empty string `""` when nil (not omitted from JSON).
7. **`CountedQuantity`, `Variance`, `CountedBy`, `CountedDate`** are all empty strings when nil/zero.

### Seed Data Dependency Chain
```
users → regions (query existing) → cities → streets → warehouses → zones →
inventory locations → contact infos → brands → product categories → products →
inventory items → cycle count sessions → cycle count items → roles → user roles → table access
```

### `dbtest` Helpers
- `dbtest.StringPointer(s string) *string`
- `dbtest.BoolPointer(b bool) *bool`
- `dbtest.IntPointer(i int) *int` — check if this exists; if not, create inline `func intPtr(i int) *int { return &i }`

### Imports You'll Need (both packages)
```go
import (
    "context"
    "fmt"
    "net/http"
    "sort"
    "testing"

    "github.com/google/go-cmp/cmp"
    "github.com/google/uuid"
    "github.com/timmaaaz/ichor/api/sdk/http/apitest"
    "github.com/timmaaez/ichor/app/sdk/errs"
    "github.com/timmaaez/ichor/app/sdk/query"
    "github.com/timmaaez/ichor/business/sdk/dbtest"
    "github.com/timmaaez/ichor/business/sdk/page"
    // Domain-specific imports below
)
```

---

## Task 1: Cycle Count Session — Entry Point and Seed Data

**Files:**
- Create: `api/cmd/services/ichor/tests/inventory/cyclecountsessionapi/cyclecountsession_test.go`
- Create: `api/cmd/services/ichor/tests/inventory/cyclecountsessionapi/seed_test.go`

### Step 1: Create `seed_test.go`

- [ ] **Step 1a: Write the seed file**

This file creates all test data. Seed 2 users (1 regular, 1 admin), full dependency chain, 4 cycle count sessions, and permissions restricting the regular user to read-only on the session table.

```go
package cyclecountsessionapi_test

import (
	"context"
	"fmt"

	"github.com/timmaaaz/ichor/api/domain/http/inventory/cyclecountsessionapi"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/inventory/cyclecountsessionapp"
	"github.com/timmaaez/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaez/ichor/business/domain/geography/citybus"
	"github.com/timmaaez/ichor/business/domain/geography/regionbus"
	"github.com/timmaaez/ichor/business/domain/geography/streetbus"
	"github.com/timmaaez/ichor/business/domain/inventory/cyclecountsessionbus"
	"github.com/timmaaez/ichor/business/domain/permissions/rolebus"
	"github.com/timmaaez/ichor/business/domain/permissions/tableaccessbus"
	"github.com/timmaaez/ichor/business/domain/permissions/userrolebus"
	"github.com/timmaaez/ichor/business/domain/users/userbus"
	"github.com/timmaaez/ichor/business/sdk/dbtest"
	"github.com/timmaaez/ichor/business/sdk/page"

	"github.com/timmaaez/ichor/foundation/security/auth"
)

func insertSeedData(db *dbtest.Database, ath *auth.Auth) (apitest.SeedData, error) {
	ctx := context.Background()
	busDomain := db.BusDomain

	// -------------------------------------------------------------------------
	// Users
	// -------------------------------------------------------------------------

	usrs, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users: %w", err)
	}
	tu1 := apitest.User{
		User:  usrs[0],
		Token: apitest.Token(busDomain.User, ath, usrs[0].Email.Address),
	}

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding admins: %w", err)
	}
	tu2 := apitest.User{
		User:  admins[0],
		Token: apitest.Token(busDomain.User, ath, admins[0].Email.Address),
	}

	// -------------------------------------------------------------------------
	// Cycle Count Sessions (only FK is created_by → users)
	// -------------------------------------------------------------------------

	createdByIDs := []uuid.UUID{tu2.ID}
	sessions, err := cyclecountsessionbus.TestSeedCycleCountSessions(ctx, 4, createdByIDs, busDomain.CycleCountSession)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding cycle count sessions: %w", err)
	}

	// -------------------------------------------------------------------------
	// Permissions: tu1 (User) = read-only on sessions table
	// -------------------------------------------------------------------------

	roles, err := rolebus.TestSeedRoles(ctx, 2, busDomain.Role)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding roles: %w", err)
	}

	roleIDs := make([]uuid.UUID, len(roles))
	for i, r := range roles {
		roleIDs[i] = r.ID
	}

	userIDs := []uuid.UUID{tu1.ID, tu2.ID}

	_, err = userrolebus.TestSeedUserRoles(ctx, userIDs, roleIDs, busDomain.UserRole)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding user roles: %w", err)
	}

	tas, err := tableaccessbus.TestSeedTableAccess(ctx, roleIDs, busDomain.TableAccess)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding table access: %w", err)
	}

	for _, ta := range tas {
		if ta.TableName == cyclecountsessionapi.RouteTable {
			update := tableaccessbus.UpdateTableAccess{
				CanCreate: dbtest.BoolPointer(false),
				CanUpdate: dbtest.BoolPointer(false),
				CanDelete: dbtest.BoolPointer(false),
				CanRead:   dbtest.BoolPointer(true),
			}
			if _, err := busDomain.TableAccess.Update(ctx, ta, update); err != nil {
				return apitest.SeedData{}, fmt.Errorf("updating table access: %w", err)
			}
		}
	}

	// -------------------------------------------------------------------------
	// Return
	// -------------------------------------------------------------------------

	return apitest.SeedData{
		Admins:             []apitest.User{tu2},
		Users:              []apitest.User{tu1},
		CycleCountSessions: cyclecountsessionapp.ToAppCycleCountSessions(sessions),
	}, nil
}
```

**IMPORTANT NOTE:** The `uuid` import is `"github.com/google/uuid"`. The `auth` import path should match what `picktaskapi/seed_test.go` uses — check the exact import path. It may be `"github.com/timmaaez/ichor/foundation/security/auth"` or similar. Copy the exact import from the picktask seed file.

- [ ] **Step 1b: Verify the seed file compiles**

Run:
```bash
go build ./api/cmd/services/ichor/tests/inventory/cyclecountsessionapi/...
```
Expected: Build succeeds (no test execution yet).

Fix any import path issues.

### Step 2: Create `cyclecountsession_test.go`

- [ ] **Step 2a: Write the entry point**

```go
package cyclecountsessionapi_test

import (
	"testing"

	"github.com/timmaaez/ichor/api/sdk/http/apitest"
)

func Test_CycleCountSession(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_CycleCountSession")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// Query
	test.Run(t, query200(sd), "query-200")
	test.Run(t, queryByID200(sd), "query-by-id-200")
	test.Run(t, queryByID404(sd), "query-by-id-404")

	// Create
	test.Run(t, create200(sd), "create-200")
	test.Run(t, create400(sd), "create-400")
	test.Run(t, create401(sd), "create-401")

	// Update
	test.Run(t, update200(sd), "update-200")
	test.Run(t, update400(sd), "update-400")
	test.Run(t, update401(sd), "update-401")
	test.Run(t, update404(sd), "update-404")

	// Delete
	test.Run(t, delete200(sd), "delete-200")
	test.Run(t, delete401(sd), "delete-401")
	test.Run(t, delete404(sd), "delete-404")
}
```

- [ ] **Step 2b: Verify it compiles (will fail — query200 etc don't exist yet)**

This won't compile until all test functions exist. Move to the next tasks and come back to verify the full build once all files are created.

---

## Task 2: Cycle Count Session — Query Tests

**Files:**
- Create: `api/cmd/services/ichor/tests/inventory/cyclecountsessionapi/query_test.go`

- [ ] **Step 1: Write query tests**

```go
package cyclecountsessionapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaez/ichor/api/sdk/http/apitest"
	"github.com/timmaaez/ichor/app/domain/inventory/cyclecountsessionapp"
	"github.com/timmaaez/ichor/app/sdk/errs"
	"github.com/timmaaez/ichor/app/sdk/query"
)

func query200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "all",
			URL:        "/v1/inventory/cycle-count-sessions?rows=10&page=1",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			GotResp:    &query.Result[cyclecountsessionapp.CycleCountSession]{},
			ExpResp: &query.Result[cyclecountsessionapp.CycleCountSession]{
				Items:       sd.CycleCountSessions,
				Total:       len(sd.CycleCountSessions),
				Page:        1,
				RowsPerPage: 10,
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func queryByID200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", sd.CycleCountSessions[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			GotResp:    &cyclecountsessionapp.CycleCountSession{},
			ExpResp:    &sd.CycleCountSessions[0],
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func queryByID404(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "not-found",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", uuid.NewString()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusNotFound,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.NotFound, "cycle count session not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
```

---

## Task 3: Cycle Count Session — Create Tests

**Files:**
- Create: `api/cmd/services/ichor/tests/inventory/cyclecountsessionapi/create_test.go`

- [ ] **Step 1: Write create tests**

```go
package cyclecountsessionapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/timmaaez/ichor/api/sdk/http/apitest"
	"github.com/timmaaez/ichor/app/domain/inventory/cyclecountsessionapp"
	"github.com/timmaaez/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/inventory/cycle-count-sessions",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &cyclecountsessionapp.NewCycleCountSession{
				Name: "Integration Test Session",
			},
			GotResp: &cyclecountsessionapp.CycleCountSession{},
			ExpResp: &cyclecountsessionapp.CycleCountSession{
				Name:          "Integration Test Session",
				Status:        "draft",
				CreatedBy:     sd.Admins[0].ID.String(),
				CompletedDate: "",
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*cyclecountsessionapp.CycleCountSession)
				expResp := exp.(*cyclecountsessionapp.CycleCountSession)

				// Server-assigned fields
				expResp.ID = gotResp.ID
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "missing-name",
			URL:        "/v1/inventory/cycle-count-sessions",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &cyclecountsessionapp.NewCycleCountSession{
				Name: "",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"name","error":"name is a required field"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "empty-token",
			URL:        "/v1/inventory/cycle-count-sessions",
			Token:      "&nbsp;",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad-sig",
			URL:        "/v1/inventory/cycle-count-sessions",
			Token:      sd.Admins[0].Token + "A",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "no-create-permission",
			URL:        "/v1/inventory/cycle-count-sessions",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Input: &cyclecountsessionapp.NewCycleCountSession{
				Name: "Should Fail",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "user does not have permission CREATE for table: inventory.cycle_count_sessions"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
```

**NOTE:** Session create has no FK besides `created_by` (injected from auth), so a 409 FK test is not applicable. Skip `create409`.

---

## Task 4: Cycle Count Session — Update Tests (Table-Driven)

**Files:**
- Create: `api/cmd/services/ichor/tests/inventory/cyclecountsessionapi/update_test.go`

- [ ] **Step 1: Write table-driven update tests**

```go
package cyclecountsessionapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaez/ichor/api/sdk/http/apitest"
	"github.com/timmaaez/ichor/app/domain/inventory/cyclecountsessionapp"
	"github.com/timmaaez/ichor/app/sdk/errs"
	"github.com/timmaaez/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "name-change",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", sd.CycleCountSessions[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &cyclecountsessionapp.UpdateCycleCountSession{
				Name: dbtest.StringPointer("Updated Session Name"),
			},
			GotResp: &cyclecountsessionapp.CycleCountSession{},
			ExpResp: &cyclecountsessionapp.CycleCountSession{
				ID:            sd.CycleCountSessions[0].ID,
				Name:          "Updated Session Name",
				Status:        "draft",
				CreatedBy:     sd.CycleCountSessions[0].CreatedBy,
				CompletedDate: "",
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*cyclecountsessionapp.CycleCountSession)
				expResp := exp.(*cyclecountsessionapp.CycleCountSession)
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate
				return cmp.Diff(gotResp, expResp)
			},
		},
		{
			Name:       "draft-to-in-progress",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", sd.CycleCountSessions[1].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &cyclecountsessionapp.UpdateCycleCountSession{
				Status: dbtest.StringPointer("in_progress"),
			},
			GotResp: &cyclecountsessionapp.CycleCountSession{},
			ExpResp: &cyclecountsessionapp.CycleCountSession{
				ID:            sd.CycleCountSessions[1].ID,
				Name:          sd.CycleCountSessions[1].Name,
				Status:        "in_progress",
				CreatedBy:     sd.CycleCountSessions[1].CreatedBy,
				CompletedDate: "",
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*cyclecountsessionapp.CycleCountSession)
				expResp := exp.(*cyclecountsessionapp.CycleCountSession)
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate
				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func update400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "invalid-status",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", sd.CycleCountSessions[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &cyclecountsessionapp.UpdateCycleCountSession{
				Status: dbtest.StringPointer("not_a_valid_status"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `parse status: invalid status "not_a_valid_status"`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func update401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "empty-token",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", sd.CycleCountSessions[0].ID),
			Token:      "&nbsp;",
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad-sig",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", sd.CycleCountSessions[0].ID),
			Token:      sd.Admins[0].Token + "A",
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "no-update-permission",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", sd.CycleCountSessions[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			Input: &cyclecountsessionapp.UpdateCycleCountSession{
				Name: dbtest.StringPointer("Should Fail"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "user does not have permission UPDATE for table: inventory.cycle_count_sessions"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func update404(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "not-found",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", uuid.NewString()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusNotFound,
			Input: &cyclecountsessionapp.UpdateCycleCountSession{
				Name: dbtest.StringPointer("Does Not Exist"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.NotFound, "cycle count session not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
```

---

## Task 5: Cycle Count Session — Standalone Multi-Step Update Tests

**Files:**
- Modify: `api/cmd/services/ichor/tests/inventory/cyclecountsessionapi/update_test.go` (append to file)

These are standalone test functions — each gets its own `apitest.StartTest` and `insertSeedData`.

- [ ] **Step 1: Add `TestUpdate200Cancel` — draft → cancelled**

Append to `update_test.go`:

```go
// TestUpdate200Cancel tests the draft → cancelled status transition.
func TestUpdate200Cancel(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_CycleCountSession_Cancel")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// Session[2] is in draft — cancel it.
	test.Run(t, []apitest.Table{
		{
			Name:       "draft-to-cancelled",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", sd.CycleCountSessions[2].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &cyclecountsessionapp.UpdateCycleCountSession{
				Status: dbtest.StringPointer("cancelled"),
			},
			GotResp: &cyclecountsessionapp.CycleCountSession{},
			ExpResp: &cyclecountsessionapp.CycleCountSession{
				ID:            sd.CycleCountSessions[2].ID,
				Name:          sd.CycleCountSessions[2].Name,
				Status:        "cancelled",
				CreatedBy:     sd.CycleCountSessions[2].CreatedBy,
				CompletedDate: "",
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*cyclecountsessionapp.CycleCountSession)
				expResp := exp.(*cyclecountsessionapp.CycleCountSession)
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate
				return cmp.Diff(gotResp, expResp)
			},
		},
	}, "cancel")
}
```

- [ ] **Step 2: Add `TestUpdate400TerminalState` — completing an already-completed session**

Append to `update_test.go`. This tests the TOCTOU scenario from the prompt: completing an already-completed session returns FailedPrecondition.

```go
// TestUpdate400TerminalState tests that transitioning from a terminal state returns FailedPrecondition.
func TestUpdate400TerminalState(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_CycleCountSession_TerminalState")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	session := sd.CycleCountSessions[2]

	// Step 1: Cancel the session (draft → cancelled).
	test.Run(t, []apitest.Table{
		{
			Name:       "cancel",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", session.ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &cyclecountsessionapp.UpdateCycleCountSession{
				Status: dbtest.StringPointer("cancelled"),
			},
			GotResp: &cyclecountsessionapp.CycleCountSession{},
			CmpFunc: func(got, exp any) string { return "" },
		},
	}, "cancel")

	// Step 2: Try to transition from cancelled → in_progress. Should fail.
	test.Run(t, []apitest.Table{
		{
			Name:       "transition-from-cancelled",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", session.ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &cyclecountsessionapp.UpdateCycleCountSession{
				Status: dbtest.StringPointer("in_progress"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.FailedPrecondition, "session is already cancelled and cannot be transitioned"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}, "terminal-state")
}
```

- [ ] **Step 3: Add `TestUpdate400CompleteFromDraft` — cannot complete a draft session**

Append to `update_test.go`:

```go
// TestUpdate400CompleteFromDraft tests that completing a session directly from draft returns FailedPrecondition.
func TestUpdate400CompleteFromDraft(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_CycleCountSession_CompleteFromDraft")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	session := sd.CycleCountSessions[2]

	// Try to complete directly from draft — should fail.
	test.Run(t, []apitest.Table{
		{
			Name:       "complete-from-draft",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", session.ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &cyclecountsessionapp.UpdateCycleCountSession{
				Status: dbtest.StringPointer("completed"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.FailedPrecondition, "session must be in_progress to complete, current status: draft"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}, "complete-from-draft")
}
```

---

## Task 6: Cycle Count Session — Complete Flow (Critical Integration Test)

**Files:**
- Modify: `api/cmd/services/ichor/tests/inventory/cyclecountsessionapi/update_test.go` (append to file)

This is the most important test. It exercises the full lifecycle: create session → add items → count items → approve variance → complete session → verify inventory adjustments are created AND approved.

- [ ] **Step 1: Add required imports to update_test.go**

Make sure the import block includes everything needed for the complete flow test. The file will need these additional imports beyond what's already there:

```go
import (
	"context"
	"testing"
	// ... (existing imports) ...
	"github.com/timmaaez/ichor/app/domain/inventory/cyclecountitemapp"
	"github.com/timmaaez/ichor/business/domain/inventory/inventoryadjustmentbus"
	"github.com/timmaaez/ichor/business/sdk/page"
)
```

- [ ] **Step 2: Add `TestUpdate200CompleteFlow`**

Append to `update_test.go`. This test needs its own seed data that includes the full dependency chain for items (products + inventory locations).

```go
// TestUpdate200CompleteFlow tests the full cycle count lifecycle:
// create session → add items → count items → approve variance → complete session
// → verify inventory adjustments created AND approved (not pending).
func TestUpdate200CompleteFlow(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_CycleCountSession_CompleteFlow")

	sd, err := insertCompleteFlowSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	session := sd.CycleCountSessions[0]

	// -------------------------------------------------------------------------
	// Step 1: Transition session from draft → in_progress
	// -------------------------------------------------------------------------
	test.Run(t, []apitest.Table{
		{
			Name:       "to-in-progress",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", session.ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &cyclecountsessionapp.UpdateCycleCountSession{
				Status: dbtest.StringPointer("in_progress"),
			},
			GotResp: &cyclecountsessionapp.CycleCountSession{},
			CmpFunc: func(got, exp any) string { return "" },
		},
	}, "step1-in-progress")

	// -------------------------------------------------------------------------
	// Step 2: Create a cycle count item for this session
	// -------------------------------------------------------------------------
	var createdItem cyclecountitemapp.CycleCountItem
	test.Run(t, []apitest.Table{
		{
			Name:       "create-item",
			URL:        "/v1/inventory/cycle-count-items",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &cyclecountitemapp.NewCycleCountItem{
				SessionID:      session.ID,
				ProductID:      sd.Products[0].ID,
				LocationID:     sd.InventoryLocations[0].ID,
				SystemQuantity: "100",
			},
			GotResp: &createdItem,
			CmpFunc: func(got, exp any) string { return "" },
		},
	}, "step2-create-item")

	// -------------------------------------------------------------------------
	// Step 3: Count the item (set counted_quantity, triggering auto-variance)
	// -------------------------------------------------------------------------
	test.Run(t, []apitest.Table{
		{
			Name:       "count-item",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-items/%s", createdItem.ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &cyclecountitemapp.UpdateCycleCountItem{
				CountedQuantity: dbtest.StringPointer("95"),
			},
			GotResp: &cyclecountitemapp.CycleCountItem{},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*cyclecountitemapp.CycleCountItem)
				// Verify auto-computed variance: 95 - 100 = -5
				if gotResp.Variance != "-5" {
					return fmt.Sprintf("expected variance -5, got %s", gotResp.Variance)
				}
				// Verify counted_by was auto-injected
				if gotResp.CountedBy == "" {
					return "expected counted_by to be auto-injected"
				}
				// Verify counted_date was auto-injected
				if gotResp.CountedDate == "" {
					return "expected counted_date to be auto-injected"
				}
				return ""
			},
		},
	}, "step3-count-item")

	// -------------------------------------------------------------------------
	// Step 4: Approve the variance (pending → variance_approved)
	// -------------------------------------------------------------------------
	test.Run(t, []apitest.Table{
		{
			Name:       "approve-variance",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-items/%s", createdItem.ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &cyclecountitemapp.UpdateCycleCountItem{
				Status: dbtest.StringPointer("variance_approved"),
			},
			GotResp: &cyclecountitemapp.CycleCountItem{},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*cyclecountitemapp.CycleCountItem)
				if gotResp.Status != "variance_approved" {
					return fmt.Sprintf("expected status variance_approved, got %s", gotResp.Status)
				}
				return ""
			},
		},
	}, "step4-approve-variance")

	// -------------------------------------------------------------------------
	// Step 5: Complete the session (with a simultaneous name change to verify it's preserved)
	// -------------------------------------------------------------------------
	test.Run(t, []apitest.Table{
		{
			Name:       "complete-session",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", session.ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &cyclecountsessionapp.UpdateCycleCountSession{
				Name:   dbtest.StringPointer("Final Session Name"),
				Status: dbtest.StringPointer("completed"),
			},
			GotResp: &cyclecountsessionapp.CycleCountSession{},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*cyclecountsessionapp.CycleCountSession)
				if gotResp.Status != "completed" {
					return fmt.Sprintf("expected status completed, got %s", gotResp.Status)
				}
				if gotResp.Name != "Final Session Name" {
					return fmt.Sprintf("expected name 'Final Session Name', got %s", gotResp.Name)
				}
				if gotResp.CompletedDate == "" {
					return "expected completed_date to be set"
				}
				return ""
			},
		},
	}, "step5-complete")

	// -------------------------------------------------------------------------
	// Step 6: Verify inventory adjustments were created AND approved
	// -------------------------------------------------------------------------
	ctx := context.Background()
	busDomain := test.DB.BusDomain

	reasonCode := inventoryadjustmentbus.ReasonCodeCycleCount
	filter := inventoryadjustmentbus.QueryFilter{
		ReasonCode: &reasonCode,
	}

	adjs, err := busDomain.InventoryAdjustment.Query(ctx, filter, inventoryadjustmentbus.DefaultOrderBy, page.MustParse("1", "10"))
	if err != nil {
		t.Fatalf("querying inventory adjustments: %s", err)
	}

	if len(adjs) == 0 {
		t.Fatal("expected at least one inventory adjustment to be created")
	}

	for _, adj := range adjs {
		if adj.ApprovalStatus != inventoryadjustmentbus.ApprovalStatusApproved {
			t.Errorf("expected adjustment %s to be approved, got %s", adj.ID, adj.ApprovalStatus)
		}
		if adj.ReasonCode != inventoryadjustmentbus.ReasonCodeCycleCount {
			t.Errorf("expected reason_code cycle_count, got %s", adj.ReasonCode)
		}
		if adj.QuantityChange != -5 {
			t.Errorf("expected quantity_change -5, got %d", adj.QuantityChange)
		}
	}

	// -------------------------------------------------------------------------
	// Step 7: Verify TOCTOU — completing again returns FailedPrecondition
	// -------------------------------------------------------------------------
	test.Run(t, []apitest.Table{
		{
			Name:       "already-completed",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", session.ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &cyclecountsessionapp.UpdateCycleCountSession{
				Status: dbtest.StringPointer("completed"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.FailedPrecondition, "session is already completed and cannot be transitioned"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}, "step7-already-completed")
}
```

- [ ] **Step 3: Add `insertCompleteFlowSeedData` to `seed_test.go`**

This seed function needs the extended dependency chain for items (products + inventory locations). Add this function to the bottom of `seed_test.go`:

```go
// insertCompleteFlowSeedData creates seed data for the complete flow test.
// This includes products and inventory locations needed for cycle count items.
func insertCompleteFlowSeedData(db *dbtest.Database, ath *auth.Auth) (apitest.SeedData, error) {
	ctx := context.Background()
	busDomain := db.BusDomain

	// -------------------------------------------------------------------------
	// Users
	// -------------------------------------------------------------------------

	usrs, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users: %w", err)
	}
	tu1 := apitest.User{
		User:  usrs[0],
		Token: apitest.Token(busDomain.User, ath, usrs[0].Email.Address),
	}

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding admins: %w", err)
	}
	tu2 := apitest.User{
		User:  admins[0],
		Token: apitest.Token(busDomain.User, ath, admins[0].Email.Address),
	}

	// -------------------------------------------------------------------------
	// Geography → Warehouse → Zones → Locations
	// -------------------------------------------------------------------------

	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying regions: %w", err)
	}
	regionIDs := make([]uuid.UUID, len(regions))
	for i, r := range regions {
		regionIDs[i] = r.ID
	}

	ctys, err := citybus.TestSeedCities(ctx, 1, regionIDs, busDomain.City)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding cities: %w", err)
	}
	ctyIDs := make([]uuid.UUID, len(ctys))
	for i, c := range ctys {
		ctyIDs[i] = c.ID
	}

	strs, err := streetbus.TestSeedStreets(ctx, 1, ctyIDs, busDomain.Street)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding streets: %w", err)
	}
	strIDs := make([]uuid.UUID, len(strs))
	for i, s := range strs {
		strIDs[i] = s.ID
	}

	warehouses, err := warehousebus.TestSeedWarehouses(ctx, 1, tu2.ID, strIDs, busDomain.Warehouse)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding warehouses: %w", err)
	}
	warehouseIDs := make([]uuid.UUID, len(warehouses))
	for i, w := range warehouses {
		warehouseIDs[i] = w.ID
	}

	zones, err := zonebus.TestSeedZone(ctx, 1, warehouseIDs, busDomain.Zones)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding zones: %w", err)
	}
	zoneIDs := make([]uuid.UUID, len(zones))
	for i, z := range zones {
		zoneIDs[i] = z.ID
	}

	inventoryLocations, err := inventorylocationbus.TestSeedInventoryLocations(ctx, 2, warehouseIDs, zoneIDs, busDomain.InventoryLocation)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding inventory locations: %w", err)
	}

	// -------------------------------------------------------------------------
	// Products
	// -------------------------------------------------------------------------

	tzs, err := busDomain.Timezone.Query(ctx, timezonebus.QueryFilter{}, timezonebus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying timezones: %w", err)
	}
	tzIDs := make([]uuid.UUID, len(tzs))
	for i, tz := range tzs {
		tzIDs[i] = tz.ID
	}

	contacts, err := contactinfosbus.TestSeedContactInfos(ctx, 1, strIDs, tzIDs, busDomain.ContactInfos)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding contacts: %w", err)
	}
	contactIDs := make([]uuid.UUID, len(contacts))
	for i, c := range contacts {
		contactIDs[i] = c.ID
	}

	brands, err := brandbus.TestSeedBrands(ctx, 1, contactIDs, busDomain.Brand)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding brands: %w", err)
	}
	brandIDs := make([]uuid.UUID, len(brands))
	for i, b := range brands {
		brandIDs[i] = b.ID
	}

	pcs, err := productcategorybus.TestSeedProductCategories(ctx, 1, busDomain.ProductCategory)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding product categories: %w", err)
	}
	pcIDs := make([]uuid.UUID, len(pcs))
	for i, pc := range pcs {
		pcIDs[i] = pc.ID
	}

	products, err := productbus.TestSeedProducts(ctx, 2, brandIDs, pcIDs, busDomain.Product)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding products: %w", err)
	}

	// -------------------------------------------------------------------------
	// Cycle Count Session (1 session for the complete flow)
	// -------------------------------------------------------------------------

	sessions, err := cyclecountsessionbus.TestSeedCycleCountSessions(ctx, 1, []uuid.UUID{tu2.ID}, busDomain.CycleCountSession)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding sessions: %w", err)
	}

	// -------------------------------------------------------------------------
	// Permissions
	// -------------------------------------------------------------------------

	roles, err := rolebus.TestSeedRoles(ctx, 2, busDomain.Role)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding roles: %w", err)
	}

	roleIDs := make([]uuid.UUID, len(roles))
	for i, r := range roles {
		roleIDs[i] = r.ID
	}

	_, err = userrolebus.TestSeedUserRoles(ctx, []uuid.UUID{tu1.ID, tu2.ID}, roleIDs, busDomain.UserRole)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding user roles: %w", err)
	}

	_, err = tableaccessbus.TestSeedTableAccess(ctx, roleIDs, busDomain.TableAccess)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding table access: %w", err)
	}

	// -------------------------------------------------------------------------
	// Return
	// -------------------------------------------------------------------------

	return apitest.SeedData{
		Admins:             []apitest.User{tu2},
		Users:              []apitest.User{tu1},
		CycleCountSessions: cyclecountsessionapp.ToAppCycleCountSessions(sessions),
		Products:           productapp.ToAppProducts(products),
		InventoryLocations: inventorylocationapp.ToAppInventoryLocations(inventoryLocations),
	}, nil
}
```

**IMPORTANT:** You'll need to add these imports to `seed_test.go`:
```go
import (
	// ... existing imports ...
	"github.com/timmaaez/ichor/app/domain/inventory/inventorylocationapp"
	"github.com/timmaaez/ichor/app/domain/inventory/productapp"
	"github.com/timmaaez/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaez/ichor/business/domain/geography/timezonebus"
	"github.com/timmaaez/ichor/business/domain/inventory/brandbus"
	"github.com/timmaaez/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaez/ichor/business/domain/inventory/productbus"
	"github.com/timmaaez/ichor/business/domain/inventory/productcategorybus"
	"github.com/timmaaez/ichor/business/domain/inventory/warehousebus"
	"github.com/timmaaez/ichor/business/domain/inventory/zonebus"
)
```

**NOTE:** Check the exact import paths by looking at the picktask seed file's imports. The `productapp` import might be at `app/domain/products/productapp` or similar. Same for `inventorylocationapp`. Verify against existing code.

---

## Task 7: Cycle Count Session — Delete Tests

**Files:**
- Create: `api/cmd/services/ichor/tests/inventory/cyclecountsessionapi/delete_test.go`

- [ ] **Step 1: Write delete tests**

```go
package cyclecountsessionapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaez/ichor/api/sdk/http/apitest"
	"github.com/timmaaez/ichor/app/sdk/errs"
	"github.com/timmaaez/ichor/business/sdk/dbtest"
)

func delete200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", sd.CycleCountSessions[3].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodDelete,
			StatusCode: http.StatusNoContent,
		},
	}
}

func delete401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "empty-token",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", sd.CycleCountSessions[3].ID),
			Token:      "&nbsp;",
			Method:     http.MethodDelete,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad-sig",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", sd.CycleCountSessions[3].ID),
			Token:      sd.Admins[0].Token + "A",
			Method:     http.MethodDelete,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "no-delete-permission",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", sd.CycleCountSessions[3].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodDelete,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission DELETE for table: inventory.cycle_count_sessions"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func delete404(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "not-found",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-sessions/%s", uuid.NewString()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodDelete,
			StatusCode: http.StatusNotFound,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.NotFound, "cycle count session not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
```

**NOTE:** `delete200` uses index `[3]` to avoid conflicts with other tests that mutate indices `[0]`, `[1]`, `[2]`. The `dbtest` import is needed only if you have `dbtest.StringPointer` calls — since delete has no body, you may not need it. Remove unused imports.

---

## Task 8: Cycle Count Session — Build and Test

- [ ] **Step 1: Build the package**

Run:
```bash
go build ./api/cmd/services/ichor/tests/inventory/cyclecountsessionapi/...
```
Expected: Build succeeds. Fix any import issues.

- [ ] **Step 2: Run the tests**

Run:
```bash
go test -v -count=1 ./api/cmd/services/ichor/tests/inventory/cyclecountsessionapi/...
```
Expected: All tests PASS. Fix any failures.

- [ ] **Step 3: Commit**

```bash
git add api/cmd/services/ichor/tests/inventory/cyclecountsessionapi/
git commit -m "test(cyclecountsession): add integration tests for cycle count session API"
```

---

## Task 9: Cycle Count Item — Entry Point and Seed Data

**Files:**
- Create: `api/cmd/services/ichor/tests/inventory/cyclecountitemapi/cyclecountitem_test.go`
- Create: `api/cmd/services/ichor/tests/inventory/cyclecountitemapi/seed_test.go`

### Step 1: Create `seed_test.go`

- [ ] **Step 1a: Write the seed file**

The item seed needs the full dependency chain: users → geography → warehouse → products → sessions → items → permissions.

```go
package cyclecountitemapi_test

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaez/ichor/api/domain/http/inventory/cyclecountitemapi"
	"github.com/timmaaez/ichor/api/sdk/http/apitest"
	"github.com/timmaaez/ichor/app/domain/inventory/cyclecountitemapp"
	"github.com/timmaaez/ichor/app/domain/inventory/cyclecountsessionapp"
	"github.com/timmaaez/ichor/app/domain/inventory/inventorylocationapp"
	"github.com/timmaaez/ichor/app/domain/products/productapp"
	"github.com/timmaaez/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaez/ichor/business/domain/geography/citybus"
	"github.com/timmaaez/ichor/business/domain/geography/regionbus"
	"github.com/timmaaez/ichor/business/domain/geography/streetbus"
	"github.com/timmaaez/ichor/business/domain/geography/timezonebus"
	"github.com/timmaaez/ichor/business/domain/inventory/brandbus"
	"github.com/timmaaez/ichor/business/domain/inventory/cyclecountitembus"
	"github.com/timmaaez/ichor/business/domain/inventory/cyclecountsessionbus"
	"github.com/timmaaez/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaez/ichor/business/domain/inventory/productbus"
	"github.com/timmaaez/ichor/business/domain/inventory/productcategorybus"
	"github.com/timmaaez/ichor/business/domain/inventory/warehousebus"
	"github.com/timmaaez/ichor/business/domain/inventory/zonebus"
	"github.com/timmaaez/ichor/business/domain/permissions/rolebus"
	"github.com/timmaaez/ichor/business/domain/permissions/tableaccessbus"
	"github.com/timmaaez/ichor/business/domain/permissions/userrolebus"
	"github.com/timmaaez/ichor/business/domain/users/userbus"
	"github.com/timmaaez/ichor/business/sdk/dbtest"
	"github.com/timmaaez/ichor/business/sdk/page"
	"github.com/timmaaez/ichor/foundation/security/auth"
)

func insertSeedData(db *dbtest.Database, ath *auth.Auth) (apitest.SeedData, error) {
	ctx := context.Background()
	busDomain := db.BusDomain

	// -------------------------------------------------------------------------
	// Users
	// -------------------------------------------------------------------------

	usrs, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding users: %w", err)
	}
	tu1 := apitest.User{
		User:  usrs[0],
		Token: apitest.Token(busDomain.User, ath, usrs[0].Email.Address),
	}

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding admins: %w", err)
	}
	tu2 := apitest.User{
		User:  admins[0],
		Token: apitest.Token(busDomain.User, ath, admins[0].Email.Address),
	}

	// -------------------------------------------------------------------------
	// Geography → Warehouse → Zones → Locations
	// -------------------------------------------------------------------------

	regions, err := busDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying regions: %w", err)
	}
	regionIDs := make([]uuid.UUID, len(regions))
	for i, r := range regions {
		regionIDs[i] = r.ID
	}

	ctys, err := citybus.TestSeedCities(ctx, 1, regionIDs, busDomain.City)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding cities: %w", err)
	}
	ctyIDs := make([]uuid.UUID, len(ctys))
	for i, c := range ctys {
		ctyIDs[i] = c.ID
	}

	strs, err := streetbus.TestSeedStreets(ctx, 1, ctyIDs, busDomain.Street)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding streets: %w", err)
	}
	strIDs := make([]uuid.UUID, len(strs))
	for i, s := range strs {
		strIDs[i] = s.ID
	}

	warehouses, err := warehousebus.TestSeedWarehouses(ctx, 1, tu2.ID, strIDs, busDomain.Warehouse)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding warehouses: %w", err)
	}
	warehouseIDs := make([]uuid.UUID, len(warehouses))
	for i, w := range warehouses {
		warehouseIDs[i] = w.ID
	}

	zones, err := zonebus.TestSeedZone(ctx, 1, warehouseIDs, busDomain.Zones)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding zones: %w", err)
	}
	zoneIDs := make([]uuid.UUID, len(zones))
	for i, z := range zones {
		zoneIDs[i] = z.ID
	}

	inventoryLocations, err := inventorylocationbus.TestSeedInventoryLocations(ctx, 2, warehouseIDs, zoneIDs, busDomain.InventoryLocation)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding inventory locations: %w", err)
	}
	locationIDs := make([]uuid.UUID, len(inventoryLocations))
	for i, loc := range inventoryLocations {
		locationIDs[i] = loc.ID
	}

	// -------------------------------------------------------------------------
	// Products
	// -------------------------------------------------------------------------

	tzs, err := busDomain.Timezone.Query(ctx, timezonebus.QueryFilter{}, timezonebus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("querying timezones: %w", err)
	}
	tzIDs := make([]uuid.UUID, len(tzs))
	for i, tz := range tzs {
		tzIDs[i] = tz.ID
	}

	contacts, err := contactinfosbus.TestSeedContactInfos(ctx, 1, strIDs, tzIDs, busDomain.ContactInfos)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding contacts: %w", err)
	}
	contactIDs := make([]uuid.UUID, len(contacts))
	for i, c := range contacts {
		contactIDs[i] = c.ID
	}

	brands, err := brandbus.TestSeedBrands(ctx, 1, contactIDs, busDomain.Brand)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding brands: %w", err)
	}
	brandIDs := make([]uuid.UUID, len(brands))
	for i, b := range brands {
		brandIDs[i] = b.ID
	}

	pcs, err := productcategorybus.TestSeedProductCategories(ctx, 1, busDomain.ProductCategory)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding product categories: %w", err)
	}
	pcIDs := make([]uuid.UUID, len(pcs))
	for i, pc := range pcs {
		pcIDs[i] = pc.ID
	}

	products, err := productbus.TestSeedProducts(ctx, 2, brandIDs, pcIDs, busDomain.Product)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding products: %w", err)
	}
	productIDs := make([]uuid.UUID, len(products))
	for i, p := range products {
		productIDs[i] = p.ID
	}

	// -------------------------------------------------------------------------
	// Sessions → Items
	// -------------------------------------------------------------------------

	sessions, err := cyclecountsessionbus.TestSeedCycleCountSessions(ctx, 1, []uuid.UUID{tu2.ID}, busDomain.CycleCountSession)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding sessions: %w", err)
	}
	sessionIDs := make([]uuid.UUID, len(sessions))
	for i, s := range sessions {
		sessionIDs[i] = s.ID
	}

	items, err := cyclecountitembus.TestSeedCycleCountItems(ctx, 4, sessionIDs, productIDs, locationIDs, busDomain.CycleCountItem)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding cycle count items: %w", err)
	}

	// -------------------------------------------------------------------------
	// Permissions
	// -------------------------------------------------------------------------

	roles, err := rolebus.TestSeedRoles(ctx, 2, busDomain.Role)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding roles: %w", err)
	}

	roleIDs := make([]uuid.UUID, len(roles))
	for i, r := range roles {
		roleIDs[i] = r.ID
	}

	_, err = userrolebus.TestSeedUserRoles(ctx, []uuid.UUID{tu1.ID, tu2.ID}, roleIDs, busDomain.UserRole)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding user roles: %w", err)
	}

	tas, err := tableaccessbus.TestSeedTableAccess(ctx, roleIDs, busDomain.TableAccess)
	if err != nil {
		return apitest.SeedData{}, fmt.Errorf("seeding table access: %w", err)
	}

	for _, ta := range tas {
		if ta.TableName == cyclecountitemapi.RouteTable {
			update := tableaccessbus.UpdateTableAccess{
				CanCreate: dbtest.BoolPointer(false),
				CanUpdate: dbtest.BoolPointer(false),
				CanDelete: dbtest.BoolPointer(false),
				CanRead:   dbtest.BoolPointer(true),
			}
			if _, err := busDomain.TableAccess.Update(ctx, ta, update); err != nil {
				return apitest.SeedData{}, fmt.Errorf("updating table access: %w", err)
			}
		}
	}

	// -------------------------------------------------------------------------
	// Return
	// -------------------------------------------------------------------------

	return apitest.SeedData{
		Admins:             []apitest.User{tu2},
		Users:              []apitest.User{tu1},
		CycleCountSessions: cyclecountsessionapp.ToAppCycleCountSessions(sessions),
		CycleCountItems:    cyclecountitemapp.ToAppCycleCountItems(items),
		Products:           productapp.ToAppProducts(products),
		InventoryLocations: inventorylocationapp.ToAppInventoryLocations(inventoryLocations),
	}, nil
}
```

**IMPORTANT:** Verify the exact import paths for `productapp` and `inventorylocationapp` by checking the picktask seed imports. They may be under `app/domain/products/productapp` or `app/domain/inventory/productapp` — check the actual codebase. Same for `warehousebus` — it might be under `business/domain/warehouse/warehousebus` not `business/domain/inventory/warehousebus`.

- [ ] **Step 1b: Create `cyclecountitem_test.go`**

```go
package cyclecountitemapi_test

import (
	"testing"

	"github.com/timmaaez/ichor/api/sdk/http/apitest"
)

func Test_CycleCountItem(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_CycleCountItem")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// Query
	test.Run(t, query200(sd), "query-200")
	test.Run(t, queryByID200(sd), "query-by-id-200")
	test.Run(t, queryByID404(sd), "query-by-id-404")

	// Create
	test.Run(t, create200(sd), "create-200")
	test.Run(t, create400(sd), "create-400")
	test.Run(t, create401(sd), "create-401")
	test.Run(t, create409(sd), "create-409")

	// Update
	test.Run(t, update200(sd), "update-200")
	test.Run(t, update400(sd), "update-400")
	test.Run(t, update401(sd), "update-401")
	test.Run(t, update404(sd), "update-404")

	// Delete
	test.Run(t, delete200(sd), "delete-200")
	test.Run(t, delete401(sd), "delete-401")
	test.Run(t, delete404(sd), "delete-404")
}
```

---

## Task 10: Cycle Count Item — Query Tests

**Files:**
- Create: `api/cmd/services/ichor/tests/inventory/cyclecountitemapi/query_test.go`

- [ ] **Step 1: Write query tests**

```go
package cyclecountitemapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaez/ichor/api/sdk/http/apitest"
	"github.com/timmaaez/ichor/app/domain/inventory/cyclecountitemapp"
	"github.com/timmaaez/ichor/app/sdk/errs"
	"github.com/timmaaez/ichor/app/sdk/query"
)

func query200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "all",
			URL:        "/v1/inventory/cycle-count-items?rows=10&page=1",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			GotResp:    &query.Result[cyclecountitemapp.CycleCountItem]{},
			ExpResp: &query.Result[cyclecountitemapp.CycleCountItem]{
				Items:       sd.CycleCountItems,
				Total:       len(sd.CycleCountItems),
				Page:        1,
				RowsPerPage: 10,
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func queryByID200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-items/%s", sd.CycleCountItems[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusOK,
			GotResp:    &cyclecountitemapp.CycleCountItem{},
			ExpResp:    &sd.CycleCountItems[0],
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func queryByID404(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "not-found",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-items/%s", uuid.NewString()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodGet,
			StatusCode: http.StatusNotFound,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.NotFound, "cycle count item not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
```

---

## Task 11: Cycle Count Item — Create Tests

**Files:**
- Create: `api/cmd/services/ichor/tests/inventory/cyclecountitemapi/create_test.go`

- [ ] **Step 1: Write create tests**

```go
package cyclecountitemapi_test

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaez/ichor/api/sdk/http/apitest"
	"github.com/timmaaez/ichor/app/domain/inventory/cyclecountitemapp"
	"github.com/timmaaez/ichor/app/sdk/errs"
)

func create200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/inventory/cycle-count-items",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &cyclecountitemapp.NewCycleCountItem{
				SessionID:      sd.CycleCountSessions[0].ID,
				ProductID:      sd.Products[0].ID,
				LocationID:     sd.InventoryLocations[0].ID,
				SystemQuantity: "50",
			},
			GotResp: &cyclecountitemapp.CycleCountItem{},
			ExpResp: &cyclecountitemapp.CycleCountItem{
				SessionID:       sd.CycleCountSessions[0].ID,
				ProductID:       sd.Products[0].ID,
				LocationID:      sd.InventoryLocations[0].ID,
				SystemQuantity:  "50",
				CountedQuantity: "",
				Variance:        "",
				Status:          "pending",
				CountedBy:       "",
				CountedDate:     "",
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*cyclecountitemapp.CycleCountItem)
				expResp := exp.(*cyclecountitemapp.CycleCountItem)

				// Server-assigned fields
				expResp.ID = gotResp.ID
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func create400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "missing-session-id",
			URL:        "/v1/inventory/cycle-count-items",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &cyclecountitemapp.NewCycleCountItem{
				SessionID:      "",
				ProductID:      sd.Products[0].ID,
				LocationID:     sd.InventoryLocations[0].ID,
				SystemQuantity: "50",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"sessionId","error":"sessionId is a required field"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-product-id",
			URL:        "/v1/inventory/cycle-count-items",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &cyclecountitemapp.NewCycleCountItem{
				SessionID:      sd.CycleCountSessions[0].ID,
				ProductID:      "",
				LocationID:     sd.InventoryLocations[0].ID,
				SystemQuantity: "50",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"productId","error":"productId is a required field"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-location-id",
			URL:        "/v1/inventory/cycle-count-items",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &cyclecountitemapp.NewCycleCountItem{
				SessionID:      sd.CycleCountSessions[0].ID,
				ProductID:      sd.Products[0].ID,
				LocationID:     "",
				SystemQuantity: "50",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"locationId","error":"locationId is a required field"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "missing-system-quantity",
			URL:        "/v1/inventory/cycle-count-items",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusBadRequest,
			Input: &cyclecountitemapp.NewCycleCountItem{
				SessionID:      sd.CycleCountSessions[0].ID,
				ProductID:      sd.Products[0].ID,
				LocationID:     sd.InventoryLocations[0].ID,
				SystemQuantity: "",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `validate: [{"field":"systemQuantity","error":"systemQuantity is a required field"}]`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "empty-token",
			URL:        "/v1/inventory/cycle-count-items",
			Token:      "&nbsp;",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad-sig",
			URL:        "/v1/inventory/cycle-count-items",
			Token:      sd.Admins[0].Token + "A",
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "no-create-permission",
			URL:        "/v1/inventory/cycle-count-items",
			Token:      sd.Users[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusUnauthorized,
			Input: &cyclecountitemapp.NewCycleCountItem{
				SessionID:      sd.CycleCountSessions[0].ID,
				ProductID:      sd.Products[0].ID,
				LocationID:     sd.InventoryLocations[0].ID,
				SystemQuantity: "50",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "user does not have permission CREATE for table: inventory.cycle_count_items"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create409(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "fk-violation-bad-session",
			URL:        "/v1/inventory/cycle-count-items",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusConflict,
			Input: &cyclecountitemapp.NewCycleCountItem{
				SessionID:      uuid.NewString(),
				ProductID:      sd.Products[0].ID,
				LocationID:     sd.InventoryLocations[0].ID,
				SystemQuantity: "50",
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Aborted, "create: namedexeccontext: foreign key violation"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
```

---

## Task 12: Cycle Count Item — Update Tests

**Files:**
- Create: `api/cmd/services/ichor/tests/inventory/cyclecountitemapi/update_test.go`

- [ ] **Step 1: Write update tests**

```go
package cyclecountitemapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaez/ichor/api/sdk/http/apitest"
	"github.com/timmaaez/ichor/app/domain/inventory/cyclecountitemapp"
	"github.com/timmaaez/ichor/app/sdk/errs"
	"github.com/timmaaez/ichor/business/sdk/dbtest"
)

func update200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "count-item",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-items/%s", sd.CycleCountItems[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &cyclecountitemapp.UpdateCycleCountItem{
				CountedQuantity: dbtest.StringPointer("8"),
			},
			GotResp: &cyclecountitemapp.CycleCountItem{},
			ExpResp: &cyclecountitemapp.CycleCountItem{
				ID:              sd.CycleCountItems[0].ID,
				SessionID:       sd.CycleCountItems[0].SessionID,
				ProductID:       sd.CycleCountItems[0].ProductID,
				LocationID:      sd.CycleCountItems[0].LocationID,
				SystemQuantity:  sd.CycleCountItems[0].SystemQuantity,
				CountedQuantity: "8",
				Status:          sd.CycleCountItems[0].Status,
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*cyclecountitemapp.CycleCountItem)
				expResp := exp.(*cyclecountitemapp.CycleCountItem)

				// Server-assigned fields
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate

				// Auto-injected fields when counted_quantity is set
				expResp.CountedBy = gotResp.CountedBy
				expResp.CountedDate = gotResp.CountedDate

				// Auto-computed variance: 8 - systemQuantity
				// SystemQuantity was set by seed (i+1)*10, so items[0] = 10
				// Variance = 8 - 10 = -2
				expResp.Variance = gotResp.Variance

				return cmp.Diff(gotResp, expResp)
			},
		},
		{
			Name:       "status-to-counted",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-items/%s", sd.CycleCountItems[1].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusOK,
			Input: &cyclecountitemapp.UpdateCycleCountItem{
				Status: dbtest.StringPointer("counted"),
			},
			GotResp: &cyclecountitemapp.CycleCountItem{},
			ExpResp: &cyclecountitemapp.CycleCountItem{
				ID:              sd.CycleCountItems[1].ID,
				SessionID:       sd.CycleCountItems[1].SessionID,
				ProductID:       sd.CycleCountItems[1].ProductID,
				LocationID:      sd.CycleCountItems[1].LocationID,
				SystemQuantity:  sd.CycleCountItems[1].SystemQuantity,
				CountedQuantity: "",
				Variance:        "",
				Status:          "counted",
				CountedBy:       "",
				CountedDate:     "",
			},
			CmpFunc: func(got, exp any) string {
				gotResp := got.(*cyclecountitemapp.CycleCountItem)
				expResp := exp.(*cyclecountitemapp.CycleCountItem)
				expResp.CreatedDate = gotResp.CreatedDate
				expResp.UpdatedDate = gotResp.UpdatedDate
				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func update400(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "invalid-status",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-items/%s", sd.CycleCountItems[0].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusBadRequest,
			Input: &cyclecountitemapp.UpdateCycleCountItem{
				Status: dbtest.StringPointer("not_a_valid_status"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.InvalidArgument, `parse status: invalid status "not_a_valid_status"`),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func update401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "empty-token",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-items/%s", sd.CycleCountItems[0].ID),
			Token:      "&nbsp;",
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad-sig",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-items/%s", sd.CycleCountItems[0].ID),
			Token:      sd.Admins[0].Token + "A",
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "no-update-permission",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-items/%s", sd.CycleCountItems[0].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusUnauthorized,
			Input: &cyclecountitemapp.UpdateCycleCountItem{
				CountedQuantity: dbtest.StringPointer("5"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.Unauthenticated, "user does not have permission UPDATE for table: inventory.cycle_count_items"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func update404(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "not-found",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-items/%s", uuid.NewString()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPut,
			StatusCode: http.StatusNotFound,
			Input: &cyclecountitemapp.UpdateCycleCountItem{
				CountedQuantity: dbtest.StringPointer("5"),
			},
			GotResp: &errs.Error{},
			ExpResp: errs.Newf(errs.NotFound, "cycle count item not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
```

---

## Task 13: Cycle Count Item — Delete Tests

**Files:**
- Create: `api/cmd/services/ichor/tests/inventory/cyclecountitemapi/delete_test.go`

- [ ] **Step 1: Write delete tests**

```go
package cyclecountitemapi_test

import (
	"fmt"
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaez/ichor/api/sdk/http/apitest"
	"github.com/timmaaez/ichor/app/sdk/errs"
)

func delete200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-items/%s", sd.CycleCountItems[3].ID),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodDelete,
			StatusCode: http.StatusNoContent,
		},
	}
}

func delete401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "empty-token",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-items/%s", sd.CycleCountItems[3].ID),
			Token:      "&nbsp;",
			Method:     http.MethodDelete,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "error parsing token: token contains an invalid number of segments"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad-sig",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-items/%s", sd.CycleCountItems[3].ID),
			Token:      sd.Admins[0].Token + "A",
			Method:     http.MethodDelete,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "authentication failed : bindings results[[{[true] map[x:false]}]] ok[true]"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "no-delete-permission",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-items/%s", sd.CycleCountItems[3].ID),
			Token:      sd.Users[0].Token,
			Method:     http.MethodDelete,
			StatusCode: http.StatusUnauthorized,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.Unauthenticated, "user does not have permission DELETE for table: inventory.cycle_count_items"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func delete404(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:       "not-found",
			URL:        fmt.Sprintf("/v1/inventory/cycle-count-items/%s", uuid.NewString()),
			Token:      sd.Admins[0].Token,
			Method:     http.MethodDelete,
			StatusCode: http.StatusNotFound,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.NotFound, "cycle count item not found"),
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
```

---

## Task 14: Cycle Count Item — Build and Test

- [ ] **Step 1: Build the package**

Run:
```bash
go build ./api/cmd/services/ichor/tests/inventory/cyclecountitemapi/...
```
Expected: Build succeeds. Fix any import issues.

- [ ] **Step 2: Run the tests**

Run:
```bash
go test -v -count=1 ./api/cmd/services/ichor/tests/inventory/cyclecountitemapi/...
```
Expected: All tests PASS. Fix any failures.

- [ ] **Step 3: Commit**

```bash
git add api/cmd/services/ichor/tests/inventory/cyclecountitemapi/
git commit -m "test(cyclecountitem): add integration tests for cycle count item API"
```

---

## Task 15: Final Verification

- [ ] **Step 1: Run both test suites together**

Run:
```bash
go test -v -count=1 ./api/cmd/services/ichor/tests/inventory/cyclecountsessionapi/... ./api/cmd/services/ichor/tests/inventory/cyclecountitemapi/...
```
Expected: All tests PASS.

- [ ] **Step 2: Verify go build for the entire project**

Run:
```bash
go build ./...
```
Expected: Build succeeds with no errors.

- [ ] **Step 3: Final commit (if any fixes were needed)**

Only if changes were required during verification.

---

## Import Path Verification Checklist

Before writing any file, verify these import paths by checking the picktask seed file or grepping the codebase:

- [ ] `auth` package: `foundation/security/auth` or `business/sdk/auth`? → Check `picktaskapi/seed_test.go`
- [ ] `productapp`: `app/domain/products/productapp` or `app/domain/inventory/productapp`? → Check `picktaskapi/seed_test.go`
- [ ] `inventorylocationapp`: path? → Check `picktaskapi/seed_test.go`
- [ ] `warehousebus`: `business/domain/warehouse/warehousebus` or `business/domain/inventory/warehousebus`? → Check `picktaskapi/seed_test.go`
- [ ] `zonebus`: same question → Check `picktaskapi/seed_test.go`
- [ ] `brandbus`: path? → Check `picktaskapi/seed_test.go`
- [ ] `timezonebus`: path? → Check `picktaskapi/seed_test.go`
- [ ] `contactinfosbus`: path? → Check `picktaskapi/seed_test.go`
- [ ] `dbtest.StringPointer` exists? Or is it `dbtest.StrPointer`? → Check usage in picktask tests
- [ ] `dbtest.IntPointer` exists? → If not, write `func intPtr(i int) *int { return &i }` locally
- [ ] `query.Result` — exact import path for `query` package
- [ ] `inventoryadjustmentbus.DefaultOrderBy` exists? → Check the order.go file for `DefaultOrderBy`

These paths use `timmaaez` in the plan as written — the executing agent MUST verify each import against the actual codebase before writing files. Copy exact imports from `picktaskapi/seed_test.go`.
