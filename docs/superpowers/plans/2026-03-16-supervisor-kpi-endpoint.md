# Supervisor KPI Endpoint Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.
>
> **Worktree:** Create a worktree before executing: `create a worktree for blocker-009-kpi and execute this plan`

**Goal:** Add a `GET /v1/inventory/supervisor/kpis` endpoint returning aggregated counts (pending approvals, open adjustments, pending transfers, open inspections) for the supervisor dashboard.

**Architecture:** New `supervisorkpiapp` (app layer) + `supervisorkpiapi` (API layer). The app holds references to multiple business buses and calls their `Count()` methods with appropriate filters. Read-only — no transactions needed.

**Tech Stack:** Go 1.23, Ardan Labs service architecture

---

## KPI Fields

| KPI Field                  | Source Bus               | Filter                                                    |
|----------------------------|--------------------------|-----------------------------------------------------------|
| `pending_approvals`        | `approvalrequestbus`     | `QueryFilter{Status: ptr("pending")}`                     |
| `pending_adjustments`      | `inventoryadjustmentbus` | `QueryFilter{ApprovalStatus: ptr("pending")}`             |
| `pending_transfers`        | `transferorderbus`       | `QueryFilter{Status: ptr("pending")}`                     |
| `open_inspections`         | `inspectionbus`          | `QueryFilter{Status: ptr("open")}`                        |
| `pending_put_away_tasks`   | `putawaytaskbus`         | `QueryFilter{Status: &putawaytaskbus.Statuses.Pending}`   |
| `active_alerts`            | `alertbus`               | `QueryFilter{Status: ptr("active")}`                      |

---

## Step 1: Create app layer — model

- [ ] Create `app/domain/inventory/supervisorkpiapp/model.go`

```go
package supervisorkpiapp

import "encoding/json"

// KPIs represents the aggregated supervisor dashboard counts.
type KPIs struct {
	PendingApprovals   int `json:"pending_approvals"`
	PendingAdjustments int `json:"pending_adjustments"`
	PendingTransfers   int `json:"pending_transfers"`
	OpenInspections    int `json:"open_inspections"`
	PendingPutAwayTasks int `json:"pending_put_away_tasks"`
	ActiveAlerts       int `json:"active_alerts"`
}

// Encode implements the web.Encoder interface.
func (k KPIs) Encode() ([]byte, string, error) {
	data, err := json.Marshal(k)
	return data, "application/json", err
}
```

**Verify:** `go build ./app/domain/inventory/supervisorkpiapp/...`

---

## Step 2: Create app layer — business logic

- [ ] Create `app/domain/inventory/supervisorkpiapp/supervisorkpiapp.go`

