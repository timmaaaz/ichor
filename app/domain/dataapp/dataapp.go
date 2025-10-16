// Package data maintains the app layer api for the tablebuilder domain.
package dataapp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/mid"
	"github.com/timmaaaz/ichor/business/sdk/convert"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/business/sdk/tablebuilder"
)

// App manages the set of app layer api functions for the tablebuilder domain.
type App struct {
	configStore *tablebuilder.ConfigStore
	tableStore  *tablebuilder.Store
	auth        *auth.Auth
}

// NewApp constructs a tablebuilder app API for use.
func NewApp(configStore *tablebuilder.ConfigStore, tableStore *tablebuilder.Store) *App {
	return &App{
		configStore: configStore,
		tableStore:  tableStore,
	}
}

// NewAppWithAuth constructs a tablebuilder app API for use with auth support.
func NewAppWithAuth(configStore *tablebuilder.ConfigStore, tableStore *tablebuilder.Store, ath *auth.Auth) *App {
	return &App{
		auth:        ath,
		configStore: configStore,
		tableStore:  tableStore,
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
	config, err := toBusPageConfig(app)
	if err != nil {
		return PageConfig{}, fmt.Errorf("to bus page config: %w", err)
	}

	stored, err := a.configStore.CreatePageConfig(ctx, config)
	if err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return PageConfig{}, errs.New(errs.Aborted, errors.New("page configuration name already exists"))
		}
		return PageConfig{}, fmt.Errorf("create: config[%s]: %w", app.Name, err)
	}

	return toAppPageConfig(*stored), nil
}

// CreatePageTabConfig adds a new page tab configuration to the system.
func (a *App) CreatePageTabConfig(ctx context.Context, app NewPageTabConfig) (PageTabConfig, error) {
	config, err := toBusPageTabConfig(app)
	if err != nil {
		return PageTabConfig{}, fmt.Errorf("to bus page tab config: %w", err)
	}

	stored, err := a.configStore.CreatePageTabConfig(ctx, config)
	if err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return PageTabConfig{}, errs.New(errs.Aborted, errors.New("page tab configuration already exists"))
		}
		return PageTabConfig{}, fmt.Errorf("create: page tab config[%s]: %w", app.Label, err)
	}

	return ToAppPageTabConfig(*stored), nil
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
	upc, err := toBusUpdatePageConfig(app)
	if err != nil {
		return PageConfig{}, fmt.Errorf("updatepageconfig: %w", err)
	}

	// Get current config to preserve unchanged fields
	current, err := a.configStore.QueryPageByID(ctx, id)
	if err != nil {
		if errors.Is(err, tablebuilder.ErrNotFound) {
			return PageConfig{}, errs.New(errs.NotFound, err)
		}
		return PageConfig{}, errs.Newf(errs.Internal, "get config: %s", err)
	}

	err = convert.PopulateSameTypes(upc, current)
	if err != nil {
		return PageConfig{}, fmt.Errorf("populate struct: %w", err)
	}

	// Apply updates
	pageConfig, err := a.configStore.UpdatePageConfig(ctx, *current)
	if err != nil {
		return PageConfig{}, fmt.Errorf("update page config: %w", err)
	}

	return toAppPageConfig(*pageConfig), nil
}

// UpdatePageTabConfig updates an existing page tab configuration.
func (a *App) UpdatePageTabConfig(ctx context.Context, id uuid.UUID, app UpdatePageTabConfig) (PageTabConfig, error) {
	upc, err := toBusUpdatePageTabConfig(app)
	if err != nil {
		return PageTabConfig{}, fmt.Errorf("updatepagetabconfig: %w", err)
	}

	// Get current config to preserve unchanged fields
	current, err := a.configStore.QueryPageTabConfigByID(ctx, id)
	if err != nil {
		if errors.Is(err, tablebuilder.ErrNotFound) {
			return PageTabConfig{}, errs.New(errs.NotFound, err)
		}
		return PageTabConfig{}, errs.Newf(errs.Internal, "get config: %s", err)
	}

	// err = convert.PopulateSameTypes(upc, &current)
	err = convert.PopulateSameTypes(upc, current)
	if err != nil {
		return PageTabConfig{}, fmt.Errorf("populate struct: %w", err)
	}

	// Apply updates
	pageTabConfig, err := a.configStore.UpdatePageTabConfig(ctx, *current)
	if err != nil {
		return PageTabConfig{}, fmt.Errorf("update page tab config: %w", err)
	}

	return ToAppPageTabConfig(*pageTabConfig), nil
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
	if err := a.configStore.DeletePageConfig(ctx, id); err != nil {
		if errors.Is(err, tablebuilder.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "delete page config: %s", err)
	}

	return nil
}

