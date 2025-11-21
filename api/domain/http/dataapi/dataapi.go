package dataapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/dataapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	dataapp *dataapp.App
}

func newAPI(dataapp *dataapp.App) *api {
	return &api{
		dataapp: dataapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app dataapp.NewTableConfig
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	tConfig, err := api.dataapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return tConfig
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app dataapp.UpdateTableConfig
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	id := web.Param(r, "table_config_id")
	parsed, err := uuid.Parse(id)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	tConfig, err := api.dataapp.Update(ctx, parsed, app)
	if err != nil {
		return errs.NewError(err)
	}

	return tConfig
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	id := web.Param(r, "table_config_id")
	parsed, err := uuid.Parse(id)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := api.dataapp.Delete(ctx, parsed); err != nil {
		return errs.NewError(err)
	}

	return nil
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	id := web.Param(r, "table_config_id")
	parsed, err := uuid.Parse(id)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	tableConfig, err := api.dataapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return tableConfig
}

func (api *api) queryByName(ctx context.Context, r *http.Request) web.Encoder {
	name := web.Param(r, "name")

	tableConfig, err := api.dataapp.QueryByName(ctx, name)
	if err != nil {
		return errs.NewError(err)
	}

	return tableConfig
}

func (api *api) queryByUser(ctx context.Context, r *http.Request) web.Encoder {
	id := web.Param(r, "user_id")
	parsed, err := uuid.Parse(id)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	tableConfigs, err := api.dataapp.QueryByUser(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return tableConfigs
}

func (api *api) queryAll(ctx context.Context, r *http.Request) web.Encoder {
	tableConfigs, err := api.dataapp.QueryAll(ctx)
	if err != nil {
		return errs.NewError(err)
	}

	return tableConfigs
}

func (api *api) executeQuery(ctx context.Context, r *http.Request) web.Encoder {
	id := web.Param(r, "table_config_id")
	parsed, err := uuid.Parse(id)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	// Query filters/pagination stuff
	var app dataapp.TableQuery
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	tableData, err := api.dataapp.ExecuteQuery(ctx, parsed, app)
	if err != nil {
		return errs.NewError(err)
	}

	return tableData
}

func (api *api) executeQueryByName(ctx context.Context, r *http.Request) web.Encoder {
	name := web.Param(r, "name")

	var app dataapp.TableQuery
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	tableData, err := api.dataapp.ExecuteQueryByName(ctx, name, app)
	if err != nil {
		return errs.NewError(err)
	}

	return tableData
}

func (api *api) executeQueryCountByID(ctx context.Context, r *http.Request) web.Encoder {
	id := web.Param(r, "table_config_id")
	parsed, err := uuid.Parse(id)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	// Query filters/pagination stuff - allow empty body
	var app dataapp.TableQuery
	if err := web.Decode(r, &app); err != nil {
		// Only return error if it's not an empty body or EOF
		if r.ContentLength > 0 && !errors.Is(err, io.EOF) {
			return errs.New(errs.InvalidArgument, err)
		}
	}

	count, err := api.dataapp.ExecuteQueryCountByID(ctx, parsed, app)
	if err != nil {
		return errs.NewError(err)
	}

	return count
}

func (api *api) executeQueryCountByName(ctx context.Context, r *http.Request) web.Encoder {
	name := web.Param(r, "name")

	var app dataapp.TableQuery
	if err := web.Decode(r, &app); err != nil {
		// Only return error if it's not an empty body or EOF
		if r.ContentLength > 0 && !errors.Is(err, io.EOF) {
			return errs.New(errs.InvalidArgument, err)
		}
	}

	count, err := api.dataapp.ExecuteQueryCountByName(ctx, name, app)
	if err != nil {
		return errs.NewError(err)
	}

	return count
}

func (api *api) validateConfig(ctx context.Context, r *http.Request) web.Encoder {
	var app dataapp.NewTableConfig
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := api.dataapp.ValidateConfig(ctx, app); err != nil {
		return errs.NewError(err)
	}

	return nil
}

// =============================================================================
// PageConfig handlers

func (api *api) createPageConfig(ctx context.Context, r *http.Request) web.Encoder {
	var app dataapp.NewPageConfig
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	pageConfig, err := api.dataapp.CreatePageConfig(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return pageConfig
}

func (api *api) updatePageConfig(ctx context.Context, r *http.Request) web.Encoder {
	var app dataapp.UpdatePageConfig
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	id := web.Param(r, "page_config_id")
	parsed, err := uuid.Parse(id)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	pageConfig, err := api.dataapp.UpdatePageConfig(ctx, parsed, app)
	if err != nil {
		return errs.NewError(err)
	}

	return pageConfig
}

func (api *api) deletePageConfig(ctx context.Context, r *http.Request) web.Encoder {
	id := web.Param(r, "page_config_id")
	parsed, err := uuid.Parse(id)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := api.dataapp.DeletePageConfig(ctx, parsed); err != nil {
		return errs.NewError(err)
	}

	return nil
}

func (api *api) queryFullPageByName(ctx context.Context, r *http.Request) web.Encoder {
	name := web.Param(r, "name")

	fullPageConfig, err := api.dataapp.QueryFullPageByName(ctx, name)
	if err != nil {
		return errs.NewError(err)
	}

	return fullPageConfig
}

func (api *api) queryFullPageByNameAndUserID(ctx context.Context, r *http.Request) web.Encoder {
	name := web.Param(r, "name")
	userIDStr := web.Param(r, "user_id")

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	fullPageConfig, err := api.dataapp.QueryFullPageByNameAndUserID(ctx, name, userID)
	if err != nil {
		return errs.NewError(err)
	}

	return fullPageConfig
}

func (api *api) queryFullPageByID(ctx context.Context, r *http.Request) web.Encoder {
	id := web.Param(r, "page_config_id")
	parsed, err := uuid.Parse(id)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	fullPageConfig, err := api.dataapp.QueryFullPageByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return fullPageConfig
}

// =============================================================================
// Export/Import handlers

func (api *api) exportTableConfigs(ctx context.Context, r *http.Request) web.Encoder {
	var req struct {
		IDs []string `json:"ids"`
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if len(req.IDs) == 0 {
		return errs.New(errs.InvalidArgument, errs.Newf(errs.InvalidArgument, "ids cannot be empty"))
	}

	pkg, err := api.dataapp.ExportByIDs(ctx, req.IDs)
	if err != nil {
		return errs.NewError(err)
	}

	return pkg
}

func (api *api) importTableConfigs(ctx context.Context, r *http.Request) web.Encoder {
	var pkg dataapp.ImportPackage
	if err := web.Decode(r, &pkg); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	result, err := api.dataapp.ImportTableConfigs(ctx, pkg)
	if err != nil {
		return errs.NewError(err)
	}

	return result
}
