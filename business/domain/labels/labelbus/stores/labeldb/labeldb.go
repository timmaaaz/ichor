// Package labeldb provides the Postgres implementation of labelbus.Storer.
package labeldb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/labels/labelbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for label_catalog database access.
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

// NewWithTx constructs a new Store value replacing the sqlx DB value with one
// currently inside a transaction.
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (labelbus.Storer, error) {
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

// Create inserts a new label catalog row.
func (s *Store) Create(ctx context.Context, lc labelbus.LabelCatalog) error {
	const q = `
	INSERT INTO inventory.label_catalog (
		id, code, type, entity_ref, payload_json, created_date
	) VALUES (
		:id, :code, :type, :entity_ref, :payload_json, :created_date
	)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBLabelCatalog(lc)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", labelbus.ErrUniqueCode)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Update replaces an existing label catalog row.
func (s *Store) Update(ctx context.Context, lc labelbus.LabelCatalog) error {
	const q = `
	UPDATE
		inventory.label_catalog
	SET
		code = :code,
		type = :type,
		entity_ref = :entity_ref,
		payload_json = :payload_json
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBLabelCatalog(lc)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", labelbus.ErrUniqueCode)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Delete removes a label catalog row.
func (s *Store) Delete(ctx context.Context, lc labelbus.LabelCatalog) error {
	const q = `
	DELETE FROM
		inventory.label_catalog
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBLabelCatalog(lc)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}
	return nil
}

// Query retrieves a page of label catalog rows.
func (s *Store) Query(ctx context.Context, filter labelbus.QueryFilter, orderBy order.By, pg page.Page) ([]labelbus.LabelCatalog, error) {
	data := map[string]any{
		"offset":        (pg.Number() - 1) * pg.RowsPerPage(),
		"rows_per_page": pg.RowsPerPage(),
	}

	const q = `
	SELECT
		id, code, type, entity_ref, payload_json, created_date
	FROM
		inventory.label_catalog`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var rows []labelCatalog
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &rows); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusLabelCatalogs(rows), nil
}

// Count returns the number of label catalog rows matching the filter.
func (s *Store) Count(ctx context.Context, filter labelbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		COUNT(1) AS count
	FROM
		inventory.label_catalog`

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

// QueryByID retrieves a single label catalog row by ID.
func (s *Store) QueryByID(ctx context.Context, id uuid.UUID) (labelbus.LabelCatalog, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: id.String(),
	}

	const q = `
	SELECT
		id, code, type, entity_ref, payload_json, created_date
	FROM
		inventory.label_catalog
	WHERE
		id = :id`

	var row labelCatalog
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &row); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return labelbus.LabelCatalog{}, labelbus.ErrNotFound
		}
		return labelbus.LabelCatalog{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusLabelCatalog(row), nil
}

// QueryByCode retrieves a single label catalog row by its stable code.
func (s *Store) QueryByCode(ctx context.Context, code string) (labelbus.LabelCatalog, error) {
	data := struct {
		Code string `db:"code"`
	}{
		Code: code,
	}

	const q = `
	SELECT
		id, code, type, entity_ref, payload_json, created_date
	FROM
		inventory.label_catalog
	WHERE
		code = :code`

	var row labelCatalog
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &row); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return labelbus.LabelCatalog{}, labelbus.ErrNotFound
		}
		return labelbus.LabelCatalog{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusLabelCatalog(row), nil
}
