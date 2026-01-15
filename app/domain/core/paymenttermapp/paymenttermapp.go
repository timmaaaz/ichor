package paymenttermapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/core/paymenttermbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer api functions for the payment term domain.
type App struct {
	paymentTermBus *paymenttermbus.Business
	auth           *auth.Auth
}

// NewApp constructs a payment term app API for use.
func NewApp(paymentTermBus *paymenttermbus.Business) *App {
	return &App{
		paymentTermBus: paymentTermBus,
	}
}

// NewAppWithAuth constructs a payment term app API for use with auth support.
func NewAppWithAuth(paymentTermBus *paymenttermbus.Business, ath *auth.Auth) *App {
	return &App{
		auth:           ath,
		paymentTermBus: paymentTermBus,
	}
}

// Create adds a new payment term to the system.
func (a *App) Create(ctx context.Context, app NewPaymentTerm) (PaymentTerm, error) {
	paymentTerm, err := a.paymentTermBus.Create(ctx, ToBusNewPaymentTerm(app))
	if err != nil {
		if errors.Is(err, paymenttermbus.ErrUniqueEntry) {
			return PaymentTerm{}, errs.New(errs.Aborted, paymenttermbus.ErrUniqueEntry)
		}
		return PaymentTerm{}, errs.Newf(errs.Internal, "create: payment term[%+v]: %s", paymentTerm, err)
	}

	return ToAppPaymentTerm(paymentTerm), nil
}

// Update updates an existing payment term.
func (a *App) Update(ctx context.Context, app UpdatePaymentTerm, id uuid.UUID) (PaymentTerm, error) {
	upt := ToBusUpdatePaymentTerm(app)

	pt, err := a.paymentTermBus.QueryByID(ctx, id)
	if err != nil {
		return PaymentTerm{}, errs.Newf(errs.NotFound, "update: payment term[%s]: %s", id, err)
	}

	paymentTerm, err := a.paymentTermBus.Update(ctx, pt, upt)
	if err != nil {
		if errors.Is(err, paymenttermbus.ErrNotFound) {
			return PaymentTerm{}, errs.New(errs.NotFound, err)
		}
		return PaymentTerm{}, errs.Newf(errs.Internal, "update: payment term[%+v]: %s", paymentTerm, err)
	}

	return ToAppPaymentTerm(paymentTerm), nil
}

// Delete removes an existing payment term.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	pt, err := a.paymentTermBus.QueryByID(ctx, id)
	if err != nil {
		return errs.Newf(errs.NotFound, "delete: payment term[%s]: %s", id, err)
	}

	if err := a.paymentTermBus.Delete(ctx, pt); err != nil {
		return errs.Newf(errs.Internal, "delete: payment term[%+v]: %s", pt, err)
	}

	return nil
}

// Query returns a list of payment terms.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[PaymentTerm], error) {
	page, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[PaymentTerm]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[PaymentTerm]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[PaymentTerm]{}, errs.NewFieldsError("orderby", err)
	}

	pts, err := a.paymentTermBus.Query(ctx, filter, orderBy, page)
	if err != nil {
		return query.Result[PaymentTerm]{}, errs.Newf(errs.Internal, "query: %s", err)
	}

	total, err := a.paymentTermBus.Count(ctx, filter)
	if err != nil {
		return query.Result[PaymentTerm]{}, errs.Newf(errs.Internal, "count: %s", err)
	}

	return query.NewResult(ToAppPaymentTerms(pts), total, page), nil
}

// QueryByID returns a single payment term based on the id.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (PaymentTerm, error) {
	pt, err := a.paymentTermBus.QueryByID(ctx, id)
	if err != nil {
		return PaymentTerm{}, errs.Newf(errs.NotFound, "query: payment term[%s]: %s", id, err)
	}

	return ToAppPaymentTerm(pt), nil
}
