# Notifications on the Alert Backbone — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the workflow `send_notification` action deliver real, durable, user-visible notifications by retargeting it onto the already-working alert pipeline, then remove the dead parallel notification stack.

**Architecture:** A "notification" is a single-user, informational *alert*. The `send_notification` handler stops publishing to the consumer-less `QueueTypeNotification` and instead builds an `alertbus.Alert{AlertType:"notification", Severity:<priority>, Status:active}` with one `recipient_type='user'` recipient per recipient, persists it via `alertBus.Create` + `CreateRecipients`, and publishes over the shared `PublishAlertToRecipients` seam — exactly like `create_alert`. This rides the existing `alertbus → RabbitMQ(QueueTypeAlert) → AlertConsumer → AlertHub.BroadcastToUser → WebSocket → frontend alert store` path, which the frontend already consumes. The now-orphaned `workflow.notifications` table, its bus/store/app/inbox-API, and the `QueueTypeNotification` queue are deleted.

**Tech Stack:** Go 1.23, PostgreSQL (multi-schema), Temporal workflow engine, RabbitMQ, Vue 3 frontend (no change required).

## Global Constraints

- **Keep the `send_notification` action type.** This plan keeps it as a distinct (now thin) handler backed by `alertbus`; it does NOT prune the action-type allow-lists (`workflowsaveapp/model.go`, `validation.go`, `agenttools/definitions.go`, the seed grant). If the decision flips to full deprecation, that becomes a separate plan.
- **Never edit existing migrations** — the table DROP is a NEW migration version appended to `business/sdk/migrate/sql/migrate.sql`.
- **Never run `go test ./...`** — run only the changed packages.
- **Business layer is source of truth**; keep layers pure (no HTTP in business, no business logic in API).
- **Frontend requires no change to function** — a `low`-severity alert already renders as a silent inbox entry (`vue/ichor/src/stores/alert.ts:244-254` only toasts critical/high). An optional cosmetic relabel is Phase 5; it is NOT part of the definition of done.
- **Definition of done:** `send_notification` persists + delivers via the alert pipeline AND the dead notification stack (`workflow.notifications` table, `notificationbus`, `notificationdb`, `notificationapp`, `notificationinboxapi`, `QueueTypeNotification`) is removed, with the affected packages building and their tests green. The `notificationsapi` `/summary` endpoint (reads alerts+approvals — NOT `workflow.notifications`) is KEPT.

---

## Phase 0: Worktree & branch setup

PR #174 (`fix/notification-inbox-persistence`, commit `998e245d`) takes the opposite approach (persist to the dead `workflow.notifications` table). This plan supersedes it — **PR #174 should be closed.** Do this work on a fresh branch off `main`.

### Task 0.1: Create the worktree

- [ ] **Step 1: Create an isolated worktree off main**

Use the `superpowers:using-git-worktrees` skill (or the project `worktree` skill). Target branch: `fix/notifications-on-alert-backbone`, based on `main`.

- [ ] **Step 2: Confirm baseline builds**

Run: `go build ./business/sdk/workflow/... ./api/cmd/services/ichor/build/all/...`
Expected: builds clean (this is the pre-change baseline).

- [ ] **Step 3: Copy this plan into the worktree and commit**

```bash
mkdir -p docs/superpowers/plans
# copy this file to docs/superpowers/plans/2026-06-23-notifications-on-alert-backbone.md
git add docs/superpowers/plans/2026-06-23-notifications-on-alert-backbone.md
git commit -m "docs(workflow): plan notifications-on-alert-backbone"
```

---

## Phase 1: Retarget the `send_notification` handler onto `alertbus`

This is the core change. The handler gains an `*alertbus.Business` dependency (mirroring `CreateAlertHandler`) and its `Execute` builds a single-user informational alert instead of publishing to the dead queue.

### Task 1.1: Add the failing persistence test

**Files:**
- Test: `api/cmd/services/ichor/tests/workflow/actionhandlers/comms_test.go` (add a subtest next to the existing `create_alert` tests)

