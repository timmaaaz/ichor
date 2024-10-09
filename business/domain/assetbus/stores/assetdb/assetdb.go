package assetdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/assetbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for assets database access.
type Store struct {
	log *logger.Logger
	db  sqlx.ExtContext
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (assetbus.Storer, error) {
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
func (s *Store) Create(ctx context.Context, ass assetbus.Asset) error {
	const q = `
    INSERT INTO assets (
        asset_id, type_id, condition_id, name, est_price, maintenance_interval,
        life_expectancy, model_number, is_enabled, date_created,
        date_updated, created_by, updated_by
    ) VALUES (
        :asset_id, :type_id, :condition_id, :name, :est_price, :maintenance_interval,
        :life_expectancy, :model_number, :is_enabled, :date_created,
        :date_updated, :created_by, :updated_by
    )   
    `

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBAsset(ass)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", assetbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Update replaces an asset document in the database.
func (s *Store) Update(ctx context.Context, ass assetbus.Asset) error {
	const q = `
	UPDATE
		assets
	SET
		asset_id = :asset_id,
		type_id = :type_id,
		condition_id = :condition_id,
		name = :name,
		est_price = :est_price,
		price = :price,
		maintenance_interval = :maintenance_interval,
		life_expectancy = :life_expectancy,
		model_number = :model_number,
		is_enabled = :is_enabled,
		date_created = :date_created,
		date_updated = :date_updated,
		created_by = :created_by,
		updated_by = :updated_by
	WHERE
		asset_id = :asset_id

	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBAsset(ass)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", assetbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Delete removes an asset from the database.
func (s *Store) Delete(ctx context.Context, ass assetbus.Asset) error {
	const q = `
	DELETE FROM
		assets
	WHERE
		asset_id = :asset_id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBAsset(ass)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of assets from the database.
func (s *Store) Query(ctx context.Context, filter assetbus.QueryFilter, orderBy order.By, page page.Page) ([]assetbus.Asset, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
    SELECT
        asset_id, type_id, condition_id, name, est_price, maintenance_interval,
        life_expectancy, model_number, is_enabled, date_created,
        date_updated, created_by, updated_by
    FROM
        assets`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var assets []asset
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &assets); err != nil {
		return nil, fmt.Errorf("namedselectcontext: %w", err)
	}

	return toBusAssets(assets)
}

// Count returns the number of assets in the database.
func (s *Store) Count(ctx context.Context, filter assetbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM
        assets`

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
func (s *Store) QueryByID(ctx context.Context, assetID uuid.UUID) (assetbus.Asset, error) {
	data := struct {
		ID string `db:"asset_id"`
	}{
		ID: assetID.String(),
	}

	const q = `
    SELECT
        asset_id, city_id, line_1, line_2, postal_code
    FROM
        assets
    WHERE
        asset_id = :asset_id
    `
	var ass asset

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &ass); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return assetbus.Asset{}, assetbus.ErrNotFound
		}
		return assetbus.Asset{}, fmt.Errorf("querystruct: %w", err)
	}

	return toBusAsset(ass)
}
