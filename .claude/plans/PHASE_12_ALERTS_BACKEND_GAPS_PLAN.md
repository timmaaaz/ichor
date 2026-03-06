# Phase 12: Workflow Alerts & Notifications — Backend Gap Analysis

**Date:** 2026-02-26
**Source:** Frontend repo `ichor-floor-worker-phase-12` — audit of all 14 Phase 12 endpoints
**Purpose:** Drive backend work needed before Phase 12 frontend can be completed

---

## What's Already Solid

All 14 specified endpoints exist and are fully wired. The core CRUD, WebSocket hub, RabbitMQ consumer, and database schema are production-grade. The WS payload already carries `title`, `severity`, and `message` — enough to render toast notifications without a secondary REST fetch. Don't touch any of this.

**Alert endpoints (all implemented):**
- `GET /v1/workflow/alerts` — admin list with filters
- `GET /v1/workflow/alerts/mine` — authenticated user's alerts
- `GET /v1/workflow/alerts/{id}` — single alert detail
- `POST /v1/workflow/alerts/{id}/acknowledge` — acknowledge (with optional notes)
- `POST /v1/workflow/alerts/{id}/dismiss` — dismiss
- `POST /v1/workflow/alerts/acknowledge-selected` — bulk ack, returns `{count, skipped}`
- `POST /v1/workflow/alerts/acknowledge-all` — ack all active
- `POST /v1/workflow/alerts/dismiss-selected` — bulk dismiss
- `POST /v1/workflow/alerts/dismiss-all` — dismiss all active (skips acknowledged)
- `GET /v1/workflow/alerts/ws` — WebSocket with JWT via `?token=`, RabbitMQ-backed

**Approval endpoints (all implemented):**
- `GET /v1/workflow/approvals` — admin list
- `GET /v1/workflow/approvals/mine` — user's pending approvals
- `GET /v1/workflow/approvals/{id}` — single approval detail
- `POST /v1/workflow/approvals/{id}/resolve` — approve/reject with optional reason

---

## GAP 1 — Missing Notification Summary Endpoint

**Severity: Blocker for notification bell UX**
**Effort: Small**

The Phase 12 spec calls for a "notification bell with unread badge (persistent across app)." Today there is no lightweight endpoint for this — the frontend must make two paginated list calls just to show two numbers.

**Current workaround (inefficient):**
```
GET /v1/workflow/alerts/mine?status=active&rows=1   → read response.total
GET /v1/workflow/approvals/mine?status=pending&rows=1 → read response.total
```
Two round trips, fetches and discards actual records, on every page load and every poll cycle.

**What to build:**
```
GET /v1/workflow/notifications/summary
```

**Response:**
```json
{
  "alerts": {
    "active": 7,
    "by_severity": {
      "critical": 1,
      "high": 2,
      "medium": 3,
      "low": 1
    }
  },
  "approvals": {
    "pending": 3
  }
}
```

**Implementation path:**
- New handler in a `notificationsapi` package (or extend `alertapi`)
- Calls `alertBus.CountMine(ctx, userID, roleIDs, filter{status:"active"})` — method already exists (`alertbus.go:180`)
- Add a `CountMineBySeverity()` method to `alertbus` that returns `map[string]int` — or execute a single `SELECT severity, COUNT(*) FROM workflow.alerts JOIN workflow.alert_recipients ... GROUP BY severity` query
- Calls `approvalBus.CountMine(ctx, userID, filter{status:"pending"})` — verify method exists; if not, add mirroring `alertBus.CountMine`
- One HTTP call replaces two, no result data transferred

**Also consider:** Push this over WebSocket when counts change so the bell badge updates in real-time without polling.

---

## GAP 2 — Missing Acknowledgment Details in Alert Response

**Severity: High — affects audit trail display and supervisor UX**
**Effort: Small**

The `workflow.alert_acknowledgments` table records who acknowledged an alert, when, and with what notes. None of this is surfaced in the API response. The alert response only changes its `status` to `"acknowledged"`.

**Current response (after ack):**
```json
{ "status": "acknowledged", "updatedDate": "2026-02-25T..." }
```

**What the supervisor dashboard needs:**
```json
{
  "status": "acknowledged",
  "acknowledgments": [
    {
      "acknowledgedBy": "Jane Smith",
      "acknowledgedById": "uuid",
      "acknowledgedDate": "2026-02-25T14:32:00Z",
      "notes": "Checked and confirmed — reorder placed"
    }
  ]
}
```