**Interfaces:**
- Consumes: the existing test harness in `comms_test.go` — a `dbtest.Database`, a real `alertbus.Business` (built via `alertbus.NewBusiness(db.Log, alertdb.NewStore(db.Log, db.DB))`), and `workflow.ActionExecutionContext`. **Read the existing `create_alert` subtest in this file and mirror its harness setup** (DB, bus construction, execCtx, recipient user IDs from seeded users).
- Produces: nothing (test only).

- [ ] **Step 1: Write the failing test**

Add a subtest that constructs the handler WITH a real alert bus and asserts a `workflow.alerts` row of type `notification` plus one recipient per user is created. Use seeded user UUIDs (see `docs/arch/seeding.md` for available users).

```go
// In comms_test.go, alongside the create_alert subtests.
t.Run("send_notification persists a single-user informational alert", func(t *testing.T) {
	alertBus := alertbus.NewBusiness(db.Log, alertdb.NewStore(db.Log, db.DB))
	handler := communication.NewSendNotificationHandler(db.Log, alertBus, nil) // nil queue: persistence only

	ruleID := uuid.New()
	cfg := []byte(`{"recipients":["` + seededUserID.String() + `"],"priority":"low","message":"Your report is ready","title":"Report"}`)
	execCtx := workflow.ActionExecutionContext{RuleID: &ruleID, RawData: map[string]any{}}

	out, err := handler.Execute(context.Background(), cfg, execCtx)
	require.NoError(t, err)
	require.Equal(t, "sent", out.(map[string]interface{})["status"])

	// Scope the query to this rule to stay isolated from concurrent subtests
	// (see docs/arch/workflow-engine.md ⚠ alert source_rule_id propagation).
	alerts, err := alertBus.Query(context.Background(),
		alertbus.QueryFilter{SourceRuleID: &ruleID},
		alertbus.DefaultOrderBy, page.MustParse("1", "10"))
	require.NoError(t, err)
	require.Len(t, alerts, 1)
	require.Equal(t, "notification", alerts[0].AlertType)
	require.Equal(t, alertbus.StatusActive, alerts[0].Status)
	require.Equal(t, "low", alerts[0].Severity)
})
```

> Verify the exact `alertbus.QueryFilter` field names, `DefaultOrderBy`, and `page` helper against the sibling `create_alert` test — copy whatever that test uses.

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./api/cmd/services/ichor/tests/workflow/actionhandlers/ -run 'send_notification persists' -v`
Expected: **does not compile yet** — `NewSendNotificationHandler` currently takes 2 args, not 3. (This is the first RED; the signature change in Task 1.2 fixes compilation, and the assertion proves the behavior.)

### Task 1.2: Rewrite the handler to emit an alert

**Files:**
- Modify: `business/sdk/workflow/workflowactions/communication/notification.go` (entire file — struct, constructor, Execute)

**Interfaces:**
- Produces: `func NewSendNotificationHandler(log *logger.Logger, alertBus *alertbus.Business, workflowQueue *rabbitmq.WorkflowQueue) *SendNotificationHandler` — the new 3-arg constructor every call site in Phase 2 must use.
- Consumes: `resolveTemplateVars` and `PublishAlertToRecipients` (same `communication` package, already defined in `alert.go` / `publish.go`); `alertbus.Alert`, `alertbus.AlertRecipient`, `alertbus.SeverityLow`, `alertbus.StatusActive` (`business/domain/workflow/alertbus/model.go:33-58`).

- [ ] **Step 1: Replace the struct + constructor**

Replace `notification.go:16-27` with:

```go
// SendNotificationHandler handles send_notification actions. A notification is a
// single-user, informational alert that rides the durable alert pipeline.
type SendNotificationHandler struct {
	log           *logger.Logger
	alertBus      *alertbus.Business
	workflowQueue *rabbitmq.WorkflowQueue
}

