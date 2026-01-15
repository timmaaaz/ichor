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

// QueryReferencingTables returns all tables that have foreign keys pointing to the given table.
func (b *Business) QueryReferencingTables(ctx context.Context, schema, table string) ([]ReferencingTable, error) {
	ctx, span := otel.AddSpan(ctx, "business.introspectionbus.queryreferencingtables")
	defer span.End()

	const q = `
	SELECT
		tc.table_schema AS schema,
		tc.table_name AS table,
		kcu.column_name AS fk_column,
		tc.constraint_name AS constraint_name
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
		AND ccu.table_schema = :schema
		AND ccu.table_name = :table
	ORDER BY
		tc.table_schema, tc.table_name`

	data := struct {
		Schema string `db:"schema"`
		Table  string `db:"table"`
	}{
		Schema: schema,
		Table:  table,
	}

	var tables []ReferencingTable
	if err := sqldb.NamedQuerySlice(ctx, b.log, b.db, q, data, &tables); err != nil {
		return nil, fmt.Errorf("query referencing tables: %w", err)
	}

	return tables, nil
}

// QueryEnumTypes returns all ENUM types in a given schema.
func (b *Business) QueryEnumTypes(ctx context.Context, schema string) ([]EnumType, error) {
	ctx, span := otel.AddSpan(ctx, "business.introspectionbus.queryenumtypes")
	defer span.End()

	const q = `
	SELECT
		t.typname AS enum_name,
		n.nspname AS schema_name,
		array_agg(e.enumlabel ORDER BY e.enumsortorder) AS values
	FROM pg_type t
	JOIN pg_namespace n ON t.typnamespace = n.oid
	JOIN pg_enum e ON t.oid = e.enumtypid
	WHERE t.typtype = 'e'
	  AND n.nspname = :schema
	GROUP BY t.typname, n.nspname
	ORDER BY t.typname`

	data := struct {
		Schema string `db:"schema"`
	}{
		Schema: schema,
	}

	var enums []EnumType
	if err := sqldb.NamedQuerySlice(ctx, b.log, b.db, q, data, &enums); err != nil {
		return nil, fmt.Errorf("query enum types: %w", err)
	}

	return enums, nil
}

// QueryEnumValues returns all values for a specific ENUM type.
func (b *Business) QueryEnumValues(ctx context.Context, schema, enumName string) ([]EnumValue, error) {
	ctx, span := otel.AddSpan(ctx, "business.introspectionbus.queryenumvalues")
	defer span.End()

	const q = `
	SELECT
		e.enumlabel AS value,
		e.enumsortorder AS sort_order
	FROM pg_type t
	JOIN pg_namespace n ON t.typnamespace = n.oid
	JOIN pg_enum e ON t.oid = e.enumtypid
	WHERE t.typtype = 'e'
	  AND n.nspname = :schema
	  AND t.typname = :enum_name
	ORDER BY e.enumsortorder`

	data := struct {
		Schema   string `db:"schema"`
		EnumName string `db:"enum_name"`
	}{
		Schema:   schema,
		EnumName: enumName,
	}

	var values []EnumValue
	if err := sqldb.NamedQuerySlice(ctx, b.log, b.db, q, data, &values); err != nil {
		return nil, fmt.Errorf("query enum values: %w", err)
	}

	return values, nil
}

// QueryEnumLabels returns labels for a specific enum from the config.enum_labels table.
func (b *Business) QueryEnumLabels(ctx context.Context, enumName string) ([]EnumLabel, error) {
	ctx, span := otel.AddSpan(ctx, "business.introspectionbus.queryenumlabels")
	defer span.End()

	const q = `
	SELECT
		enum_name,
		value,
		label,
		sort_order
	FROM config.enum_labels
	WHERE enum_name = :enum_name
	ORDER BY sort_order`

	data := struct {
		EnumName string `db:"enum_name"`
	}{
		EnumName: enumName,
	}

	var labels []EnumLabel
	if err := sqldb.NamedQuerySlice(ctx, b.log, b.db, q, data, &labels); err != nil {
		return nil, fmt.Errorf("query enum labels: %w", err)
	}

	return labels, nil
}

// QueryEnumOptions returns ready-to-use dropdown options by merging enum values with labels.
// If a label exists in config.enum_labels, it's used; otherwise, the value is title-cased.
func (b *Business) QueryEnumOptions(ctx context.Context, schema, enumName string) ([]EnumOption, error) {
	ctx, span := otel.AddSpan(ctx, "business.introspectionbus.queryenumoptions")
	defer span.End()

	// Get enum values from PostgreSQL
	values, err := b.QueryEnumValues(ctx, schema, enumName)
	if err != nil {
		return nil, fmt.Errorf("query enum values: %w", err)
	}

	// Get labels from config.enum_labels table
	fullEnumName := schema + "." + enumName
	labels, err := b.QueryEnumLabels(ctx, fullEnumName)
	if err != nil {
		return nil, fmt.Errorf("query enum labels: %w", err)
	}

	// Build a map of value -> label for quick lookup
	labelMap := make(map[string]EnumLabel)
	for _, l := range labels {
		labelMap[l.Value] = l
	}

	// Merge values with labels
	options := make([]EnumOption, len(values))
	for i, v := range values {
		opt := EnumOption{
			Value:     v.Value,
			SortOrder: v.SortOrder,
		}

		if label, ok := labelMap[v.Value]; ok {
			opt.Label = label.Label
			// Use label's sort_order if defined (non-default)
			if label.SortOrder != 1000 {
				opt.SortOrder = label.SortOrder
			}
		} else {
			// Title-case the value as fallback
			opt.Label = titleCase(v.Value)
		}

		options[i] = opt
	}

	return options, nil
}

// titleCase converts a snake_case or lowercase string to Title Case.
func titleCase(s string) string {
	if s == "" {
		return s
	}
	// Simple title case: capitalize first letter
	result := make([]byte, len(s))
	capitalizeNext := true
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '_' || c == '-' || c == ' ' {
			result[i] = ' '
			capitalizeNext = true
		} else if capitalizeNext && c >= 'a' && c <= 'z' {
			result[i] = c - 32 // Convert to uppercase
			capitalizeNext = false
		} else {
			result[i] = c
			capitalizeNext = false
		}
	}
	return string(result)
}