```go
package supervisorkpiapp

import (
	"context"

	"github.com/timmaaaz/ichor/business/domain/inventory/inspectionbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryadjustmentbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/putawaytaskbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// App manages the supervisor KPI aggregation use case.
type App struct {
	log                    *logger.Logger
	approvalRequestBus     *approvalrequestbus.Business
	inventoryAdjustmentBus *inventoryadjustmentbus.Business
	transferOrderBus       *transferorderbus.Business
	inspectionBus          *inspectionbus.Business
	putAwayTaskBus         *putawaytaskbus.Business
	alertBus               *alertbus.Business
}

// NewApp constructs a supervisor KPI app.
func NewApp(
	log *logger.Logger,
	approvalRequestBus *approvalrequestbus.Business,
	inventoryAdjustmentBus *inventoryadjustmentbus.Business,
	transferOrderBus *transferorderbus.Business,
	inspectionBus *inspectionbus.Business,
	putAwayTaskBus *putawaytaskbus.Business,
	alertBus *alertbus.Business,
) *App {
	return &App{
		log:                    log,
		approvalRequestBus:     approvalRequestBus,
		inventoryAdjustmentBus: inventoryAdjustmentBus,
		transferOrderBus:       transferOrderBus,
		inspectionBus:          inspectionBus,
		putAwayTaskBus:         putAwayTaskBus,
		alertBus:               alertBus,
	}
}

// Query returns aggregated KPI counts for the supervisor dashboard.
func (a *App) Query(ctx context.Context) (KPIs, error) {
	var kpis KPIs

	pendingStatus := "pending"
	openStatus := "open"
	activeStatus := "active"
	pendingPATStatus := putawaytaskbus.Statuses.Pending

	// Pending approval requests
	approvalCount, err := a.approvalRequestBus.Count(ctx, approvalrequestbus.QueryFilter{
		Status: &pendingStatus,
	})
	if err != nil {
		return KPIs{}, err
	}
	kpis.PendingApprovals = approvalCount

	// Pending inventory adjustments
	adjustmentCount, err := a.inventoryAdjustmentBus.Count(ctx, inventoryadjustmentbus.QueryFilter{
		ApprovalStatus: &pendingStatus,
	})
	if err != nil {
		return KPIs{}, err
	}
	kpis.PendingAdjustments = adjustmentCount

	// Pending transfer orders
	transferCount, err := a.transferOrderBus.Count(ctx, transferorderbus.QueryFilter{
		Status: &pendingStatus,
	})
	if err != nil {
		return KPIs{}, err
	}
	kpis.PendingTransfers = transferCount

	// Open inspections
	inspectionCount, err := a.inspectionBus.Count(ctx, inspectionbus.QueryFilter{
		Status: &openStatus,
	})
	if err != nil {
		return KPIs{}, err
	}
	kpis.OpenInspections = inspectionCount

	// Pending put-away tasks
	putAwayCount, err := a.putAwayTaskBus.Count(ctx, putawaytaskbus.QueryFilter{
		Status: &pendingPATStatus,
	})
	if err != nil {
		return KPIs{}, err
	}
	kpis.PendingPutAwayTasks = putAwayCount

	// Active alerts
	alertCount, err := a.alertBus.Count(ctx, alertbus.QueryFilter{
		Status: &activeStatus,
	})
	if err != nil {
		return KPIs{}, err
	}
	kpis.ActiveAlerts = alertCount

	return kpis, nil
}
```

**Verify:** `go build ./app/domain/inventory/supervisorkpiapp/...`

---

## Step 3: Create API layer — handler

- [ ] Create `api/domain/http/inventory/supervisorkpiapi/supervisorkpiapi.go`

```go
package supervisorkpiapi

import (
	"context"
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/inventory/supervisorkpiapp"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	supervisorKPIApp *supervisorkpiapp.App
}

func newAPI(app *supervisorkpiapp.App) *api {
	return &api{supervisorKPIApp: app}
}

func (a *api) query(ctx context.Context, r *http.Request) web.Encoder {
	kpis, err := a.supervisorKPIApp.Query(ctx)
	if err != nil {
		return errs.Newf(errs.Internal, "query kpis: %s", err)
	}

	return kpis
}
```

> **Note:** The `errs` import will be `github.com/timmaaaz/ichor/app/sdk/errs`. Fix the import and error handling to match the codebase pattern — check a nearby handler like `scanapi` or `putawaytaskapi` for the exact `errs` usage.

**Verify:** `go build ./api/domain/http/inventory/supervisorkpiapi/...`

---

## Step 4: Create API layer — routes

- [ ] Create `api/domain/http/inventory/supervisorkpiapi/routes.go`

```go
package supervisorkpiapi

import (
	"net/http"

	"github.com/timmaaaz/ichor/api/sdk/http/mid"
	"github.com/timmaaaz/ichor/app/domain/inventory/supervisorkpiapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/domain/core/permissionsbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inspectionbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryadjustmentbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/putawaytaskbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/web"
)

// Config holds all dependencies needed by the supervisor KPI API routes.
type Config struct {
	Log                    *logger.Logger
	ApprovalRequestBus     *approvalrequestbus.Business
	InventoryAdjustmentBus *inventoryadjustmentbus.Business
	TransferOrderBus       *transferorderbus.Business
	InspectionBus          *inspectionbus.Business
	PutAwayTaskBus         *putawaytaskbus.Business
	AlertBus               *alertbus.Business
	AuthClient             *authclient.Client
	PermissionsBus         *permissionsbus.Business
}

// RouteTable is the table name used for permission lookups.
// Uses the inventory_adjustments table since this is a cross-cutting read-only
// endpoint — any user with read access to inventory adjustments can view KPIs.
const RouteTable = "inventory.inventory_adjustments"

// Routes registers supervisor KPI HTTP routes.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := mid.Authenticate(cfg.AuthClient)
	a := newAPI(supervisorkpiapp.NewApp(
		cfg.Log,
		cfg.ApprovalRequestBus,
		cfg.InventoryAdjustmentBus,
		cfg.TransferOrderBus,
		cfg.InspectionBus,
		cfg.PutAwayTaskBus,
		cfg.AlertBus,
	))

	app.HandlerFunc(http.MethodGet, version, "/inventory/supervisor/kpis", a.query, authen,
		mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))
}
```

