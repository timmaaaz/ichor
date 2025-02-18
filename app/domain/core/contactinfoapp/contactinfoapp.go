package contactinfoapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfobus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the  contact info domain.
type App struct {
	contactinfobus *contactinfobus.Business
	auth           *auth.Auth
}

// NewApp constructs a  contact info app API for use.
func NewApp(contactinfobus *contactinfobus.Business) *App {
	return &App{
		contactinfobus: contactinfobus,
	}
}

// NewAppWithAuth constructs a  contact info app API for use with auth support.
func NewAppWithAuth(contactinfobus *contactinfobus.Business, ath *auth.Auth) *App {
	return &App{
		auth:           ath,
		contactinfobus: contactinfobus,
	}
}

// Create adds a new  contact info to the system.
func (a *App) Create(ctx context.Context, app NewContactInfo) (ContactInfo, error) {
	na, err := toBusNewContactInfo(app)
	if err != nil {
		return ContactInfo{}, errs.New(errs.InvalidArgument, err)
	}

	ass, err := a.contactinfobus.Create(ctx, na)
	if err != nil {
		if errors.Is(err, contactinfobus.ErrUniqueEntry) {
			return ContactInfo{}, errs.New(errs.Aborted, contactinfobus.ErrUniqueEntry)
		}
		return ContactInfo{}, errs.Newf(errs.Internal, "create:  contact info[%+v]: %s", ass, err)
	}

	return ToAppContactInfo(ass), err
}

// Update updates an existing  contact info.
func (a *App) Update(ctx context.Context, app UpdateContactInfo, id uuid.UUID) (ContactInfo, error) {
	us, err := toBusUpdateContactInfo(app)
	if err != nil {
		return ContactInfo{}, errs.New(errs.InvalidArgument, err)
	}

	st, err := a.contactinfobus.QueryByID(ctx, id)
	if err != nil {
		return ContactInfo{}, errs.New(errs.NotFound, contactinfobus.ErrNotFound)
	}

	contactInfo, err := a.contactinfobus.Update(ctx, st, us)
	if err != nil {
		if errors.Is(err, contactinfobus.ErrNotFound) {
			return ContactInfo{}, errs.New(errs.NotFound, err)
		}
		return ContactInfo{}, errs.Newf(errs.Internal, "update:  contact info[%+v]: %s", contactInfo, err)
	}

	return ToAppContactInfo(contactInfo), nil
}

// Delete removes an existing  contact info.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	st, err := a.contactinfobus.QueryByID(ctx, id)
	if err != nil {
		return errs.New(errs.NotFound, contactinfobus.ErrNotFound)
	}

	err = a.contactinfobus.Delete(ctx, st)
	if err != nil {
		return errs.Newf(errs.Internal, "delete:  contact info[%+v]: %s", st, err)
	}

	return nil
}

// Query returns a list of  contact infos based on the filter, order and page.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[ContactInfo], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[ContactInfo]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[ContactInfo]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[ContactInfo]{}, errs.NewFieldsError("orderby", err)
	}

	contactInfos, err := a.contactinfobus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[ContactInfo]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.contactinfobus.Count(ctx, filter)
	if err != nil {
		return query.Result[ContactInfo]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppContactInfos(contactInfos), total, page), nil
}

// QueryByID retrieves a single contact info by its id.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (ContactInfo, error) {
	contactInfo, err := a.contactinfobus.QueryByID(ctx, id)
	if err != nil {
		return ContactInfo{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppContactInfo(contactInfo), nil
}
