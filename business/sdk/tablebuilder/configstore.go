package tablebuilder

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// ConfigStore manages table configuration storage in the database
type ConfigStore struct {
	log *logger.Logger
	db  sqlx.ExtContext
}

// NewConfigStore creates a new configuration store
func NewConfigStore(log *logger.Logger, db *sqlx.DB) *ConfigStore {
	return &ConfigStore{
		log: log,
		db:  db,
	}
}

// Create saves a new table configuration
func (s *ConfigStore) Create(ctx context.Context, name, description string, config *Config, userID uuid.UUID) (*StoredConfig, error) {
	// Validate the configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	// Marshal config to JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("marshal config: %w", err)
	}

	// Create stored config
	stored := StoredConfig{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Config:      configJSON,
		CreatedBy:   userID,
		UpdatedBy:   userID,
		CreatedDate: time.Now().UTC(),
		UpdatedDate: time.Now().UTC(),
	}

	// Insert into database
	const q = `
		INSERT INTO config.table_configs (
			id, name, description, config,
			created_by, updated_by, created_date, updated_date
		) VALUES (
			:id, :name, :description, :config,
			:created_by, :updated_by, :created_date, :updated_date
		)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, stored); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return nil, fmt.Errorf("configuration name already exists: %w", err)
		}
		return nil, fmt.Errorf("insert config: %w", err)
	}

	return &stored, nil
}

// CreatePageConfig saves a new page configuration.
func (s *ConfigStore) CreatePageConfig(ctx context.Context, pc PageConfig) (*PageConfig, error) {
	pc.ID = uuid.New()

	const q = `
		INSERT INTO config.page_configs (
			id, page_id, user_id, is_default
		) VALUES (
			:id, :page_id, :user_id, :is_default
		)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, pc); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return nil, fmt.Errorf("page config already exists: %w", err)
		}
		return nil, fmt.Errorf("insert page config: %w", err)
	}

	return &pc, nil
}

// CreatePageTabConfig saves a new page tab configuration.
func (s *ConfigStore) CreatePageTabConfig(ctx context.Context, ptc PageTabConfig) (*PageTabConfig, error) {

	ptc.ID = uuid.New()

	const q = `
		INSERT INTO config.page_tab_configs (
			id, page_id, label, config_id, is_default, tab_order
		) VALUES (
			:id, :page_id, :label, :config_id, :is_default, :tab_order
		)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, ptc); err != nil {
		return nil, fmt.Errorf("insert page tab config: %w", err)
	}

	return &ptc, nil
}

// Update updates an existing table configuration
func (s *ConfigStore) Update(ctx context.Context, id uuid.UUID, name, description string, config *Config, userID uuid.UUID) (*StoredConfig, error) {
	// Validate the configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	// Marshal config to JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("marshal config: %w", err)
	}

	// Update stored config
	stored := StoredConfig{
		ID:          id,
		Name:        name,
		Description: description,
		Config:      configJSON,
		UpdatedBy:   userID,
		UpdatedDate: time.Now(),
	}

	const q = `
		UPDATE config.table_configs SET
			name = :name,
			description = :description,
			config = :config,
			updated_by = :updated_by,
			updated_date = :updated_date
		WHERE id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, stored); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return nil, fmt.Errorf("configuration name already exists: %w", err)
		}
		return nil, fmt.Errorf("update config: %w", err)
	}

	// Fetch and return the updated config
	return s.QueryByID(ctx, id)
}

// UpdatePageConfig updates an existing page configuration.
func (s *ConfigStore) UpdatePageConfig(ctx context.Context, pc PageConfig) (*PageConfig, error) {
	const q = `
		UPDATE config.page_configs SET
			page_id = :page_id,
			user_id = :user_id,
			is_default = :is_default
		WHERE id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, pc); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update page config: %w", err)
	}

	return &pc, nil
}

// UpdatePageTabConfig updates an existing page tab configuration.
func (s *ConfigStore) UpdatePageTabConfig(ctx context.Context, ptc PageTabConfig) (*PageTabConfig, error) {
	const q = `
		UPDATE config.page_tab_configs SET
			page_id = :page_id,
			label = :label,
			config_id = :config_id,
			is_default = :is_default,
			tab_order = :tab_order
		WHERE id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, ptc); err != nil {
		return nil, fmt.Errorf("update page tab config: %w", err)
	}

	return &ptc, nil
}

// Delete removes a table configuration
func (s *ConfigStore) Delete(ctx context.Context, id uuid.UUID) error {
	data := struct {
		ID uuid.UUID `db:"id"`
	}{
		ID: id,
	}

	const q = `DELETE FROM config.table_configs WHERE id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("delete config: %w", err)
	}

	return nil
}

// DeletePageConfig removes a page configuration.
func (s *ConfigStore) DeletePageConfig(ctx context.Context, id uuid.UUID) error {
	data := struct {
		ID uuid.UUID `db:"id"`
	}{
		ID: id,
	}

	const q = `DELETE FROM config.page_configs WHERE id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("delete page config: %w", err)
	}

	return nil
}

