# Twin-Site Payload Drift Audit

**Date:** 2026-04-23
**Scope:** Ichor Go codebase. Read-only audit.
**Trigger:** PR #126 fixed a twin-site drift bug in `api/domain/http/workflow/approvalapi` where primary + retry paths each constructed a Temporal `Result: map[string]any{...}` inline and the retry path silently dropped `resolved_by` / `reason`.

**Pattern recap.** Two or more code paths build inline literals (`map[string]any`, struct-with-untyped-field, JSON payload) that a single downstream reader consumes via string keys or reflection. No shared type, helper, or schema forces the shapes to agree. Fields drift silently when one site grows and the other does not. `map[string]any` and `interface{}` erase shape, so the compiler cannot help. Tests usually assert behavior at each site in isolation, not parity between sites.

**Methodology.** Six parallel investigation agents, one per surface. Each read the authoritative arch doc first, then grepped for construction sites + downstream consumers. Top findings spot-checked against the source before elevation in the ranking below.

---

## Top 3 candidates to fix first

1. **`approvalapi.publishApprovalResolved` WebSocket payload routing mismatch** (`api/domain/http/workflow/approvalapi/approvalapi.go:348`).
   The publisher sets `msg.Type = "approval_resolved"` with fields at the top of `msg.Payload` (`approvalId`, `status`, `resolvedBy`, `ruleId`, `actionName`). The WebSocket consumer (`consumer.go:99-101`) only switches on `"alert_updated"`; every other type — including `"approval_resolved"` — falls through to the default branch and is delivered as WS type `"alert"` with the raw approval payload inside. Frontend `"alert"` subscribers expect keys `id`, `severity`, `message`, `alertType`, etc.; they receive `approvalId`, `resolvedBy`, `ruleId` instead and every access silently returns `undefined`. This is the *same class* of bug as PR #126 and has a live production impact: approval-resolved WebSocket notifications are structurally broken for any frontend keying on the alert schema.
   **Why first:** HIGH severity, UNGUARDED, actively on the hot path, one-line audit surface.

2. **`seek_approval` approval-created notification never reaches the WebSocket hub** (`business/sdk/workflow/workflowactions/approval/seek.go:216-272`, confirmed by direct read).
   `createApprovalAlert` writes a row to `alertBus` and attaches recipients, but (unlike `CreateAlertHandler.publishAlertToWebSocket` in `business/sdk/workflow/workflowactions/communication/alert.go:247`) it never publishes to RabbitMQ / the hub. Result: an approver user has no real-time notification that a new approval request is waiting for them. They must poll. This is a twin-site by omission — two code paths that both create alerts and *should* both push, but only one does.
   **Why second:** HIGH severity, UNGUARDED, affects the approval UX directly.

3. **`alertapi.testAlert` ships two divergent shapes of the same alert** (`api/domain/http/workflow/alertapi/alertapi.go:350-363` vs `alertapi.go:379`).
   The handler returns `toAppAlert(alert)` to the HTTP caller (14 fields) and publishes an inline `map[string]interface{}` to the WebSocket channel (9 fields). Missing from the WS push: `context`, `sourceRuleId`, `sourceRuleName`, `actionUrl`, `expiresDate`. The frontend alert tray consumes both channels; if it updates store state from the WS push (the common SSE/WS pattern), those 5 fields silently vanish from reactive state until a refetch. **Zero tests** exercise this endpoint at all.
   **Why third:** HIGH severity, UNGUARDED, and it's an exact structural match for the PR #126 pattern — canonical `toAppX` function exists, a second site hand-rolls a near-duplicate.

The following sections enumerate findings surface-by-surface.

---

## Surface 1: Delegate event publishing & registrations

Arch doc: `docs/arch/delegate.md`. Scope combined surfaces 1 + 7 from the master brief (emitters + handler registrations share the same `DelegateEventParams` contract).

