package pageconfigdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/config/pageconfigbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for page config database access.
type Store struct {
	log *logger.Logger
	db  sqlx.ExtContext
}

// NewStore constructs the API for data access.
func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// NewWithTx constructs a new Store value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (pageconfigbus.Storer, error) {
	ec, err := sqldb.GetExtContext(tx)
	if err != nil {
		return nil, err
	}

	store := Store{
		log: s.log,
		db:  ec,
	}

	return &store, nil
}

// Create inserts a new page configuration into the database.
func (s *Store) Create(ctx context.Context, config pageconfigbus.PageConfig) error {
	const q = `
		INSERT INTO config.page_configs (
			id, name, user_id, is_default
		) VALUES (
			:id, :name, :user_id, :is_default
		)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBPageConfig(config)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: page config already exists: %w", err)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update modifies an existing page configuration in the database.
func (s *Store) Update(ctx context.Context, config pageconfigbus.PageConfig) error {
	const q = `
		UPDATE config.page_configs SET
			name = :name,
			user_id = :user_id,
			is_default = :is_default
		WHERE id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBPageConfig(config)); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return fmt.Errorf("namedexeccontext: %w", pageconfigbus.ErrNotFound)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a page configuration from the database.
func (s *Store) Delete(ctx context.Context, configID uuid.UUID) error {
	data := struct {
		ID uuid.UUID `db:"id"`
	}{
		ID: configID,
	}

	const q = `DELETE FROM config.page_configs WHERE id = :id`

	rowsAffected, err := sqldb.NamedExecContextWithCount(ctx, s.log, s.db, q, data)
	if err != nil {
		return fmt.Errorf("namedexeccontextwithcount: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("namedexeccontextwithcount: %w", pageconfigbus.ErrNotFound)
	}

	return nil
}

// Query retrieves a list of page configurations based on filters.
func (s *Store) Query(ctx context.Context, filter pageconfigbus.QueryFilter, orderBy order.By, pageReq page.Page) ([]pageconfigbus.PageConfig, error) {
	data := map[string]any{
		"offset":        (pageReq.Number() - 1) * pageReq.RowsPerPage(),
		"rows_per_page": pageReq.RowsPerPage(),
	}

	const q = `
	SELECT
		id, name, user_id, is_default
	FROM
		config.page_configs`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbConfigs []dbPageConfig
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbConfigs); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusPageConfigs(dbConfigs), nil
}

// Count returns the total number of page configurations matching the filter.
func (s *Store) Count(ctx context.Context, filter pageconfigbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		COUNT(1) AS count
	FROM
		config.page_configs`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("namedquerystruct: %w", err)
	}

	return count.Count, nil
}

// QueryByID retrieves a single page configuration by ID.
func (s *Store) QueryByID(ctx context.Context, configID uuid.UUID) (pageconfigbus.PageConfig, error) {
	data := struct {
		ID uuid.UUID `db:"id"`
	}{
		ID: configID,
	}

	const q = `
		SELECT
			id, name, user_id, is_default
		FROM
			config.page_configs
		WHERE
			id = :id`

	var dbConfig dbPageConfig
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbConfig); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return pageconfigbus.PageConfig{}, fmt.Errorf("namedquerystruct: %w", pageconfigbus.ErrNotFound)
		}
		return pageconfigbus.PageConfig{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusPageConfig(dbConfig), nil
}

// QueryByName retrieves the default page configuration by name.
// This returns the default page config that serves as a fallback for all users.
func (s *Store) QueryByName(ctx context.Context, name string) (pageconfigbus.PageConfig, error) {
	data := struct {
		Name      string `db:"name"`
		IsDefault bool   `db:"is_default"`
	}{
		Name:      name,
		IsDefault: true,
	}

	const q = `
		SELECT
			id, name, user_id, is_default
		FROM
			config.page_configs
		WHERE
			name = :name
			AND is_default = :is_default`

	var dbConfig dbPageConfig
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbConfig); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return pageconfigbus.PageConfig{}, fmt.Errorf("namedquerystruct: %w", pageconfigbus.ErrNotFound)
		}
		return pageconfigbus.PageConfig{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusPageConfig(dbConfig), nil
}

// QueryByNameAndUserID retrieves a page configuration by name and user ID
func (s *Store) QueryByNameAndUserID(ctx context.Context, name string, userID uuid.UUID) (pageconfigbus.PageConfig, error) {
	data := struct {
		Name   string    `db:"name"`
		UserID uuid.UUID `db:"user_id"`
	}{
		Name:   name,
		UserID: userID,
	}

	const q = `
		SELECT
			id, name, user_id, is_default
		FROM
			config.page_configs
		WHERE
			name = :name
			AND user_id = :user_id`

	var dbConfig dbPageConfig
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbConfig); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return pageconfigbus.PageConfig{}, fmt.Errorf("namedquerystruct: %w", pageconfigbus.ErrNotFound)
		}
		return pageconfigbus.PageConfig{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusPageConfig(dbConfig), nil
}

// QueryAll retrieves all page configurations from the database.
func (s *Store) QueryAll(ctx context.Context) ([]pageconfigbus.PageConfig, error) {
	data := struct{}{}

	const q = `
	SELECT
		id, name, user_id, is_default
	FROM
		config.page_configs
	ORDER BY
		name`

	var dbConfigs []dbPageConfig
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbConfigs); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusPageConfigs(dbConfigs), nil
}
