package permissionsbus

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/permissions/rolebus"
	"github.com/timmaaaz/ichor/business/domain/permissions/tableaccessbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/userrolebus"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound              = errors.New("role not found")
	ErrUnique                = errors.New("organizational unit is not unique")
	ErrAuthenticationFailure = errors.New("authentication failed")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	QueryUserPermissions(ctx context.Context, userID uuid.UUID) (UserPermissions, error)
	ClearCache()
}

// Business manages the set of APIs for user access.
type Business struct {
	log            *logger.Logger
	del            *delegate.Delegate
	storer         Storer
	RolesBus       *rolebus.Business
	UserRolesBus   *userrolebus.Business
	TableAccessBus *tableaccessbus.Business
}

// NewBusiness constructs a user business API for use.
func NewBusiness(log *logger.Logger, del *delegate.Delegate, storer Storer, urb *userrolebus.Business, tab *tableaccessbus.Business, rb *rolebus.Business) *Business {
	b := &Business{
		log:            log,
		del:            del,
		storer:         storer,
		RolesBus:       rb,
		UserRolesBus:   urb,
		TableAccessBus: tab,
	}

	// Register the handler as a closure that calls the ClearCache method on this business instance
	del.Register(rolebus.DomainName, rolebus.ActionUpdated, func(ctx context.Context, data delegate.Data) error {
		return b.ClearCache(ctx, data)
	})
	del.Register(rolebus.DomainName, rolebus.ActionDeleted, func(ctx context.Context, data delegate.Data) error {
		return b.ClearCache(ctx, data)
	})

	del.Register(tableaccessbus.DomainName, tableaccessbus.ActionUpdated, func(ctx context.Context, data delegate.Data) error {
		return b.ClearCache(ctx, data)
	})
	del.Register(tableaccessbus.DomainName, tableaccessbus.ActionDeleted, func(ctx context.Context, data delegate.Data) error {
		return b.ClearCache(ctx, data)
	})

	del.Register(userrolebus.DomainName, userrolebus.ActionUpdated, func(ctx context.Context, data delegate.Data) error {
		return b.ClearCache(ctx, data)
	})
	del.Register(userrolebus.DomainName, userrolebus.ActionDeleted, func(ctx context.Context, data delegate.Data) error {
		return b.ClearCache(ctx, data)
	})

	return b
}

// NewWithTx constructs a new business value that will use the
// specified transaction in any store related calls.
func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (*Business, error) {
	storer, err := b.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	bus := Business{
		log:    b.log,
		storer: storer,
	}

	return &bus, nil
}

// ClearCache clears the cache of the business.
func (b *Business) ClearCache(ctx context.Context, data delegate.Data) error {
	b.storer.ClearCache()
	return nil
}

// QueryUserPermissions retrieves the permissions for the specified user.
func (b *Business) QueryUserPermissions(ctx context.Context, userID uuid.UUID) (UserPermissions, error) {
	// QueryUserRoles
	var userRole *userrolebus.UserRole
	tmp, err := b.UserRolesBus.QueryByUserID(ctx, userID)
	if err != nil {
		return UserPermissions{}, err
	}
	if tmp != (userrolebus.UserRole{}) {
		userRole = &tmp
	}

	var role rolebus.Role
	if userRole != nil {
		role, err = b.RolesBus.QueryByID(ctx, userRole.RoleID)
		if err != nil {
			return UserPermissions{}, err
		}
	}

	var tables []tableaccessbus.TableAccess
	if userRole != nil {
		// TableAccesses
		tables, err = b.TableAccessBus.Query(
			ctx,
			tableaccessbus.QueryFilter{RoleID: &userRole.RoleID},
			tableaccessbus.DefaultOrderBy,
			page.MustParse("1", "100"),
		)
		if err != nil {
			return UserPermissions{}, err
		}
	}

	tableAccesses := make(map[string]tableaccessbus.TableAccess, len(tables))
	for _, table := range tables {
		tableAccesses[table.TableName] = table
	}

	userPerms := UserPermissions{
		RoleName:    role.Name,
		UserID:      userID,
		Role:        userRole,
		TableAccess: tableAccesses,
	}

	userPermJson, err := json.MarshalIndent(userPerms, "", "  ")
	if err != nil {
		return UserPermissions{}, err
	}
	atmp := string(userPermJson)
	_ = atmp

	return userPerms, nil
}
