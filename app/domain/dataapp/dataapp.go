// Package data maintains the app layer api for the tablebuilder domain.
package dataapp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/config/pageactionapp"
	"github.com/timmaaaz/ichor/app/domain/config/pageconfigapp"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/mid"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/business/sdk/tablebuilder"
)

// App manages the set of app layer api functions for the tablebuilder domain.
type App struct {
	configStore   *tablebuilder.ConfigStore
	tableStore    *tablebuilder.Store
	auth          *auth.Auth
	pageactionapp *pageactionapp.App
	pageconfigapp *pageconfigapp.App
}

// NewApp constructs a tablebuilder app API for use.
func NewApp(configStore *tablebuilder.ConfigStore, tableStore *tablebuilder.Store, pageactionapp *pageactionapp.App, pageconfigapp *pageconfigapp.App) *App {
	return &App{
		configStore:   configStore,
		tableStore:    tableStore,
		pageactionapp: pageactionapp,
		pageconfigapp: pageconfigapp,
	}
}

// NewAppWithAuth constructs a tablebuilder app API for use with auth support.
func NewAppWithAuth(configStore *tablebuilder.ConfigStore, tableStore *tablebuilder.Store, ath *auth.Auth, pageactionapp *pageactionapp.App, pageconfigapp *pageconfigapp.App) *App {
	return &App{
		auth:          ath,
		configStore:   configStore,
		tableStore:    tableStore,
		pageactionapp: pageactionapp,
		pageconfigapp: pageconfigapp,
	}
}

// Create adds a new table configuration to the system.
func (a *App) Create(ctx context.Context, app NewTableConfig) (TableConfig, error) {
	config, err := toBusNewTableConfig(app)
	if err != nil {
		return TableConfig{}, errs.New(errs.InvalidArgument, err)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return TableConfig{}, errs.Newf(errs.Internal, "user missing in context: %s", err)
	}

	stored, err := a.configStore.Create(ctx, app.Name, app.Description, config, userID)
	if err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return TableConfig{}, errs.New(errs.Aborted, errors.New("configuration name already exists"))
		}
		return TableConfig{}, errs.Newf(errs.Internal, "create: config[%s]: %s", app.Name, err)
	}

	return ToAppTableConfig(*stored), nil
}

// CreatePageConfig adds a new page configuration to the system.
func (a *App) CreatePageConfig(ctx context.Context, app NewPageConfig) (PageConfig, error) {
	return a.pageconfigapp.Create(ctx, app)
}

// Update updates an existing table configuration.
func (a *App) Update(ctx context.Context, id uuid.UUID, app UpdateTableConfig) (TableConfig, error) {
	// Get current config to preserve unchanged fields
	current, err := a.configStore.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, tablebuilder.ErrNotFound) {
			return TableConfig{}, errs.New(errs.NotFound, err)
		}
		return TableConfig{}, errs.Newf(errs.Internal, "get config: %s", err)
	}

	// Apply updates
	name := current.Name
	if app.Name != nil {
		name = *app.Name
	}

	description := current.Description
	if app.Description != nil {
		description = *app.Description
	}

	config, err := toBusUpdateTableConfig(app)
	if err != nil {
		return TableConfig{}, errs.New(errs.InvalidArgument, err)
	}

	// If no config update, use existing
	if config == nil {
		var existingConfig tablebuilder.Config
		if err := json.Unmarshal(current.Config, &existingConfig); err != nil {
			return TableConfig{}, errs.Newf(errs.Internal, "unmarshal existing config: %s", err)
		}
		config = &existingConfig
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return TableConfig{}, errs.Newf(errs.Internal, "user missing in context: %s", err)
	}

	stored, err := a.configStore.Update(ctx, id, name, description, config, userID)
	if err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return TableConfig{}, errs.New(errs.Aborted, errors.New("configuration name already exists"))
		}
		return TableConfig{}, errs.Newf(errs.Internal, "update: configID[%s]: %s", id, err)
	}

	return ToAppTableConfig(*stored), nil
}

// UpdatePageConfig updates an existing page configuration.
func (a *App) UpdatePageConfig(ctx context.Context, id uuid.UUID, app UpdatePageConfig) (PageConfig, error) {
	return a.pageconfigapp.Update(ctx, app, id)
}

