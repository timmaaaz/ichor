package branddb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/business/domain/inventory/core/brandbus"
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (brandbus.Storer, error) {
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

// Create inserts a new brand into the database.
func (s *Store) Create(ctx context.Context, brand brandbus.Brand) error {
	const q = `
    INSERT INTO brands (
        brand_id, name, contact_info_id, created_date, updated_date
    ) VALUES (
		:brand_id, :name, :contact_info_id, :created_date, :updated_date
	)
    `

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBBrand(brand)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", brandbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", brandbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Update replaces a brand document in the database.
func (s *Store) Update(ctx context.Context, ass brandbus.Brand) error {
	const q = `
	UPDATE
		brands
	SET
		name = :name,
        contact_info_id = :contact_info_id,
        updated_date = :updated_date
	WHERE
		brand_id = :brand_id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBBrand(ass)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", brandbus.ErrUniqueEntry)
		}
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", brandbus.ErrForeignKeyViolation)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Delete removes an brand from the database.
func (s *Store) Delete(ctx context.Context, brand brandbus.Brand) error {
	const q = `
	DELETE FROM
		brands
	WHERE
		brand_id = :brand_id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBBrand(brand)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of brands from the database.
func (s *Store) Query(ctx context.Context, filter brandbus.QueryFilter, orderBy order.By, page page.Page) ([]brandbus.Brand, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
    SELECT
		brand_id, name, contact_info_id, created_date, updated_date
    FROM
        brands`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var ci []brand
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &ci); err != nil {
		return nil, fmt.Errorf("namedselectcontext: %w", err)
	}

	return toBusBrands(ci), nil
}

// Count returns the number of assets in the database.
func (s *Store) Count(ctx context.Context, filter brandbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM
        brands`

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
func (s *Store) QueryByID(ctx context.Context, userBrandID uuid.UUID) (brandbus.Brand, error) {
	data := struct {
		ID string `db:"brand_id"`
	}{
		ID: userBrandID.String(),
	}

	const q = `
    SELECT
        brand_id, name, contact_info_id, created_date, updated_date
    FROM
        brands
    WHERE
        brand_id = :brand_id
    `
	var ci brand

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &ci); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return brandbus.Brand{}, brandbus.ErrNotFound
		}
		return brandbus.Brand{}, fmt.Errorf("querystruct: %w", err)
	}

	return toBusBrand(ci), nil
}
