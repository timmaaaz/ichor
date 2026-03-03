package settingsapp

import (
	"context"
	"errors"
	"fmt"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/query"
	"github.com/timmaaaz/ichor/business/domain/config/settingsbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// App manages the set of app layer APIs for settings.
type App struct {
	settingsbus *settingsbus.Business
}

// NewApp constructs a settings app API for use.
func NewApp(settingsbus *settingsbus.Business) *App {
	return &App{
		settingsbus: settingsbus,
	}
}

// Create creates a new setting.
func (a *App) Create(ctx context.Context, app NewSetting) (Setting, error) {
	newSetting := toBusNewSetting(app)

	s, err := a.settingsbus.Create(ctx, newSetting)
	if err != nil {
		if errors.Is(err, settingsbus.ErrUniqueEntry) {
			return Setting{}, errs.New(errs.AlreadyExists, err)
		}
		return Setting{}, fmt.Errorf("create: %w", err)
	}

	return ToAppSetting(s), nil
}

// Update updates an existing setting.
func (a *App) Update(ctx context.Context, key string, app UpdateSetting) (Setting, error) {
	updateSetting := toBusUpdateSetting(app)

	s, err := a.settingsbus.QueryByKey(ctx, key)
	if err != nil {
		if errors.Is(err, settingsbus.ErrNotFound) {
			return Setting{}, errs.New(errs.NotFound, err)
		}
		return Setting{}, fmt.Errorf("update [queryByKey]: %w", err)
	}

	s, err = a.settingsbus.Update(ctx, s, updateSetting)
	if err != nil {
		return Setting{}, fmt.Errorf("update: %w", err)
	}

	return ToAppSetting(s), nil
}

// Delete deletes an existing setting.
func (a *App) Delete(ctx context.Context, key string) error {
	s, err := a.settingsbus.QueryByKey(ctx, key)
	if err != nil {
		if errors.Is(err, settingsbus.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return fmt.Errorf("delete [queryByKey]: %w", err)
	}

	if err := a.settingsbus.Delete(ctx, s); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves settings based on filter/order/page.
func (a *App) Query(ctx context.Context, qp QueryParams) (query.Result[Setting], error) {
	pg, err := page.Parse(qp.Page, qp.Rows)
	if err != nil {
		return query.Result[Setting]{}, errs.NewFieldsError("page", err)
	}

	filter, err := parseFilter(qp)
	if err != nil {
		return query.Result[Setting]{}, errs.NewFieldsError("filter", err)
	}

	orderBy, err := order.Parse(orderByFields, qp.OrderBy, defaultOrderBy)
	if err != nil {
		return query.Result[Setting]{}, errs.NewFieldsError("orderBy", err)
	}

	results, err := a.settingsbus.Query(ctx, filter, orderBy, pg)
	if err != nil {
		return query.Result[Setting]{}, errs.Newf(errs.Internal, "query: %v", err)
	}

	total, err := a.settingsbus.Count(ctx, filter)
	if err != nil {
		return query.Result[Setting]{}, errs.Newf(errs.Internal, "count: %v", err)
	}

	return query.NewResult(ToAppSettings(results), total, pg), nil
}

// QueryByKey retrieves a setting by its key.
func (a *App) QueryByKey(ctx context.Context, key string) (Setting, error) {
	s, err := a.settingsbus.QueryByKey(ctx, key)
	if err != nil {
		if errors.Is(err, settingsbus.ErrNotFound) {
			return Setting{}, errs.New(errs.NotFound, err)
		}
		return Setting{}, fmt.Errorf("querybykey: %w", err)
	}

	return ToAppSetting(s), nil
}