// Delete removes a table configuration from the system.
func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	if err := a.configStore.Delete(ctx, id); err != nil {
		if errors.Is(err, tablebuilder.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "delete: configID[%s]: %s", id, err)
	}

	return nil
}

// DeletePageConfig removes a page configuration from the system.
func (a *App) DeletePageConfig(ctx context.Context, id uuid.UUID) error {
	return a.pageconfigapp.Delete(ctx, id)
}

// QueryByID returns a table configuration by its ID.
func (a *App) QueryByID(ctx context.Context, id uuid.UUID) (TableConfig, error) {
	stored, err := a.configStore.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, tablebuilder.ErrNotFound) {
			return TableConfig{}, errs.New(errs.NotFound, err)
		}
		return TableConfig{}, errs.Newf(errs.Internal, "querybyid: %s", err)
	}

	return ToAppTableConfig(*stored), nil
}

// QueryByName returns a table configuration by its name.
func (a *App) QueryByName(ctx context.Context, name string) (TableConfig, error) {
	stored, err := a.configStore.QueryByName(ctx, name)
	if err != nil {
		if errors.Is(err, tablebuilder.ErrNotFound) {
			return TableConfig{}, errs.New(errs.NotFound, err)
		}
		return TableConfig{}, errs.Newf(errs.Internal, "querybyname: %s", err)
	}

	return ToAppTableConfig(*stored), nil
}

// QueryByUser returns all table configurations created by a user.
func (a *App) QueryByUser(ctx context.Context, userID uuid.UUID) (TableConfigList, error) {
	configs, err := a.configStore.QueryByUser(ctx, userID)
	if err != nil {
		return TableConfigList{}, errs.Newf(errs.Internal, "querybyuser: %s", err)
	}

	return ToAppTableConfigList(configs), nil
}

// QueryAll returns all table configurations from the system.
func (a *App) QueryAll(ctx context.Context) (TableConfigList, error) {
	configs, err := a.configStore.QueryAll(ctx)
	if err != nil {
		return TableConfigList{}, errs.Newf(errs.Internal, "queryall: %s", err)
	}

	return ToAppTableConfigList(configs), nil
}

// QueryFullPageByName returns the default page configuration by its name.
// This retrieves the default page config which serves as a fallback for all users.
// Only one default page config is allowed per page name (enforced by database constraint).
func (a *App) QueryFullPageByName(ctx context.Context, name string) (FullPageConfig, error) {
	unescaped, err := url.QueryUnescape(name)
	if err != nil {
		return FullPageConfig{}, errs.Newf(errs.InvalidArgument, "invalid page name: %s", err)
	}

	page, err := a.pageconfigapp.QueryByName(ctx, unescaped)
	if err != nil {
		return FullPageConfig{}, err
	}

	// Parse the ID from string to UUID for querying actions
	pageID, err := uuid.Parse(page.ID)
	if err != nil {
		return FullPageConfig{}, errs.Newf(errs.Internal, "parse page config id: %s", err)
	}

	// Fetch page actions
	actions, err := a.pageactionapp.QueryByPageConfigID(ctx, pageID)
	if err != nil {
		return FullPageConfig{}, errs.Newf(errs.Internal, "query page actions: %s", err)
	}

	return FullPageConfig{
		PageConfig:  page,
		PageActions: actions,
	}, nil
}

// QueryFullPageByNameAndUserID returns a user-specific page configuration by name and user ID.
// This retrieves a specific user's customized version of a page (e.g., Jake's version of the orders page).
// Multiple users can have configs with the same page name, but only one per user+name combination.
func (a *App) QueryFullPageByNameAndUserID(ctx context.Context, name string, userID uuid.UUID) (FullPageConfig, error) {
	unescaped, err := url.QueryUnescape(name)
	if err != nil {
		return FullPageConfig{}, errs.Newf(errs.InvalidArgument, "invalid page name: %s", err)
	}

	page, err := a.pageconfigapp.QueryByNameAndUserID(ctx, unescaped, userID)
	if err != nil {
		return FullPageConfig{}, err
	}

	// Parse the ID from string to UUID for querying actions
	pageID, err := uuid.Parse(page.ID)
	if err != nil {
		return FullPageConfig{}, errs.Newf(errs.Internal, "parse page config id: %s", err)
	}

	// Fetch page actions
	actions, err := a.pageactionapp.QueryByPageConfigID(ctx, pageID)
	if err != nil {
		return FullPageConfig{}, errs.Newf(errs.Internal, "query page actions: %s", err)
	}

	return FullPageConfig{
		PageConfig:  page,
		PageActions: actions,
	}, nil
}

