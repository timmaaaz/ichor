// Package actionpermissionsapp provides the application layer for managing
// action permissions that control which roles can execute workflow actions manually.
package actionpermissionsapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/workflow/actionpermissionsbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer API functions for action permissions.
type App struct {
	actionPermBus *actionpermissionsbus.Business
}

// NewApp constructs an action permissions app API for use.
func NewApp(actionPermBus *actionpermissionsbus.Business) *App {
	return &App{
		actionPermBus: actionPermBus,
	}
}

// Create adds a new action permission to the system.
func (a *App) Create(ctx context.Context, app NewActionPermission) (ActionPermission, error) {
	roleID, err := uuid.Parse(app.RoleID)
	if err != nil {
		return ActionPermission{}, errs.NewFieldsError("roleId", err)
	}

	nap := actionpermissionsbus.NewActionPermission{
		RoleID:      roleID,
		ActionType:  app.ActionType,
		IsAllowed:   app.IsAllowed,
		Constraints: app.Constraints,
	}

	ap, err := a.actionPermBus.Create(ctx, nap)
	if err != nil {
		if errors.Is(err, actionpermissionsbus.ErrUnique) {
			return ActionPermission{}, errs.New(errs.Aborted, actionpermissionsbus.ErrUnique)
		}
		return ActionPermission{}, errs.Newf(errs.Internal, "create: action permission[%+v]: %s", nap, err)
	}

	return ToAppActionPermission(ap), nil
}

// Update modifies an existing action permission.
func (a *App) Update(ctx context.Context, app UpdateActionPermission, id uuid.UUID) (ActionPermission, error) {
	ap, err := a.actionPermBus.QueryByID(ctx, id)
	if err != nil {
		return ActionPermission{}, errs.New(errs.NotFound, actionpermissionsbus.ErrNotFound)
	}

	uap := actionpermissionsbus.UpdateActionPermission{
		IsAllowed:   app.IsAllowed,
		Constraints: app.Constraints,
	}

	updated, err := a.actionPermBus.Update(ctx, ap, uap)
	if err != nil {
		if errors.Is(err, actionpermissionsbus.ErrNotFound) {
			return ActionPermission{}, errs.New(errs.NotFound, actionpermissionsbus.ErrNotFound)
		}
		return ActionPermission{}, errs.Newf(errs.Internal, "update: action permission[%+v]: %s", updated, err)
	}

	return ToAppActionPermission(updated), nil
}

// Delete removes an action permission from the system.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	ap, err := a.actionPermBus.QueryByID(ctx, id)
	if err != nil {
		return errs.New(errs.NotFound, actionpermissionsbus.ErrNotFound)
	}

	if err := a.actionPermBus.Delete(ctx, ap); err != nil {
		return errs.Newf(errs.Internal, "delete: action permission[%+v]: %s", ap, err)
	}

	return nil
}

// Query retrieves a list of action permissions based on filter criteria.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[ActionPermission], error) {
	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[ActionPermission]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[ActionPermission]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[ActionPermission]{}, errs.NewFieldsError("orderby", err)
	}

	perms, err := a.actionPermBus.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return query.Result[ActionPermission]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.actionPermBus.Count(ctx, filter)
	if err != nil {
		return query.Result[ActionPermission]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppActionPermissions(perms), total, pg), nil
}

// QueryByID finds an action permission by its ID.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (ActionPermission, error) {
	ap, err := a.actionPermBus.QueryByID(ctx, id)
	if err != nil {
		return ActionPermission{}, errs.Newf(errs.NotFound, "querybyid: %s", err)
	}

	return ToAppActionPermission(ap), nil
}