**Implementation path:**
- Add `Acknowledgments []AlertAcknowledgmentVM` to `alertapi.Alert` response struct (`api/domain/http/workflow/alertapi/model.go`)
- Add `AlertAcknowledgmentVM` struct with `acknowledgedBy` (name), `acknowledgedById`, `acknowledgedDate`, `notes`
- In `alertapi.go`, after fetching recipients, fetch acknowledgments via `alertBus.QueryAcknowledgments(ctx, alertID)` — add this method to `alertbus` and `alertdb`
- JOIN with `core.users` to resolve name + email like `buildEnrichedRecipients` does today (`alertapi.go:551`)
- Apply `omitempty` so it is absent on `active` alerts — zero overhead for the common case

---

## GAP 3 — Missing Approver Name Enrichment on Approval Response

**Severity: High — the approval detail screen cannot show "who else needs to approve"**
**Effort: Small**

The `approvers` array in the approval response is `["uuid1", "uuid2", ...]` — raw UUIDs. The alert API enriches recipients with name + email (`alertapi.go:551`); the approval API does not.

**Current:**
```json
{ "approvers": ["3f2a...", "7c1b..."] }
```

**What the approval detail screen needs:**
```json
{
  "approvers": [
    { "id": "3f2a...", "name": "Marcus Chen", "email": "m.chen@...", "type": "user" },
    { "id": "7c1b...", "name": "Warehouse Supervisor", "type": "role" }
  ],
  "resolvedByDetails": { "id": "...", "name": "Jane Smith", "email": "..." }
}
```

**Implementation path:**
- Add `Approvers []ApproverVM` to the `Approval` HTTP model (`api/domain/http/workflow/approvalapi/model.go`)
- Add `ResolvedByDetails *ApproverVM` for the resolved case
- Reuse the `buildEnrichedRecipients` pattern from `alertapi.go:551`
- Approver IDs are user UUIDs (verify no role UUIDs in approvals), so lookup is simpler than alert recipients

---

## GAP 4 — Missing Rule Name Inline on Alert & Approval Responses

**Severity: Medium — requires N extra round-trips to show human context**
**Effort: Small**

Both alert and approval responses omit the rule name. To show "triggered by rule: _Low Stock Reorder Alert_", the frontend must call `GET /v1/workflow/rules/{sourceRuleId}` per record — an N+1 problem on list views.

The execution endpoint (`GET /v1/workflow/executions/{id}`) already resolves `rule_name` via LEFT JOIN as a pattern to follow.

**What to add to the alert response:**
```json
{ "sourceRuleName": "Low Stock Reorder Alert" }
```

**What to add to the approval response:**
```json
{ "ruleName": "Purchase Order Approval — High Value" }
```

**Implementation path:**
- In `alertdb.go`, extend `Query` and `QueryByUserID` SELECT to LEFT JOIN `workflow.automation_rules` on `source_rule_id`, include `rules.name AS source_rule_name`
- Add `SourceRuleName string \`json:"sourceRuleName,omitempty"\`` to `alertapi.Alert`
- Same pattern for `approvaldb.go` — LEFT JOIN on `rule_id` to get `rule_name`
- Zero extra round trips, one SQL JOIN per query

---

## GAP 5 — Alert Expiry Not Enforced in Queries

**Severity: Medium — stale alerts pollute the feed and inflate badge count**
**Effort: Small**

Expired alerts (`expires_date < NOW()`) still appear in `QueryMine` results with `status = 'active'` because no SQL filter excludes them. The partial index `idx_alerts_expires_date WHERE expires_date IS NOT NULL` already exists for this but is unused in queries.

**Two fixes needed:**

**Fix A — Expire-aware query filter (immediate):**
In `applyFilterWithJoin()` and `applyFilter()` in `alertdb.go`, when `Status = 'active'` is set, add:
```sql
AND (a.expires_date IS NULL OR a.expires_date > NOW())
```

**Fix B — Background expiry job:**
A periodic cleanup job (Temporal cron or Go ticker) that runs:
```sql
UPDATE workflow.alerts
SET status = 'resolved', updated_date = NOW()
WHERE expires_date IS NOT NULL
  AND expires_date < NOW()
  AND status = 'active'
```
Keeps DB in consistent state so badge counts are accurate without query-time filtering.

