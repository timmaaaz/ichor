package mid

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/homebus"
	"github.com/timmaaaz/ichor/business/domain/permissions/permissionsbus"
	"github.com/timmaaaz/ichor/business/domain/productbus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
)

// ErrInvalidID represents a condition where the id is not a uuid.
var ErrInvalidID = errors.New("ID is not in its proper form")

// Authorize validates authorization via the auth service.
func Authorize(ctx context.Context, client *authclient.Client, rule string, next HandlerFunc) Encoder {
	userID, err := GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	auth := authclient.Authorize{
		Claims: GetClaims(ctx),
		UserID: userID,
		Rule:   rule,
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Authorize opa roles
	if err := client.Authorize(ctx, auth); err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	return next(ctx)
}

// AuthorizeTable validates authorization via the auth service with table information.
func AuthorizeTable(ctx context.Context, client *authclient.Client, permissionsBus *permissionsbus.Business, tableInfo *TableInfo, rule string, next HandlerFunc) Encoder {
	userID, err := GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	auth := authclient.Authorize{
		Claims: GetClaims(ctx),
		UserID: userID,
		Rule:   rule,
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Authorize opa roles
	if err := client.Authorize(ctx, auth); err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	zeroValue := TableInfo{}
	if tableInfo != nil && *tableInfo != zeroValue {
		// Authorize on our permissions
		perms, err := permissionsBus.QueryUserPermissions(ctx, userID)
		if err != nil {
			return errs.New(errs.Unauthenticated, fmt.Errorf("query user permissions: %w", err))
		}

		// If we have table information in the context, check table permissions
		if *tableInfo != zeroValue {
			if !hasTablePermission(perms, *tableInfo) {
				return errs.New(errs.Unauthenticated, fmt.Errorf("user %s lacks permission for %s on table %s", userID, tableInfo.Action, tableInfo.Name))
			}
		}
	}

	// Add table info to context
	ctx = setTableInfo(ctx, tableInfo)

	// Add restricted columns to context
	rcs, err := permissionsBus.RestrictedColumns.QueryAll(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, fmt.Errorf("query restricted columns: %w", err))
	}
	ctx = setRestrictedColumns(ctx, rcs)

	// Call the standard Authorize middleware with the enhanced context
	return Authorize(ctx, client, rule, next)
}

// hasTablePermission checks if the user has the required permission for the specified table
func hasTablePermission(userPerms permissionsbus.UserPermissions, tableInfo TableInfo) bool {
	// Search through all roles assigned to the user
	for _, role := range userPerms.Roles {
		// Check each table access in this role
		for _, tableAccess := range role.Tables {
			if strings.EqualFold(tableAccess.TableName, tableInfo.Name) {
				// Check specific permission based on the action
				switch tableInfo.Action {
				case permissionsbus.Actions.Create:
					if tableAccess.CanCreate {
						return true
					}
				case permissionsbus.Actions.Read:
					if tableAccess.CanRead {
						return true
					}
				case permissionsbus.Actions.Update:
					if tableAccess.CanUpdate {
						return true
					}
				case permissionsbus.Actions.Delete:
					if tableAccess.CanDelete {
						return true
					}
				}
			}
		}
	}
	return false
}

// AuthorizeUser executes the specified role and extracts the specified
// user from the DB if a user id is specified in the call. Depending on the rule
// specified, the userid from the claims may be compared with the specified
// user id.
func AuthorizeUser(ctx context.Context, client *authclient.Client, userBus *userbus.Business, rule string, id string, next HandlerFunc) Encoder {
	var userID uuid.UUID

	if id != "" {
		var err error
		userID, err = uuid.Parse(id)
		if err != nil {
			return errs.New(errs.Unauthenticated, ErrInvalidID)
		}

		usr, err := userBus.QueryByID(ctx, userID)
		if err != nil {
			switch {
			case errors.Is(err, userbus.ErrNotFound):
				return errs.New(errs.Unauthenticated, err)
			default:
				return errs.Newf(errs.Unauthenticated, "querybyid: userID[%s]: %s", userID, err)
			}
		}

		ctx = setUser(ctx, usr)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	auth := authclient.Authorize{
		Claims: GetClaims(ctx),
		UserID: userID,
		Rule:   rule,
	}

	if err := client.Authorize(ctx, auth); err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	return next(ctx)
}

// AuthorizeProduct executes the specified role and extracts the specified
// product from the DB if a product id is specified in the call. Depending on
// the rule specified, the userid from the claims may be compared with the
// specified user id from the product.
func AuthorizeProduct(ctx context.Context, client *authclient.Client, productBus *productbus.Business, id string, next HandlerFunc) Encoder {
	var userID uuid.UUID

	if id != "" {
		var err error
		productID, err := uuid.Parse(id)
		if err != nil {
			return errs.New(errs.Unauthenticated, ErrInvalidID)
		}

		prd, err := productBus.QueryByID(ctx, productID)
		if err != nil {
			switch {
			case errors.Is(err, productbus.ErrNotFound):
				return errs.New(errs.Unauthenticated, err)
			default:
				return errs.Newf(errs.Internal, "querybyid: productID[%s]: %s", productID, err)
			}
		}

		userID = prd.UserID
		ctx = setProduct(ctx, prd)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	auth := authclient.Authorize{
		UserID: userID,
		Claims: GetClaims(ctx),
		Rule:   auth.RuleAdminOrSubject,
	}

	if err := client.Authorize(ctx, auth); err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	return next(ctx)
}

// AuthorizeHome executes the specified role and extracts the specified
// home from the DB if a home id is specified in the call. Depending on
// the rule specified, the userid from the claims may be compared with the
// specified user id from the home.
func AuthorizeHome(ctx context.Context, client *authclient.Client, homeBus *homebus.Business, id string, next HandlerFunc) Encoder {
	var userID uuid.UUID

	if id != "" {
		var err error
		homeID, err := uuid.Parse(id)
		if err != nil {
			return errs.New(errs.Unauthenticated, ErrInvalidID)
		}

		hme, err := homeBus.QueryByID(ctx, homeID)
		if err != nil {
			switch {
			case errors.Is(err, homebus.ErrNotFound):
				return errs.New(errs.Unauthenticated, err)
			default:
				return errs.Newf(errs.Unauthenticated, "querybyid: homeID[%s]: %s", homeID, err)
			}
		}

		userID = hme.UserID
		ctx = setHome(ctx, hme)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	auth := authclient.Authorize{
		Claims: GetClaims(ctx),
		UserID: userID,
		Rule:   auth.RuleAdminOrSubject,
	}

	if err := client.Authorize(ctx, auth); err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	return next(ctx)
}