| # | Site A (emitter) | Site B / Consumer | Consumer | Severity | Test-guarded? |
|---|---|---|---|---|---|
| 1 | `approvalrequestbus/event.go:22-52` — `ActionUpdatedParms` struct omits `UserID`; emitted by `approvalrequestbus.go:143` | Conforming consumer `business/sdk/workflow/temporal/delegatehandler.go:50-64` (`json.Unmarshal` into `workflow.DelegateEventParams`, reads `params.UserID`) | `delegatehandler.go:63` assigns `params.UserID` → `TriggerEvent.UserID` | MEDIUM (latent — `approvalrequest` domain is not currently wired via `RegisterDomain`, so mismatched payload is not actively delivered) | UNGUARDED — `delegate_test.go:62-90` checks `capturedDomain`/`capturedAction` only, never decodes `RawParams` |
| 2 | `settingsbus/event.go` — all three `ActionXxxParms` structs use `Key string` instead of `EntityID uuid.UUID`, omit `UserID` | Same `delegatehandler.go` consumer shape | `params.EntityID` → `uuid.Nil`; `params.UserID` → `uuid.Nil` | LOW — no `RegisterDomain` for `"config.settings"` in `all.go` | UNGUARDED — no delegate-specific test for settingsbus |
| 3 | `business/sdk/workflow/trigger.go:516` registers a consumer for `ActionRuleDeleted` | `workflowbus.go` never emits `ActionRuleDeleted` | Handler (`trigger.go:507-511`) only calls `RefreshRules`, ignores `RawParams` | LOW — dead registration; can't trigger | N/A |

**Summary.** The latent risk is the `approvalrequestbus` payload: the moment someone adds `workflow.RegisterDomain` for `"approvalrequest"`, every approval-resolve event silently delivers `UserID = uuid.Nil` to the workflow engine. Testing hygiene would not catch this. The other two findings are defensively harmless today.

**Checked and cleared.**
- ~68 other registered domains (ordersbus, userbus, assetbus, productbus, warehousebus, etc.) all use the standard `ActionCreatedParms` / `ActionUpdatedParms` / `ActionDeletedParms` shape with `EntityID`, `UserID`, `Entity`, `BeforeEntity`. All `delegate.Call` sites go through the `ActionXxxData` helper — no inline map literals.
- `workflowbus` `ActionRuleChangedParms` — single emitter, single consumer that ignores payload beyond `data.Action`.
- `scenariobus` — registered; standard shape; `EntityID`/`UserID` explicitly `uuid.Nil` as documented limitation.

---

## Surface 2: Temporal activity payloads

Arch doc: `docs/arch/workflow-engine.md`. Scope: `temporal.ActionActivityOutput` and `AsyncCompleter` construction sites beyond approvalapi.

| # | Site A | Site B | Consumer | Severity | Test-guarded? |
|---|---|---|---|---|---|
| 1 | `seek_approval` async completion via `approvalapi.completeAndClear` → `buildResolveResult` (keys: `output`, `approval_id`, `resolved_by`, `reason`) `approvalapi.go:207-213` | `seek_approval` synchronous-fallback `Execute()` path (keys: `approval_id`, `output`, `status`) `seek.go:208-212` | `GraphExecutor` / `MergedContext` reads `result["output"]` for edge routing; downstream steps read other keys from `ActionResults` | MEDIUM — sync path drops `resolved_by` / `reason`; only bites if a workflow runs in manual-execution mode and a downstream step templates those fields | PARTIALLY GUARDED — `resolve_test.go:173` checks the async shape; `seek_test.go:529` checks `approval_id` but never compares the two shapes |
| 2 | `seek_approval.Execute()` stub path (nil `approvalRequestBus`): `{"output":"approved","status":"stub"}` `seek.go:163-165` | `seek_approval.Execute()` real path: `{"approval_id":..., "output":"pending", "status":"pending"}` `seek.go:208-212` | Same `GraphExecutor` — same `Execute()` method, two inline branches | MEDIUM — stub path has no `approval_id`; template resolution like `{{seek_approval_0.approval_id}}` silently renders empty | UNGUARDED — no cross-branch test |
| 3 | `send_email.Execute()` payload `email.go:130-136` | No second writer — `IsAsync()=true` but no `StartAsync`, routed through `ExecuteActionActivity` via `selectActivityFunc` | N/A | LOW — stale annotation; a future dev could register it async and create a twin | GUARDED by `allocate_inventory_async_test.go:33-37` (routing asserted) |
| 4 | `allocate_inventory.Execute()` inline map `allocate.go:377-386` | Same story — `IsAsync()=true` but no `StartAsync`, routed sync | N/A | LOW — same stale-annotation risk | GUARDED by `allocate_inventory_async_test.go:84-114` |