---

## GAP 6 — Missing Date Range Filters on Alert Queries

**Severity: Medium — affects supervisor dashboard and alert history views**
**Effort: Small**

The `QueryFilter` struct (`alertbus/filter.go`) has no date fields. Supervisors need to query "alerts from the last 7 days." Currently impossible without fetching all pages and filtering client-side.

**What to add to `alertbus/filter.go`:**
```go
CreatedAfter  *time.Time
CreatedBefore *time.Time
```

**Implementation path:**
- Add `created_after` and `created_before` to `alertbus.QueryFilter`
- Extend `applyFilter()` and `applyFilterWithJoin()` in `alertdb.go` to emit:
  ```sql
  AND a.created_date >= :created_after   -- when set
  AND a.created_date <= :created_before  -- when set
  ```
- Expose as `createdAfter` and `createdBefore` query params in `alertapi/filter.go`
- Existing index `idx_alerts_created_date` already covers this — no new indexes needed

---

## GAP 7 — Missing Source Entity Route Hint

**Severity: Medium — deep-links are a core Phase 12 feature**
**Effort: Small (migration + one field)**

The alert spec says: `"Low stock for SKU-4422 in Zone B" → taps to inventory item`. The backend returns `sourceEntityName: "inventory.products"` and `sourceEntityId: "uuid"`. The frontend has no way to map `"inventory.products"` → `/inventory/products/{id}` without a hardcoded client-side lookup table.

**Recommended approach — Add `frontend_route` to entity type registry:**

Add a nullable `frontend_route TEXT` column to `workflow.entity_types` (e.g., `/inventory/products/{id}`):
```json
GET /v1/workflow/entity-types
→ [{ "id": "...", "name": "inventory.products", "frontend_route": "/inventory/products/{id}", ... }]
```

The frontend fetches this once on startup and caches it. Substitutes `{id}` with `sourceEntityId` when rendering alert deep-links.

**Implementation path:**
- Migration: `ALTER TABLE workflow.entity_types ADD COLUMN frontend_route TEXT;`
- Backfill known entity types with their frontend routes
- Add `FrontendRoute string \`json:"frontend_route,omitempty"\`` to `EntityType` struct in `referenceapi/model.go`
- No handler changes needed — the existing `GET /v1/workflow/entity-types` endpoint already returns the full struct

---

## GAP 8 — WebSocket Has No Status-Change Events

**Severity: Medium — creates stale UI state**
**Effort: Medium**

The WebSocket only fires on new alert creation. When another user acknowledges an alert you're also a recipient of, your UI doesn't update until next poll or page refresh. On the supervisor dashboard showing a team's live alert feed, this creates stale state.

**New WS message types to add:**

```json
{ "type": "alert_updated", "payload": { "id": "...", "status": "acknowledged", "updatedDate": "..." } }
{ "type": "approval_updated", "payload": { "id": "...", "status": "approved", "resolvedBy": "...", "resolvedDate": "..." } }
```

**Implementation path:**
- Add `MessageTypeAlertUpdated = "alert_updated"` and `MessageTypeApprovalUpdated = "approval_updated"` to `foundation/websocket/message.go`
- After `alertBus.Acknowledge()` or `Dismiss()` completes, publish a lightweight `alert_updated` RabbitMQ message (same queue, same routing) with `id`, `status`, `updatedDate`
- The existing `AlertConsumer` broadcasts to the same recipient targeting as the original `alert` message
- After `approvalBus.Resolve()`, publish `approval_updated` with `id`, `status`, `resolvedBy`, `resolvedDate`
- Frontend updates records in-place without re-fetching the full list

---

## GAP 9 — No Composite DB Index for Primary Query Pattern

**Severity: Low-Medium — performance concern at scale**
**Effort: Trivial (migration only)**

The most common query — "my active alerts, newest first" — JOINs `workflow.alert_recipients` → `workflow.alerts` with `status = 'active'` ORDER BY `created_date DESC`. At scale (thousands of alerts), a composite covering index would help:

```sql
CREATE INDEX idx_alerts_status_created ON workflow.alerts(status, created_date DESC);
```

Migration-only change, no code impact.

---

## GAP 10 — `dismiss-all` Skips Acknowledged Alerts

**Severity: Low — behavioral quirk to document or fix**
**Effort: Trivial**

