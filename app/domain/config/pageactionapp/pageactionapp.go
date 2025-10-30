package pageactionapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/config/pageactionbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the page action domain.
type App struct {
	pageactionbus *pageactionbus.Business
	db            *sqlx.DB
}

// NewApp constructs a page action app API for use.
func NewApp(pageactionbus *pageactionbus.Business) *App {
	return &App{
		pageactionbus: pageactionbus,
	}
}

// NewAppWithDB constructs a page action app API with database support for transactions.
func NewAppWithDB(pageactionbus *pageactionbus.Business, db *sqlx.DB) *App {
	return &App{
		pageactionbus: pageactionbus,
		db:            db,
	}
}

// =============================================================================
// Create Methods
// =============================================================================

// CreateButton adds a new button action to the system.
func (a *App) CreateButton(ctx context.Context, app NewButtonAction) (PageAction, error) {
	nba, err := toBusNewButtonAction(app)
	if err != nil {
		return PageAction{}, err
	}

	action, err := a.pageactionbus.CreateButton(ctx, nba)
	if err != nil {
		if errors.Is(err, pageactionbus.ErrUniqueEntry) {
			return PageAction{}, errs.New(errs.Aborted, pageactionbus.ErrUniqueEntry)
		}
		if errors.Is(err, pageactionbus.ErrForeignKeyViolation) {
			return PageAction{}, errs.New(errs.FailedPrecondition, pageactionbus.ErrForeignKeyViolation)
		}
		return PageAction{}, errs.Newf(errs.Internal, "create button: %s", err)
	}

	return ToAppPageAction(action), nil
}

// CreateDropdown adds a new dropdown action with items to the system.
func (a *App) CreateDropdown(ctx context.Context, app NewDropdownAction) (PageAction, error) {
	nda, err := toBusNewDropdownAction(app)
	if err != nil {
		return PageAction{}, err
	}

	action, err := a.pageactionbus.CreateDropdown(ctx, nda)
	if err != nil {
		if errors.Is(err, pageactionbus.ErrUniqueEntry) {
			return PageAction{}, errs.New(errs.Aborted, pageactionbus.ErrUniqueEntry)
		}
		if errors.Is(err, pageactionbus.ErrForeignKeyViolation) {
			return PageAction{}, errs.New(errs.FailedPrecondition, pageactionbus.ErrForeignKeyViolation)
		}
		return PageAction{}, errs.Newf(errs.Internal, "create dropdown: %s", err)
	}

	return ToAppPageAction(action), nil
}

// CreateSeparator adds a new separator action to the system.
func (a *App) CreateSeparator(ctx context.Context, app NewSeparatorAction) (PageAction, error) {
	nsa, err := toBusNewSeparatorAction(app)
	if err != nil {
		return PageAction{}, err
	}

	action, err := a.pageactionbus.CreateSeparator(ctx, nsa)
	if err != nil {
		if errors.Is(err, pageactionbus.ErrUniqueEntry) {
			return PageAction{}, errs.New(errs.Aborted, pageactionbus.ErrUniqueEntry)
		}
		if errors.Is(err, pageactionbus.ErrForeignKeyViolation) {
			return PageAction{}, errs.New(errs.FailedPrecondition, pageactionbus.ErrForeignKeyViolation)
		}
		return PageAction{}, errs.Newf(errs.Internal, "create separator: %s", err)
	}

	return ToAppPageAction(action), nil
}

// =============================================================================
// Update Methods
// =============================================================================

// UpdateButton updates an existing button action.
func (a *App) UpdateButton(ctx context.Context, app UpdateButtonAction, id uuid.UUID) (PageAction, error) {
	uba, err := toBusUpdateButtonAction(app)
	if err != nil {
		return PageAction{}, err
	}

	action, err := a.pageactionbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, pageactionbus.ErrNotFound) {
			return PageAction{}, errs.New(errs.NotFound, pageactionbus.ErrNotFound)
		}
		return PageAction{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	if action.ActionType != pageactionbus.ActionTypeButton {
		return PageAction{}, errs.Newf(errs.InvalidArgument, "action %s is not a button", id)
	}

	updated, err := a.pageactionbus.UpdateButton(ctx, action, uba)
	if err != nil {
		if errors.Is(err, pageactionbus.ErrNotFound) {
			return PageAction{}, errs.New(errs.NotFound, pageactionbus.ErrNotFound)
		}
		return PageAction{}, errs.Newf(errs.Internal, "update button: %s", err)
	}

	return ToAppPageAction(updated), nil
}

// UpdateDropdown updates an existing dropdown action.
func (a *App) UpdateDropdown(ctx context.Context, app UpdateDropdownAction, id uuid.UUID) (PageAction, error) {
	uda, err := toBusUpdateDropdownAction(app)
	if err != nil {
		return PageAction{}, err
	}

	action, err := a.pageactionbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, pageactionbus.ErrNotFound) {
			return PageAction{}, errs.New(errs.NotFound, pageactionbus.ErrNotFound)
		}
		return PageAction{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	if action.ActionType != pageactionbus.ActionTypeDropdown {
		return PageAction{}, errs.Newf(errs.InvalidArgument, "action %s is not a dropdown", id)
	}

	updated, err := a.pageactionbus.UpdateDropdown(ctx, action, uda)
	if err != nil {
		if errors.Is(err, pageactionbus.ErrNotFound) {
			return PageAction{}, errs.New(errs.NotFound, pageactionbus.ErrNotFound)
		}
		return PageAction{}, errs.Newf(errs.Internal, "update dropdown: %s", err)
	}

	return ToAppPageAction(updated), nil
}

