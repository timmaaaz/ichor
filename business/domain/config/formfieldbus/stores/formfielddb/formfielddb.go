package formfielddb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/config/formfieldbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// Store manages the set of APIs for form fields database access.
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
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (formfieldbus.Storer, error) {
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

// Create inserts a new form field into the database.
func (s *Store) Create(ctx context.Context, field formfieldbus.FormField) error {
	// Validate that the schema.table combination exists
	const checkTable = `
	SELECT EXISTS (
		SELECT 1
		FROM information_schema.tables
		WHERE table_schema = :schema
		AND table_name = :table
	)`

	exists := struct {
		Exists bool `db:"exists"`
	}{}

	checkData := map[string]any{
		"schema": field.EntitySchema,
		"table":  field.EntityTable,
	}

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, checkTable, checkData, &exists); err != nil {
		return fmt.Errorf("namedquerystruct: %w", err)
	}

	if !exists.Exists {
		return fmt.Errorf("schema[%s] table[%s]: %w", field.EntitySchema, field.EntityTable, formfieldbus.ErrNonexistentTableName)
	}

	const q = `
	INSERT INTO config.form_fields (
		id, form_id, entity_id, entity_schema, entity_table, name, label, field_type, field_order, required, config
	) VALUES (
		:id, :form_id, :entity_id, :entity_schema, :entity_table, :name, :label, :field_type, :field_order, :required, :config
	)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBFormField(field)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", formfieldbus.ErrUniqueEntry)
		}
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", formfieldbus.ErrForeignKeyViolation)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update replaces a form field document in the database.
func (s *Store) Update(ctx context.Context, field formfieldbus.FormField) error {
	// Validate that the schema.table combination exists
	const checkTable = `
	SELECT EXISTS (
		SELECT 1
		FROM information_schema.tables
		WHERE table_schema = :schema
		AND table_name = :table
	)`

	exists := struct {
		Exists bool `db:"exists"`
	}{}

	checkData := map[string]any{
		"schema": field.EntitySchema,
		"table":  field.EntityTable,
	}

	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, checkTable, checkData, &exists); err != nil {
		return fmt.Errorf("namedquerystruct: %w", err)
	}

	if !exists.Exists {
		return fmt.Errorf("schema[%s] table[%s]: %w", field.EntitySchema, field.EntityTable, formfieldbus.ErrNonexistentTableName)
	}

	const q = `
	UPDATE
		config.form_fields
	SET
		form_id = :form_id,
		entity_id = :entity_id,
		entity_schema = :entity_schema,
		entity_table = :entity_table,
		name = :name,
		label = :label,
		field_type = :field_type,
		field_order = :field_order,
		required = :required,
		config = :config
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBFormField(field)); err != nil {
		if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", formfieldbus.ErrUniqueEntry)
		}
		if errors.Is(err, sqldb.ErrForeignKeyViolation) {
			return fmt.Errorf("namedexeccontext: %w", formfieldbus.ErrForeignKeyViolation)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a form field from the database.
func (s *Store) Delete(ctx context.Context, field formfieldbus.FormField) error {
	const q = `
	DELETE FROM
		config.form_fields
	WHERE
		id = :id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBFormField(field)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of form fields from the database.
func (s *Store) Query(ctx context.Context, filter formfieldbus.QueryFilter, orderBy order.By, page page.Page) ([]formfieldbus.FormField, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
		id, form_id, entity_id, entity_schema, entity_table, name, label, field_type, field_order, required, config
	FROM
		config.form_fields`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbFields []formField
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbFields); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusFormFields(dbFields), nil
}

// Count returns the number of form fields in the database.
func (s *Store) Count(ctx context.Context, filter formfieldbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		COUNT(1) AS count
	FROM
		config.form_fields`

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

// QueryByID retrieves a single form field from the database by its ID.
func (s *Store) QueryByID(ctx context.Context, fieldID uuid.UUID) (formfieldbus.FormField, error) {
	data := struct {
		ID string `db:"id"`
	}{
		ID: fieldID.String(),
	}

	const q = `
	SELECT
		id, form_id, entity_id, entity_schema, entity_table, name, label, field_type, field_order, required, config
	FROM
		config.form_fields
	WHERE
		id = :id`

	var dbField formField
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbField); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return formfieldbus.FormField{}, formfieldbus.ErrNotFound
		}
		return formfieldbus.FormField{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toBusFormField(dbField), nil
}

// QueryByFormID retrieves all fields for a specific form, ordered by field_order.
func (s *Store) QueryByFormID(ctx context.Context, formID uuid.UUID) ([]formfieldbus.FormField, error) {
	data := struct {
		FormID string `db:"form_id"`
	}{
		FormID: formID.String(),
	}

	const q = `
	SELECT
		id, form_id, entity_id, entity_schema, entity_table, name, label, field_type, field_order, required, config
	FROM
		config.form_fields
	WHERE
		form_id = :form_id
	ORDER BY
		field_order ASC, name ASC`

	var dbFields []formField
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbFields); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusFormFields(dbFields), nil
}