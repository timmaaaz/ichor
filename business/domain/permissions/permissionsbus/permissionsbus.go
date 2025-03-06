package permissionsbus

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/permissions/crossunitpermissionsbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/organizationalunitbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/orgunitcolumnaccessbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/restrictedcolumnbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/rolebus"
	"github.com/timmaaaz/ichor/business/domain/permissions/tableaccessbus"
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
}

// Business manages the set of APIs for user access.
type Business struct {
	log                    *logger.Logger
	storer                 Storer
	RestrictedColumnsBus   *restrictedcolumnbus.Business
	OrgUntitsBus           *organizationalunitbus.Business
	TableAccessBus         *tableaccessbus.Business
	CrossUnitPermissionBus *crossunitpermissionsbus.Business
	RolesBus               *rolebus.Business
	OrgUnitColumnAccessBus *orgunitcolumnaccessbus.Business
}

// NewBusiness constructs a user business API for use.
func NewBusiness(log *logger.Logger, storer Storer, rcb *restrictedcolumnbus.Business, oub *organizationalunitbus.Business, tab *tableaccessbus.Business, cupb *crossunitpermissionsbus.Business, rb *rolebus.Business, oucb *orgunitcolumnaccessbus.Business) *Business {
	return &Business{
		log:                    log,
		storer:                 storer,
		RestrictedColumnsBus:   rcb,
		OrgUntitsBus:           oub,
		TableAccessBus:         tab,
		CrossUnitPermissionBus: cupb,
		RolesBus:               rb,
		OrgUnitColumnAccessBus: oucb,
	}
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

// QueryUserPermissions retrieves the permissions for the specified user.
func (b *Business) QueryUserPermissions(ctx context.Context, userID uuid.UUID) (UserPermissions, error) {
	// perms, err := b.storer.QueryUserPermissions(ctx, userID)

	// if err != nil {
	// 	return UserPermissions{}, fmt.Errorf("query user permissions: %w", err)
	// }

	// QueryRoles
	_, err := b.RolesBus.QueryAll(ctx)
	if err != nil {
		return UserPermissions{}, err
	}

	return UserPermissions{}, nil
}
