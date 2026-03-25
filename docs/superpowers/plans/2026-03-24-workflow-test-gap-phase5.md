# Phase 5: Integration Test Completeness

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add integration tests for `notificationsapi`, constructor tests for `ResendEmailClient`, audit conditional skips in CI, and assess `actionpermissionsapi` status.

**Architecture:** Integration tests use `apitest.Table` pattern for HTTP endpoints. ResendEmailClient tests are pure unit tests. The conditional skip audit is investigative (read-only).

**Tech Stack:** Go testing, `apitest.Table`, `cmp.diff`

**Spec:** `docs/superpowers/specs/2026-03-24-workflow-test-gap-remediation-design.md` (Phase 5)

---

### Task 1: notificationsapi Integration Tests

**Files:**
- Create: `api/cmd/services/ichor/tests/workflow/notificationsapi/notification_test.go`
- Create: `api/cmd/services/ichor/tests/workflow/notificationsapi/seed_test.go`
- Reference: `api/domain/http/workflow/notificationsapi/notificationsapi.go`
- Reference: `api/domain/http/workflow/notificationsapi/route.go` (GET `/v1/workflow/notifications/summary`)
- Reference: `api/domain/http/workflow/notificationsapi/model.go`
- Pattern: `api/cmd/services/ichor/tests/workflow/alertapi/` (follow this test structure)

- [ ] **Step 1: Read the existing integration test pattern**

Read these files to understand the apitest pattern:
- `api/cmd/services/ichor/tests/workflow/alertapi/alert_test.go`
- `api/cmd/services/ichor/tests/workflow/alertapi/seed_test.go`
- `api/cmd/services/ichor/tests/workflow/alertapi/query_test.go`

Key things to note:
- How `apitest.Table` is structured
- How the test DB and app are initialized
- How seed data is created for the test
- How authentication tokens are obtained

- [ ] **Step 2: Read the notificationsapi model**

Read `api/domain/http/workflow/notificationsapi/model.go` to understand `NotificationSummary`, `AlertSummary`, `ApprovalSummary` response structures.

- [ ] **Step 3: Discovery — study the apitest pattern before writing code**

This is a discovery-first task. Before writing any test code:

1. Read `api/cmd/services/ichor/tests/workflow/alertapi/alert_test.go` completely
2. Read `api/cmd/services/ichor/tests/workflow/alertapi/seed_test.go` completely
3. Read `api/cmd/services/ichor/tests/workflow/approvalapi/seed_test.go`

Understand how the test harness, auth tokens, seed data, and `apitest.Table` entries work.

- [ ] **Step 4: Write seed_test.go and notification_test.go**

Based on what you learned in Step 3, implement:

**seed_test.go:** Create seed function with:
- A test user with roles (for authentication)
- 2-3 alerts with varying severities, user as recipient, status=active
- 1-2 approval requests with user as approver, status=pending
- 1 resolved approval (to verify it's excluded from count)

**notification_test.go:** Using `apitest.Table` pattern:
- `GET /v1/workflow/notifications/summary` with valid auth → 200, verify alert counts match seeded data, verify pending approval count
- `GET /v1/workflow/notifications/summary` without auth → 401

The endpoint is:
- `GET /v1/workflow/notifications/summary` (requires authentication)
- Returns `NotificationSummary` with `Alerts` (severity breakdown) and `Approvals` (pending count)

- [ ] **Step 5: Run tests**

Run: `go test ./api/cmd/services/ichor/tests/workflow/notificationsapi/... -v -count=1`

- [ ] **Step 6: Commit**

```
git add api/cmd/services/ichor/tests/workflow/notificationsapi/
git commit -m "test(notificationsapi): add integration tests for notification summary endpoint"
```

---

### Task 2: ResendEmailClient Constructor Tests

**Files:**
- Create: `business/sdk/workflow/workflowactions/communication/resend_test.go`
- Reference: `business/sdk/workflow/workflowactions/communication/resend.go`

- [ ] **Step 1: Write constructor tests**

```go
package communication_test

import (
	"testing"

	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/communication"
)

func TestNewResendEmailClient_EmptyAPIKey(t *testing.T) {
	client := communication.NewResendEmailClient(communication.ResendConfig{
		APIKey: "",
		From:   "test@example.com",
	})
	if client != nil {
		t.Fatal("expected nil for empty API key")
	}
}

func TestNewResendEmailClient_EmptyFrom(t *testing.T) {
	client := communication.NewResendEmailClient(communication.ResendConfig{
		APIKey: "re_test_key",
		From:   "",
	})
	if client != nil {
		t.Fatal("expected nil for empty From")
	}
}

func TestNewResendEmailClient_BothEmpty(t *testing.T) {
	client := communication.NewResendEmailClient(communication.ResendConfig{})
	if client != nil {
		t.Fatal("expected nil for empty config")
	}
}

func TestNewResendEmailClient_ValidConfig(t *testing.T) {
	client := communication.NewResendEmailClient(communication.ResendConfig{
		APIKey: "re_test_key_12345",
		From:   "noreply@example.com",
	})
	if client == nil {
		t.Fatal("expected non-nil client for valid config")
	}
}
```

Uses external test package since only exported APIs are tested.

- [ ] **Step 2: Run tests**

Run: `go test ./business/sdk/workflow/workflowactions/communication/... -run TestNewResendEmailClient -v -count=1`
Expected: All PASS.

- [ ] **Step 3: Commit**

```
git add business/sdk/workflow/workflowactions/communication/resend_test.go
git commit -m "test(resend): add constructor graceful degradation tests"
```

---

### Task 3: Conditional Skip Audit

This is investigative — no code changes expected.

- [ ] **Step 1: Check CI configuration for test mode**

Look for Makefile targets and CI config:
- `Makefile` — check `test`, `test-only`, `test-race` targets for `-short` flag
- `.github/workflows/` or equivalent CI config
- Check if seed data is populated before integration tests run

Key questions:
1. Does CI run with `-short`? If yes, `workflow_replay_test.go` integration tests never run in CI.
2. Does CI seed the database? If not, `actionhandlers/inventory_test.go` always skips.
3. Are there at least 3 trigger types seeded? If not, `workflowsaveapi/errors_test.go` always skips.

- [ ] **Step 2: Document findings**

Run:
```bash
grep -n "short" Makefile
grep -rn "short" .github/ || true
grep -n "seed" Makefile
```

- [ ] **Step 3: Create a findings document or commit message**

If skips are effectively dead code in CI, document this as a follow-up item. Do not fix in this phase — just identify.

```
git commit --allow-empty -m "audit: document conditional skip findings in workflow tests

Findings:
- [document what you found about -short mode in CI]
- [document what you found about seed data availability]
- [document whether these tests actually run]"
```

---

### Task 4: actionpermissionsapi Assessment

- [ ] **Step 1: Check if API endpoints exist**

```bash
grep -rn "actionpermissions" api/domain/http/ || true
grep -rn "actionpermissions" api/cmd/services/ichor/build/
```

- [ ] **Step 2: Check route registration**

Read `api/cmd/services/ichor/build/all/all.go` to see if `actionpermissionsapi` routes are registered.

- [ ] **Step 3: Document assessment**

If no API layer exists:
- Document that `actionpermissionsbus` is internal-only (used by `ActionService` permission checks)
- No integration tests needed — covered by Phase 1 unit tests

If API endpoints exist but have no tests:
- Add integration tests following the `apitest.Table` pattern
- This is "discovered work" per the spec

- [ ] **Step 4: Commit findings**

```
git commit --allow-empty -m "audit: actionpermissionsapi assessment

[Document whether API layer exists and whether integration tests are needed]"
```
