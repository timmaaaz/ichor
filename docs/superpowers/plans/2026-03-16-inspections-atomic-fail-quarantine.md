# Inspections Atomic Fail + Quarantine Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.
>
> **Worktree:** Create a worktree before executing: `create a worktree for blocker-008-inspections and execute this plan`

**Goal:** Add a composite `POST /v1/inventory/quality-inspections/{id}/fail` endpoint that atomically fails an inspection and (optionally) quarantines the associated lot in a single transaction.

**Architecture:** Extend `inspectionapp.App` with `lotTrackingsBus` and `db *sqlx.DB`. The `Fail` method runs a 2-way atomic write (update inspection status + update lot quality_status) inside a ReadCommitted transaction. Also add a CHECK constraint on inspection status via migration v2.16.

**Tech Stack:** Go 1.23, PostgreSQL, Ardan Labs service architecture

---

## Step 1: Migration — Add CHECK constraint on inspection status

**File:** `business/sdk/migrate/sql/migrate.sql`

Append after the v2.14 block (line ~2259):

```sql
-- Version: 2.16
-- Description: Add CHECK constraint on inspection status values
ALTER TABLE inventory.quality_inspections
    ADD CONSTRAINT quality_inspections_status_check
    CHECK (status IN ('pending', 'passed', 'failed'));
```

- [ ] Add migration v2.16 to `business/sdk/migrate/sql/migrate.sql`
- [ ] Verify: `grep 'Version: 2.16' business/sdk/migrate/sql/migrate.sql`

**Commit:** `feat(migration): add CHECK constraint on inspection status (v2.16)`

---

## Step 2: App-layer model — Add FailInspection request + response

**File:** `app/domain/inventory/inspectionapp/model.go`

Add after the `UpdateInspection` block (~line 207):

```go
// FailInspection represents the request to fail an inspection and optionally
// quarantine the associated lot.
type FailInspection struct {
	Notes         string `json:"notes" validate:"omitempty"`
	QuarantineLot bool   `json:"quarantine_lot"`
}

func (app *FailInspection) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app FailInspection) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

// FailInspectionResult is the composite response returned after atomically
// failing an inspection and (optionally) quarantining the lot.
type FailInspectionResult struct {
	Inspection Inspection `json:"inspection"`
	LotStatus  string     `json:"lot_status"`
}

func (r FailInspectionResult) Encode() ([]byte, string, error) {
	data, err := json.Marshal(r)
	return data, "application/json", err
}
```

- [ ] Add `FailInspection`, `FailInspectionResult` types to `app/domain/inventory/inspectionapp/model.go`
- [ ] Verify: `go build ./app/domain/inventory/inspectionapp/...`

**Commit:** `feat(inspectionapp): add FailInspection request and result models`

---

## Step 3: App-layer — Extend App struct and add Fail method

**File:** `app/domain/inventory/inspectionapp/inspectionapp.go`

### 3a. Add imports and extend struct

Add new imports:
```go
"database/sql"

"github.com/jmoiron/sqlx"
"github.com/timmaaaz/ichor/business/domain/inventory/lottrackingsbus"
```

Replace the `App` struct and constructors:

```go
type App struct {
	inspectionbus  *inspectionbus.Business
	lotTrackingsBus *lottrackingsbus.Business
	db             *sqlx.DB
	auth           *auth.Auth
}

func NewApp(inspectionbus *inspectionbus.Business) *App {
	return &App{
		inspectionbus: inspectionbus,
	}
}

func NewAppWithAuth(inspectionbus *inspectionbus.Business, auth *auth.Auth) *App {
	return &App{
		inspectionbus: inspectionbus,
		auth:          auth,
	}
}

// NewAppWithTx constructs an App with dependencies needed for transactional
// composite operations (e.g., fail + quarantine).
func NewAppWithTx(inspectionbus *inspectionbus.Business, lotTrackingsBus *lottrackingsbus.Business, db *sqlx.DB) *App {
	return &App{
		inspectionbus:   inspectionbus,
		lotTrackingsBus: lotTrackingsBus,
		db:              db,
	}
}
```

### 3b. Add Fail method

