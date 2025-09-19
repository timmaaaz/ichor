package commentapp

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/hr/commentbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the approval status domain.
type App struct {
	userapprovalcommentbus *commentbus.Business
	auth                   *auth.Auth
}

// NewApp constructs an approval status app API for use.
func NewApp(userapprovalcommentbus *commentbus.Business) *App {
	return &App{
		userapprovalcommentbus: userapprovalcommentbus,
	}
}

// NewAppWithAuth constructs an approval status app API for use with auth support.
func NewAppWithAuth(userapprovalcommentbus *commentbus.Business, ath *auth.Auth) *App {
	return &App{
		auth:                   ath,
		userapprovalcommentbus: userapprovalcommentbus,
	}
}

// Create adds a new approval status to the system
func (a *App) Create(ctx context.Context, app NewUserApprovalComment) (UserApprovalComment, error) {
	nas, err := toBusNewUserApprovalComment(app)
	if err != nil {
		return UserApprovalComment{}, errs.New(errs.InvalidArgument, err)
	}

	as, err := a.userapprovalcommentbus.Create(ctx, nas)
	if err != nil {
		return UserApprovalComment{}, err
	}

	return ToAppUserApprovalComment(as), nil
}

// Update updates an existing approval status
func (a *App) Update(ctx context.Context, app UpdateUserApprovalComment, id uuid.UUID) (UserApprovalComment, error) {
	uas := toBusUpdateUserApprovalComment(app)

	as, err := a.userapprovalcommentbus.QueryByID(ctx, id)
	if err != nil {
		return UserApprovalComment{}, errs.New(errs.NotFound, commentbus.ErrNotFound)
	}

	updated, err := a.userapprovalcommentbus.Update(ctx, as, uas)
	if err != nil {
		if errors.Is(err, commentbus.ErrNotFound) {
			return UserApprovalComment{}, errs.New(errs.NotFound, err)
		}
		return UserApprovalComment{}, errs.Newf(errs.Internal, "update: user approval comment[%+v]: %s", updated, err)
	}

	return ToAppUserApprovalComment(updated), nil
}

// Delete removes an existing approval status
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	as, err := a.userapprovalcommentbus.QueryByID(ctx, id)
	if err != nil {
		return errs.New(errs.NotFound, commentbus.ErrNotFound)
	}

	err = a.userapprovalcommentbus.Delete(ctx, as)
	if err != nil {
		return errs.Newf(errs.Internal, "delete user approval status[%+v]: %s", as, err)
	}

	return nil
}

// Query returns a list of approval statuses based on the filter, order and page
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[UserApprovalComment], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[UserApprovalComment]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[UserApprovalComment]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[UserApprovalComment]{}, errs.NewFieldsError("orderby", err)
	}

	as, err := a.userapprovalcommentbus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[UserApprovalComment]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.userapprovalcommentbus.Count(ctx, filter)
	if err != nil {
		return query.Result[UserApprovalComment]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppUserApprovalComments(as), total, page), nil
}

// QueryByID retrieves the approval status by ID
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (UserApprovalComment, error) {
	as, err := a.userapprovalcommentbus.QueryByID(ctx, id)
	if err != nil {
		return UserApprovalComment{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppUserApprovalComment(as), nil
}
