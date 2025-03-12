package permissionsbus

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/permissions/crossunitpermissionsbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/organizationalunitbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/orgunitcolumnaccessbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/restrictedcolumnbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/rolebus"
	"github.com/timmaaaz/ichor/business/domain/permissions/tableaccessbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/userorganizationbus"
	"github.com/timmaaaz/ichor/business/domain/permissions/userrolebus"
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
}

// Business manages the set of APIs for user access.
type Business struct {
	log                    *logger.Logger
	storer                 Storer
	RolesBus               *rolebus.Business
	UserRolesBus           *userrolebus.Business
	UserOrganizationsBus   *userorganizationbus.Business
	RestrictedColumnsBus   *restrictedcolumnbus.Business
	OrgUntitsBus           *organizationalunitbus.Business
	TableAccessBus         *tableaccessbus.Business
	CrossUnitPermissionBus *crossunitpermissionsbus.Business
	OrgUnitColumnAccessBus *orgunitcolumnaccessbus.Business
}

// NewBusiness constructs a user business API for use.
func NewBusiness(log *logger.Logger, storer Storer, rcb *restrictedcolumnbus.Business, uob *userorganizationbus.Business, urb *userrolebus.Business, oub *organizationalunitbus.Business, tab *tableaccessbus.Business, cupb *crossunitpermissionsbus.Business, rb *rolebus.Business, oucb *orgunitcolumnaccessbus.Business) *Business {
	return &Business{
		log:                    log,
		storer:                 storer,
		RolesBus:               rb,
		UserRolesBus:           urb,
		RestrictedColumnsBus:   rcb,
		OrgUntitsBus:           oub,
		TableAccessBus:         tab,
		CrossUnitPermissionBus: cupb,
		OrgUnitColumnAccessBus: oucb,
		UserOrganizationsBus:   uob,
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

	// UserOrganizations
	var userOrgUnit *organizationalunitbus.OrganizationalUnit
	tmpUsrOrg, err := b.UserOrganizationsBus.QueryByUserID(ctx, userID)
	if err != nil {
		if !errors.Is(err, userorganizationbus.ErrNotFound) {
			return UserPermissions{}, err
		}
	}
	if tmpUsrOrg != (userorganizationbus.UserOrganization{}) {
		tmpOrgUnit, err := b.OrgUntitsBus.Query(
			ctx, organizationalunitbus.QueryFilter{ID: &tmpUsrOrg.OrganizationalUnitID},
			organizationalunitbus.DefaultOrderBy,
			page.MustParse("1", "1"),
		)
		if err != nil {
			// return UserPermissions{}, err

		}
		if len(tmpOrgUnit) > 0 {
			userOrgUnit = &tmpOrgUnit[0]
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
		OrgUnit:     userOrgUnit,
	}

	userPermJson, err := json.MarshalIndent(userPerms, "", "  ")
	if err != nil {
		return UserPermissions{}, err
	}
	atmp := string(userPermJson)
	_ = atmp

	return userPerms, nil
}