`DismissMultiple` in `alertdb.go` uses `WHERE status = 'active'` — so "dismiss all" silently skips already-acknowledged alerts. The response `count` will be lower than the visible alert count in the UI, which is confusing if the user sees acknowledged alerts in the feed.

**Fix options:**
1. Change `DismissMultiple` to also dismiss `'acknowledged'` alerts: `IN ('active', 'acknowledged')`
2. Document it as intentional in API comments and surface `skipped` count meaning in docs
3. Add a `clear-all` endpoint that does both ack + dismiss in one pass

---

## GAP 11 — No PWA Push Notification Infrastructure

**Severity: Low for Phase 12 launch, High for floor worker completeness**
**Effort: Large (standalone feature)**

Currently: **in-app only**. No VAPID key setup, no service worker endpoint, no subscription storage, no FCM/APNs integration. Floor workers who close the app between scans miss critical alerts entirely.

**Minimum viable addition:**
- `POST /v1/notifications/push-subscription` — store `PushSubscription` from browser Web Push API
- `GET /v1/notifications/push-subscription` — check if current device is subscribed
- `DELETE /v1/notifications/push-subscription` — unsubscribe
- New `webpush` package wrapping a Go web push library (e.g., `github.com/SherClockHolmes/webpush-go`)
- Trigger from the `create_alert` action handler after DB write + RabbitMQ publish

**Schema needed:**
```sql
CREATE TABLE core.push_subscriptions (
    id           UUID PRIMARY KEY,
    user_id      UUID NOT NULL REFERENCES core.users(id) ON DELETE CASCADE,
    endpoint     TEXT NOT NULL,
    p256dh       TEXT NOT NULL,
    auth         TEXT NOT NULL,
    device_name  TEXT,
    created_date TIMESTAMP NOT NULL
);
```

This is Phase 12+ scope — plan as a follow-up.

---

## Additional Observations (No Code Changes Needed)

- **`alert_type` is free-form string** — no server-side enum validation. Convention-based values (`"inventory_warning"`, `"approval_required"`, etc.) should be documented per deployment. The frontend cannot enumerate valid types without documentation.
- **`context` JSONB template vars are not resolved** — `{{product_id}}` in the `context` field is stored literally. Only `title` and `message` get template variable substitution. Rule authors must pre-resolve context values or the frontend receives raw placeholder strings.
- **`GET /v1/workflow/approvals/{id}` has no ownership guard** — any authenticated user can read any approval by UUID. Intentional for email deep-link patterns, but differs from `/mine` scoping. Document this behavior.
- **Temporal `Complete()` is best-effort** — if the Temporal callback fails after a successful DB resolve, the HTTP response still returns 200. Without a timeout/reconciliation mechanism, workflows could get stuck. Backend operational concern, not a frontend issue.
- **Missed WS messages are not replayed** — the hub is purely in-memory. On reconnect, the canonical recovery is `GET /v1/workflow/alerts/mine`. The frontend must call this after every WS reconnection to resync state.

---

## Prioritized Build Order

| # | Gap | Effort | Impact | Ship with Phase 12? |
|---|---|---|---|---|
| 1 | `GET /v1/workflow/notifications/summary` endpoint | Small | Critical (badge UX) | **Yes** |
| 2 | Acknowledgment details in alert response | Small | High | **Yes** |
| 3 | Approver name enrichment on approval response | Small | High | **Yes** |
| 4 | Rule name inline on alert + approval responses | Small | Medium | **Yes** |
| 5 | Alert expiry query filter (`expires_date > NOW()`) | Small | Medium | **Yes** |
| 6 | Date range filters (`createdAfter`, `createdBefore`) | Small | Medium | **Yes** |
| 7 | `frontend_route` on entity type registry | Small | Medium | **Yes** |
| 8 | WS `alert_updated` / `approval_updated` message types | Medium | High | **Yes** |
| 9 | Composite index `(status, created_date DESC)` | Trivial | Perf | **Yes** |
| 10 | Background alert expiry job | Medium | Medium | Optional |
| 11 | `dismiss-all` behavioral fix or documentation | Trivial | Low | **Yes** |
| 12 | PWA push notification infrastructure | Large | High (floor workers) | Phase 12+ |

Items 1–9 + 11 are all targeted, low-risk changes. Items 1–7 are pure additions with no breaking changes to existing response shapes (all new fields are `omitempty`). Item 8 is additive to the WebSocket protocol. Item 9 is a migration-only change.
