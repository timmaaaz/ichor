package currencydb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/core/currencybus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for currency database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (currencybus.Storer, error) {
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

// Create adds a new currency to the system
func (s *Store) Create(ctx context.Context, c currencybus.Currency) error {
	const q = `
	INSERT INTO core.currencies (
		id, code, name, symbol, locale, decimal_places, is_active, sort_order,
		created_by, created_date, updated_by, updated_date
	) VALUES (
		:id, :code, :name, :symbol, :locale, :decimal_places, :is_active, :sort_order,
		:created_by, :created_date, :updated_by, :updated_date
	)
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBCurrency(c)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", currencybus.ErrUnique)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update modifies a currency in the system
func (s *Store) Update(ctx context.Context, c currencybus.Currency) error {
	const q = `
	UPDATE
		core.currencies
	SET
		code = :code,
		name = :name,
		symbol = :symbol,
		locale = :locale,
		decimal_places = :decimal_places,
		is_active = :is_active,
		sort_order = :sort_order,
		updated_by = :updated_by,
		updated_date = :updated_date
	WHERE
		id = :id
	`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBCurrency(c)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", currencybus.ErrUnique)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a currency from the system
func (s *Store) Delete(ctx context.Context, c currencybus.Currency) error {
	const q = `
	DELETE FROM
		core.currencies
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBCurrency(c)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of currencies from the system
func (s *Store) Query(ctx context.Context, filter currencybus.QueryFilter, orderBy order.By, page page.Page) ([]currencybus.Currency, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		id, code, name, symbol, locale, decimal_places, is_active, sort_order,
		created_by, created_date, updated_by, updated_date
	FROM
		core.currencies`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbCurrencies []currency
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbCurrencies); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusCurrencies(dbCurrencies), nil
}

// Count returns the total number of currencies in the DB.
func (s *Store) Count(ctx context.Context, filter currencybus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		COUNT(1) AS count
	FROM
		core.currencies`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("db: %w", err)
	}

	return count.Count, nil
}

// QueryByID retrieves a single currency from the system by its ID.
func (s *Store) QueryByID(ctx context.Context, currencyID uuid.UUID) (currencybus.Currency, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: currencyID.String(),
	}

	const q = `
	SELECT
		id, code, name, symbol, locale, decimal_places, is_active, sort_order,
		created_by, created_date, updated_by, updated_date
	FROM
		core.currencies
	WHERE
		id = :id`

	var dbCurrency currency
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbCurrency); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return currencybus.Currency{}, fmt.Errorf("db: %w", currencybus.ErrNotFound)
		}
		return currencybus.Currency{}, fmt.Errorf("db: %w", err)
	}

	return toBusCurrency(dbCurrency), nil
}

// QueryByIDs retrieves a list of currencies from the system by their IDs.
func (s *Store) QueryByIDs(ctx context.Context, currencyIDs []uuid.UUID) ([]currencybus.Currency, error) {
	uuidStrings := make([]string, len(currencyIDs))
	for i, id := range currencyIDs {
		uuidStrings[i] = id.String()
	}

	data := struct {
		CurrencyIDs []string `db:"currency_ids"`
	}{
		CurrencyIDs: uuidStrings,
	}

	const q = `
	SELECT
		id, code, name, symbol, locale, decimal_places, is_active, sort_order,
		created_by, created_date, updated_by, updated_date
	FROM
		core.currencies
	WHERE
		id IN (:currency_ids)`

	var dbCurrencies []currency
	if err := sqldb.NamedQuerySliceUsingIn(ctx, s.log, s.db, q, data, &dbCurrencies); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusCurrencies(dbCurrencies), nil
}

// QueryAll retrieves all currencies from the system.
func (s *Store) QueryAll(ctx context.Context) ([]currencybus.Currency, error) {
	const q = `
	SELECT
		id, code, name, symbol, locale, decimal_places, is_active, sort_order,
		created_by, created_date, updated_by, updated_date
	FROM
		core.currencies
	ORDER BY
		sort_order ASC`

	var dbCurrencies []currency
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, struct{}{}, &dbCurrencies); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusCurrencies(dbCurrencies), nil
}
