package introspectionbus

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

// Business manages the introspection operations.
type Business struct {
	log *logger.Logger
	db  *sqlx.DB
}

// NewBusiness constructs a Business for introspection.
func NewBusiness(log *logger.Logger, db *sqlx.DB) *Business {
	return &Business{
		log: log,
		db:  db,
	}
}

// QuerySchemas returns all database schemas (excluding system schemas).
func (b *Business) QuerySchemas(ctx context.Context) ([]Schema, error) {
	ctx, span := otel.AddSpan(ctx, "business.introspectionbus.queryschemas")
	defer span.End()

	const q = `
	SELECT
		schema_name AS name
	FROM
		information_schema.schemata
	WHERE
		schema_name NOT IN ('pg_catalog', 'information_schema', 'pg_toast')
	ORDER BY
		schema_name`

	var schemas []Schema
	if err := sqldb.NamedQuerySlice(ctx, b.log, b.db, q, struct{}{}, &schemas); err != nil {
		return nil, fmt.Errorf("query schemas: %w", err)
	}

	return schemas, nil
}

// QueryTables returns all tables in a given schema.
func (b *Business) QueryTables(ctx context.Context, schema string) ([]Table, error) {
	ctx, span := otel.AddSpan(ctx, "business.introspectionbus.querytables")
	defer span.End()

	const q = `
	SELECT
		t.table_schema AS schema,
		t.table_name AS name,
		CAST(c.reltuples AS bigint) AS row_count_estimate
	FROM
		information_schema.tables t
	LEFT JOIN
		pg_class c ON c.relname = t.table_name
	LEFT JOIN
		pg_namespace n ON n.oid = c.relnamespace AND n.nspname = t.table_schema
	WHERE
		t.table_schema = :schema
		AND t.table_type = 'BASE TABLE'
	ORDER BY
		t.table_name`

	data := struct {
		Schema string `db:"schema"`
	}{
		Schema: schema,
	}

	var tables []Table
	if err := sqldb.NamedQuerySlice(ctx, b.log, b.db, q, data, &tables); err != nil {
		return nil, fmt.Errorf("query tables: %w", err)
	}

	return tables, nil
}

// QueryColumns returns all columns for a given table.
func (b *Business) QueryColumns(ctx context.Context, schema, table string) ([]Column, error) {
	ctx, span := otel.AddSpan(ctx, "business.introspectionbus.querycolumns")
	defer span.End()

	const q = `
	SELECT
		c.column_name AS name,
		c.data_type AS data_type,
		c.is_nullable = 'YES' AS is_nullable,
		COALESCE(c.column_default, '') AS default_value,
		EXISTS(
			SELECT 1
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage kcu
				ON tc.constraint_name = kcu.constraint_name
				AND tc.table_schema = kcu.table_schema
			WHERE tc.constraint_type = 'PRIMARY KEY'
				AND tc.table_schema = :schema
				AND tc.table_name = :table
				AND kcu.column_name = c.column_name
		) AS is_primary_key
	FROM
		information_schema.columns c
	WHERE
		c.table_schema = :schema
		AND c.table_name = :table
	ORDER BY
		c.ordinal_position`

	data := struct {
		Schema string `db:"schema"`
		Table  string `db:"table"`
	}{
		Schema: schema,
		Table:  table,
	}

	var columns []Column
	if err := sqldb.NamedQuerySlice(ctx, b.log, b.db, q, data, &columns); err != nil {
		return nil, fmt.Errorf("query columns: %w", err)
	}

	return columns, nil
}

// QueryRelationships returns all foreign key relationships for a given table.
func (b *Business) QueryRelationships(ctx context.Context, schema, table string) ([]Relationship, error) {
	ctx, span := otel.AddSpan(ctx, "business.introspectionbus.queryrelationships")
	defer span.End()

	const q = `
	SELECT
		tc.constraint_name AS foreign_key_name,
		kcu.column_name AS column_name,
		ccu.table_schema AS referenced_schema,
		ccu.table_name AS referenced_table,
		ccu.column_name AS referenced_column,
		'many-to-one' AS relationship_type
	FROM
		information_schema.table_constraints tc
	JOIN
		information_schema.key_column_usage kcu
		ON tc.constraint_name = kcu.constraint_name
		AND tc.table_schema = kcu.table_schema
	JOIN
		information_schema.constraint_column_usage ccu
		ON tc.constraint_name = ccu.constraint_name
	WHERE
		tc.constraint_type = 'FOREIGN KEY'
		AND tc.table_schema = :schema
		AND tc.table_name = :table
	ORDER BY
		kcu.ordinal_position`

	data := struct {
		Schema string `db:"schema"`
		Table  string `db:"table"`
	}{
		Schema: schema,
		Table:  table,
	}

	var relationships []Relationship
	if err := sqldb.NamedQuerySlice(ctx, b.log, b.db, q, data, &relationships); err != nil {
		return nil, fmt.Errorf("query relationships: %w", err)
	}

	return relationships, nil
}