// DeletePageTabConfig removes a page tab configuration.
func (s *ConfigStore) DeletePageTabConfig(ctx context.Context, id uuid.UUID) error {
	data := struct {
		ID uuid.UUID `db:"id"`
	}{
		ID: id,
	}

	const q = `DELETE FROM config.page_tab_configs WHERE id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("delete page tab config: %w", err)
	}

	return nil
}

// QueryByID retrieves a configuration by ID
func (s *ConfigStore) QueryByID(ctx context.Context, id uuid.UUID) (*StoredConfig, error) {
	data := struct {
		ID uuid.UUID `db:"id"`
	}{
		ID: id,
	}

	const q = `
		SELECT
			id, name, description, config,
			created_by, updated_by, created_date, updated_date
		FROM
			config.table_configs
		WHERE
			id = :id`

	var stored StoredConfig
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &stored); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query config: %w", err)
	}

	return &stored, nil
}

// QueryByName retrieves a configuration by name
func (s *ConfigStore) QueryByName(ctx context.Context, name string) (*StoredConfig, error) {
	data := struct {
		Name string `db:"name"`
	}{
		Name: name,
	}

	const q = `
		SELECT
			id, name, description, config,
			created_by, updated_by, created_date, updated_date
		FROM
			config.table_configs
		WHERE
			name = :name`

	var stored StoredConfig
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &stored); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query config: %w", err)
	}

	return &stored, nil
}

// QueryByUser retrieves all configurations created by a user
func (s *ConfigStore) QueryByUser(ctx context.Context, userID uuid.UUID) ([]StoredConfig, error) {
	data := struct {
		UserID uuid.UUID `db:"user_id"`
	}{
		UserID: userID,
	}

	const q = `
		SELECT
			id, name, description, config,
			created_by, updated_by, created_date, updated_date
		FROM
			config.table_configs
		WHERE
			created_by = :user_id
		ORDER BY
			updated_date DESC`

	var configs []StoredConfig
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &configs); err != nil {
		return nil, fmt.Errorf("query configs: %w", err)
	}

	return configs, nil
}

// QueryPageByName retrieves a page configuration by name
func (s *ConfigStore) QueryPageByName(ctx context.Context, name string) (*PageConfig, error) {
	data := struct {
		Name string `db:"name"`
	}{
		Name: name,
	}

	const q = `
		SELECT
			id, name, user_id, is_default
		FROM
			config.page_configs
		WHERE
			name = :name`

	var pc PageConfig
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &pc); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query page config: %w", err)
	}

	return &pc, nil
}

// QueryPageByID retrieves a page configuration by ID
func (s *ConfigStore) QueryPageByID(ctx context.Context, id uuid.UUID) (*PageConfig, error) {
	data := struct {
		ID uuid.UUID `db:"id"`
	}{
		ID: id,
	}

	const q = `
		SELECT
			id, name, user_id, is_default
		FROM
			config.page_configs
		WHERE
			id = :id`

	var pc PageConfig
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &pc); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query page config: %w", err)
	}

	return &pc, nil
}

// QueryPageTabConfigByID retrieves a page tab configuration by ID
func (s *ConfigStore) QueryPageTabConfigByID(ctx context.Context, id uuid.UUID) (*PageTabConfig, error) {
	data := struct {
		ID uuid.UUID `db:"id"`
	}{
		ID: id,
	}

	const q = `
		SELECT
			id, page_id, label, config_id, is_default, tab_order
		FROM
			config.page_tab_configs
		WHERE
			id = :id`

	var ptc PageTabConfig
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &ptc); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query page tab config: %w", err)
	}

	return &ptc, nil
}

// QueryPageTabConfigsByPageID retrieves all page tab configurations for a given page ID
func (s *ConfigStore) QueryPageTabConfigsByPageID(ctx context.Context, pageID uuid.UUID) ([]PageTabConfig, error) {
	data := struct {
		PageID uuid.UUID `db:"page_id"`
	}{
		PageID: pageID,
	}

	const q = `
		SELECT
			id, page_id, label, config_id, is_default, tab_order
		FROM
			config.page_tab_configs
		WHERE
			page_id = :page_id
		ORDER BY
			tab_order ASC`

	var ptcs []PageTabConfig
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &ptcs); err != nil {
		return nil, fmt.Errorf("query page tab configs: %w", err)
	}

	return ptcs, nil
}

// LoadConfig loads a configuration and returns the parsed Config
func (s *ConfigStore) LoadConfig(ctx context.Context, id uuid.UUID) (*Config, error) {
	stored, err := s.QueryByID(ctx, id)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(stored.Config, &config); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &config, nil
}

// LoadConfigByName loads a configuration by name and returns the parsed Config
func (s *ConfigStore) LoadConfigByName(ctx context.Context, name string) (*Config, error) {
	stored, err := s.QueryByName(ctx, name)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(stored.Config, &config); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &config, nil
}

// ValidateStoredConfig validates a stored configuration
func (s *ConfigStore) ValidateStoredConfig(ctx context.Context, id uuid.UUID) error {
	config, err := s.LoadConfig(ctx, id)
	if err != nil {
		return err
	}

	return config.Validate()
}
