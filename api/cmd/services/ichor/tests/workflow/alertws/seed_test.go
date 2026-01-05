package alertws_test

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/business/domain/core/rolebus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/core/userrolebus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

// AlertWSSeedData contains test users, roles, and their relationships
// for WebSocket integration testing.
type AlertWSSeedData struct {
	// Users with tokens for WebSocket authentication
	Users []apitest.User

	// Admin users with elevated permissions
	Admins []apitest.User

	// Roles for testing role-based broadcasts
	Roles []rolebus.Role

	// UserRoleAssignments maps user index to their role indices
	// User 0: role 0
	// User 1: role 0, role 1 (multiple roles)
	// User 2: no roles
}

// insertSeedData creates test users with roles for WebSocket testing.
// The role assignments are designed to test various broadcast scenarios:
// - User-targeted broadcasts
// - Role-targeted broadcasts (single and multiple users per role)
// - Users without roles
func insertSeedData(db *dbtest.Database, ath *auth.Auth) (AlertWSSeedData, error) {
	ctx := context.Background()
	busDomain := db.BusDomain

	// =========================================================================
	// Create Users
	// =========================================================================

	// Create 3 regular users
	users, err := userbus.TestSeedUsersWithNoFKs(ctx, 3, userbus.Roles.User, busDomain.User)
	if err != nil {
		return AlertWSSeedData{}, fmt.Errorf("seeding users: %w", err)
	}

	// Create 1 admin user
	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, busDomain.User)
	if err != nil {
		return AlertWSSeedData{}, fmt.Errorf("seeding admins: %w", err)
	}

	// Generate JWT tokens for all users
	testUsers := make([]apitest.User, len(users))
	for i, u := range users {
		testUsers[i] = apitest.User{
			User:  u,
			Token: apitest.Token(busDomain.User, ath, u.Email.Address),
		}
	}

	testAdmins := make([]apitest.User, len(admins))
	for i, a := range admins {
		testAdmins[i] = apitest.User{
			User:  a,
			Token: apitest.Token(busDomain.User, ath, a.Email.Address),
		}
	}

	// =========================================================================
	// Create Roles
	// =========================================================================

	roles, err := rolebus.TestSeedRoles(ctx, 2, busDomain.Role)
	if err != nil {
		return AlertWSSeedData{}, fmt.Errorf("seeding roles: %w", err)
	}

	// =========================================================================
	// Assign Roles to Users
	// =========================================================================

	// User 0: role 0 only
	_, err = busDomain.UserRole.Create(ctx, userrolebus.NewUserRole{
		UserID: users[0].ID,
		RoleID: roles[0].ID,
	})
	if err != nil {
		return AlertWSSeedData{}, fmt.Errorf("assigning role 0 to user 0: %w", err)
	}

	// User 1: role 0 AND role 1 (multiple roles)
	_, err = busDomain.UserRole.Create(ctx, userrolebus.NewUserRole{
		UserID: users[1].ID,
		RoleID: roles[0].ID,
	})
	if err != nil {
		return AlertWSSeedData{}, fmt.Errorf("assigning role 0 to user 1: %w", err)
	}

	_, err = busDomain.UserRole.Create(ctx, userrolebus.NewUserRole{
		UserID: users[1].ID,
		RoleID: roles[1].ID,
	})
	if err != nil {
		return AlertWSSeedData{}, fmt.Errorf("assigning role 1 to user 1: %w", err)
	}

	// User 2: no roles (tests users without role-based access)

	return AlertWSSeedData{
		Users:  testUsers,
		Admins: testAdmins,
		Roles:  roles,
	}, nil
}

// UserID returns the UUID of a user at the given index.
func (sd AlertWSSeedData) UserID(index int) uuid.UUID {
	if index < 0 || index >= len(sd.Users) {
		return uuid.Nil
	}
	return sd.Users[index].ID
}

// UserToken returns the JWT token for a user at the given index.
func (sd AlertWSSeedData) UserToken(index int) string {
	if index < 0 || index >= len(sd.Users) {
		return ""
	}
	return sd.Users[index].Token
}

// RoleID returns the UUID of a role at the given index.
func (sd AlertWSSeedData) RoleID(index int) uuid.UUID {
	if index < 0 || index >= len(sd.Roles) {
		return uuid.Nil
	}
	return sd.Roles[index].ID
}

// AdminToken returns the JWT token for the first admin user.
func (sd AlertWSSeedData) AdminToken() string {
	if len(sd.Admins) == 0 {
		return ""
	}
	return sd.Admins[0].Token
}