// DeletePageTabConfig removes a page tab configuration from the system.
func (a *App) DeletePageTabConfig(ctx context.Context, id uuid.UUID) error {
	if err := a.configStore.DeletePageTabConfig(ctx, id); err != nil {
		if errors.Is(err, tablebuilder.ErrNotFound) {
			return errs.New(errs.NotFound, err)
		}
		return errs.Newf(errs.Internal, "delete page tab config: %s", err)
	}

	return nil
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

// QueryFullPageByName returns the default page configuration by its name, including tabs.
// This retrieves the default page config which serves as a fallback for all users.
// Only one default page config is allowed per page name (enforced by database constraint).
func (a *App) QueryFullPageByName(ctx context.Context, name string) (FullPageConfig, error) {
	unescaped, err := url.QueryUnescape(name)
	if err != nil {
		return FullPageConfig{}, errs.Newf(errs.InvalidArgument, "invalid page name: %s", err)
	}

	storedPage, err := a.configStore.QueryPageByName(ctx, unescaped)
	if err != nil {
		if errors.Is(err, tablebuilder.ErrNotFound) {
			return FullPageConfig{}, errs.New(errs.NotFound, err)
		}
		return FullPageConfig{}, errs.Newf(errs.Internal, "query page by name: %s", err)
	}
	page := toAppPageConfig(*storedPage)

	storedTabs, err := a.configStore.QueryPageTabConfigsByPageID(ctx, storedPage.ID)
	if err != nil {
		return FullPageConfig{}, errs.Newf(errs.Internal, "query page tabs by page id: %s", err)
	}

	return FullPageConfig{
		PageConfig: page,
		PageTabs:   ToAppPageTabConfigs(storedTabs),
	}, nil
}

// QueryFullPageByNameAndUserID returns a user-specific page configuration by name and user ID, including tabs.
// This retrieves a specific user's customized version of a page (e.g., Jake's version of the orders page).
// Multiple users can have configs with the same page name, but only one per user+name combination.
func (a *App) QueryFullPageByNameAndUserID(ctx context.Context, name string, userID uuid.UUID) (FullPageConfig, error) {
	unescaped, err := url.QueryUnescape(name)
	if err != nil {
		return FullPageConfig{}, errs.Newf(errs.InvalidArgument, "invalid page name: %s", err)
	}

	storedPage, err := a.configStore.QueryPageByNameAndUserID(ctx, unescaped, userID)
	if err != nil {
		if errors.Is(err, tablebuilder.ErrNotFound) {
			return FullPageConfig{}, errs.New(errs.NotFound, err)
		}
		return FullPageConfig{}, errs.Newf(errs.Internal, "query page by name and user id: %s", err)
	}
	page := toAppPageConfig(*storedPage)

	storedTabs, err := a.configStore.QueryPageTabConfigsByPageID(ctx, storedPage.ID)
	if err != nil {
		return FullPageConfig{}, errs.Newf(errs.Internal, "query page tabs by page id: %s", err)
	}

	return FullPageConfig{
		PageConfig: page,
		PageTabs:   ToAppPageTabConfigs(storedTabs),
	}, nil
}

// QueryFullPageByID returns a full page configuration by its ID, including tabs.
func (a *App) QueryFullPageByID(ctx context.Context, id uuid.UUID) (FullPageConfig, error) {
	storedPage, err := a.configStore.QueryPageByID(ctx, id)
	if err != nil {
		if errors.Is(err, tablebuilder.ErrNotFound) {
			return FullPageConfig{}, errs.New(errs.NotFound, err)
		}
		return FullPageConfig{}, errs.Newf(errs.Internal, "query page by id: %s", err)
	}
	page := toAppPageConfig(*storedPage)

	storedTabs, err := a.configStore.QueryPageTabConfigsByPageID(ctx, storedPage.ID)
	if err != nil {
		return FullPageConfig{}, errs.Newf(errs.Internal, "query page tabs by page id: %s", err)
	}

	return FullPageConfig{
		PageConfig: page,
		PageTabs:   ToAppPageTabConfigs(storedTabs),
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