// NewSendNotificationHandler creates a new send notification handler.
// alertBus persists the notification as a single-user alert (nil = no
// persistence, e.g. validation-only registries).
func NewSendNotificationHandler(log *logger.Logger, alertBus *alertbus.Business, workflowQueue *rabbitmq.WorkflowQueue) *SendNotificationHandler {
	return &SendNotificationHandler{
		log:           log,
		alertBus:      alertBus,
		workflowQueue: workflowQueue,
	}
}
```

- [ ] **Step 2: Update imports**

Replace the import block (`notification.go:3-13`) — drop nothing, ADD the alertbus import (uuid/time/json/fmt/context/logger/rabbitmq/workflow all still used):

```go
import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/rabbitmq"
)
```

- [ ] **Step 3: Replace Execute (`notification.go:77-136`) with the alert-emitting version**

```go
func (h *SendNotificationHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext) (interface{}, error) {
	var cfg struct {
		Recipients []string `json:"recipients"`
		Priority   string   `json:"priority"`
		Message    string   `json:"message"`
		Title      string   `json:"title,omitempty"`
	}
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse notification config: %w", err)
	}

	now := time.Now()

	// A notification is a single-user informational alert riding the same durable
	// pipeline (alertbus -> RabbitMQ -> AlertHub -> WebSocket) the frontend already
	// consumes. Severity carries the config priority: low = silent inbox entry,
	// high/critical interrupt (frontend toast gating). alert_type "notification"
	// lets the UI treat these distinctly later. SourceEntityID is deliberately
	// left nil so personal notifications are NOT collapsed in the frontend's alert
	// bundling view (bundle key = source_entity_id:alert_type).
	severity := cfg.Priority
	if severity == "" {
		severity = alertbus.SeverityLow
	}

	sourceRuleID := uuid.Nil
	if execCtx.RuleID != nil {
		sourceRuleID = *execCtx.RuleID
	}

	alert := alertbus.Alert{
		ID:           uuid.New(),
		AlertType:    "notification",
		Severity:     severity,
		Title:        resolveTemplateVars(cfg.Title, execCtx.RawData),
		Message:      resolveTemplateVars(cfg.Message, execCtx.RawData),
		Context:      json.RawMessage(`{}`),
		SourceRuleID: sourceRuleID,
		Status:       alertbus.StatusActive,
		CreatedDate:  now,
		UpdatedDate:  now,
	}

	// Notifications target users only (no role fan-out). Validate all UUIDs first.
	var recipients []alertbus.AlertRecipient
	for _, u := range cfg.Recipients {
		uid, err := uuid.Parse(u)
		if err != nil {
			return nil, fmt.Errorf("invalid recipient UUID %q: %w", u, err)
		}
		recipients = append(recipients, alertbus.AlertRecipient{
			ID:            uuid.New(),
			AlertID:       alert.ID,
			RecipientType: "user",
			RecipientID:   uid,
			CreatedDate:   now,
		})
	}

	// Graceful degradation when no alert bus is wired (validation-only registries).
	if h.alertBus == nil {
		h.log.Warn(ctx, "send_notification: alertBus not configured, skipping notification")
		return map[string]interface{}{
			"notification_id": uuid.Nil.String(),
			"status":          "skipped",
			"recipients":      0,
		}, nil
	}

	if err := h.alertBus.Create(ctx, alert); err != nil {
		return nil, fmt.Errorf("create notification alert: %w", err)
	}
	if err := h.alertBus.CreateRecipients(ctx, recipients); err != nil {
		return nil, fmt.Errorf("create notification recipients: %w", err)
	}

	// Publish for real-time WebSocket delivery via the shared alert publish seam.
	if h.workflowQueue != nil {
		PublishAlertToRecipients(ctx, h.workflowQueue, h.log, alert, recipients)
	}

	h.log.Info(ctx, "send_notification executed",
		"notification_id", alert.ID,
		"recipients", len(recipients),
		"severity", severity)

	return map[string]interface{}{
		"notification_id": alert.ID.String(),
		"status":          "sent",
		"recipients":      len(recipients),
	}, nil
}
```

> Verify `alertbus.SeverityLow` exists (the enum values are `low/medium/high/critical`; see how `alert.go:111` references `alertbus.SeverityMedium`). If the constant name differs, the validated literal `"low"` is equivalent.

- [ ] **Step 4: Confirm the package builds (call sites still 2-arg — expect breakage, fixed in Phase 2)**

Run: `go build ./business/sdk/workflow/workflowactions/communication/`
Expected: PASS (the `communication` package itself compiles; the 2-arg callers live in other packages and are fixed in Phase 2).

### Task 1.3: Add the nil-bus graceful-degradation unit test

**Files:**
- Modify: `business/sdk/workflow/workflowactions/communication/notification_test.go`

- [ ] **Step 1: Update the existing handler construction to 3-arg and add a skip test**

Change `notification_test.go:20` from `NewSendNotificationHandler(log, nil)` to `NewSendNotificationHandler(log, nil, nil)`, and add:

```go
func Test_SendNotification_NilBus_Skips(t *testing.T) {
	log := logger.New(os.Stdout, logger.LevelInfo, "notif", func(context.Context) string { return "00000000-0000-0000-0000-000000000000" })
	handler := communication.NewSendNotificationHandler(log, nil, nil)
	cfg := []byte(`{"recipients":["` + uuid.NewString() + `"],"priority":"low","message":"hi"}`)
	out, err := handler.Execute(context.Background(), cfg, workflow.ActionExecutionContext{RawData: map[string]any{}})
	require.NoError(t, err)
	require.Equal(t, "skipped", out.(map[string]interface{})["status"])
}
```

- [ ] **Step 2: Run the unit tests**

Run: `go test ./business/sdk/workflow/workflowactions/communication/ -v`
Expected: PASS (validate + nil-bus skip). The DB-backed persistence test from Task 1.1 runs in Phase 2 once call sites compile.

---

## Phase 2: Wire the `alertBus` dependency at every call site

The 3-arg signature breaks 6 call sites. Fix all of them so the whole tree builds and the worker/server paths get a real alert bus.

### Task 2.1: Fix the two `register.go` sites

**Files:**
- Modify: `business/sdk/workflow/workflowactions/register.go:133` and `:276`

- [ ] **Step 1: RegisterAll (worker path) — pass the real alert bus**

Change `register.go:133` to mirror the `create_alert` line directly beneath it:

```go
registry.Register(communication.NewSendNotificationHandler(config.Log, config.Buses.Alert, config.QueueClient))
```

- [ ] **Step 2: RegisterCoreActions (server core path) — nil bus, graceful degradation**

Change `register.go:276` to:

```go
registry.Register(communication.NewSendNotificationHandler(log, nil, nil))
```

- [ ] **Step 3: Build**

Run: `go build ./business/sdk/workflow/workflowactions/`
Expected: PASS.

### Task 2.2: Add the server runtime upgrade in `all.go`

**Files:**
- Modify: `api/cmd/services/ichor/build/all/all.go` (immediately after the `create_alert` upgrade at `:578`)

**Interfaces:**
- Consumes: `alertBus` (already constructed at `all.go:522`).

- [ ] **Step 1: Register the upgraded send_notification handler**

After `all.go:578` (`actionRegistry.Register(communication.NewCreateAlertHandler(cfg.Log, alertBus, nil))`), add:

```go
	// Upgrade send_notification handler with the real alert bus.
	// The core registration uses a nil bus (graceful degradation); replace it here
	// so manual execution and triggered notifications persist + deliver as alerts.
	actionRegistry.Register(communication.NewSendNotificationHandler(cfg.Log, alertBus, nil))
