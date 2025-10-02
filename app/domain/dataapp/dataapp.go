// Package data maintains the app layer api for the tablebuilder domain.
package dataapp

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/mid"
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
