package assetconditiondb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/assetconditionbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for asset condition database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (assetconditionbus.Storer, error) {
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

// Create inserts a new asset condition into the database.
func (s *Store) Create(ctx context.Context, ac assetconditionbus.AssetCondition) error {
	const q = `
    INSERT INTO asset_condition (
        asset_condition_id, name
    ) VALUES (
        :asset_condition_id, :name
    )
    `
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBAssetCondition(ac)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", assetconditionbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update replaces an asset condition in the database
func (s *Store) Update(ctx context.Context, ac assetconditionbus.AssetCondition) error {
	const q = `
	UPDATE asset_condition
	SET 
        name = :name
	WHERE 
		asset_condition_id = :asset_condition_id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBAssetCondition(ac)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", assetconditionbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes an asset condition from the database.
func (s *Store) Delete(ctx context.Context, as assetconditionbus.AssetCondition) error {
	const q = `
	DELETE FROM
		asset_condition
	WHERE
		asset_condition_id = :asset_condition_id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBAssetCondition(as)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of asset conditions from the database.
func (s *Store) Query(ctx context.Context, filter assetconditionbus.QueryFilter, orderBy order.By, page page.Page) ([]assetconditionbus.AssetCondition, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT 
		asset_condition_id, name
	FROM
		asset_condition
	`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var assetCondition []assetCondition

	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &assetCondition); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusAssetConditions(assetCondition), nil
}

// Count returns the total number of asset conditions
func (s *Store) Count(ctx context.Context, filter assetconditionbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM
        asset_condition`

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

// QueryByID finds the asset condition by the specified ID.
func (s *Store) QueryByID(ctx context.Context, aprvlStatusID uuid.UUID) (assetconditionbus.AssetCondition, error) {
	data := struct {
		ID string `db:"asset_condition_id"`
	}{
		ID: aprvlStatusID.String(),
	}

	const q = `
    SELECT
        asset_condition_id, name
    FROM
        asset_condition
    WHERE
        asset_condition_id = :asset_condition_id
    `

	var assetCondition assetCondition
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &assetCondition); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return assetconditionbus.AssetCondition{}, fmt.Errorf("db: %w", assetconditionbus.ErrNotFound)
		}
		return assetconditionbus.AssetCondition{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusAssetCondition(assetCondition), nil
}
