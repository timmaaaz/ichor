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