```go
// Fail atomically fails an inspection and optionally quarantines the associated lot.
// Both writes happen inside a single ReadCommitted transaction.
func (a *App) Fail(ctx context.Context, id uuid.UUID, app FailInspection) (FailInspectionResult, error) {
	// 1. Look up the inspection.
	inspection, err := a.inspectionbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, inspectionbus.ErrNotFound) {
			return FailInspectionResult{}, errs.New(errs.NotFound, err)
		}
		return FailInspectionResult{}, fmt.Errorf("fail [querybyid]: %w", err)
	}

	// Guard: cannot fail an already-failed inspection.
	if inspection.Status == "failed" {
		return FailInspectionResult{}, errs.Newf(errs.FailedPrecondition, "inspection is already failed")
	}

	// 2. Begin transaction.
	tx, err := a.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return FailInspectionResult{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 3. Update inspection status to "failed" inside the transaction.
	inspBusTx, err := a.inspectionbus.NewWithTx(tx)
	if err != nil {
		return FailInspectionResult{}, fmt.Errorf("new inspection tx: %w", err)
	}

	failedStatus := "failed"
	updateInsp := inspectionbus.UpdateInspection{
		Status: &failedStatus,
		Notes:  &app.Notes,
	}

	updated, err := inspBusTx.Update(ctx, inspection, updateInsp)
	if err != nil {
		return FailInspectionResult{}, fmt.Errorf("update inspection: %w", err)
	}

	// 4. Optionally quarantine the lot.
	lotStatus := ""
	if app.QuarantineLot {
		lot, err := a.lotTrackingsBus.QueryByID(ctx, inspection.LotID)
		if err != nil {
			if errors.Is(err, lottrackingsbus.ErrNotFound) {
				return FailInspectionResult{}, errs.Newf(errs.NotFound, "lot %s not found", inspection.LotID)
			}
			return FailInspectionResult{}, fmt.Errorf("query lot: %w", err)
		}

		ltBusTx, err := a.lotTrackingsBus.NewWithTx(tx)
		if err != nil {
			return FailInspectionResult{}, fmt.Errorf("new lottrackings tx: %w", err)
		}

		quarantined := "quarantined"
		updateLot := lottrackingsbus.UpdateLotTrackings{
			QualityStatus: &quarantined,
		}

		lot, err = ltBusTx.Update(ctx, lot, updateLot)
		if err != nil {
			return FailInspectionResult{}, fmt.Errorf("quarantine lot: %w", err)
		}

		lotStatus = lot.QualityStatus
	}

	// 5. Commit.
	if err := tx.Commit(); err != nil {
		return FailInspectionResult{}, fmt.Errorf("commit transaction: %w", err)
	}

	return FailInspectionResult{
		Inspection: ToAppInspection(updated),
		LotStatus:  lotStatus,
	}, nil
}
```

- [ ] Add `database/sql`, `github.com/jmoiron/sqlx`, `lottrackingsbus` imports
- [ ] Add `lotTrackingsBus *lottrackingsbus.Business` and `db *sqlx.DB` fields to `App` struct
- [ ] Add `NewAppWithTx` constructor
- [ ] Add `Fail` method
- [ ] Verify: `go build ./app/domain/inventory/inspectionapp/...`

**Commit:** `feat(inspectionapp): add atomic Fail method with quarantine support`

---

## Step 4: API-layer — Add fail handler and route

### 4a. Handler

**File:** `api/domain/http/inventory/inspectionapi/inspectionapi.go`

Add after the `queryByID` handler:

```go
func (api *api) fail(ctx context.Context, r *http.Request) web.Encoder {
	var app inspectionapp.FailInspection
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	inspectionID := web.Param(r, "inspection_id")
	parsed, err := uuid.Parse(inspectionID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	result, err := api.inspectionapp.Fail(ctx, parsed, app)
	if err != nil {
		return errs.NewError(err)
	}

	return result
}
```

- [ ] Add `fail` handler to `api/domain/http/inventory/inspectionapi/inspectionapi.go`

### 4b. Route registration

**File:** `api/domain/http/inventory/inspectionapi/routes.go`

Add imports for `lottrackingsbus` and `sqlx`. Update `Config`:

```go
type Config struct {
	Log             *logger.Logger
	InspectionBus   *inspectionbus.Business
	LotTrackingsBus *lottrackingsbus.Business
	DB              *sqlx.DB
	AuthClient      *authclient.Client
	PermissionsBus  *permissionsbus.Business
}
```

Update the `Routes` function to use `NewAppWithTx` and register the fail route:

