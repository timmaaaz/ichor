package contactinfosapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the  contact info domain.
type App struct {
	contactinfosbus *contactinfosbus.Business
	auth            *auth.Auth
}

// NewApp constructs a  contact info app API for use.
func NewApp(contactinfosbus *contactinfosbus.Business) *App {
	return &App{
		contactinfosbus: contactinfosbus,
	}
}

// NewAppWithAuth constructs a  contact info app API for use with auth support.
func NewAppWithAuth(contactinfosbus *contactinfosbus.Business, ath *auth.Auth) *App {
	return &App{
		auth:            ath,
		contactinfosbus: contactinfosbus,
	}
}

// Create adds a new  contact info to the system.
func (a *App) Create(ctx context.Context, app NewContactInfos) (ContactInfos, error) {
	na, err := toBusNewContactInfos(app)
	if err != nil {
		return ContactInfos{}, errs.New(errs.InvalidArgument, err)
	}

	ass, err := a.contactinfosbus.Create(ctx, na)
	if err != nil {
		if errors.Is(err, contactinfosbus.ErrUniqueEntry) {
			return ContactInfos{}, errs.New(errs.Aborted, contactinfosbus.ErrUniqueEntry)
		}
		return ContactInfos{}, errs.Newf(errs.Internal, "create:  contact info[%+v]: %s", ass, err)
	}

	return ToAppContactInfo(ass), err
}

// Update updates an existing  contact info.
func (a *App) Update(ctx context.Context, app UpdateContactInfos, id uuid.UUID) (ContactInfos, error) {
	us, err := toBusUpdateContactInfos(app)
	if err != nil {
		return ContactInfos{}, errs.New(errs.InvalidArgument, err)
	}

	st, err := a.contactinfosbus.QueryByID(ctx, id)
	if err != nil {
		return ContactInfos{}, errs.New(errs.NotFound, contactinfosbus.ErrNotFound)
	}

	contactInfos, err := a.contactinfosbus.Update(ctx, st, us)
	if err != nil {
		if errors.Is(err, contactinfosbus.ErrNotFound) {
			return ContactInfos{}, errs.New(errs.NotFound, err)
		}
		return ContactInfos{}, errs.Newf(errs.Internal, "update:  contact info[%+v]: %s", contactInfos, err)
	}

	return ToAppContactInfo(contactInfos), nil
}

// Delete removes an existing  contact info.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	st, err := a.contactinfosbus.QueryByID(ctx, id)
	if err != nil {
		return errs.New(errs.NotFound, contactinfosbus.ErrNotFound)
	}

	err = a.contactinfosbus.Delete(ctx, st)
	if err != nil {
		return errs.Newf(errs.Internal, "delete:  contact info[%+v]: %s", st, err)
	}

	return nil
}

// Query returns a list of  contact infos based on the filter, order and page.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[ContactInfos], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[ContactInfos]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[ContactInfos]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[ContactInfos]{}, errs.NewFieldsError("orderby", err)
	}

	contactInfoss, err := a.contactinfosbus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[ContactInfos]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.contactinfosbus.Count(ctx, filter)
	if err != nil {
		return query.Result[ContactInfos]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppContactInfos(contactInfoss), total, page), nil
}

// QueryByID retrieves a single contact info by its id.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (ContactInfos, error) {
	contactInfos, err := a.contactinfosbus.QueryByID(ctx, id)
	if err != nil {
		return ContactInfos{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppContactInfo(contactInfos), nil
}