```

- [ ] **Step 2: Build**

Run: `go build ./api/cmd/services/ichor/build/all/`
Expected: PASS.

### Task 2.3: Fix the remaining call sites (admin, apitest, integration test)

**Files:**
- Modify: `api/cmd/tooling/admin/commands/validateworkflows.go:84`
- Modify: `api/sdk/http/apitest/workflow.go:58`
- Modify: `api/cmd/services/ichor/tests/workflow/actionhandlers/comms_test.go:356`

- [ ] **Step 1: Admin validation registry — nil bus (validation only)**

`validateworkflows.go:84`: `communication.NewSendNotificationHandler(nil, nil, nil)`

- [ ] **Step 2: apitest infra — real alert bus so integration tests can assert delivery**

`apitest/workflow.go:58`: build an alert bus inline (mirror how the file constructs `alertbus`/`alertdb` for create_alert) and pass it:

```go
registry.Register(communication.NewSendNotificationHandler(db.Log,
	alertbus.NewBusiness(db.Log, alertdb.NewStore(db.Log, db.DB)), nil))
```

> Add the `alertbus`/`alertdb` imports if not already present in this file.

- [ ] **Step 3: Integration test registry — `comms_test.go:356`**

Change `NewSendNotificationHandler(db.Log, nil)` to pass the real alert bus already in scope in that test (or `alertbus.NewBusiness(db.Log, alertdb.NewStore(db.Log, db.DB))`).

- [ ] **Step 4: Whole-tree build + vet**

Run: `go build ./... && go vet ./api/cmd/services/ichor/tests/workflow/actionhandlers/`
Expected: PASS (no remaining 2-arg callers).

- [ ] **Step 5: Run the Phase 1 persistence test (now GREEN)**

Run: `go test ./api/cmd/services/ichor/tests/workflow/actionhandlers/ -run 'send_notification persists' -v`
Expected: PASS.

- [ ] **Step 6: Commit Phases 1–2**

```bash
git add business/sdk/workflow/workflowactions/communication/ business/sdk/workflow/workflowactions/register.go api/cmd/services/ichor/build/all/all.go api/cmd/tooling/admin/commands/validateworkflows.go api/sdk/http/apitest/workflow.go api/cmd/services/ichor/tests/workflow/actionhandlers/comms_test.go
git commit -m "feat(workflow): retarget send_notification onto the alert pipeline"
```

---

## Phase 3: End-to-end delivery verification

Prove a triggered `send_notification` reaches a user over the alert WebSocket path (not just the DB).

### Task 3.1: Add/extend an integration test for WS delivery

**Files:**
- Modify: an existing workflow integration test that already exercises the alert WS path (search `tests/workflow` for `BroadcastToUser` / alert WS assertions; mirror it).

- [ ] **Step 1: Assert a `send_notification` rule delivers to the target user**

Mirror the existing `create_alert` end-to-end assertion: fire a rule whose action is `send_notification` with one user recipient; assert (a) a `workflow.alerts` row with `alert_type='notification'` exists and (b) the alert publish path was invoked for `user:{recipientID}` (use whatever stub/spy the sibling alert WS test uses).

- [ ] **Step 2: Run it**

Run: `go test ./api/cmd/services/ichor/tests/workflow/... -run Notification -v`
Expected: PASS.

- [ ] **Step 3: Commit**

```bash
git commit -am "test(workflow): e2e send_notification delivers via alert WS path"
```

---

## Phase 4: Cleanup — remove the dead notification stack (REQUIRED)

The retarget orphans the entire `workflow.notifications` persistence stack and the `QueueTypeNotification` queue. **The plan is not done until these are gone.** Each deletion is gated by a grep that proves nothing live references it.

> KEEP `notificationsapi` (the `/v1/workflow/notifications/summary` endpoint) — it reads `alertBus`+`approvalBus`, not `workflow.notifications`. Verify before deleting anything that `notificationsapi` does NOT import `notificationapp`/`notificationbus`.

### Task 4.1: Remove the inbox REST API

**Files:**
- Delete: `api/domain/http/workflow/notificationinboxapi/` (whole dir incl. tests)
- Modify: `api/cmd/services/ichor/build/all/all.go` (remove the `notificationinboxapi` route registration ~`:1444` and its import)

- [ ] **Step 1: Prove the inbox endpoints have no other callers**

```bash
grep -rn "notificationinboxapi" --include=*.go . | grep -v "/notificationinboxapi/"
grep -rn "v1/workflow/notifications\b" vue/ichor/src   # expect: no live calls (only commented-out)
```
Expected: only the registration in `all.go`. If anything else appears, stop and reassess.

- [ ] **Step 2: Delete the package and remove the route + import**

```bash
git rm -r api/domain/http/workflow/notificationinboxapi/
```
Then remove the `notificationinboxapi.Routes(...)` call (~`all.go:1444`) and its import line.

- [ ] **Step 3: Build**

Run: `go build ./api/...`
Expected: PASS.

### Task 4.2: Remove the app + bus + store layers

**Files:**
- Delete: `app/domain/workflow/notificationapp/`
- Delete: `business/domain/workflow/notificationbus/` (incl. `stores/notificationdb/` and `testutil.go`)
- Modify: `api/cmd/services/ichor/build/all/all.go` (remove `notificationBus` construction at `:523` and its imports)

- [ ] **Step 1: Prove nothing live references them**

```bash
grep -rn "notificationapp\|notificationbus\|notificationdb\|TestSeedNotifications" --include=*.go . \
  | grep -vE "/notificationapp/|/notificationbus/|/notificationdb/"