// QueryFullPageByID returns a full page configuration by its ID.
func (a *App) QueryFullPageByID(ctx context.Context, id uuid.UUID) (FullPageConfig, error) {
	page, err := a.pageconfigapp.QueryByID(ctx, id)
	if err != nil {
		return FullPageConfig{}, err
	}

	// Fetch page actions
	actions, err := a.pageactionapp.QueryByPageConfigID(ctx, id)
	if err != nil {
		return FullPageConfig{}, errs.Newf(errs.Internal, "query page actions: %s", err)
	}

	return FullPageConfig{
		PageConfig:  page,
		PageActions: actions,
	}, nil
}

// ExecuteQuery executes a table query with the specified configuration.
func (a *App) ExecuteQuery(ctx context.Context, id uuid.UUID, app TableQuery) (TableData, error) {
	// Load the configuration
	config, err := a.configStore.LoadConfig(ctx, id)
	if err != nil {
		if errors.Is(err, tablebuilder.ErrNotFound) {
			return TableData{}, errs.New(errs.NotFound, err)
		}
		return TableData{}, errs.Newf(errs.Internal, "load config: %s", err)
	}

	// Convert to business query params
	params := toBusTableQuery(app)

	// Execute the query
	data, err := a.tableStore.FetchTableData(ctx, config, params)
	if err != nil {
		return TableData{}, errs.Newf(errs.Internal, "execute query: %s", err)
	}

	return toAppTableData(data), nil
}

// ExecuteQueryByName executes a table query using a configuration name.
func (a *App) ExecuteQueryByName(ctx context.Context, name string, app TableQuery) (TableData, error) {
	// Load the configuration by name
	config, err := a.configStore.LoadConfigByName(ctx, name)
	if err != nil {
		if errors.Is(err, tablebuilder.ErrNotFound) {
			return TableData{}, errs.New(errs.NotFound, err)
		}
		return TableData{}, errs.Newf(errs.Internal, "load config by name: %s", err)
	}

	// Convert to business query params
	params := toBusTableQuery(app)

	// Execute the query
	data, err := a.tableStore.FetchTableData(ctx, config, params)
	if err != nil {
		return TableData{}, errs.Newf(errs.Internal, "execute query: %s", err)
	}

	return toAppTableData(data), nil
}

func (a *App) ExecuteQueryCountByID(ctx context.Context, id uuid.UUID, app TableQuery) (Count, error) {
	// Load the configuration
	config, err := a.configStore.LoadConfig(ctx, id)
	if err != nil {
		if errors.Is(err, tablebuilder.ErrNotFound) {
			return Count{}, errs.New(errs.NotFound, err)
		}
		return Count{}, errs.Newf(errs.Internal, "load config: %s", err)
	}

	// Convert to business query params
	params := toBusTableQuery(app)

	// Execute the query
	count, err := a.tableStore.FetchTableDataCount(ctx, config, params)
	if err != nil {
		return Count{}, errs.Newf(errs.Internal, "execute query count: %s", err)
	}

	return Count{Count: count}, nil
}

func (a *App) ExecuteQueryCountByName(ctx context.Context, name string, app TableQuery) (Count, error) {
	unescaped, err := url.QueryUnescape(name)
	if err != nil {
		return Count{}, errs.Newf(errs.InvalidArgument, "invalid config name: %s", err)
	}

	// Load the configuration by name
	config, err := a.configStore.LoadConfigByName(ctx, unescaped)
	if err != nil {
		if errors.Is(err, tablebuilder.ErrNotFound) {
			return Count{}, errs.New(errs.NotFound, err)
		}
		return Count{}, errs.Newf(errs.Internal, "load config by name: %s", err)
	}

	// Convert to business query params
	params := toBusTableQuery(app)

	// Execute the query
	count, err := a.tableStore.FetchTableDataCount(ctx, config, params)
	if err != nil {
		return Count{}, errs.Newf(errs.Internal, "execute query count: %s", err)
	}

	return Count{Count: count}, nil
}