```go
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	api := newAPI(inspectionapp.NewAppWithTx(cfg.InspectionBus, cfg.LotTrackingsBus, cfg.DB))

	// ... existing routes unchanged ...

	app.HandlerFunc(http.MethodPost, version, "/inventory/quality-inspections/{inspection_id}/fail", api.fail, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Update, auth.RuleAny))
}
```

- [ ] Add `LotTrackingsBus` and `DB` to `Config` struct in `routes.go`
- [ ] Add `lottrackingsbus` and `sqlx` imports to `routes.go`
- [ ] Change `NewApp` to `NewAppWithTx` call in `Routes`
- [ ] Add `POST .../fail` route
- [ ] Verify: `go build ./api/domain/http/inventory/inspectionapi/...`

**Commit:** `feat(inspectionapi): add POST /fail route and handler`

---

## Step 5: Wiring — Update all.go and crud.go

### 5a. all.go

**File:** `api/cmd/services/ichor/build/all/all.go` (line ~1026)

Update the `inspectionapi.Routes` call to pass the new dependencies:

```go
inspectionapi.Routes(app, inspectionapi.Config{
	InspectionBus:   inspectionBus,
	LotTrackingsBus: lotTrackingsBus,
	DB:              cfg.DB,
	AuthClient:      cfg.AuthClient,
	Log:             cfg.Log,
	PermissionsBus:  permissionsBus,
})
```

`lotTrackingsBus` is already instantiated at line ~420. `cfg.DB` is the existing `*sqlx.DB`.

### 5b. crud.go

**File:** `api/cmd/services/ichor/build/crud/crud.go` (line ~539)

Same update — add `LotTrackingsBus` and `DB` fields. `lotTrackingsBus` should already be instantiated in crud.go; if not, instantiate it following the same pattern as all.go.

- [ ] Update `inspectionapi.Routes` call in `all/all.go` with `LotTrackingsBus` and `DB`
- [ ] Update `inspectionapi.Routes` call in `crud/crud.go` with `LotTrackingsBus` and `DB`
- [ ] Add `lottrackingsbus` import if not already present in crud.go
- [ ] Verify: `go build ./api/cmd/services/ichor/build/...`

**Commit:** `feat(build): wire lotTrackingsBus and DB into inspectionapi config`

---

## Step 6: Integration test — fail endpoint

**File:** `api/cmd/services/ichor/tests/inventory/inspectionapi/fail_test.go` (new file)

```go
package inspectionapi_test

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/inventory/inspectionapp"
)

func fail200(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:  "fail-with-quarantine",
			Token: sd.Admins[0].Token,
			URL:   "/v1/inventory/quality-inspections/" + sd.Inspections[0].InspectionID + "/fail",
			Method: http.MethodPost,
			Body: &inspectionapp.FailInspection{
				Notes:         "Contamination detected",
				QuarantineLot: true,
			},
			StatusCode: http.StatusOK,
			GotResp:    &inspectionapp.FailInspectionResult{},
			CmpFunc: func(got any, exp any) string {
				result := got.(*inspectionapp.FailInspectionResult)
				if result.Inspection.Status != "failed" {
					return "expected inspection status 'failed', got '" + result.Inspection.Status + "'"
				}
				if result.LotStatus != "quarantined" {
					return "expected lot_status 'quarantined', got '" + result.LotStatus + "'"
				}
				return ""
			},
		},
	}
}

func fail200NoQuarantine(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:  "fail-without-quarantine",
			Token: sd.Admins[0].Token,
			URL:   "/v1/inventory/quality-inspections/" + sd.Inspections[1].InspectionID + "/fail",
			Method: http.MethodPost,
			Body: &inspectionapp.FailInspection{
				Notes:         "Minor defect noted",
				QuarantineLot: false,
			},
			StatusCode: http.StatusOK,
			GotResp:    &inspectionapp.FailInspectionResult{},
			CmpFunc: func(got any, exp any) string {
				result := got.(*inspectionapp.FailInspectionResult)
				if result.Inspection.Status != "failed" {
					return "expected inspection status 'failed', got '" + result.Inspection.Status + "'"
				}
				if result.LotStatus != "" {
					return "expected empty lot_status, got '" + result.LotStatus + "'"
				}
				return ""
			},
		},
	}
}

func fail401(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:  "fail-unauthorized",
			Token: sd.Users[0].Token,
			URL:   "/v1/inventory/quality-inspections/" + sd.Inspections[2].InspectionID + "/fail",
			Method: http.MethodPost,
			Body: &inspectionapp.FailInspection{
				Notes:         "Should not work",
				QuarantineLot: true,
			},
			StatusCode: http.StatusForbidden,
		},
	}
}

func fail404(sd apitest.SeedData) []apitest.Table {
	return []apitest.Table{
		{
			Name:  "fail-not-found",
			Token: sd.Admins[0].Token,
			URL:   "/v1/inventory/quality-inspections/00000000-0000-0000-0000-000000000000/fail",
			Method: http.MethodPost,
			Body: &inspectionapp.FailInspection{
				Notes:         "Does not exist",
				QuarantineLot: false,
			},
			StatusCode: http.StatusNotFound,
		},
	}
}
```

