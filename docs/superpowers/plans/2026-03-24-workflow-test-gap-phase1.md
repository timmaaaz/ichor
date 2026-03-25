# Phase 1: Security & Permissions Tests

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add unit tests for `actionpermissionsbus` (CRUD + permission checks) and `data/tables.go` (SQL validation whitelist/regex).

**Architecture:** Two independent test files. `actionpermissionsbus` tests need real Postgres via `dbtest`. `tables_test.go` is pure unit tests with no dependencies.

**Tech Stack:** Go testing, `dbtest`, `cmp.diff`, `unitest`

**Spec:** `docs/superpowers/specs/2026-03-24-workflow-test-gap-remediation-design.md` (Phase 1)

---

### Task 1: actionpermissionsbus CRUD Tests

**Files:**
- Create: `business/domain/workflow/actionpermissionsbus/actionpermissionsbus_test.go`
- Reference: `business/domain/workflow/actionpermissionsbus/actionpermissionsbus.go`
- Reference: `business/domain/workflow/actionpermissionsbus/testutil.go`
- Reference: `business/domain/workflow/actionpermissionsbus/model.go`
- Pattern: `business/domain/workflow/alertbus/alertbus_test.go` (follow this test structure)

- [ ] **Step 1: Create test file with seed data function**

The test file follows the pattern from `alertbus_test.go`: one top-level `Test_ActionPermissions` that creates a DB, seeds data, then runs subtests via `unitest.Run`.

```go
package actionpermissionsbus_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"sort"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/actionpermissionsbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_ActionPermissions(t *testing.T) {
	db := dbtest.NewDatabase(t, "Test_ActionPermissions")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unitest.Run(t, crudTests(db.BusDomain, sd), "crud")
	unitest.Run(t, canUserExecuteTests(db.BusDomain, sd), "canUserExecute")
	unitest.Run(t, getAllowedActionsTests(db.BusDomain, sd), "getAllowedActions")
}

type seedData struct {
	Users       []userbus.User
	Permissions []actionpermissionsbus.ActionPermission
	RoleIDs     []uuid.UUID
}

func insertSeedData(busDomain dbtest.BusDomain) (seedData, error) {
	ctx := context.Background()

	users, err := userbus.TestSeedUsersWithNoFKs(ctx, 2, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding users: %w", err)
	}

	// IMPORTANT: userbus.User does NOT have RoleIDs []uuid.UUID.
	// You must obtain real role UUIDs from the database. Check how other
	// test files get role UUIDs — likely via busDomain.Role or by querying
	// core.roles / core.user_roles tables. Read the rolebus or userrolebus
	// package to find the right approach. The key is: actionpermissionsbus
	// needs a real uuid.UUID from the roles table, not a userbus.Role string.
	//
	// Example pattern (adapt based on what actually exists):
	//   roles, err := rolebus.TestSeedRoles(ctx, 2, busDomain.Role)
	//   roleID1 := roles[0].ID
	//   roleID2 := roles[1].ID
	//
	// Or query existing seeded roles:
	//   allRoles, err := busDomain.Role.Query(ctx, rolebus.QueryFilter{}, ...)
	//   roleID1 := allRoles[0].ID
	//
	// Read business/domain/core/rolebus/ and business/sdk/dbtest/dbtest.go
	// to determine the correct approach before writing this seed function.
	var roleID1, roleID2 uuid.UUID // Replace with actual role UUID retrieval

	// Seed permissions: role1 can send_email and create_alert, role2 can send_email
	perms := make([]actionpermissionsbus.ActionPermission, 0)

	p1, err := actionpermissionsbus.TestSeedActionPermission(ctx, busDomain.ActionPermissions, roleID1, "send_email", true)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding perm1: %w", err)
	}
	perms = append(perms, p1)

	p2, err := actionpermissionsbus.TestSeedActionPermission(ctx, busDomain.ActionPermissions, roleID1, "create_alert", true)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding perm2: %w", err)
	}
	perms = append(perms, p2)

	p3, err := actionpermissionsbus.TestSeedActionPermission(ctx, busDomain.ActionPermissions, roleID2, "send_email", true)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding perm3: %w", err)
	}
	perms = append(perms, p3)

	// Seed a denied permission
	p4, err := actionpermissionsbus.TestSeedActionPermission(ctx, busDomain.ActionPermissions, roleID2, "update_field", false)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding perm4: %w", err)
	}
	perms = append(perms, p4)

	return seedData{
		Users:       users,
		Permissions: perms,
		RoleIDs:     []uuid.UUID{roleID1, roleID2},
	}, nil
}
```

