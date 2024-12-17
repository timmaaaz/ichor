package userassetdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/assetbus"
	"github.com/timmaaaz/ichor/business/domain/userassetbus"
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (userassetbus.Storer, error) {
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
func (s *Store) Create(ctx context.Context, ass userassetbus.UserAsset) error {
	const q = `
    INSERT INTO user_assets (
        user_asset_id, asset_id, condition_id, user_id, approved_by, approval_status_id, fulfillment_status_id,
		date_received, last_maintenance
    ) VALUES (
        :user_asset_id, :asset_id,  :condition_id, :user_id, :approved_by, :approval_status_id, :fulfillment_status_id,
		:date_received, :last_maintenance
	)   
    `

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBUserAsset(ass)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", assetbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Update replaces a user asset document in the database.
func (s *Store) Update(ctx context.Context, ass userassetbus.UserAsset) error {
	const q = `
	UPDATE
		user_assets
	SET
		user_asset_id = :user_asset_id,
		asset_id = :asset_id,
		condition_id = :condition_id,
		user_id = :user_id,
        approved_by = :approved_by,
        approval_status_id = :approval_status_id,
        fulfillment_status_id = :fulfillment_status_id,
        date_received = :date_received,
		last_maintenance = :last_maintenance
	WHERE
		user_asset_id = :user_asset_id

	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBUserAsset(ass)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", assetbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Delete removes an user asset from the database.
func (s *Store) Delete(ctx context.Context, ass userassetbus.UserAsset) error {
	const q = `
	DELETE FROM
		user_assets
	WHERE
		user_asset_id = :user_asset_id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBUserAsset(ass)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of user assets from the database.
func (s *Store) Query(ctx context.Context, filter userassetbus.QueryFilter, orderBy order.By, page page.Page) ([]userassetbus.UserAsset, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
    SELECT
        user_asset_id, condition_id, user_id, asset_id, approved_by, approval_status_id, fulfillment_status_id,
		date_received, last_maintenance
    FROM
        user_assets`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var assets []userAsset
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &assets); err != nil {
		return nil, fmt.Errorf("namedselectcontext: %w", err)
	}

	return toBusUserAssets(assets), nil
}

// Count returns the number of assets in the database.
func (s *Store) Count(ctx context.Context, filter userassetbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM
        user_assets`

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
func (s *Store) QueryByID(ctx context.Context, userAssetID uuid.UUID) (userassetbus.UserAsset, error) {
	data := struct {
		ID string `db:"user_asset_id"`
	}{
		ID: userAssetID.String(),
	}

	const q = `
    SELECT
        user_asset_id, condition_id, user_id, asset_id, approved_by, approval_status_id, fulfillment_status_id,
		date_received, last_maintenance
    FROM
        user_assets
    WHERE
        user_asset_id = :user_asset_id
    `
	var ass userAsset

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &ass); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return userassetbus.UserAsset{}, userassetbus.ErrNotFound
		}
		return userassetbus.UserAsset{}, fmt.Errorf("querystruct: %w", err)
	}

	return toBusUserAsset(ass), nil
}
