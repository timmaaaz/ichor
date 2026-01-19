package pageconfigapp

import (
	"context"
	"errors"
	"time"

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

// QueryAll retrieves all page configurations from the system.
func (a *App) QueryAll(ctx context.Context) (PageConfigs, error) {
	configs, err := a.pageConfigBus.QueryAll(ctx)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "queryall: %s", err)
	}

	return PageConfigs(ToAppPageConfigs(configs)), nil
}

// =============================================================================
// Export/Import Methods

// ExportByIDs exports page configs by IDs as a JSON package.
func (a *App) ExportByIDs(ctx context.Context, configIDs []string) (ExportPackage, error) {
	// Convert string IDs to UUIDs
	uuids := make([]uuid.UUID, len(configIDs))
	for i, id := range configIDs {
		uid, err := uuid.Parse(id)
		if err != nil {
			return ExportPackage{}, errs.Newf(errs.InvalidArgument, "invalid config ID %s: %s", id, err)
		}
		uuids[i] = uid
	}

	// Export from business layer
	results, err := a.pageConfigBus.ExportByIDs(ctx, uuids)
	if err != nil {
		return ExportPackage{}, errs.Newf(errs.Internal, "export: %s", err)
	}

	// Convert to app models
	var packages []PageConfigPackage
	for _, result := range results {
		packages = append(packages, toAppPageConfigWithRelations(result))
	}

	return ExportPackage{
		Version:    "1.0",
		Type:       "page-configs",
		ExportedAt: time.Now().Format(time.RFC3339),
		Count:      len(packages),
		Data:       packages,
	}, nil
}

// ImportPageConfigs imports page configs from a JSON package.
func (a *App) ImportPageConfigs(ctx context.Context, pkg ImportPackage) (ImportResult, error) {
	// Validate package
	if err := pkg.Validate(); err != nil {
		return ImportResult{}, err
	}

	// Convert app models to business models
	var busPackages []pageconfigbus.PageConfigWithRelations
	for i, configPkg := range pkg.Data {
		busPkg, err := ToBusPageConfigWithRelations(configPkg)
		if err != nil {
			return ImportResult{
				Errors: []string{err.Error()},
			}, errs.Newf(errs.InvalidArgument, "convert page config %d: %s", i, err)
		}
		busPackages = append(busPackages, busPkg)
	}

	// Import via business layer
	stats, err := a.pageConfigBus.ImportPageConfigs(ctx, busPackages, pkg.Mode)
	if err != nil {
		return ImportResult{
			Errors: []string{err.Error()},
		}, errs.Newf(errs.Internal, "import: %s", err)
	}

	return ImportResult{
		ImportedCount: stats.ImportedCount,
		SkippedCount:  stats.SkippedCount,
		UpdatedCount:  stats.UpdatedCount,
	}, nil
}

// =============================================================================
// JSON Blob Import/Export/Validation Methods

// ValidateBlob validates a page config JSON blob.
// Returns app layer ValidationResult (with Encode()) for HTTP response.
func (a *App) ValidateBlob(ctx context.Context, blob []byte) (ValidationResult, error) {
	// Call business layer (returns business types)
	busResult, err := a.pageConfigBus.ValidateImportBlob(ctx, blob)
	if err != nil {
		return ValidationResult{}, errs.Newf(errs.Internal, "validate: %s", err)
	}

	// Convert business → app types
	return toAppValidationResult(busResult), nil
}

// ImportBlob imports a page config from JSON blob.
// Returns app layer ImportStats (with Encode()) for HTTP response.
func (a *App) ImportBlob(ctx context.Context, blob []byte, mode string) (ImportStats, error) {
	// Validate first
	busResult, err := a.pageConfigBus.ValidateImportBlob(ctx, blob)
	if err != nil {
		return ImportStats{}, errs.Newf(errs.Internal, "validate: %s", err)
	}

	if !busResult.Valid {
		// Return validation errors in error response
		return ImportStats{}, errs.New(errs.InvalidArgument, errors.New("validation failed"))
	}

	// Import into database (returns business types)
	busStats, err := a.pageConfigBus.ImportBlob(ctx, blob, mode)
	if err != nil {
		return ImportStats{}, errs.Newf(errs.Internal, "import: %s", err)
	}

	// Convert business → app types
	return toAppImportStats(busStats), nil
}

// ExportBlob exports a page config as JSON blob.
func (a *App) ExportBlob(ctx context.Context, configID uuid.UUID) (pageconfigbus.PageConfigWithRelations, error) {
	// Use existing ExportByIDs method
	results, err := a.pageConfigBus.ExportByIDs(ctx, []uuid.UUID{configID})
	if err != nil {
		return pageconfigbus.PageConfigWithRelations{}, errs.Newf(errs.Internal, "export: %s", err)
	}

	if len(results) == 0 {
		return pageconfigbus.PageConfigWithRelations{}, errs.New(errs.NotFound, errors.New("page config not found"))
	}

	return results[0], nil
}

// ExportBlobAsApp exports a page config as app layer type (snake_case JSON).
func (a *App) ExportBlobAsApp(ctx context.Context, configID uuid.UUID) (PageConfigPackage, error) {
	busPkg, err := a.ExportBlob(ctx, configID)
	if err != nil {
		return PageConfigPackage{}, err
	}
	return toAppPageConfigWithRelations(busPkg), nil
}