**Verify:** `go build ./api/domain/http/inventory/supervisorkpiapi/...`

---

## Step 5: Wire into all.go

- [ ] Edit `api/cmd/services/ichor/build/all/all.go`

**5a. Add import** (in the import block, near other inventory api imports around line 51):
```go
"github.com/timmaaaz/ichor/api/domain/http/inventory/supervisorkpiapi"
```

**5b. Add Routes call** (after the `putawaytaskapi.Routes(...)` block, around line 1060):
```go
supervisorkpiapi.Routes(app, supervisorkpiapi.Config{
    Log:                    cfg.Log,
    ApprovalRequestBus:     approvalRequestBus,
    InventoryAdjustmentBus: inventoryAdjustmentBus,
    TransferOrderBus:       transferOrderBus,
    InspectionBus:          inspectionBus,
    PutAwayTaskBus:         putAwayTaskBus,
    AlertBus:               alertBus,
    AuthClient:             cfg.AuthClient,
    PermissionsBus:         permissionsBus,
})
```

All six bus variables already exist in all.go:
- `approvalRequestBus` — line 461
- `inventoryAdjustmentBus` — line 437
- `transferOrderBus` — line 440
- `inspectionBus` — line ~448 (find exact)
- `putAwayTaskBus` — line 438
- `alertBus` — line 460

No new `supervisorkpiapp` import is needed in all.go because the app is constructed inside the `supervisorkpiapi.Routes()` function.

**Verify:** `go build ./api/cmd/services/ichor/build/all/...`

---

## Step 6: Build and manual test

- [ ] Full build verification:
```bash
go build ./app/domain/inventory/supervisorkpiapp/...
go build ./api/domain/http/inventory/supervisorkpiapi/...
go build ./api/cmd/services/ichor/build/all/...
go build ./api/cmd/services/ichor/...
```

- [ ] Manual smoke test (requires running service):
```bash
make token
export TOKEN=<token>
curl -H "Authorization: Bearer $TOKEN" http://localhost:3000/v1/inventory/supervisor/kpis | jq .
```

Expected response shape:
```json
{
  "pending_approvals": 0,
  "pending_adjustments": 0,
  "pending_transfers": 0,
  "open_inspections": 0,
  "pending_put_away_tasks": 0,
  "active_alerts": 0
}
```

---

## Files Created/Modified Summary

| Action   | File Path                                                        |
|----------|------------------------------------------------------------------|
| CREATE   | `app/domain/inventory/supervisorkpiapp/model.go`                 |
| CREATE   | `app/domain/inventory/supervisorkpiapp/supervisorkpiapp.go`      |
| CREATE   | `api/domain/http/inventory/supervisorkpiapi/supervisorkpiapi.go` |
| CREATE   | `api/domain/http/inventory/supervisorkpiapi/routes.go`           |
| MODIFY   | `api/cmd/services/ichor/build/all/all.go` (import + Routes call) |

---

## Notes

- **No business layer changes.** All six buses already have `Count(ctx, QueryFilter)` methods.
- **No migration.** This is a read-only aggregation endpoint over existing tables.
- **No new tests required for MVP.** The endpoint is a thin aggregation layer calling existing tested `Count()` methods. Integration tests can be added later if needed.
- **Permission routing:** Uses `inventory.inventory_adjustments` as the RouteTable. If a dedicated permission is needed later, add a row to the permissions seed and update the constant.
- **Parallelization opportunity:** The six `Count()` calls are independent and could be run concurrently with `errgroup`. This is a future optimization — sequential calls are simpler and sufficient for MVP since each is a single `SELECT COUNT(*)`.
