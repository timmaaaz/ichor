package validassetdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/assets/validassetbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for assets database access.
type Store struct {
	log          *logger.Logger
	db           sqlx.ExtContext
	columnFilter *sqldb.ColumnFilter
}

// NewStore constructs the api for data access.
func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// NewWithTx constructs a new Store value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (validassetbus.Storer, error) {
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

// Create inserts a new asset into the database.
func (s *Store) Create(ctx context.Context, ass validassetbus.ValidAsset) error {
	const q = `
    INSERT INTO assets.valid_assets (
        id, type_id, name, est_price, maintenance_interval,
        life_expectancy, serial_number, model_number, is_enabled, created_date,
        updated_date, created_by, updated_by
    ) VALUES (
        :id, :type_id, :name, :est_price, :maintenance_interval,
        :life_expectancy, :serial_number, :model_number, :is_enabled, :created_date,
        :updated_date, :created_by, :updated_by
    )   
    `

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBAsset(ass)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", validassetbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Update replaces an asset document in the database.
func (s *Store) Update(ctx context.Context, ass validassetbus.ValidAsset) error {
	const q = `
	UPDATE
		assets.valid_assets
	SET
		id = :id,
		type_id = :type_id,
		name = :name,
		est_price = :est_price,
		price = :price,
		maintenance_interval = :maintenance_interval,
		life_expectancy = :life_expectancy,
		serial_number = :serial_number,
		model_number = :model_number,
		is_enabled = :is_enabled,
		created_date = :created_date,
		updated_date = :updated_date,
		created_by = :created_by,
		updated_by = :updated_by
	WHERE
		id = :id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBAsset(ass)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", validassetbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Delete removes an asset from the database.
func (s *Store) Delete(ctx context.Context, ass validassetbus.ValidAsset) error {
	const q = `
	DELETE FROM
		assets.valid_assets
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBAsset(ass)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of assets from the database.
func (s *Store) Query(ctx context.Context, filter validassetbus.QueryFilter, orderBy order.By, page page.Page) ([]validassetbus.ValidAsset, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		id,
		type_id,
		name,
		est_price,
		price,
		maintenance_interval,
		life_expectancy,
		serial_number,
		model_number,
		is_enabled,
		created_date,
		updated_date,
		created_by,
		updated_by
	FROM
		assets.valid_assets
	`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var assets []validAsset
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &assets); err != nil {
		return nil, fmt.Errorf("namedselectcontext: %w", err)
	}

	return toBusAssets(assets)
}

// Count returns the number of assets in the database.
func (s *Store) Count(ctx context.Context, filter validassetbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM
        assets.valid_assets`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("namedqueryint: %w", err)
	}

	return count.Count, nil
}

// QueryByID retrieves a single asset from the database by its ID.
func (s *Store) QueryByID(ctx context.Context, assetID uuid.UUID) (validassetbus.ValidAsset, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: assetID.String(),
	}

	const q = `
    SELECT
		id,
		type_id,
		name,
		est_price,
		price,
		maintenance_interval,
		life_expectancy,
		serial_number,
		model_number,
		is_enabled,
		created_date,
		updated_date,
		created_by,
		updated_by
    FROM
        assets.valid_assets
    WHERE
        id = :id`

	var ass validAsset

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &ass); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return validassetbus.ValidAsset{}, validassetbus.ErrNotFound
		}
		return validassetbus.ValidAsset{}, fmt.Errorf("querystruct: %w", err)
	}

	return toBusAsset(ass)
}