Note: Check if `busDomain.ActionPermissions` exists on the `dbtest.BusDomain` struct. If it does not, you will need to check how other bus domains are wired in `dbtest` and add it. This is "discovered work" per the spec.

- [ ] **Step 2: Write CRUD subtest function**

```go
func crudTests(busDomain dbtest.BusDomain, sd seedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "query",
			ExpFunc: func(ctx context.Context) any {
				// Query all permissions from DB
				filter := actionpermissionsbus.QueryFilter{}
				perms, err := busDomain.ActionPermissions.Query(ctx, filter, actionpermissionsbus.DefaultOrderBy, page.MustParse("1", "100"))
				if err != nil {
					return err
				}
				return perms
			},
			CmpFunc: func(ctx context.Context, got any) string {
				if err, ok := got.(error); ok {
					return fmt.Sprintf("query error: %s", err)
				}
				perms := got.([]actionpermissionsbus.ActionPermission)
				if len(perms) < len(sd.Permissions) {
					return fmt.Sprintf("expected at least %d permissions, got %d", len(sd.Permissions), len(perms))
				}
				// Verify seeded permissions are in the result set
				for _, exp := range sd.Permissions {
					found := false
					for _, p := range perms {
						if p.ID == exp.ID {
							if diff := cmp.Diff(exp, p); diff != "" {
								return fmt.Sprintf("permission %s mismatch:\n%s", exp.ID, diff)
							}
							found = true
							break
						}
					}
					if !found {
						return fmt.Sprintf("seeded permission %s not found in query results", exp.ID)
					}
				}
				return ""
			},
		},
		{
			Name: "queryByID",
			ExpFunc: func(ctx context.Context) any {
				return sd.Permissions[0]
			},
			CmpFunc: func(ctx context.Context, got any) string {
				exp := got.(actionpermissionsbus.ActionPermission)
				perm, err := busDomain.ActionPermissions.QueryByID(ctx, exp.ID)
				if err != nil {
					return fmt.Sprintf("queryByID error: %s", err)
				}
				return cmp.Diff(exp, perm)
			},
		},
		{
			Name: "queryByRoleAndAction",
			ExpFunc: func(ctx context.Context) any {
				return sd.Permissions[0]
			},
			CmpFunc: func(ctx context.Context, got any) string {
				exp := got.(actionpermissionsbus.ActionPermission)
				perm, err := busDomain.ActionPermissions.QueryByRoleAndAction(ctx, exp.RoleID, exp.ActionType)
				if err != nil {
					return fmt.Sprintf("queryByRoleAndAction error: %s", err)
				}
				return cmp.Diff(exp, perm)
			},
		},
		{
			Name: "create-update-delete",
			ExpFunc: func(ctx context.Context) any {
				return nil
			},
			CmpFunc: func(ctx context.Context, got any) string {
				// Create
				newPerm, err := busDomain.ActionPermissions.Create(ctx, actionpermissionsbus.NewActionPermission{
					RoleID:      sd.RoleIDs[0],
					ActionType:  "call_webhook",
					IsAllowed:   true,
					Constraints: json.RawMessage("{}"),
				})
				if err != nil {
					return fmt.Sprintf("create error: %s", err)
				}
				if newPerm.ActionType != "call_webhook" {
					return fmt.Sprintf("expected action_type call_webhook, got %s", newPerm.ActionType)
				}

				// Update
				isAllowed := false
				updated, err := busDomain.ActionPermissions.Update(ctx, newPerm, actionpermissionsbus.UpdateActionPermission{
					IsAllowed: &isAllowed,
				})
				if err != nil {
					return fmt.Sprintf("update error: %s", err)
				}
				if updated.IsAllowed != false {
					return "expected IsAllowed=false after update"
				}

				// Delete
				if err := busDomain.ActionPermissions.Delete(ctx, updated); err != nil {
					return fmt.Sprintf("delete error: %s", err)
				}

				// Verify deleted
				_, err = busDomain.ActionPermissions.QueryByID(ctx, newPerm.ID)
				if err == nil {
					return "expected error after delete, got nil"
				}
				return ""
			},
		},
		{
			Name: "count",
			ExpFunc: func(ctx context.Context) any {
				return len(sd.Permissions)
			},
			CmpFunc: func(ctx context.Context, got any) string {
				expCount := got.(int)
				count, err := busDomain.ActionPermissions.Count(ctx, actionpermissionsbus.QueryFilter{})
				if err != nil {
					return fmt.Sprintf("count error: %s", err)
				}
				if count < expCount {
					return fmt.Sprintf("expected at least %d, got %d", expCount, count)
				}
				return ""
			},
		},
	}
}
```

- [ ] **Step 3: Write CanUserExecuteAction subtest function**

