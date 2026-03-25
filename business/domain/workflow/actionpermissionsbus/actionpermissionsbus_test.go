package actionpermissionsbus_test

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/rolebus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/actionpermissionsbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_ActionPermissions(t *testing.T) {
	t.Parallel()

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

	// Get the first seeded role and create a second one. The migration only
	// seeds one role (ZZZADMIN), so we need to create another for multi-role tests.
	roles, err := busDomain.Role.Query(ctx, rolebus.QueryFilter{}, rolebus.DefaultOrderBy, page.MustParse("1", "5"))
	if err != nil {
		return seedData{}, fmt.Errorf("querying roles: %w", err)
	}
	if len(roles) < 1 {
		return seedData{}, fmt.Errorf("expected at least 1 role, got %d", len(roles))
	}

	roleID1 := roles[0].ID

	// Create a second role for multi-role permission tests.
	role2, err := busDomain.Role.Create(ctx, rolebus.NewRole{
		Name:        "test_workflow_role",
		Description: "Test role for action permissions",
	})
	if err != nil {
		return seedData{}, fmt.Errorf("creating test role: %w", err)
	}
	roleID2 := role2.ID

	perms := make([]actionpermissionsbus.ActionPermission, 0)

	// Use unique action type names to avoid collisions with seed data that
	// may already grant permissions to the admin role.
	// role1: test_action_a (allowed), test_action_b (allowed)
	p1, err := actionpermissionsbus.TestSeedActionPermission(ctx, busDomain.ActionPermissions, roleID1, "test_action_a", true)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding perm1: %w", err)
	}
	perms = append(perms, p1)

	p2, err := actionpermissionsbus.TestSeedActionPermission(ctx, busDomain.ActionPermissions, roleID1, "test_action_b", true)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding perm2: %w", err)
	}
	perms = append(perms, p2)

	// role2: test_action_a (allowed), test_action_denied (denied)
	p3, err := actionpermissionsbus.TestSeedActionPermission(ctx, busDomain.ActionPermissions, roleID2, "test_action_a", true)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding perm3: %w", err)
	}
	perms = append(perms, p3)

	p4, err := actionpermissionsbus.TestSeedActionPermission(ctx, busDomain.ActionPermissions, roleID2, "test_action_denied", false)
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

// =========================================================================