```
Expected: only `all.go:523` (construction). If `dbtest.go` wires a notification bus into `BusDomain`, remove that too (and note it here). If anything else references it, stop.

- [ ] **Step 2: Delete and unwire**

```bash
git rm -r app/domain/workflow/notificationapp/ business/domain/workflow/notificationbus/
```
Remove the `notificationBus := notificationbus.NewBusiness(...)` line (`all.go:523`) and the now-unused `notificationbus`/`notificationdb` imports.

- [ ] **Step 3: Build whole tree**

Run: `go build ./...`
Expected: PASS.

### Task 4.3: Remove the `QueueTypeNotification` queue

**Files:**
- Modify: `foundation/rabbitmq/client.go` (remove the `QueueTypeNotification` const `:316` and its `wq.queues[QueueTypeNotification]` config `:414-422`)

- [ ] **Step 1: Prove no producer/consumer remains**

```bash
grep -rn "QueueTypeNotification" --include=*.go . | grep -v "/foundation/rabbitmq/client.go"
```
Expected: ZERO (the only producer was the old `notification.go`, already rewritten in Phase 1).

- [ ] **Step 2: Remove the const + config block, build**

Run: `go build ./foundation/rabbitmq/ ./...`
Expected: PASS.

### Task 4.4: Drop the `workflow.notifications` table (new migration)

**Files:**
- Modify: `business/sdk/migrate/sql/migrate.sql` (append a NEW version block — never edit existing)

- [ ] **Step 1: Find the current max migration version**

```bash
grep -n "^-- Version:" business/sdk/migrate/sql/migrate.sql | tail -3
```
Note the highest version N.

- [ ] **Step 2: Append a new version that drops the table + its grant**

At the END of `migrate.sql`, add (replace `<N+1>` with the next version):

```sql
-- Version: <N+1>
-- Description: Drop dead workflow.notifications inbox (send_notification now rides the alert pipeline)
DROP TABLE IF EXISTS workflow.notifications;
```

> Also remove the `send_notification`-era notification table GRANT if present (search `migrate.sql` for `workflow.notifications` grants and drop via the new version if any remain — `DROP TABLE` removes the table but a stale `GRANT ... ON workflow.notifications` line in an *earlier* version is harmless once the table is gone; do NOT edit the earlier version).

- [ ] **Step 3: Sanity-check migration applies (changed pkg only)**

Run the migration smoke per `docs/arch/sqldb.md` / project commands (e.g. `make migrate` against a throwaway DB, or the migrate unit test if one exists). Expected: applies clean; `\d workflow.notifications` → does not exist.

- [ ] **Step 4: Commit cleanup**

```bash
git add -A
git commit -m "chore(workflow): remove dead notification stack (table, bus, inbox API, queue)"
```

### Task 4.5: Sweep for stragglers

- [ ] **Step 1: Final dead-reference grep**

```bash
grep -rn "workflow.notifications\|NotificationDelivery\|notification_deliveries\|QueueTypeNotification" --include=*.go .
```
Expected: no LIVE references. If `workflow.NewNotificationDelivery` (`business/sdk/workflow/models.go:454`) or a `notification_deliveries` table is genuinely unused (tests only), remove it in this task; if anything real remains, note it and leave it.

- [ ] **Step 2: Whole-tree build**

Run: `go build ./...`
Expected: PASS.

---

## Phase 5 (OPTIONAL, cosmetic): Frontend read/unread labeling

The frontend already works unchanged (low-severity notification alerts render silently in the bell/Notification Center). This phase is pure relabeling and is NOT part of the definition of done.

### Task 5.1: Relabel acknowledge→"Mark read" for the notification feel (optional)

**Files (vue/ichor):** `src/components/alerts/AlertListItem.vue:193,209`, `src/components/alerts/AlertBundleItem.vue:154,157`, `src/pages/alerts/index.vue:482,492,503`, `src/components/alerts/AlertDropdown.vue:51`

- [ ] **Step 1:** Change button labels/aria "Acknowledge"→"Mark read", "Dismiss"→"Remove" (no store/type/backend change). Run `npm run lint && npm run type-check` in `vue/ichor`. Skip entirely if not wanted.

---

## Phase 6: Docs + ship

### Task 6.1: Update arch docs

**Files:**
- Modify: `docs/arch/workflow-alerts.md` — document that `send_notification` is a single-user informational alert riding the alert pipeline; remove any implication of a separate notification store.
- Modify: `docs/workflow/README.md` — update the `send_notification` handler catalog entry.

- [ ] **Step 1:** Edit the docs to reflect the unified pipeline. Note `alert_type='notification'` as the distinguishing marker and the "leave source_entity_id nil to avoid bundling" rule.

- [ ] **Step 2: Commit**

```bash
git commit -am "docs(workflow): document send_notification on the alert backbone"
```

### Task 6.2: Final verification + close PR #174

- [ ] **Step 1: Build + changed-package tests**

```bash
go build ./...
go test ./business/sdk/workflow/workflowactions/communication/ ./api/cmd/services/ichor/tests/workflow/actionhandlers/ ./api/cmd/services/ichor/tests/workflow/...
```
Expected: PASS.

- [ ] **Step 2:** Open the PR for `fix/notifications-on-alert-backbone`; in the description, note it supersedes PR #174 and request #174 be closed.

---

## OPTIONAL APPENDIX: Idempotency / dedup (separate decision)

Not required for this plan and NOT in the definition of done. The alert path (and therefore the retargeted notification) mints `uuid.New()` with no dedup key (`alert.go:128`; no `ON CONFLICT` in `alertdb` create), so a Temporal activity retry can duplicate. This is **pre-existing `create_alert` behavior** — the retarget does not make it worse. If you want to fix it, do it once for BOTH paths:

- Add a dedup column to `workflow.alerts` (e.g. `dedup_key VARCHAR(255) UNIQUE`) via a new migration; build it from `ExecutionID + ruleID + recipient` (mirror `reserve_inventory.go:193`); `ON CONFLICT (dedup_key) DO NOTHING` in `alertdb` create. RED test: execute the same action twice (simulated retry) → assert one alert row. This benefits `create_alert` too.

---

## Self-Review

- **Spec coverage:** retarget (Phase 1–2) ✓, delivery proof (Phase 3) ✓, cleanup as a required phase (Phase 4) ✓, frontend no-op confirmed + optional polish (Phase 5) ✓, docs (Phase 6) ✓, supersede PR #174 (Phase 0 + 6.2) ✓, idempotency surfaced as out-of-scope decision (appendix) ✓.
- **Open assumption to confirm with the requester:** keep `send_notification` as a thin handler (chosen) vs. fully deprecate it in favor of `create_alert`. If deprecation is preferred, Phases 1–3 collapse into "delete the handler + prune action-type allow-lists" and Phase 4 still applies.
- **Type consistency:** new constructor `NewSendNotificationHandler(log, alertBus, workflowQueue)` is used identically at all 6 call sites; Execute uses `alertbus.Alert`/`alertbus.AlertRecipient` exactly as `alert.go` does.
- **Verify-at-execution notes:** `alertbus.SeverityLow` constant name; `alertbus.QueryFilter`/`DefaultOrderBy`/`page` helpers (copy from the sibling create_alert test); whether `dbtest.go` wires `notificationBus` into `BusDomain` (unwire if so); the current max migration version.
