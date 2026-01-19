package pageconfigapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/config/pageconfigapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	pageConfigApp *pageconfigapp.App
}

func newAPI(pageConfigApp *pageconfigapp.App) *api {
	return &api{
		pageConfigApp: pageConfigApp,
	}
}

// create adds a new page configuration to the system.
func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app pageconfigapp.NewPageConfig
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	config, err := api.pageConfigApp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return config
}

// update modifies an existing page configuration.
func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app pageconfigapp.UpdatePageConfig
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	configID, err := uuid.Parse(web.Param(r, "config_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	config, err := api.pageConfigApp.Update(ctx, app, configID)
	if err != nil {
		return errs.NewError(err)
	}

	return config
}

// delete removes a page configuration from the system.
func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	configID, err := uuid.Parse(web.Param(r, "config_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := api.pageConfigApp.Delete(ctx, configID); err != nil {
		return errs.NewError(err)
	}

	return nil
}

// queryByID retrieves a single page configuration by ID.
func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	configID, err := uuid.Parse(web.Param(r, "config_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	config, err := api.pageConfigApp.QueryByID(ctx, configID)
	if err != nil {
		return errs.NewError(err)
	}

	return config
}

// queryByName retrieves the default page configuration by name.
func (api *api) queryByName(ctx context.Context, r *http.Request) web.Encoder {
	name := web.Param(r, "name")

	config, err := api.pageConfigApp.QueryByName(ctx, name)
	if err != nil {
		return errs.NewError(err)
	}

	return config
}

// queryAll retrieves all page configurations from the system.
func (api *api) queryAll(ctx context.Context, r *http.Request) web.Encoder {
	configs, err := api.pageConfigApp.QueryAll(ctx)
	if err != nil {
		return errs.NewError(err)
	}

	return configs
}

// =============================================================================
// Export/Import handlers

func (api *api) exportPageConfigs(ctx context.Context, r *http.Request) web.Encoder {
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

	pkg, err := api.pageConfigApp.ExportByIDs(ctx, req.IDs)
	if err != nil {
		return errs.NewError(err)
	}

	return pkg
}

func (api *api) importPageConfigs(ctx context.Context, r *http.Request) web.Encoder {
	var pkg pageconfigapp.ImportPackage
	if err := web.Decode(r, &pkg); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	result, err := api.pageConfigApp.ImportPageConfigs(ctx, pkg)
	if err != nil {
		return errs.NewError(err)
	}

	return result
}

// =============================================================================
// JSON Blob Import/Export/Validation handlers

const maxImportBlobSize = 10 * 1024 * 1024 // 10MB

// validateBlob handles POST /v1/config/page-configs/validate
func (api *api) validateBlob(ctx context.Context, r *http.Request) web.Encoder {
	// Check content length before reading body
	if r.ContentLength > maxImportBlobSize {
		return errs.Newf(errs.InvalidArgument,
			"request body exceeds maximum size of %d bytes", maxImportBlobSize)
	}

	// Read raw JSON body
	blob, err := io.ReadAll(r.Body)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}
	defer r.Body.Close()

	// App layer returns app.ValidationResult (with Encode())
	result, err := api.pageConfigApp.ValidateBlob(ctx, blob)
	if err != nil {
		return errs.NewError(err)
	}

	return result // App layer ValidationResult implements web.Encoder
}

// importBlob handles POST /v1/config/page-configs/import?mode={skip|replace|merge}
func (api *api) importBlob(ctx context.Context, r *http.Request) web.Encoder {
	// Check content length before reading body
	if r.ContentLength > maxImportBlobSize {
		return errs.Newf(errs.InvalidArgument,
			"request body exceeds maximum size of %d bytes", maxImportBlobSize)
	}

	// Read raw JSON body
	blob, err := io.ReadAll(r.Body)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}
	defer r.Body.Close()

	// Get mode from query param, default to "replace"
	mode := r.URL.Query().Get("mode")
	if mode == "" {
		mode = "replace"
	}

	// Validate mode parameter
	if mode != "skip" && mode != "replace" && mode != "merge" {
		return errs.Newf(errs.InvalidArgument,
			"invalid mode: %s (must be: skip, replace, or merge)", mode)
	}

	// App layer returns app.ImportStats (with Encode())
	stats, err := api.pageConfigApp.ImportBlob(ctx, blob, mode)
	if err != nil {
		return errs.NewError(err)
	}

	return stats // App layer ImportStats implements web.Encoder
}

// exportBlob handles GET /v1/config/page-configs/{config_id}/export
func (api *api) exportBlob(ctx context.Context, r *http.Request) web.Encoder {
	configID, err := uuid.Parse(web.Param(r, "config_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	// Export blob returns app layer type with snake_case JSON
	pkg, err := api.pageConfigApp.ExportBlobAsApp(ctx, configID)
	if err != nil {
		return errs.NewError(err)
	}

	return pkg // PageConfigPackage implements Encode()
}

// rawJSON is a simple wrapper to return raw JSON bytes
type rawJSON []byte

func (r rawJSON) Encode() ([]byte, string, error) {
	return r, "application/json", nil
}
