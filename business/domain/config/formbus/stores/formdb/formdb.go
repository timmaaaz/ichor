package formdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/config/formbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for forms database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (formbus.Storer, error) {
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

// Create inserts a new form into the database.
func (s *Store) Create(ctx context.Context, form formbus.Form) error {
	const q = `
	INSERT INTO config.forms (
		id, name, is_reference_data, allow_inline_create
	) VALUES (
		:id, :name, :is_reference_data, :allow_inline_create
	)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBForm(form)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", formbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update replaces a form document in the database.
func (s *Store) Update(ctx context.Context, form formbus.Form) error {
	const q = `
	UPDATE
		config.forms
	SET
		name = :name,
		is_reference_data = :is_reference_data,
		allow_inline_create = :allow_inline_create
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBForm(form)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", formbus.ErrUniqueEntry)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a form from the database.
func (s *Store) Delete(ctx context.Context, form formbus.Form) error {
	const q = `
	DELETE FROM
		config.forms
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBForm(form)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of forms from the database.
func (s *Store) Query(ctx context.Context, filter formbus.QueryFilter, orderBy order.By, page page.Page) ([]formbus.Form, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		id, name, is_reference_data, allow_inline_create
	FROM
		config.forms`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbForms []form
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbForms); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusForms(dbForms), nil
}

// Count returns the number of forms in the database.
func (s *Store) Count(ctx context.Context, filter formbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		COUNT(1) AS count
	FROM
		config.forms`

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

// QueryByID retrieves a single form from the database by its ID.
func (s *Store) QueryByID(ctx context.Context, formID uuid.UUID) (formbus.Form, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: formID.String(),
	}

	const q = `
	SELECT
		id, name, is_reference_data, allow_inline_create
	FROM
		config.forms
	WHERE
		id = :id`

	var dbForm form
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbForm); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return formbus.Form{}, formbus.ErrNotFound
		}
		return formbus.Form{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusForm(dbForm), nil
}

// QueryByName retrieves a form by its unique name.
func (s *Store) QueryByName(ctx context.Context, name string) (formbus.Form, error) {
	data := struct {
		Name string `db:"name"`
	}{
		Name: name,
	}

	const q = `
	SELECT
		id, name, is_reference_data, allow_inline_create
	FROM
		config.forms
	WHERE
		name = :name`

	var dbForm form
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbForm); err != nil{
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return formbus.Form{}, formbus.ErrNotFound
		}
		return formbus.Form{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusForm(dbForm), nil
}

// QueryAll retrieves all forms from the database.
func (s *Store) QueryAll(ctx context.Context) ([]formbus.Form, error) {
	data := struct{}{}

	const q = `
	SELECT
		id, name, is_reference_data, allow_inline_create
	FROM
		config.forms
	ORDER BY
		name`

	var dbForms []form
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbForms); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusForms(dbForms), nil
}