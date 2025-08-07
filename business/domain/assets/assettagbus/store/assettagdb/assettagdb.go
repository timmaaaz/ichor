package assettagdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/assets/assettagbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for streets database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (assettagbus.Storer, error) {
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

// Create inserts a new asset tag into the database.
func (s *Store) Create(ctx context.Context, t assettagbus.AssetTag) error {
	const q = `
    INSERT INTO asset_tags (
        id, valid_asset_id, tag_id
    ) VALUES (
        :id, :valid_asset_id, :tag_id
    )
    `
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBAssetTag(t)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", assettagbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update modifies data about an asset tag in the database.
func (s *Store) Update(ctx context.Context, t assettagbus.AssetTag) error {
	const q = `
    UPDATE 
        asset_tags
    SET
        valid_asset_id = :valid_asset_id,
        tag_id = :tag_id
    WHERE
        id = :id
    `
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBAssetTag(t)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", assettagbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes an tag from the database.
func (s *Store) Delete(ctx context.Context, at assettagbus.AssetTag) error {
	const q = `
    DELETE FROM
        asset_tags
    WHERE
        id = :id
    `
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBAssetTag(at)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of existing asset_tags from the database.
func (s *Store) Query(ctx context.Context, filter assettagbus.QueryFilter, orderBy order.By, page page.Page) ([]assettagbus.AssetTag, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
    SELECT
        id, valid_asset_id, tag_id
    FROM
        asset_tags `

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbATs []assetTag
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbATs); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusAssetTags(dbATs), nil
}

// Count returns the total number of asset_tags in the DB.
func (s *Store) Count(ctx context.Context, filter assettagbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM
        asset_tags`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("namedquerysingle: %w", err)
	}

	return count.Count, nil
}

// QueryByID retrieves a single tag by its id.
func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (assettagbus.AssetTag, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: id.String(),
	}

	const q = `
    SELECT
        id, valid_asset_id, tag_id
    FROM
        asset_tags
    WHERE
        id = :id
    `

	var dbAT assetTag
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbAT); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return assettagbus.AssetTag{}, fmt.Errorf("db: %w", assettagbus.ErrNotFound)
		}
		return assettagbus.AssetTag{}, fmt.Errorf("querystruct: %w", err)
	}

	return toBusAssetTag(dbAT), nil
}
