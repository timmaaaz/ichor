package inventorytransactiondb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorytransactionbus"
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

func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (inventorytransactionbus.Storer, error) {
	ec, err := sqldb.GetExtContext(tx)
	if err != nil {
		return nil, err
	}

	return &Store{
		log: s.log,
		db:  ec,
	}, nil
}

func (s *Store) Create(ctx context.Context, it inventorytransactionbus.InventoryTransaction) error {
	const q = `
	INSERT INTO inventory.inventory_transactions (
		id, product_id, location_id, user_id, transaction_type, reference_number, 
		quantity, transaction_date, created_date, updated_date
	) VALUES (
		:id, :product_id, :location_id, :user_id, :transaction_type, :reference_number, 
        :quantity, :transaction_date, :created_date, :updated_date
    )
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBInventoryTransaction(it)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", inventorytransactionbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", inventorytransactionbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil

}

func (s *Store) Update(ctx context.Context, it inventorytransactionbus.InventoryTransaction) error {
	const q = `
    UPDATE
        inventory.inventory_transactions
    SET
        id = :id,
        product_id = :product_id,
        location_id = :location_id,
        user_id = :user_id,
        transaction_type = :transaction_type,
        reference_number = :reference_number,
        quantity = :quantity,
        transaction_date = :transaction_date,
        updated_date = :updated_date
    WHERE
        id = :id
    `

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBInventoryTransaction(it)); err != nil {
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", inventorytransactionbus.ErrForeignKeyViolation)
		}
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", inventorytransactionbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)

	}

	return nil
}

func (s *Store) Delete(ctx context.Context, transaction inventorytransactionbus.InventoryTransaction) error {
	const q = `
    DELETE FROM
        inventory.inventory_transactions
    WHERE
        id = :id
    `

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBInventoryTransaction(transaction)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

func (s *Store) Query(ctx context.Context, filter inventorytransactionbus.QueryFilter, orderBy order.By, page page.Page) ([]inventorytransactionbus.InventoryTransaction, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
	    id, product_id, location_id, user_id, transaction_type, reference_number, 
        quantity, transaction_date, created_date, updated_date
	FROM 
	    inventory.inventory_transactions
		`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbInvTran []inventoryTransaction

	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbInvTran); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusInventoryTransactions(dbInvTran), nil

}

func (s *Store) Count(ctx context.Context, filter inventorytransactionbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT 
        COUNT(1) AS count
    FROM 
        inventory.inventory_transactions
    `

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("namedqueryrow: %w", err)
	}

	return count.Count, nil
}

func (s *Store) QueryByID(ctx context.Context, transactionID uuid.UUID) (inventorytransactionbus.InventoryTransaction, error) {
	data := struct {
		TransactionID string `db:"id"`
	}{
		TransactionID: transactionID.String(),
	}

	const q = `
    SELECT
        id, product_id, location_id, user_id, transaction_type, reference_number, 
        quantity, transaction_date, created_date, updated_date
    FROM
        inventory.inventory_transactions
    WHERE
        id = :id
    `

	var dbInvTran inventoryTransaction

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbInvTran); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return inventorytransactionbus.InventoryTransaction{}, inventorytransactionbus.ErrNotFound
		}
		return inventorytransactionbus.InventoryTransaction{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusInventoryTransaction(dbInvTran), nil
}
