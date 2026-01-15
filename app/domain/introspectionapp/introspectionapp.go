package introspectionapp

import (
	"context"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/introspectionbus"
)

// App manages the set of app layer api functions for the introspection domain.
type App struct {
	introspectionBus *introspectionbus.Business
}

// NewApp constructs an introspection app API for use.
func NewApp(introspectionBus *introspectionbus.Business) *App {
	return &App{
		introspectionBus: introspectionBus,
	}
}

// QuerySchemas returns all database schemas (excluding system schemas).
func (a *App) QuerySchemas(ctx context.Context) (Schemas, error) {
	schemas, err := a.introspectionBus.QuerySchemas(ctx)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "query schemas: %s", err)
	}

	return Schemas(ToAppSchemas(schemas)), nil
}

// QueryTables returns all tables in a given schema.
func (a *App) QueryTables(ctx context.Context, schema string) (Tables, error) {
	tables, err := a.introspectionBus.QueryTables(ctx, schema)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "query tables: %s", err)
	}

	return Tables(ToAppTables(tables)), nil
}

// QueryColumns returns all columns for a given table.
func (a *App) QueryColumns(ctx context.Context, schema, table string) (Columns, error) {
	columns, err := a.introspectionBus.QueryColumns(ctx, schema, table)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "query columns: %s", err)
	}

	return Columns(ToAppColumns(columns)), nil
}

// QueryRelationships returns all foreign key relationships for a given table.
func (a *App) QueryRelationships(ctx context.Context, schema, table string) (Relationships, error) {
	relationships, err := a.introspectionBus.QueryRelationships(ctx, schema, table)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "query relationships: %s", err)
	}

	return Relationships(ToAppRelationships(relationships)), nil
}

// QueryReferencingTables returns all tables that have foreign keys pointing to the given table.
func (a *App) QueryReferencingTables(ctx context.Context, schema, table string) (ReferencingTables, error) {
	tables, err := a.introspectionBus.QueryReferencingTables(ctx, schema, table)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "query referencing tables: %s", err)
	}

	return ReferencingTables(ToAppReferencingTables(tables)), nil
}

// QueryEnumTypes returns all ENUM types in a given schema.
func (a *App) QueryEnumTypes(ctx context.Context, schema string) (EnumTypes, error) {
	enums, err := a.introspectionBus.QueryEnumTypes(ctx, schema)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "query enum types: %s", err)
	}

	return EnumTypes(ToAppEnumTypes(enums)), nil
}

// QueryEnumValues returns all values for a specific ENUM type.
func (a *App) QueryEnumValues(ctx context.Context, schema, enumName string) (EnumValues, error) {
	values, err := a.introspectionBus.QueryEnumValues(ctx, schema, enumName)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "query enum values: %s", err)
	}

	return EnumValues(ToAppEnumValues(values)), nil
}

// QueryEnumOptions returns ready-to-use dropdown options by merging enum values with labels.
func (a *App) QueryEnumOptions(ctx context.Context, schema, enumName string) (EnumOptions, error) {
	options, err := a.introspectionBus.QueryEnumOptions(ctx, schema, enumName)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "query enum options: %s", err)
	}

	return EnumOptions(ToAppEnumOptions(options)), nil
}
