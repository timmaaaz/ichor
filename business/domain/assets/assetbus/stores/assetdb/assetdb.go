package assetdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/assets/assetbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"

	"github.com/jmoiron/sqlx"
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

// Create inserts a new user asset into the database.
func (s *Store) Create(ctx context.Context, ass assetbus.Asset) error {
	const q = `
    INSERT INTO assets.assets (
        id, valid_asset_id, last_maintenance_time, serial_number, asset_condition_id
    ) VALUES (
		:id, :valid_asset_id, :last_maintenance_time, :serial_number, :asset_condition_id
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

// Update replaces a user asset document in the database.
func (s *Store) Update(ctx context.Context, ass assetbus.Asset) error {
	const q = `
	UPDATE
		assets.assets
	SET
		id = :id,
		valid_asset_id = :valid_asset_id,
        last_maintenance_time = :last_maintenance_time,
        serial_number = :serial_number,
        asset_condition_id = :asset_condition_id
	WHERE
		id = :id

	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBAsset(ass)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", assetbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Delete removes an user asset from the database.
func (s *Store) Delete(ctx context.Context, ass assetbus.Asset) error {
	const q = `
	DELETE FROM
		assets.assets
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBAsset(ass)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of user assets from the database.
func (s *Store) Query(ctx context.Context, filter assetbus.QueryFilter, orderBy order.By, page page.Page) ([]assetbus.Asset, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
    SELECT
		id, valid_asset_id, last_maintenance_time, serial_number, asset_condition_id
    FROM
        assets.assets`

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

	return toBusAssets(assets), nil
}

// Count returns the number of assets in the database.
func (s *Store) Count(ctx context.Context, filter assetbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM
        assets.assets`

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
func (s *Store) QueryByID(ctx context.Context, userAssetID uuid.UUID) (assetbus.Asset, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: userAssetID.String(),
	}

	const q = `
    SELECT
        id, valid_asset_id, asset_condition_id, serial_number, last_maintenance_time
    FROM
        assets.assets
    WHERE
        id = :id
    `
	var ass asset

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &ass); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return assetbus.Asset{}, assetbus.ErrNotFound
		}
		return assetbus.Asset{}, fmt.Errorf("querystruct: %w", err)
	}

	return toBusAsset(ass), nil
}
