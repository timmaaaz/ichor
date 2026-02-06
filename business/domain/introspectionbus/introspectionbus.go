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
		nspname AS name
	FROM
		pg_catalog.pg_namespace
	WHERE
		nspname NOT IN ('pg_catalog', 'information_schema', 'pg_toast')
		AND nspname NOT LIKE 'pg_temp_%'
		AND nspname NOT LIKE 'pg_toast_temp_%'
	ORDER BY
		nspname`

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
		n.nspname AS schema,
		c.relname AS name,
		CAST(c.reltuples AS bigint) AS row_count_estimate
	FROM
		pg_catalog.pg_class c
	JOIN
		pg_catalog.pg_namespace n ON c.relnamespace = n.oid
	WHERE
		n.nspname = :schema
		AND c.relkind = 'r'
	ORDER BY
		c.relname`

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

// QueryColumns returns all columns for a given table, enriched with FK metadata.
func (b *Business) QueryColumns(ctx context.Context, schema, table string) ([]Column, error) {
	ctx, span := otel.AddSpan(ctx, "business.introspectionbus.querycolumns")
	defer span.End()

	const q = `
	SELECT
		a.attname AS name,
		pg_catalog.format_type(a.atttypid, a.atttypmod) AS data_type,
		NOT a.attnotnull AS is_nullable,
		COALESCE(pg_get_expr(d.adbin, d.adrelid), '') AS default_value,
		COALESCE(pk.is_pk, FALSE) AS is_primary_key,
		fk.conname IS NOT NULL AS is_foreign_key,
		fk_ns.nspname AS referenced_schema,
		fk_cl.relname AS referenced_table,
		fk_att.attname AS referenced_column
	FROM
		pg_catalog.pg_attribute a
	JOIN pg_catalog.pg_class c ON a.attrelid = c.oid
	JOIN pg_catalog.pg_namespace n ON c.relnamespace = n.oid
	-- Default values
	LEFT JOIN pg_catalog.pg_attrdef d ON a.attrelid = d.adrelid AND a.attnum = d.adnum
	-- Primary key detection
	LEFT JOIN (
		SELECT
			conrelid,
			unnest(conkey) AS attnum,
			TRUE AS is_pk
		FROM pg_catalog.pg_constraint
		WHERE contype = 'p'
	) pk ON a.attrelid = pk.conrelid AND a.attnum = pk.attnum
	-- Foreign key detection with referenced table info
	LEFT JOIN LATERAL (
		SELECT
			con.conname,
			con.confrelid,
			col_idx.ord,
			(con.confkey)[col_idx.ord] AS ref_attnum
		FROM pg_catalog.pg_constraint con
		CROSS JOIN LATERAL unnest(con.conkey) WITH ORDINALITY AS col_idx(attnum, ord)
		WHERE con.contype = 'f'
		  AND con.conrelid = a.attrelid
		  AND col_idx.attnum = a.attnum
	) fk ON TRUE
	LEFT JOIN pg_catalog.pg_class fk_cl ON fk.confrelid = fk_cl.oid
	LEFT JOIN pg_catalog.pg_namespace fk_ns ON fk_cl.relnamespace = fk_ns.oid
	LEFT JOIN pg_catalog.pg_attribute fk_att ON fk.confrelid = fk_att.attrelid AND fk.ref_attnum = fk_att.attnum
	WHERE
		n.nspname = :schema
		AND c.relname = :table
		AND a.attnum > 0           -- Exclude system columns
		AND NOT a.attisdropped     -- Exclude dropped columns
	ORDER BY
		a.attnum`

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
		con.conname AS foreign_key_name,
		att.attname AS column_name,
		ref_ns.nspname AS referenced_schema,
		ref_cl.relname AS referenced_table,
		ref_att.attname AS referenced_column,
		'many-to-one' AS relationship_type
	FROM
		pg_catalog.pg_constraint con
	JOIN pg_catalog.pg_class cl ON con.conrelid = cl.oid
	JOIN pg_catalog.pg_namespace ns ON cl.relnamespace = ns.oid
	JOIN pg_catalog.pg_class ref_cl ON con.confrelid = ref_cl.oid
	JOIN pg_catalog.pg_namespace ref_ns ON ref_cl.relnamespace = ref_ns.oid
	CROSS JOIN LATERAL unnest(con.conkey, con.confkey)
		WITH ORDINALITY AS cols(src_attnum, ref_attnum, ord)
	JOIN pg_catalog.pg_attribute att
		ON att.attrelid = con.conrelid AND att.attnum = cols.src_attnum
	JOIN pg_catalog.pg_attribute ref_att
		ON ref_att.attrelid = con.confrelid AND ref_att.attnum = cols.ref_attnum
	WHERE
		con.contype = 'f'
		AND ns.nspname = :schema
		AND cl.relname = :table
	ORDER BY
		con.conname, cols.ord`

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
		ns.nspname AS schema,
		cl.relname AS table,
		att.attname AS fk_column,
		con.conname AS constraint_name
	FROM
		pg_catalog.pg_constraint con
	JOIN pg_catalog.pg_class cl ON con.conrelid = cl.oid
	JOIN pg_catalog.pg_namespace ns ON cl.relnamespace = ns.oid
	JOIN pg_catalog.pg_class ref_cl ON con.confrelid = ref_cl.oid
	JOIN pg_catalog.pg_namespace ref_ns ON ref_cl.relnamespace = ref_ns.oid
	CROSS JOIN LATERAL unnest(con.conkey)
		WITH ORDINALITY AS cols(attnum, ord)
	JOIN pg_catalog.pg_attribute att
		ON att.attrelid = con.conrelid AND att.attnum = cols.attnum
	WHERE
		con.contype = 'f'
		AND ref_ns.nspname = :schema
		AND ref_cl.relname = :table
	ORDER BY
		ns.nspname, cl.relname`

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