// UpdateSeparator updates an existing separator action.
func (a *App) UpdateSeparator(ctx context.Context, app UpdateSeparatorAction, id uuid.UUID) (PageAction, error) {
	usa, err := toBusUpdateSeparatorAction(app)
	if err != nil {
		return PageAction{}, err
	}

	action, err := a.pageactionbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, pageactionbus.ErrNotFound) {
			return PageAction{}, errs.New(errs.NotFound, pageactionbus.ErrNotFound)
		}
		return PageAction{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	if action.ActionType != pageactionbus.ActionTypeSeparator {
		return PageAction{}, errs.Newf(errs.InvalidArgument, "action %s is not a separator", id)
	}

	updated, err := a.pageactionbus.UpdateSeparator(ctx, action, usa)
	if err != nil {
		if errors.Is(err, pageactionbus.ErrNotFound) {
			return PageAction{}, errs.New(errs.NotFound, pageactionbus.ErrNotFound)
		}
		return PageAction{}, errs.Newf(errs.Internal, "update separator: %s", err)
	}

	return ToAppPageAction(updated), nil
}

// =============================================================================
// Delete Methods
// =============================================================================

// Delete removes an existing page action.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	action, err := a.pageactionbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, pageactionbus.ErrNotFound) {
			return errs.New(errs.NotFound, pageactionbus.ErrNotFound)
		}
		return errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	if err := a.pageactionbus.Delete(ctx, action); err != nil {
		return errs.Newf(errs.Internal, "delete: %s", err)
	}

	return nil
}

// =============================================================================
// Query Methods
// =============================================================================

// Query retrieves a list of page actions from the system.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[PageAction], error) {
	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[PageAction]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[PageAction]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[PageAction]{}, errs.NewFieldsError("orderby", err)
	}

	actions, err := a.pageactionbus.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return query.Result[PageAction]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.pageactionbus.Count(ctx, filter)
	if err != nil {
		return query.Result[PageAction]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppPageActions(actions), total, pg), nil
}

// QueryByID finds the page action by the specified ID.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (PageAction, error) {
	action, err := a.pageactionbus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, pageactionbus.ErrNotFound) {
			return PageAction{}, errs.New(errs.NotFound, pageactionbus.ErrNotFound)
		}
		return PageAction{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppPageAction(action), nil
}

// QueryByPageConfigID retrieves all actions for a specific page config, grouped by type.
func (a *App) QueryByPageConfigID(ctx context.Context, pageConfigID uuid.UUID) (ActionsGroupedByType, error) {
	actions, err := a.pageactionbus.QueryByPageConfigID(ctx, pageConfigID)
	if err != nil {
		return ActionsGroupedByType{}, errs.Newf(errs.Internal, "querybypageconfigid: %s", err)
	}

	return ToAppActionsGroupedByType(actions), nil
}

// =============================================================================
// Batch Operations
// =============================================================================

// BatchCreate creates multiple page actions in a single transaction.
func (a *App) BatchCreate(ctx context.Context, app BatchCreateRequest) (PageActions, error) {
	if a.db == nil {
		return nil, errs.Newf(errs.Internal, "database not configured for batch operations")
	}

	// Begin transaction
	tx, err := a.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "begin transaction: %s", err)
	}
	defer tx.Rollback()

	// Get transactional business layer
	pageactionBusTx, err := a.pageactionbus.NewWithTx(tx)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "create tx business: %s", err)
	}

	var createdActions []pageactionbus.PageAction

	// Create each action based on type
	for i, action := range app.Actions {
		var created pageactionbus.PageAction
		var err error

		switch action.ActionType {
		case "button":
			if action.Button == nil {
				return nil, errs.Newf(errs.InvalidArgument, "action[%d]: button is nil", i)
			}
			busAction, convErr := toBusNewButtonAction(*action.Button)
			if convErr != nil {
				return nil, errs.Newf(errs.InvalidArgument, "action[%d]: %s", i, convErr)
			}
			created, err = pageactionBusTx.CreateButton(ctx, busAction)

		case "dropdown":
			if action.Dropdown == nil {
				return nil, errs.Newf(errs.InvalidArgument, "action[%d]: dropdown is nil", i)
			}
			busAction, convErr := toBusNewDropdownAction(*action.Dropdown)
			if convErr != nil {
				return nil, errs.Newf(errs.InvalidArgument, "action[%d]: %s", i, convErr)
			}
			created, err = pageactionBusTx.CreateDropdown(ctx, busAction)

		case "separator":
			if action.Separator == nil {
				return nil, errs.Newf(errs.InvalidArgument, "action[%d]: separator is nil", i)
			}
			busAction, convErr := toBusNewSeparatorAction(*action.Separator)
			if convErr != nil {
				return nil, errs.Newf(errs.InvalidArgument, "action[%d]: %s", i, convErr)
			}
			created, err = pageactionBusTx.CreateSeparator(ctx, busAction)

		default:
			return nil, errs.Newf(errs.InvalidArgument, "action[%d]: unknown action type: %s", i, action.ActionType)
		}

		if err != nil {
			if errors.Is(err, pageactionbus.ErrUniqueEntry) {
				return nil, errs.Newf(errs.Aborted, "action[%d]: %s", i, pageactionbus.ErrUniqueEntry)
			}
			if errors.Is(err, pageactionbus.ErrForeignKeyViolation) {
				return nil, errs.Newf(errs.FailedPrecondition, "action[%d]: %s", i, pageactionbus.ErrForeignKeyViolation)
			}
			return nil, errs.Newf(errs.Internal, "action[%d]: create: %s", i, err)
		}

		createdActions = append(createdActions, created)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, errs.Newf(errs.Internal, "commit: %s", err)
	}

	return PageActions(ToAppPageActions(createdActions)), nil
}
