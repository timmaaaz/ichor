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
	ErrNoPermissions         = errors.New("user has no permissions set")
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
	userRoles, err := b.UserRolesBus.Query(ctx, userrolebus.QueryFilter{UserID: &userID}, userrolebus.DefaultOrderBy, page.MustParse("1", "100"))
	if err != nil {
		return UserPermissions{}, err
	}
	if len(userRoles) == 0 {
		return UserPermissions{}, ErrNoPermissions
	}

	roleIDs := make(uuid.UUIDs, len(userRoles))
	for i, r := range userRoles {
		roleIDs[i] = r.RoleID
	}

	roles, err := b.RolesBus.QueryByIDs(ctx, roleIDs)
	if err != nil {
		return UserPermissions{}, err
	}

	if len(roles) == 0 {
		return UserPermissions{}, ErrNoPermissions
	}

	var tables []tableaccessbus.TableAccess

	tables, err = b.TableAccessBus.QueryByRoleIDs(ctx, roleIDs)
	if err != nil {
		return UserPermissions{}, err
	}

	tableAccesses := make(map[string]tableaccessbus.TableAccess, len(tables))
	for _, table := range tables {
		tableAccesses[table.TableName] = table
	}

	// Construct the UserPermissions object
	roleNames := make([]string, len(roles))
	for i, r := range roles {
		roleNames[i] = r.Name
	}

	// Combine permissions following the principle of least privilege
	combinedTableAccesses := make(map[string]tableaccessbus.TableAccess, len(tables))
	for _, table := range tables {
		// Make sure table exists in map
		if _, ok := combinedTableAccesses[table.TableName]; !ok {
			combinedTableAccesses[table.TableName] = table
		} else {
			t := combinedTableAccesses[table.TableName]
			t.CanCreate = t.CanCreate || table.CanCreate
			t.CanRead = t.CanRead || table.CanRead
			t.CanUpdate = t.CanUpdate || table.CanUpdate
			t.CanDelete = t.CanDelete || table.CanDelete
			combinedTableAccesses[table.TableName] = t
		}
	}

	userPerms := UserPermissions{
		RoleNames:   roleNames,
		UserID:      userID,
		Roles:       userRoles,
		TableAccess: combinedTableAccesses,
	}

	userPermJson, err := json.MarshalIndent(userPerms, "", "  ")
	if err != nil {
		return UserPermissions{}, err
	}
	atmp := string(userPermJson)
	_ = atmp

	return userPerms, nil
}
