package pagecontentapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/config/pagecontentapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	pageContentApp *pagecontentapp.App
}

func newAPI(pageContentApp *pagecontentapp.App) *api {
	return &api{
		pageContentApp: pageContentApp,
	}
}

// create adds a new page content block to the system.
func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app pagecontentapp.NewPageContent
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	content, err := api.pageContentApp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return content
}

// update modifies an existing page content block.
func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app pagecontentapp.UpdatePageContent
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	contentID, err := uuid.Parse(web.Param(r, "content_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	content, err := api.pageContentApp.Update(ctx, app, contentID)
	if err != nil {
		return errs.NewError(err)
	}

	return content
}

// delete removes a page content block from the system.
func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	contentID, err := uuid.Parse(web.Param(r, "content_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := api.pageContentApp.Delete(ctx, contentID); err != nil {
		return errs.NewError(err)
	}

	return nil
}

// queryByID retrieves a single page content block by ID.
func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	contentID, err := uuid.Parse(web.Param(r, "content_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	content, err := api.pageContentApp.QueryByID(ctx, contentID)
	if err != nil {
		return errs.NewError(err)
	}

	return content
}

// queryByPageConfigID retrieves all content blocks for a specific page config.
func (api *api) queryByPageConfigID(ctx context.Context, r *http.Request) web.Encoder {
	pageConfigID, err := uuid.Parse(web.Param(r, "page_config_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	contents, err := api.pageContentApp.QueryByPageConfigID(ctx, pageConfigID)
	if err != nil {
		return errs.NewError(err)
	}

	return contents
}

// queryWithChildren retrieves content blocks and nests children under parents.
func (api *api) queryWithChildren(ctx context.Context, r *http.Request) web.Encoder {
	pageConfigID, err := uuid.Parse(web.Param(r, "page_config_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	contents, err := api.pageContentApp.QueryWithChildren(ctx, pageConfigID)
	if err != nil {
		return errs.NewError(err)
	}

	return contents
}