func crudTests(busDomain dbtest.BusDomain, sd seedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "query",
			ExpResp: sd.Permissions,
			ExcFunc: func(ctx context.Context) any {
				filter := actionpermissionsbus.QueryFilter{}
				perms, err := busDomain.ActionPermissions.Query(ctx, filter, actionpermissionsbus.DefaultOrderBy, page.MustParse("1", "100"))
				if err != nil {
					return err
				}
				return perms
			},
			CmpFunc: func(got any, exp any) string {
				if err, ok := got.(error); ok {
					return fmt.Sprintf("query error: %s", err)
				}
				gotPerms := got.([]actionpermissionsbus.ActionPermission)
				expPerms := exp.([]actionpermissionsbus.ActionPermission)
				if len(gotPerms) < len(expPerms) {
					return fmt.Sprintf("expected at least %d permissions, got %d", len(expPerms), len(gotPerms))
				}
				for _, e := range expPerms {
					found := false
					for _, g := range gotPerms {
						if g.ID == e.ID {
							found = true
							break
						}
					}
					if !found {
						return fmt.Sprintf("seeded permission %s not found in query results", e.ID)
					}
				}
				return ""
			},
		},
		{
			Name:    "queryByID",
			ExpResp: sd.Permissions[0],
			ExcFunc: func(ctx context.Context) any {
				perm, err := busDomain.ActionPermissions.QueryByID(ctx, sd.Permissions[0].ID)
				if err != nil {
					return err
				}
				return perm
			},
			CmpFunc: func(got any, exp any) string {
				if err, ok := got.(error); ok {
					return fmt.Sprintf("queryByID error: %s", err)
				}
				gotPerm := got.(actionpermissionsbus.ActionPermission)
				expPerm := exp.(actionpermissionsbus.ActionPermission)
				return cmp.Diff(expPerm, gotPerm)
			},
		},
		{
			Name:    "queryByRoleAndAction",
			ExpResp: sd.Permissions[0],
			ExcFunc: func(ctx context.Context) any {
				perm, err := busDomain.ActionPermissions.QueryByRoleAndAction(ctx, sd.Permissions[0].RoleID, sd.Permissions[0].ActionType)
				if err != nil {
					return err
				}
				return perm
			},
			CmpFunc: func(got any, exp any) string {
				if err, ok := got.(error); ok {
					return fmt.Sprintf("queryByRoleAndAction error: %s", err)
				}
				gotPerm := got.(actionpermissionsbus.ActionPermission)
				expPerm := exp.(actionpermissionsbus.ActionPermission)
				return cmp.Diff(expPerm, gotPerm)
			},
		},
		{
			Name:    "create-update-delete",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				// Create
				newPerm, err := busDomain.ActionPermissions.Create(ctx, actionpermissionsbus.NewActionPermission{
					RoleID:      sd.RoleIDs[0],
					ActionType:  "test_action_webhook",
					IsAllowed:   true,
					Constraints: json.RawMessage("{}"),
				})
				if err != nil {
					return fmt.Errorf("create error: %w", err)
				}
				if newPerm.ActionType != "test_action_webhook" {
					return fmt.Errorf("expected action_type call_webhook, got %s", newPerm.ActionType)
				}

				// Update
				isAllowed := false
				updated, err := busDomain.ActionPermissions.Update(ctx, newPerm, actionpermissionsbus.UpdateActionPermission{
					IsAllowed: &isAllowed,
				})
				if err != nil {
					return fmt.Errorf("update error: %w", err)
				}
				if updated.IsAllowed != false {
					return fmt.Errorf("expected IsAllowed=false after update, got true")
				}

				// Delete
				if err := busDomain.ActionPermissions.Delete(ctx, updated); err != nil {
					return fmt.Errorf("delete error: %w", err)
				}

				// Verify deleted
				_, err = busDomain.ActionPermissions.QueryByID(ctx, newPerm.ID)
				if err == nil {
					return fmt.Errorf("expected error after delete, got nil")
				}
				return nil
			},
			CmpFunc: func(got any, exp any) string {
				if err, ok := got.(error); ok {
					return err.Error()
				}
				if got != nil {
					return fmt.Sprintf("unexpected result: %v", got)
				}
				return ""
			},
		},
		{
			Name:    "count",
			ExpResp: len(sd.Permissions),
			ExcFunc: func(ctx context.Context) any {
				count, err := busDomain.ActionPermissions.Count(ctx, actionpermissionsbus.QueryFilter{})
				if err != nil {
					return err
				}
				return count
			},
			CmpFunc: func(got any, exp any) string {
				if err, ok := got.(error); ok {
					return fmt.Sprintf("count error: %s", err)
				}
				gotCount := got.(int)
				expCount := exp.(int)
				if gotCount < expCount {
					return fmt.Sprintf("expected at least %d, got %d", expCount, gotCount)
				}
				return ""
			},
		},
	}
}

// =========================================================================

