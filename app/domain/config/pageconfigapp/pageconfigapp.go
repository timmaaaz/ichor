package pageconfigapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/config/pageconfigbus"
)

// App manages the set of app layer APIs for page configuration.
type App struct {
	pageConfigBus *pageconfigbus.Business
}

// NewApp constructs a page config app API for use.
func NewApp(pageConfigBus *pageconfigbus.Business) *App {
	return &App{
		pageConfigBus: pageConfigBus,
	}
}

// Create adds a new page configuration to the system.
func (a *App) Create(ctx context.Context, app NewPageConfig) (PageConfig, error) {
	if err := app.Validate(); err != nil {
		return PageConfig{}, err
	}

	nc, err := toBusNewPageConfig(app)
	if err != nil {
		return PageConfig{}, errs.New(errs.InvalidArgument, err)
	}

	config, err := a.pageConfigBus.Create(ctx, nc)
	if err != nil {
		return PageConfig{}, errs.Newf(errs.Internal, "create: %s", err)
	}

	return ToAppPageConfig(config), nil
}

// Update modifies an existing page configuration.
func (a *App) Update(ctx context.Context, app UpdatePageConfig, configID uuid.UUID) (PageConfig, error) {
	if err := app.Validate(); err != nil {
		return PageConfig{}, err
	}

	uc, err := toBusUpdatePageConfig(app)
	if err != nil {
		return PageConfig{}, errs.New(errs.InvalidArgument, err)
	}

	config, err := a.pageConfigBus.Update(ctx, uc, configID)
	if err != nil {
		if errors.Is(err, pageconfigbus.ErrNotFound) {
			return PageConfig{}, errs.New(errs.NotFound, pageconfigbus.ErrNotFound)
		}
		return PageConfig{}, errs.Newf(errs.Internal, "update: %s", err)
	}

	return ToAppPageConfig(config), nil
}

// Delete removes a page configuration from the system.
func (a *App) Delete(ctx context.Context, configID uuid.UUID) error {
	if err := a.pageConfigBus.Delete(ctx, configID); err != nil {
		if errors.Is(err, pageconfigbus.ErrNotFound) {
			return errs.New(errs.NotFound, pageconfigbus.ErrNotFound)
		}
		return errs.Newf(errs.Internal, "delete: %s", err)
	}

	return nil
}

// QueryByID finds a page configuration by its ID.
func (a *App) QueryByID(ctx context.Context, configID uuid.UUID) (PageConfig, error) {
	config, err := a.pageConfigBus.QueryByID(ctx, configID)
	if err != nil {
		return PageConfig{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppPageConfig(config), nil
}

// QueryByName retrieves the default page configuration by name.
func (a *App) QueryByName(ctx context.Context, name string) (PageConfig, error) {
	config, err := a.pageConfigBus.QueryByName(ctx, name)
	if err != nil {
		return PageConfig{}, errs.Newf(errs.Internal, "querybyname: %s", err)
	}

	return ToAppPageConfig(config), nil
}

// QueryByNameAndUserID retrieves a user-specific page configuration.
func (a *App) QueryByNameAndUserID(ctx context.Context, name string, userID uuid.UUID) (PageConfig, error) {
	config, err := a.pageConfigBus.QueryByNameAndUserID(ctx, name, userID)
	if err != nil {
		return PageConfig{}, errs.Newf(errs.Internal, "querybynameanduserid: %s", err)
	}

	return ToAppPageConfig(config), nil
}