```go
func canUserExecuteTests(busDomain dbtest.BusDomain, sd seedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "role-with-permission",
			ExpFunc: func(ctx context.Context) any {
				return true
			},
			CmpFunc: func(ctx context.Context, got any) string {
				// Role1 has send_email permission
				can, err := busDomain.ActionPermissions.CanUserExecuteAction(ctx, sd.Users[0].ID, "send_email", []uuid.UUID{sd.RoleIDs[0]})
				if err != nil {
					return fmt.Sprintf("error: %s", err)
				}
				if !can {
					return "expected true, got false"
				}
				return ""
			},
		},
		{
			Name: "role-without-permission",
			ExpFunc: func(ctx context.Context) any {
				return false
			},
			CmpFunc: func(ctx context.Context, got any) string {
				// Role1 does not have update_field permission
				can, err := busDomain.ActionPermissions.CanUserExecuteAction(ctx, sd.Users[0].ID, "update_field", []uuid.UUID{sd.RoleIDs[0]})
				if err != nil {
					return fmt.Sprintf("error: %s", err)
				}
				if can {
					return "expected false, got true"
				}
				return ""
			},
		},
		{
			Name: "denied-permission",
			ExpFunc: func(ctx context.Context) any {
				return false
			},
			CmpFunc: func(ctx context.Context, got any) string {
				// Role2 has update_field with is_allowed=false
				can, err := busDomain.ActionPermissions.CanUserExecuteAction(ctx, sd.Users[1].ID, "update_field", []uuid.UUID{sd.RoleIDs[1]})
				if err != nil {
					return fmt.Sprintf("error: %s", err)
				}
				if can {
					return "expected false (denied), got true"
				}
				return ""
			},
		},
		{
			Name: "multiple-roles-one-allowed",
			ExpFunc: func(ctx context.Context) any {
				return true
			},
			CmpFunc: func(ctx context.Context, got any) string {
				// Both roles, one has send_email
				can, err := busDomain.ActionPermissions.CanUserExecuteAction(ctx, sd.Users[0].ID, "send_email", sd.RoleIDs)
				if err != nil {
					return fmt.Sprintf("error: %s", err)
				}
				if !can {
					return "expected true with multiple roles, got false"
				}
				return ""
			},
		},
		{
			Name: "no-roles",
			ExpFunc: func(ctx context.Context) any {
				return false
			},
			CmpFunc: func(ctx context.Context, got any) string {
				can, err := busDomain.ActionPermissions.CanUserExecuteAction(ctx, sd.Users[0].ID, "send_email", []uuid.UUID{})
				if err != nil {
					return fmt.Sprintf("error: %s", err)
				}
				if can {
					return "expected false with no roles, got true"
				}
				return ""
			},
		},
	}
}
```

- [ ] **Step 4: Write GetAllowedActionsForRoles subtest function**

```go
func getAllowedActionsTests(busDomain dbtest.BusDomain, sd seedData) []unitest.Table {
	return []unitest.Table{
		{
			Name: "single-role",
			ExpFunc: func(ctx context.Context) any {
				return []string{"create_alert", "send_email"}
			},
			CmpFunc: func(ctx context.Context, got any) string {
				actions, err := busDomain.ActionPermissions.GetAllowedActionsForRoles(ctx, []uuid.UUID{sd.RoleIDs[0]})
				if err != nil {
					return fmt.Sprintf("error: %s", err)
				}
				sort.Strings(actions)
				exp := got.([]string)
				return cmp.Diff(exp, actions)
			},
		},
		{
			Name: "empty-roles",
			ExpFunc: func(ctx context.Context) any {
				return []string{}
			},
			CmpFunc: func(ctx context.Context, got any) string {
				actions, err := busDomain.ActionPermissions.GetAllowedActionsForRoles(ctx, []uuid.UUID{})
				if err != nil {
					return fmt.Sprintf("error: %s", err)
				}
				exp := got.([]string)
				return cmp.Diff(exp, actions)
			},
		},
		{
			Name: "multiple-roles-aggregated",
			ExpFunc: func(ctx context.Context) any {
				return nil
			},
			CmpFunc: func(ctx context.Context, got any) string {
				actions, err := busDomain.ActionPermissions.GetAllowedActionsForRoles(ctx, sd.RoleIDs)
				if err != nil {
					return fmt.Sprintf("error: %s", err)
				}
				// Should include send_email and create_alert (both allowed for role1)
				// role2 has send_email=true and update_field=false
				if len(actions) < 2 {
					return fmt.Sprintf("expected at least 2 allowed actions, got %d: %v", len(actions), actions)
				}
				return ""
			},
		},
	}
}
```

Add `"sort"` to imports.

- [ ] **Step 5: Verify the test compiles and runs**