func canUserExecuteTests(busDomain dbtest.BusDomain, sd seedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "role-with-permission",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				can, err := busDomain.ActionPermissions.CanUserExecuteAction(ctx, sd.Users[0].ID, "test_action_a", []uuid.UUID{sd.RoleIDs[0]})
				if err != nil {
					return err
				}
				return can
			},
			CmpFunc: func(got any, exp any) string {
				if err, ok := got.(error); ok {
					return fmt.Sprintf("error: %s", err)
				}
				if got.(bool) != exp.(bool) {
					return fmt.Sprintf("expected %v, got %v", exp, got)
				}
				return ""
			},
		},
		{
			Name:    "role-without-permission",
			ExpResp: false,
			ExcFunc: func(ctx context.Context) any {
				// role1 does not have update_field permission
				can, err := busDomain.ActionPermissions.CanUserExecuteAction(ctx, sd.Users[0].ID, "test_action_denied", []uuid.UUID{sd.RoleIDs[0]})
				if err != nil {
					return err
				}
				return can
			},
			CmpFunc: func(got any, exp any) string {
				if err, ok := got.(error); ok {
					return fmt.Sprintf("error: %s", err)
				}
				if got.(bool) != exp.(bool) {
					return fmt.Sprintf("expected %v, got %v", exp, got)
				}
				return ""
			},
		},
		{
			Name:    "denied-permission",
			ExpResp: false,
			ExcFunc: func(ctx context.Context) any {
				// role2 has update_field with is_allowed=false
				can, err := busDomain.ActionPermissions.CanUserExecuteAction(ctx, sd.Users[1].ID, "test_action_denied", []uuid.UUID{sd.RoleIDs[1]})
				if err != nil {
					return err
				}
				return can
			},
			CmpFunc: func(got any, exp any) string {
				if err, ok := got.(error); ok {
					return fmt.Sprintf("error: %s", err)
				}
				if got.(bool) != exp.(bool) {
					return fmt.Sprintf("expected %v (denied), got %v", exp, got)
				}
				return ""
			},
		},
		{
			Name:    "multiple-roles-one-allowed",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				can, err := busDomain.ActionPermissions.CanUserExecuteAction(ctx, sd.Users[0].ID, "test_action_a", sd.RoleIDs)
				if err != nil {
					return err
				}
				return can
			},
			CmpFunc: func(got any, exp any) string {
				if err, ok := got.(error); ok {
					return fmt.Sprintf("error: %s", err)
				}
				if got.(bool) != exp.(bool) {
					return fmt.Sprintf("expected %v with multiple roles, got %v", exp, got)
				}
				return ""
			},
		},
		{
			Name:    "no-roles",
			ExpResp: false,
			ExcFunc: func(ctx context.Context) any {
				can, err := busDomain.ActionPermissions.CanUserExecuteAction(ctx, sd.Users[0].ID, "test_action_a", []uuid.UUID{})
				if err != nil {
					return err
				}
				return can
			},
			CmpFunc: func(got any, exp any) string {
				if err, ok := got.(error); ok {
					return fmt.Sprintf("error: %s", err)
				}
				if got.(bool) != exp.(bool) {
					return fmt.Sprintf("expected %v with no roles, got %v", exp, got)
				}
				return ""
			},
		},
	}
}

// =========================================================================

func getAllowedActionsTests(busDomain dbtest.BusDomain, sd seedData) []unitest.Table {
	return []unitest.Table{
		{
			Name:    "single-role",
			ExpResp: []string{"test_action_a", "test_action_b"},
			ExcFunc: func(ctx context.Context) any {
				actions, err := busDomain.ActionPermissions.GetAllowedActionsForRoles(ctx, []uuid.UUID{sd.RoleIDs[0]})
				if err != nil {
					return err
				}
				sort.Strings(actions)
				return actions
			},
			CmpFunc: func(got any, exp any) string {
				if err, ok := got.(error); ok {
					return fmt.Sprintf("error: %s", err)
				}
				gotActions := got.([]string)
				expActions := exp.([]string)
				// The admin role may have pre-seeded permissions, so check
				// that our test actions are present rather than exact match.
				actionSet := make(map[string]bool)
				for _, a := range gotActions {
					actionSet[a] = true
				}
				for _, e := range expActions {
					if !actionSet[e] {
						return fmt.Sprintf("expected action %q not found in %v", e, gotActions)
					}
				}
				return ""
			},
		},
		{
			Name:    "empty-roles",
			ExpResp: []string{},
			ExcFunc: func(ctx context.Context) any {
				actions, err := busDomain.ActionPermissions.GetAllowedActionsForRoles(ctx, []uuid.UUID{})
				if err != nil {
					return err
				}
				return actions
			},
			CmpFunc: func(got any, exp any) string {
				if err, ok := got.(error); ok {
					return fmt.Sprintf("error: %s", err)
				}
				return cmp.Diff(exp, got)
			},
		},
		{
			Name:    "multiple-roles-aggregated",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				actions, err := busDomain.ActionPermissions.GetAllowedActionsForRoles(ctx, sd.RoleIDs)
				if err != nil {
					return err
				}
				sort.Strings(actions)
				return actions
			},
			CmpFunc: func(got any, exp any) string {
				if err, ok := got.(error); ok {
					return fmt.Sprintf("error: %s", err)
				}
				actions := got.([]string)
				// role1 has test_action_a + test_action_b (allowed)
				// role2 has test_action_a (allowed) + test_action_denied (denied)
				// Aggregated allowed: test_action_a, test_action_b
				if len(actions) < 2 {
					return fmt.Sprintf("expected at least 2 allowed actions, got %d: %v", len(actions), actions)
				}
				return ""
			},
		},
	}
}