// ValidateConfig validates a configuration without saving it.
func (a *App) ValidateConfig(ctx context.Context, app NewTableConfig) error {
	config, err := toBusNewTableConfig(app)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	// Validate using the config's built-in validation
	if err := config.Validate(); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	return nil
}

// =============================================================================
// Export/Import Methods

// ExportByIDs exports table configs by IDs as a JSON package.
func (a *App) ExportByIDs(ctx context.Context, configIDs []string) (ExportPackage, error) {
	var configs []TableConfig

	for _, idStr := range configIDs {
		id, err := parseUUID(idStr, "config ID")
		if err != nil {
			return ExportPackage{}, errs.New(errs.InvalidArgument, err)
		}

		config, err := a.configStore.QueryByID(ctx, id)
		if err != nil {
			if errors.Is(err, tablebuilder.ErrNotFound) {
				return ExportPackage{}, errs.Newf(errs.NotFound, "config %s not found", idStr)
			}
			return ExportPackage{}, errs.Newf(errs.Internal, "query config %s: %s", idStr, err)
		}

		configs = append(configs, ToAppTableConfig(*config))
	}

	return ExportPackage{
		Version:    "1.0",
		Type:       "table-configs",
		ExportedAt: time.Now().Format(time.RFC3339),
		Count:      len(configs),
		Data:       configs,
	}, nil
}

// ImportTableConfigs imports table configs from a JSON package.
func (a *App) ImportTableConfigs(ctx context.Context, pkg ImportPackage) (ImportResult, error) {
	// Validate package
	if err := pkg.Validate(); err != nil {
		return ImportResult{}, err
	}

	result := ImportResult{}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return ImportResult{}, errs.Newf(errs.Internal, "user missing in context: %s", err)
	}

	for _, config := range pkg.Data {
		// Check if config exists by name
		existing, err := a.configStore.QueryByName(ctx, config.Name)
		existsAlready := err == nil

		switch pkg.Mode {
		case "skip":
			if existsAlready {
				result.SkippedCount++
				continue
			}
			// Create new
			if err := a.createTableConfigFromImport(ctx, config, userID); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("create config %s: %s", config.Name, err))
				continue
			}
			result.ImportedCount++

		case "replace":
			if existsAlready {
				// Delete existing
				if err := a.configStore.Delete(ctx, existing.ID); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("delete existing config %s: %s", config.Name, err))
					continue
				}
				result.UpdatedCount++
			}
			// Create new
			if err := a.createTableConfigFromImport(ctx, config, userID); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("create config %s: %s", config.Name, err))
				continue
			}
			if !existsAlready {
				result.ImportedCount++
			}

		case "merge":
			if existsAlready {
				// Update existing
				if err := a.updateTableConfigFromImport(ctx, existing.ID, config, userID); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("update config %s: %s", config.Name, err))
					continue
				}
				result.UpdatedCount++
			} else {
				// Create new
				if err := a.createTableConfigFromImport(ctx, config, userID); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("create config %s: %s", config.Name, err))
					continue
				}
				result.ImportedCount++
			}
		}
	}

	return result, nil
}

func (a *App) createTableConfigFromImport(ctx context.Context, config TableConfig, userID uuid.UUID) error {
	// Parse and validate the config
	var busConfig tablebuilder.Config
	if err := json.Unmarshal(config.Config, &busConfig); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}

	if err := busConfig.Validate(); err != nil {
		return fmt.Errorf("validate config: %w", err)
	}

	// Create the config
	_, err := a.configStore.Create(ctx, config.Name, config.Description, &busConfig, userID)
	if err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return errors.New("configuration name already exists")
		}
		return fmt.Errorf("create config: %w", err)
	}

	return nil
}

func (a *App) updateTableConfigFromImport(ctx context.Context, id uuid.UUID, config TableConfig, userID uuid.UUID) error {
	// Parse and validate the config
	var busConfig tablebuilder.Config
	if err := json.Unmarshal(config.Config, &busConfig); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}

	if err := busConfig.Validate(); err != nil {
		return fmt.Errorf("validate config: %w", err)
	}

	// Update the config
	_, err := a.configStore.Update(ctx, id, config.Name, config.Description, &busConfig, userID)
	if err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return errors.New("configuration name already exists")
		}
		return fmt.Errorf("update config: %w", err)
	}

	return nil
}

func parseUUID(s string, fieldName string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("invalid %s: %w", fieldName, err)
	}
	return id, nil
}