Run: `go test ./business/domain/workflow/actionpermissionsbus/... -v -count=1`

If `busDomain.ActionPermissions` doesn't exist, check `business/sdk/dbtest/` for how `BusDomain` is structured and add the field. This is expected "discovered work."

- [ ] **Step 6: Commit**

```
git add business/domain/workflow/actionpermissionsbus/actionpermissionsbus_test.go
git commit -m "test(actionpermissionsbus): add CRUD and permission check unit tests"
```

---

### Task 2: data/tables.go Unit Tests

**Files:**
- Create: `business/sdk/workflow/workflowactions/data/tables_test.go`
- Reference: `business/sdk/workflow/workflowactions/data/tables.go`

- [ ] **Step 1: Write the test file**

Pure unit tests — no DB, no external dependencies.

```go
package data_test

import (
	"testing"

	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/data"
)

func TestIsValidColumnName(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect bool
	}{
		{"valid simple", "status", true},
		{"valid with underscore", "order_number", true},
		{"valid with digits", "field2", true},
		{"valid 63 chars", "abcdefghijklmnopqrstuvwxyz0123456789_abcdefghijklmnopqrst012345", true}, // exactly 63 chars
		{"empty string", "", false},
		{"starts with digit", "1field", false},
		{"starts with underscore", "_field", false},
		{"uppercase", "Status", false},
		{"mixed case", "orderNumber", false},
		{"sql injection semicolon", "status; DROP TABLE users", false},
		{"sql injection comment", "status -- comment", false},
		{"sql injection equals", "1=1", false},
		{"contains space", "order number", false},
		{"contains dot", "table.column", false},
		{"unicode", "naïve", false},
		{"special chars", "col@name", false},
		{"single char", "a", true},
		{"valid long", "abcdefghijklmnopqrstuvwxyz0123456789_abcdefghijklmnopqrst012345", true}, // 63 chars
		{"too long 64 chars", "abcdefghijklmnopqrstuvwxyz0123456789_abcdefghijklmnopqrstu012345", false}, // 64 chars
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := data.IsValidColumnName(tt.input)
			if got != tt.expect {
				t.Errorf("IsValidColumnName(%q) = %v, want %v", tt.input, got, tt.expect)
			}
		})
	}
}

func TestIsValidTableName(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect bool
	}{
		// Known valid entries from the whitelist
		{"valid sales.orders", "sales.orders", true},
		{"valid core.users", "core.users", true},
		{"valid inventory.inventory_items", "inventory.inventory_items", true},
		{"valid workflow.automation_rules", "workflow.automation_rules", true},
		{"valid procurement.purchase_orders", "procurement.purchase_orders", true},
		{"valid products.products", "products.products", true},
		// Unknown table names
		{"unknown table", "sales.nonexistent", false},
		{"unknown schema", "fake.users", false},
		{"empty string", "", false},
		{"no schema", "users", false},
		{"sql fragment", "sales.orders; DROP TABLE users", false},
		{"partial match", "sales.order", false},
		{"extra whitespace", " sales.orders ", false},
		{"uppercase", "Sales.Orders", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := data.IsValidTableName(tt.input)
			if got != tt.expect {
				t.Errorf("IsValidTableName(%q) = %v, want %v", tt.input, got, tt.expect)
			}
		})
	}
}

func TestIsValidOperator(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect bool
	}{
		// All valid operators
		{"equals", "equals", true},
		{"not_equals", "not_equals", true},
		{"greater_than", "greater_than", true},
		{"less_than", "less_than", true},
		{"contains", "contains", true},
		{"is_null", "is_null", true},
		{"is_not_null", "is_not_null", true},
		{"in", "in", true},
		{"not_in", "not_in", true},
		// Invalid operators
		{"empty string", "", false},
		{"sql equals", "=", false},
		{"sql like", "LIKE", false},
		{"unknown", "between", false},
		{"uppercase valid", "EQUALS", false},
		{"space padding", " equals ", false},
		{"sql fragment", "1=1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := data.IsValidOperator(tt.input)
			if got != tt.expect {
				t.Errorf("IsValidOperator(%q) = %v, want %v", tt.input, got, tt.expect)
			}
		})
	}
}
```

- [ ] **Step 2: Run the tests**

Run: `go test ./business/sdk/workflow/workflowactions/data/... -v -count=1`
Expected: All tests PASS. Fix any test cases where the expected value is wrong (e.g., the 63-char column name test — verify the exact regex boundary).

- [ ] **Step 3: Commit**

```
git add business/sdk/workflow/workflowactions/data/tables_test.go
git commit -m "test(data): add unit tests for SQL validation whitelist and regex"
```
