package transferorderdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
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

func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (transferorderbus.Storer, error) {
	ec, err := sqldb.GetExtContext(tx)
	if err != nil {
		return nil, err
	}

	return &Store{
		log: s.log,
		db:  ec,
	}, nil
}

func (s *Store) Create(ctx context.Context, transferOrder transferorderbus.TransferOrder) error {
	const q = `
	INSERT INTO inventory.transfer_orders (
	    id, product_id, from_location_id, to_location_id, requested_by, 
		approved_by, quantity, status, transfer_date, created_date, updated_date
    ) VALUES (
        :id, :product_id, :from_location_id, :to_location_id, :requested_by, 
        :approved_by, :quantity, :status, :transfer_date, :created_date, :updated_date
    )
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBTransferOrder(transferOrder)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext %w", transferorderbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext %w", transferorderbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext %w", err)
	}

	return nil

}

func (s *Store) Update(ctx context.Context, transferOrder transferorderbus.TransferOrder) error {
	const q = `
    UPDATE
        inventory.transfer_orders
    SET
        product_id = :product_id, 
		from_location_id = :from_location_id, 
		to_location_id = :to_location_id, 
        requested_by = :requested_by, 
		approved_by = :approved_by, 
		quantity = :quantity, 
        status = :status, 
		transfer_date = :transfer_date, 
		updated_date = :updated_date
    WHERE
        id = :id
    `

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBTransferOrder(transferOrder)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext %w", transferorderbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext %w", transferorderbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext %w", err)
	}

	return nil
}

func (s *Store) Delete(ctx context.Context, transferOrder transferorderbus.TransferOrder) error {
	const q = `
    DELETE FROM inventory.transfer_orders
    WHERE id = :id
    `

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBTransferOrder(transferOrder)); err != nil {
		return fmt.Errorf("namedexeccontext %w", err)
	}

	return nil
}

func (s *Store) Query(ctx context.Context, filter transferorderbus.QueryFilter, orderBy order.By, page page.Page) ([]transferorderbus.TransferOrder, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		id, product_id, from_location_id, to_location_id, requested_by, approved_by, 
		quantity, status, transfer_date, created_date, updated_date
    FROM
	    inventory.transfer_orders
		`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbTO []transferOrder
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbTO); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusTransferOrders(dbTO), nil
}

func (s *Store) Count(ctx context.Context, filter transferorderbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM
        inventory.transfer_orders
    `

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &count); err != nil {
		return 0, fmt.Errorf("namedqueryrow: %w", err)
	}

	return count.Count, nil
}

func (s *Store) QueryByID(ctx context.Context, transferOrderID uuid.UUID) (transferorderbus.TransferOrder, error) {
	data := struct {
		TransferID string `db:"id"`
	}{
		TransferID: transferOrderID.String(),
	}

	const q = `
    SELECT
        id, product_id, from_location_id, to_location_id, requested_by, approved_by, 
        quantity, status, transfer_date, created_date, updated_date
    FROM
        inventory.transfer_orders
    WHERE
        id = :id
    `

	var dbTO transferOrder
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbTO); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return transferorderbus.TransferOrder{}, transferorderbus.ErrNotFound
		}
		return transferorderbus.TransferOrder{}, fmt.Errorf("namedqueryrow: %w", err)
	}

	return toBusTransferOrder(dbTO), nil
}