**Summary.** The approvalapi primary-vs-retry pair (the PR #126 bug) is now fixed by `buildResolveResult()`. The remaining live twin-site risk is `seek_approval.Execute()` having two distinct inline maps across `{nil-bus, real-bus}` branches whose key sets diverge, and the async vs sync-fallback shapes for the same action not matching. The `IsAsync()` stale annotations on `send_email` / `allocate_inventory` are documentation bugs today — they become real twin-sites if someone implements `StartAsync` for them without realizing the sync `Execute()` contract diverges.

**Checked and cleared.**
- `approvalapi.completeAndClear` + `retryTemporalCompletion` — both call `buildResolveResult()` now; deeply guarded in `resolve_test.go:173-183`.
- `ExecuteActionActivity` return path (`activities.go:86-91`) — single construction site; `toResultMap` + `output` default injection are the only transforms.
- `ExecuteAsyncActionActivity` — returns empty `ActionActivityOutput{}` with `activity.ErrResultPending`; the real payload is supplied later by `AsyncCompleter.Complete()`, which has one construction site in `approvalapi`.
- `AsyncCompleter.Complete` / `Fail` — delegate to `client.CompleteActivity`; `Fail` passes `nil` result intentionally; no field-level drift.
- All 19 other sync-only `ActionHandler.Execute()` implementations — single return site each; no async completion twin.

---

## Surface 3: Workflow action handlers / MergedContext

Arch doc: `docs/arch/workflow-engine.md`. Scope: the 20 registered action types and the `MergedContext` they feed.

| # | Site A | Site B | Consumer | Severity | Test-guarded? |
|---|---|---|---|---|---|
| 1 | `seek.go:163-166` stub path: `{"output":"approved","status":"stub"}` (no `approval_id`) | `seek.go:208-212` real path: `{"approval_id":..., "output":"pending", "status":"pending"}` | Downstream templating `{{seek_approval_0.approval_id}}` or edge routing on `output` | HIGH — stub emits `output=approved` with no `approval_id`; downstream edge divergence + empty template | UNGUARDED — `seek_test.go:312-352` asserts each path in isolation, never cross-path key-set parity |
| 2 | `seek.go:85-88` `GetOutputPorts` declares `approved / rejected / timed_out` | `seek.go:208-212` `Execute()` emits `output=pending` | `graph_executor.go` routes on `result["output"]`; `pending` matches no edge, so `activities.go:76-78` substitutes `success`, silently bypassing approval branch logic | HIGH (in manual-execution mode; N/A for normal async-via-Temporal flow where StartAsync dispatches and Execute is not called) | UNGUARDED — no test asserts `Execute()`'s `output` value is a declared port name |
| 3 | `commit_allocation.go:186-194` returns `CommitAllocationResult` (keys: `quantity_committed`, `previous_reserved`, `new_allocated`) | `release_reservation.go:178-184` returns `ReleaseReservationResult` (keys: `quantity_released`, `previous_reserved`, `new_reserved`) | Both are inventory-lifecycle handlers; a downstream `log_audit_entry` or `evaluate_condition` templating either's `quantity_*` field | MEDIUM — conceptually symmetric twin but different field names; cross-chain templates silently get zero for the missing field | UNGUARDED — DB-side-effect-only assertions; neither test inspects the returned map keys |
| 4 | `allocate.go:377-386` `Execute()` returns map with `allocation_id` + `output` | `reserve_inventory.go:144-216` returns `ReserveInventoryResult` struct (keys: `reservation_id`, `status`, `reserved_items`, no `output`) | Downstream `evaluate_condition` templating `{{reserve_inventory.allocation_id}}` or `{{allocate_inventory.reservation_id}}` | MEDIUM — different ID key names; `output` missing on reserve path (default-injected) | UNGUARDED — `reserve_inventory_test.go` checks `result.Status` on the struct, never the map shape |

**Summary.** Both top findings center on `seek_approval`, which is the most structurally twin-site-prone handler in the registry because it has: (a) nil-bus stub branch, (b) real-bus sync branch, (c) async completion via approvalapi — three code paths, three shapes, one consumer. The inventory-pair (commit/release, allocate/reserve) drift is cosmetic today but will accumulate as more workflows chain these handlers.

**Checked and cleared** (single-writer, stable key sets, no downstream twin):
- `check_inventory`, `check_reorder_point`, `evaluate_condition`, `delay`, `log_audit_entry`, `create_entity`, `lookup_entity`, `transition_status`, `update_field`, `call_webhook`, `create_alert`, `send_email`, `send_notification`, `create_purchase_order`, `receive_inventory`, `create_put_away_task`, `resolve_approval_request` — all verified single-writer or single-function-multi-branch with consistent skeleton.

---

## Surface 4: WebSocket alert payloads

Arch doc: `docs/arch/workflow-alerts.md`.

| # | Site A | Site B | Consumer | Severity | Test-guarded? |
|---|---|---|---|---|---|
| 1 | `seek.go:216-272` `createApprovalAlert()` — persists to DB via `alertBus.Create`, **never publishes to RabbitMQ / hub** (verified) | `communication/alert.go:247` `publishAlertToWebSocket()` — correctly publishes to hub | `consumer.go:76` `handleAlert()` → `BroadcastToUser/Role/All` | HIGH — approval-request alerts silently dark in the browser; users must poll REST | UNGUARDED — no test verifies `seek_approval` delivers a WS alert |
| 2 | `approvalapi.go:348` `publishApprovalResolved()` — `msg.Type = "approval_resolved"` with top-level `approvalId`, `status`, `resolvedBy`, `ruleId`, `actionName`, `resolvedDate` | `consumer.go:99-101` — only branches on `"alert_updated"`; everything else falls through to WS type `"alert"` | `consumer.go:105` forwards raw `msg.Payload` under WS type `"alert"`; frontend `"alert"` handlers expect `id`, `severity`, `message`, `alertType` — they get `approvalId`, `resolvedBy` instead | HIGH — approval-resolved notifications structurally broken for any frontend keying on the canonical alert schema | UNGUARDED — `consumer_test.go` and `e2e_test.go` publish raw messages manually, never invoke this publisher |
| 3 | `alertapi.go:500` `publishAlertStatusChange()` — wraps `{ "alertUpdate": { "id", "status", "updatedDate" } }` | N/A — single writer for `alert_updated` | `consumer.go:100` routes on `msg.Type == "alert_updated"` | LOW — single writer, no drift risk | UNGUARDED for content, but no twin-site |

**Summary.** Two HIGH-severity live bugs in the alert pipeline. The first (`seek.go` creating a DB-only alert with no WS publish) is a *twin-site by omission*: the well-formed sibling (`CreateAlertHandler`) performs both writes, and the approval variant forgot the second. The second (`publishApprovalResolved` routing mismatch) is a textbook PR #126 pattern — emitter and consumer agreed verbally but not structurally.

**Checked and cleared.**
- `alert_updated` status-change alerts — single writer, correctly routed.
- `create_alert` workflow action — single-writer end-to-end pipeline.
- WebSocket hub targeting (`BroadcastToUser/Role/All`) — deterministic routing on `msg.UserID` / `msg.Payload["role_id"]`, no drift.

---

## Surface 5: HTTP response encoders vs. toApp* functions

Scope: `api/domain/http/**`. Arch docs: `docs/arch/domain-template.md`, `docs/arch/errs.md`.

| # | Site A (primary path) | Site B (secondary) | Consumer | Severity | Test-guarded? |
|---|---|---|---|---|---|
| 1 | `alertapi.go:379` returns `toAppAlert(alert)` (HTTP response from `testAlert`) | `alertapi.go:350-363` — inline `map[string]interface{}` published to RabbitMQ/WS from the same handler | Frontend alert tray (POST `/v1/workflow/alerts/test`) | HIGH — WS payload missing 5 canonical fields (`context`, `sourceRuleId`, `sourceRuleName`, `actionUrl`, `expiresDate`) | UNGUARDED — zero tests for this endpoint |
| 2 | `alertapi.go:232` `acknowledge` returns `toAppAlert(alert)` | `alertapi.go:500-518` `publishAlertStatusChange` — inline `{ "alertUpdate": { id, status, updatedDate } }` | Frontend tray (POST `/v1/workflow/alerts/{id}/acknowledge` & `/dismiss`) | MEDIUM — architecturally a delta-update pattern, so the shape gap is intentional; frontend must handle partial-merge correctly | PARTIALLY GUARDED — bulk endpoint asserts `BulkActionResult`; single-alert WS shape has no typed assertion |
| 3 | `approvalapi.go:199` `resolve` returns `toAppApproval(approval)` | `approvalapi.go:353-360` `publishApprovalResolved` — inline map | Frontend approval queue UI | MEDIUM — notification event rather than canonical entity payload; HTTP path is test-guarded, WS side untested | HTTP: GUARDED (`resolve_test.go` typed struct); WS: UNGUARDED |

**Summary.** Two distinct HIGH-severity patterns, one in `alertapi.testAlert` (exact PR #126 structural match — canonical `toApp*` + hand-rolled sibling), and the other a thematic companion in `acknowledge`/`dismiss` where the WS shape is an intentional delta but is not asserted anywhere. `approvalapi` has the same pattern as `alertapi.testAlert` but the approval entity is more defensively typed on the frontend.

**Checked and cleared.**
- `workflow/approvalapi` retry paths — consistently call `toAppApproval` after PR #126.
- `workflow/executionapi`, `workflow/ruleapi`, `workflow/workflowsaveapi`, `workflow/ruleapi/simulate.go` — single converter each, no secondary inline path.
- `inventory/transferorderapi`, `putawaytaskapi`, `picktaskapi`, `cyclecountitemapi`, `cyclecountsessionapi` — straight CRUD, no branching response paths.
- `sales/ordersapi`, `procurement/purchaseorderapi`, `floor/directedworkapi` — app-layer result returned directly, no inline twin.

---

## Surface 6: Form data / template processing

Arch doc: `docs/arch/form-data.md`.

| # | Site A | Site B | Consumer | Severity | Test-guarded? |
|---|---|---|---|---|---|
| 1 | `mergeFieldDefaults` (`formdataapp.go:500-588`) — parses `formfieldbus.FieldDefaultConfig` for `DefaultValue`, `DefaultValueCreate`, `DefaultValueUpdate`, `CopyFromField` + ad-hoc anonymous struct for dropdown FKs | `mergeLineItemFieldDefaults` (`formdataapp.go:623-685`) — reads the same logical fields but directly from `formfieldbus.LineItemField` | `ProcessTemplateObject` at `formdataapp.go:335`, reached by both `executeSingleOperation` and `executeArrayOperation` | HIGH — no shared extraction helper; any new default-injection field added to one function must be mirrored manually. `FieldDefaultConfig` is even a named type that `LineItemField` inlines field-by-field. Observationally symmetric today but structurally unenforced | PARTIALLY GUARDED — `TestMergeLineItemFieldDefaults` asserts injected values without `CopyFromField` cases; `TestMergeFieldDefaults_*` covers the non-line-item path including FK resolution; no test uses both functions with the same fixture to assert parity |
| 2 | `updateOrderTotals` (`formdataapp.go:808-828`) — inline `map[string]any{"id": ..., "subtotal": ..., "tax_amount": ..., "total_amount": ...}` → `executeUpdate` → `reg.DecodeUpdate` (the `ordersapp.UpdateOrder` decoder) | `executeSingleOperation` normal update path (`formdataapp.go:320-377`) — caller submits JSON whose keys must match `ordersapp.UpdateOrder` | `executeUpdate` at `formdataapp.go:457-488`, which calls `reg.DecodeUpdate(data)` | MEDIUM — if `ordersapp.UpdateOrder` renames a field (`total` → `total_amount`) or adds a required field, the hand-rolled map silently supplies the wrong key set; decoder ignores unknown keys, so this fails silently | UNGUARDED — no dedicated tests for `recalculateOrderTotalsIfNeeded`; no integration test asserts `total_amount` appears in the DB after upsert |
| 3 | `workflow.TemplateProcessor.ProcessTemplateObject` (`template.go:155`) — full engine (filters, `$me`/`$now` builtins, dot-notation, `{{expr:...}}`) | `resolveTemplateVars` (`communication/alert.go:230-243`) — 12-line private regex used in email, notification, alert action handlers; supports only flat `{{key}}` | Both consume templates written by the same form-field authors | MEDIUM — expressions with dot-notation, filters, or `$me`/`$now` silently left unresolved (returned as literal `{{...}}`) in email/notification/alert bodies, while the same template processes correctly during form-data upsert | PARTIALLY GUARDED — `email_test.go` and `notification_test.go` test flat substitution; no tests exercise dot-notation or `$me` in these handlers |

**Summary.** Finding 1 is the highest-severity because `FieldDefaultConfig` and `LineItemField` are already *explicitly* documented as sharing the same fields ("same fields as `FieldDefaultConfig` for consistency" — `model.go:210-214`), yet the extraction logic is duplicated. Finding 2 is a classic "side-door update" twin — works today, falls over the moment `UpdateOrder` evolves. Finding 3 is the most subtle: two template engines in the same project with different grammars is a drift-factory.

**Checked and cleared.**
- `formdataregistry.EntityRegistration` — named-struct contract; no bare `map[string]any` at the registry boundary.
- `TemplateProcessor` call site inside `formdataapp` — exactly one call site (`formdataapp.go:335`); no second path within the form-data pipeline.
- `buildExecutionPlan` / `executeOperation` ordering — single deterministic code path.
- `ActionExecutionContext` construction — three sites, all populate `RawData` from upstream; no template-relevant field differs.
- Registry thread-safety — `RWMutex`-guarded, read-only post-startup.

---

## Cross-surface observations

- **`seek_approval` is the single highest-risk handler in the codebase.** It appears in surfaces 2, 3, and 4 with three different twin-site patterns (sync-vs-async completion shape, stub-vs-real Execute branches, and a missing WebSocket publish). Fixing this one handler would close 4 findings.
- **`approvalapi`'s WebSocket publisher (`publishApprovalResolved`) is a PR #126 repeat.** Same shape of bug (inline map + distant consumer keyed by convention), just on the RabbitMQ/WS channel instead of the Temporal channel.
- **The canonical fix pattern (typed struct / named helper) is already in use in many places** — `buildResolveResult`, `buildWebhookResult`, the `ActionXxxData` delegate helpers, the `toApp*` HTTP encoders. The twin-site risks above are all "someone forgot to use the helper" cases, not "no helper exists" cases.
- **Tests assert shape at the primary site, not at secondary sites or cross-site.** The universal test-hygiene gap: a single-site `ExpResp` check is not a parity check.

## Suspicious but safe (catalogued so the next auditor doesn't re-check)

- `approvalapi.completeAndClear` + `retryTemporalCompletion` — shared `buildResolveResult()` since PR #126; deeply guarded.
- `ExecuteActionActivity` / `ExecuteAsyncActionActivity` — both single-writer.
- `AsyncCompleter.Complete/Fail` — thin delegation, no payload drift possible.
- ~68 registered delegate domains (ordersbus, userbus, etc.) — all conform to standard `ActionXxxParms` shape with `ActionXxxData` helpers.
- Inventory CRUD HTTP surface (`transferorderapi`, `putawaytaskapi`, `picktaskapi`, `cyclecountitemapi`, `cyclecountsessionapi`) — straight passthrough.
- Procurement/sales/floor HTTP handlers — return app-layer result directly, no inline twin.
- `call_webhook`, `transition_status`, `update_field`, `log_audit_entry`, `create_entity`, `lookup_entity` action handlers — single-writer skeletons.
- `formdataregistry` — named-struct contract at the boundary.
- WebSocket hub broadcast targeting — deterministic, no drift surface.
- `alert_updated` and `create_alert` pipelines — single-writer end-to-end.

---

*Audit complete. No code changes made.*
