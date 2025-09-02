package physicalattributedb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/business/domain/inventory/core/physicalattributebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"

	"github.com/jmoiron/sqlx"
)

// Store manages the set of APIs for product category s database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (physicalattributebus.Storer, error) {
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

// Create inserts a new product category into the database.
func (s *Store) Create(ctx context.Context, pa physicalattributebus.PhysicalAttribute) error {
	const q = `
    INSERT INTO products.physical_attributes (
        id, product_id, length, width, height, weight, weight_unit, color, size,
		material, storage_requirements, hazmat_class, shelf_life_days, created_date, updated_date
    ) VALUES (
		:id, :product_id, :length, :width, :height, :weight, :weight_unit, :color, :size,
		:material, :storage_requirements, :hazmat_class, :shelf_life_days, :created_date, :updated_date
	)
    `

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBPhysicalAttribute(pa)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", physicalattributebus.ErrUniqueEntry)
		} else if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", physicalattributebus.ErrForeignKeyViolation)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Update replaces a product category  in the database.
func (s *Store) Update(ctx context.Context, pa physicalattributebus.PhysicalAttribute) error {
	const q = `
	UPDATE
		products.physical_attributes
	SET
		id = :id,
		product_id = :product_id,
		length = :length,
		width = :width,
		height = :height,
		weight = :weight,
		weight_unit = :weight_unit,
		color = :color,
		size = :size,
		material = :material,
		storage_requirements = :storage_requirements,
		hazmat_class = :hazmat_class,
		shelf_life_days = :shelf_life_days,
		created_date = :created_date,
		updated_date = :updated_date
	WHERE
		id = :id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBPhysicalAttribute(pa)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", physicalattributebus.ErrUniqueEntry)
		} else if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", physicalattributebus.ErrForeignKeyViolation)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Delete removes an product category from the database.
func (s *Store) Delete(ctx context.Context, pa physicalattributebus.PhysicalAttribute) error {
	const q = `
	DELETE FROM
		products.physical_attributes
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBPhysicalAttribute(pa)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of product categories from the database.
func (s *Store) Query(ctx context.Context, filter physicalattributebus.QueryFilter, orderBy order.By, page page.Page) ([]physicalattributebus.PhysicalAttribute, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
    SELECT
		id, product_id, length, width, height, weight, weight_unit, color, size,
		material, storage_requirements, hazmat_class, shelf_life_days, created_date, updated_date
    FROM
        products.physical_attributes`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var pa []physicalAttribute
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &pa); err != nil {
		return nil, fmt.Errorf("namedselectcontext: %w", err)
	}

	return toBusPhysicalAttributes(pa), nil
}

// Count returns the number of product categories in the database.
func (s *Store) Count(ctx context.Context, filter physicalattributebus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM
        products.physical_attributes`

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

// QueryByID retrieves a single product category  from the database by its ID.
func (s *Store) QueryByID(ctx context.Context, userPhysicalAttributeID uuid.UUID) (physicalattributebus.PhysicalAttribute, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: userPhysicalAttributeID.String(),
	}

	const q = `
    SELECT
        id, product_id, length, width, height, weight, weight_unit, color, size,
		material, storage_requirements, hazmat_class, shelf_life_days, created_date, updated_date
    FROM
        products.physical_attributes
    WHERE
        id = :id
    `
	var pa physicalAttribute

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &pa); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return physicalattributebus.PhysicalAttribute{}, physicalattributebus.ErrNotFound
		}
		return physicalattributebus.PhysicalAttribute{}, fmt.Errorf("querystruct: %w", err)
	}

	return toBusPhysicalAttribute(pa), nil
}
