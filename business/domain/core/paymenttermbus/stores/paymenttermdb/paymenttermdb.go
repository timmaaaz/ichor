package paymenttermdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/core/paymenttermbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for payment terms database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (paymenttermbus.Storer, error) {
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

// Create inserts a new payment term into the database.
func (s *Store) Create(ctx context.Context, pt paymenttermbus.PaymentTerm) error {
	const q = `
    INSERT INTO core.payment_terms (
        id, name, description
    ) VALUES (
        :id, :name, :description
    )
    `
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBPaymentTerm(pt)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", paymenttermbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update modifies data about a payment term in the database.
func (s *Store) Update(ctx context.Context, pt paymenttermbus.PaymentTerm) error {
	const q = `
    UPDATE
        core.payment_terms
    SET
        name = :name,
        description = :description
    WHERE
        id = :id
    `
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBPaymentTerm(pt)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", paymenttermbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a payment term from the database.
func (s *Store) Delete(ctx context.Context, pt paymenttermbus.PaymentTerm) error {
	const q = `
    DELETE FROM
        core.payment_terms
    WHERE
        id = :id
    `
	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBPaymentTerm(pt)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of existing payment terms from the database.
func (s *Store) Query(ctx context.Context, filter paymenttermbus.QueryFilter, orderBy order.By, page page.Page) ([]paymenttermbus.PaymentTerm, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
    SELECT
        id, name, description
    FROM
        core.payment_terms`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbPts []paymentTerm
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbPts); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusPaymentTerms(dbPts), nil
}

// Count returns the total number of payment terms in the DB.
func (s *Store) Count(ctx context.Context, filter paymenttermbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
    SELECT
        COUNT(1) AS count
    FROM
        core.payment_terms`

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

// QueryByID retrieves a single payment term by its id.
func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (paymenttermbus.PaymentTerm, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: id.String(),
	}

	const q = `
    SELECT
        id, name, description
    FROM
        core.payment_terms
    WHERE
        id = :id
    `

	var dbPt paymentTerm
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbPt); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return paymenttermbus.PaymentTerm{}, fmt.Errorf("db: %w", paymenttermbus.ErrNotFound)
		}
		return paymenttermbus.PaymentTerm{}, fmt.Errorf("querystruct: %w", err)
	}

	return toBusPaymentTerm(dbPt), nil
}