### Register in inspectionapi_test.go

Add to `Test_Inspections`:

```go
test.Run(t, fail200(sd), "fail-200")
test.Run(t, fail200NoQuarantine(sd), "fail-200-no-quarantine")
test.Run(t, fail401(sd), "fail-401")
test.Run(t, fail404(sd), "fail-404")
```

- [ ] Create `api/cmd/services/ichor/tests/inventory/inspectionapi/fail_test.go`
- [ ] Add fail test runs to `inspectionapi_test.go`
- [ ] Verify: `go build ./api/cmd/services/ichor/tests/inventory/inspectionapi/...`
- [ ] Run: `go test ./api/cmd/services/ichor/tests/inventory/inspectionapi/... -run Test_Inspections/fail -v -count=1`

**Commit:** `test(inspectionapi): add integration tests for fail+quarantine endpoint`

---

## Step 7: Verify everything builds and passes

- [ ] `go build ./...` from repo root (full build)
- [ ] `go test ./app/domain/inventory/inspectionapp/... -v -count=1`
- [ ] `go test ./api/cmd/services/ichor/tests/inventory/inspectionapi/... -v -count=1`

**Commit:** none (verification only)

---

## Summary of changed files

| File | Change |
|------|--------|
| `business/sdk/migrate/sql/migrate.sql` | Add v2.16 CHECK constraint on `quality_inspections.status` |
| `app/domain/inventory/inspectionapp/model.go` | Add `FailInspection`, `FailInspectionResult` types |
| `app/domain/inventory/inspectionapp/inspectionapp.go` | Add `lotTrackingsBus`/`db` fields, `NewAppWithTx`, `Fail` method |
| `api/domain/http/inventory/inspectionapi/inspectionapi.go` | Add `fail` handler |
| `api/domain/http/inventory/inspectionapi/routes.go` | Add `LotTrackingsBus`/`DB` to Config, register POST .../fail route |
| `api/cmd/services/ichor/build/all/all.go` | Wire `LotTrackingsBus` and `DB` into inspectionapi Config |
| `api/cmd/services/ichor/build/crud/crud.go` | Wire `LotTrackingsBus` and `DB` into inspectionapi Config |
| `api/cmd/services/ichor/tests/inventory/inspectionapi/fail_test.go` | New: integration tests for fail endpoint |
| `api/cmd/services/ichor/tests/inventory/inspectionapi/inspectionapi_test.go` | Register fail test cases |

## Design decisions

1. **NewAppWithTx constructor** — keeps backwards compatibility. Callers that don't need the fail endpoint can still use `NewApp`/`NewAppWithAuth`. The `Fail` method will panic if `db` or `lotTrackingsBus` are nil, which is acceptable since the wiring guarantees they are set.
2. **ReadCommitted isolation** — matches the `putawaytaskapp.complete` precedent. No phantom-read risk since we look up by primary key.
3. **Status CHECK constraint** — enforces `pending`, `passed`, `failed` at the database level. The migration is idempotent-safe (ALTER TABLE ADD CONSTRAINT fails if it already exists, but migrations run exactly once).
4. **Lot read outside tx, write inside tx** — the lot lookup (`QueryByID`) uses the non-tx bus to avoid holding a row lock unnecessarily. The write (`Update`) uses the tx-wrapped bus. If the lot was deleted between read and write, the FK constraint in the write will catch it.
5. **Permission: Update action** — failing an inspection is a status change (update), not a create. Uses `permissionsbus.Actions.Update`.
6. **Guard against double-fail** — returns `FailedPrecondition` if the inspection is already `"failed"`, preventing accidental re-quarantine.
