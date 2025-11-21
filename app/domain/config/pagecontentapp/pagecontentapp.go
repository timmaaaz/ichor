package pagecontentapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/config/pagecontentbus"
)

// App manages the set of app layer APIs for page content.
type App struct {
	pageContentBus *pagecontentbus.Business
}

// NewApp constructs a page content app API for use.
func NewApp(pageContentBus *pagecontentbus.Business) *App {
	return &App{
		pageContentBus: pageContentBus,
	}
}

// Create adds a new page content block to the system.
func (a *App) Create(ctx context.Context, app NewPageContent) (PageContent, error) {
	if err := app.Validate(); err != nil {
		return PageContent{}, err
	}

	nc, err := toBusNewPageContent(app)
	if err != nil {
		return PageContent{}, errs.New(errs.InvalidArgument, err)
	}

	content, err := a.pageContentBus.Create(ctx, nc)
	if err != nil {
		return PageContent{}, errs.Newf(errs.Internal, "create: %s", err)
	}

	return ToAppPageContent(content), nil
}

// Update modifies an existing page content block.
func (a *App) Update(ctx context.Context, app UpdatePageContent, contentID uuid.UUID) (PageContent, error) {
	if err := app.Validate(); err != nil {
		return PageContent{}, err
	}

	uc := toBusUpdatePageContent(app)

	content, err := a.pageContentBus.Update(ctx, uc, contentID)
	if err != nil {
		if errors.Is(err, pagecontentbus.ErrNotFound) {
			return PageContent{}, errs.New(errs.NotFound, pagecontentbus.ErrNotFound)
		}
		return PageContent{}, errs.Newf(errs.Internal, "update: %s", err)
	}

	return ToAppPageContent(content), nil
}

// Delete removes a page content block from the system.
func (a *App) Delete(ctx context.Context, contentID uuid.UUID) error {
	if err := a.pageContentBus.Delete(ctx, contentID); err != nil {
		if errors.Is(err, pagecontentbus.ErrNotFound) {
			return errs.New(errs.NotFound, pagecontentbus.ErrNotFound)
		}
		return errs.Newf(errs.Internal, "delete: %s", err)
	}

	return nil
}

// QueryByID finds a page content block by its ID.
func (a *App) QueryByID(ctx context.Context, contentID uuid.UUID) (PageContent, error) {
	content, err := a.pageContentBus.QueryByID(ctx, contentID)
	if err != nil {
		return PageContent{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppPageContent(content), nil
}

// QueryByPageConfigID retrieves all content blocks for a specific page config.
func (a *App) QueryByPageConfigID(ctx context.Context, pageConfigID uuid.UUID) (PageContents, error) {
	contents, err := a.pageContentBus.QueryByPageConfigID(ctx, pageConfigID)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "querybypageconfigid: %s", err)
	}

	return PageContents(ToAppPageContents(contents)), nil
}

// QueryWithChildren retrieves content blocks and nests children under parents.
func (a *App) QueryWithChildren(ctx context.Context, pageConfigID uuid.UUID) (PageContents, error) {
	contents, err := a.pageContentBus.QueryWithChildren(ctx, pageConfigID)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "querywithchildren: %s", err)
	}

	return PageContents(ToAppPageContents(contents)), nil
}
