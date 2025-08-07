package inventoryadjustmentdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/movement/inventoryadjustmentbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

type Store struct {
	log *logger.Logger
	db  sqlx.ExtContext
}

func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (inventoryadjustmentbus.Storer, error) {
	ec, err := sqldb.GetExtContext(tx)
	if err != nil {
		return nil, err
	}

	return &Store{
		log: s.log,
		db:  ec,
	}, nil
}

func (s *Store) Create(ctx context.Context, ia inventoryadjustmentbus.InventoryAdjustment) error {
	const q = `
    INSERT INTO inventory_adjustments (
        id, product_id, location_id, adjusted_by, approved_by, quantity_change, reason_code, 
		notes, adjustment_date, created_date, updated_date
    ) VALUES (
        :id, :product_id, :location_id, :adjusted_by, :approved_by, :quantity_change, :reason_code, 
		:notes, :adjustment_date, :created_date, :updated_date
    )
    `

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBInventoryAdjustment(ia)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", inventoryadjustmentbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", inventoryadjustmentbus.ErrUniqueEntry)
		}
		return err
	}

	return nil
}

func (s *Store) Update(ctx context.Context, ia inventoryadjustmentbus.InventoryAdjustment) error {
	const q = `
    UPDATE
        inventory_adjustments
    SET
        id = :id,
        product_id = :product_id,
        location_id = :location_id,
        adjusted_by = :adjusted_by,
        approved_by = :approved_by,
        quantity_change = :quantity_change,
        reason_code = :reason_code,
        notes = :notes,
        adjustment_date = :adjustment_date,
        updated_date = :updated_date
    WHERE
        id = :id
    `

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBInventoryAdjustment(ia)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", inventoryadjustmentbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", inventoryadjustmentbus.ErrUniqueEntry)
		}
		return err
	}

	return nil
}

func (s *Store) Delete(ctx context.Context, inventoryAdjument inventoryadjustmentbus.InventoryAdjustment) error {
	const q = `
		DELETE FROM
		    inventory_adjustments
		WHERE
			id = :id
		`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBInventoryAdjustment(inventoryAdjument)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

func (s *Store) Query(ctx context.Context, filter inventoryadjustmentbus.QueryFilter, orderBy order.By, page page.Page) ([]inventoryadjustmentbus.InventoryAdjustment, error) {
	data := map[string]interface{}{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
	    id, product_id, location_id, adjusted_by, approved_by, quantity_change, reason_code, 
        notes, adjustment_date, created_date, updated_date
	FROM
		inventory_adjustments`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbInvAdj []inventoryAdjustment
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbInvAdj); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusInventoryAdjustments(dbInvAdj), nil
}

func (s *Store) Count(ctx context.Context, filter inventoryadjustmentbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM
        inventory_adjustments`

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

func (s *Store) QueryByID(ctx context.Context, adjustmentID uuid.UUID) (inventoryadjustmentbus.InventoryAdjustment, error) {
	data := struct {
		AdjustmentID string `db:"id"`
	}{
		AdjustmentID: adjustmentID.String(),
	}

	const q = `
	SELECT
	    id, product_id, location_id, adjusted_by, approved_by, quantity_change, reason_code, 
        notes, adjustment_date, created_date, updated_date
	FROM 
		inventory_adjustments
	WHERE
		id = :id
    `

	var dbInvAdj inventoryAdjustment
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbInvAdj); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return inventoryadjustmentbus.InventoryAdjustment{}, inventoryadjustmentbus.ErrNotFound
		}
		return inventoryadjustmentbus.InventoryAdjustment{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusInventoryAdjustment(dbInvAdj), nil
}
