package tablebuilder

import (
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/doug-martin/goqu/v9/exp"
)

// QueryBuilder builds SQL queries using goqu
type QueryBuilder struct {
	dialect goqu.DialectWrapper
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		dialect: goqu.Dialect("postgres"),
	}
}

// BuildQuery builds a SQL query from a data source configuration
func (qb *QueryBuilder) BuildQuery(ds *DataSource, params QueryParams, isPrimary bool) (string, map[string]interface{}, error) {
	// Start with the base table/view
	query := qb.dialect.From(ds.Source)

	// Build select clause - PASS THE BASE TABLE NAME
	selectCols := qb.buildSelectColumns(ds.Select, ds.Source) // <-- Added ds.Source parameter
	query = query.Select(selectCols...)

	// Apply explicit joins
	query = qb.applyJoins(query, ds.Joins)

	// Apply joins from foreign tables
	query = qb.applyForeignTableJoins(query, ds.Select.ForeignTables)

	// Apply filters
	query = qb.applyFilters(query, ds.Filters, params)

	// Apply sorting (only for primary data source)
	if isPrimary {
		query = qb.applySorting(query, ds.Sort, params)
	}

	// Apply pagination (only for primary data source)
	if isPrimary && params.Page > 0 {
		limit := params.Limit
		if limit == 0 && ds.Limit > 0 {
			limit = ds.Limit
		}
		if limit > 0 {
			offset := (params.Page - 1) * limit
			query = query.Limit(uint(limit)).Offset(uint(offset))
		}
	} else if ds.Limit > 0 {
		query = query.Limit(uint(ds.Limit))
	}

	// Generate SQL
	sql, args, err := query.ToSQL()
	if err != nil {
		return "", nil, fmt.Errorf("generate sql: %w", err)
	}

	// Convert args to map for named query
	argsMap := make(map[string]interface{})
	for i, arg := range args {
		argsMap[fmt.Sprintf("arg%d", i+1)] = arg
	}

	// Replace ? with named parameters
	sql = qb.replaceQuestionMarks(sql, len(args))

	return sql, argsMap, nil
}

// BuildCountQuery builds a count query for pagination
func (qb *QueryBuilder) BuildCountQuery(ds *DataSource, params QueryParams) (string, map[string]interface{}, error) {
	// Start with the base table/view
	query := qb.dialect.From(ds.Source)

	// Apply joins (if needed for filters)
	query = qb.applyJoins(query, ds.Joins)

	// Apply filters
	query = qb.applyFilters(query, ds.Filters, params)

	// Select COUNT(*)
	query = query.Select(goqu.COUNT("*").As("count"))

	// Generate SQL
	sql, args, err := query.ToSQL()
	if err != nil {
		return "", nil, fmt.Errorf("generate count sql: %w", err)
	}

	// Convert args to map
	argsMap := make(map[string]interface{})
	for i, arg := range args {
		argsMap[fmt.Sprintf("arg%d", i+1)] = arg
	}

	// Replace ? with named parameters
	sql = qb.replaceQuestionMarks(sql, len(args))

	return sql, argsMap, nil
}

// buildSelectColumns builds the select columns
func (qb *QueryBuilder) buildSelectColumns(config SelectConfig, baseTable string) []interface{} {
	var cols []interface{}

	for _, col := range config.Columns {
		colExpr := goqu.T(baseTable).Col(col.Name)

		if col.Alias != "" {
			cols = append(cols, colExpr.As(col.Alias))
		} else {
			// No alias - just use the column as-is
			cols = append(cols, colExpr)
		}
	}

	// Add foreign table columns
	cols = append(cols, qb.buildForeignColumns(config.ForeignTables, "")...)
	return cols
}

// buildForeignColumns recursively builds columns from foreign tables
func (qb *QueryBuilder) buildForeignColumns(foreignTables []ForeignTable, prefix string) []interface{} {
	var cols []interface{}

	for _, ft := range foreignTables {
		for _, col := range ft.Columns {
			colExpr := goqu.T(ft.Table).Col(col.Name)

			if col.Alias != "" {
				cols = append(cols, colExpr.As(col.Alias))
			} else {
				cols = append(cols, colExpr)
			}
		}

		// Recursively add nested foreign table columns
		if len(ft.ForeignTables) > 0 {
			cols = append(cols, qb.buildForeignColumns(ft.ForeignTables, ft.Table)...)
		}
	}

	return cols
}

// applyForeignTableJoins applies joins for foreign table relationships
func (qb *QueryBuilder) applyForeignTableJoins(query *goqu.SelectDataset, foreignTables []ForeignTable) *goqu.SelectDataset {
	for _, ft := range foreignTables {
		// Parse the from and to relationships
		fromParts := strings.Split(ft.RelationshipFrom, ".")
		toParts := strings.Split(ft.RelationshipTo, ".")

		var joinCond goqu.Expression
		if len(fromParts) == 2 && len(toParts) == 2 {
			// Standard table.column format
			joinCond = goqu.T(fromParts[0]).Col(fromParts[1]).Eq(
				goqu.T(toParts[0]).Col(toParts[1]),
			)
		} else if len(fromParts) == 1 && len(toParts) == 2 {
			// From is just column name (assume base table)
			joinCond = goqu.C(fromParts[0]).Eq(
				goqu.T(toParts[0]).Col(toParts[1]),
			)
		} else if len(fromParts) == 2 && len(toParts) == 1 {
			// To is just column name (assume foreign table)
			joinCond = goqu.T(fromParts[0]).Col(fromParts[1]).Eq(
				goqu.C(toParts[0]),
			)
		} else {
			// Fallback to simple column names
			joinCond = goqu.C(ft.RelationshipFrom).Eq(goqu.C(ft.RelationshipTo))
		}

		// Apply the join based on type
		switch strings.ToLower(ft.JoinType) {
		case "left":
			query = query.LeftJoin(goqu.T(ft.Table), goqu.On(joinCond))
		case "right":
			query = query.RightJoin(goqu.T(ft.Table), goqu.On(joinCond))
		case "full":
			query = query.FullJoin(goqu.T(ft.Table), goqu.On(joinCond))
		default: // "inner" or empty
			query = query.InnerJoin(goqu.T(ft.Table), goqu.On(joinCond))
		}

		// Recursively apply nested foreign table joins
		if len(ft.ForeignTables) > 0 {
			query = qb.applyForeignTableJoins(query, ft.ForeignTables)
		}
	}

	return query
}

// applyJoins applies join clauses to the query
func (qb *QueryBuilder) applyJoins(query *goqu.SelectDataset, joins []Join) *goqu.SelectDataset {
	for _, join := range joins {
		joinTable := goqu.T(join.Table)
		joinExpr := qb.BuildJoinCondition(join.On)
		joinCond := goqu.On(joinExpr) // Wrap with goqu.On()

		switch strings.ToLower(join.Type) {
		case "left":
			query = query.LeftJoin(joinTable, joinCond)
		case "right":
			query = query.RightJoin(joinTable, joinCond)
		case "full":
			query = query.FullJoin(joinTable, joinCond)
		default: // "inner"
			query = query.InnerJoin(joinTable, joinCond)
		}
	}

	// Apply joins for foreign tables
	// This would be more complex in practice, parsing the relationship field

	return query
}

// applyFilters applies filter conditions to the query
func (qb *QueryBuilder) applyFilters(query *goqu.SelectDataset, filters []Filter, params QueryParams) *goqu.SelectDataset {
	var expressions []goqu.Expression

	// Apply static filters
	for _, filter := range filters {
		expr := qb.buildFilterExpression(filter, params)
		if expr != nil {
			expressions = append(expressions, expr)
		}
	}

	// Apply dynamic filters from params
	for _, filter := range params.Filters {
		expr := qb.buildFilterExpression(filter, params)
		if expr != nil {
			expressions = append(expressions, expr)
		}
	}

	if len(expressions) > 0 {
		query = query.Where(expressions...)
	}

	return query
}

// buildFilterExpression builds a goqu expression from a filter
func (qb *QueryBuilder) buildFilterExpression(filter Filter, params QueryParams) goqu.Expression {
	// Check for dynamic value
	value := filter.Value
	if filter.Dynamic && params.Dynamic != nil {
		if dynValue, ok := params.Dynamic[filter.Column]; ok {
			value = dynValue
		}
	}

	// Skip if no value
	if value == nil {
		return nil
	}

	col := goqu.I(filter.Column)

	switch strings.ToLower(filter.Operator) {
	case "eq", "=":
		return col.Eq(value)
	case "neq", "!=", "<>":
		return col.Neq(value)
	case "gt", ">":
		return col.Gt(value)
	case "gte", ">=":
		return col.Gte(value)
	case "lt", "<":
		return col.Lt(value)
	case "lte", "<=":
		return col.Lte(value)
	case "in":
		if arr, ok := value.([]interface{}); ok {
			return col.In(arr...)
		}
		return col.In(value)
	case "like":
		return col.Like(fmt.Sprintf("%%%v%%", value))
	case "ilike":
		return col.ILike(fmt.Sprintf("%%%v%%", value))
	case "is_null":
		return col.IsNull()
	case "is_not_null":
		return col.IsNotNull()
	default:
		return col.Eq(value)
	}
}

// applySorting applies sort order to the query
func (qb *QueryBuilder) applySorting(query *goqu.SelectDataset, sorts []Sort, params QueryParams) *goqu.SelectDataset {
	// Use params sort if provided, otherwise use config sort
	sortList := params.Sort
	if len(sortList) == 0 {
		sortList = sorts
	}

	if len(sortList) == 0 {
		return query
	}

	var orderExprs []exp.OrderedExpression

	for _, sort := range sortList {
		col := goqu.I(sort.Column)
		if strings.ToLower(sort.Direction) == "desc" {
			orderExprs = append(orderExprs, col.Desc())
		} else {
			orderExprs = append(orderExprs, col.Asc())
		}
	}

	if len(orderExprs) > 0 {
		query = query.Order(orderExprs...)
	}

	return query
}

// replaceQuestionMarks replaces ? placeholders with named parameters
func (qb *QueryBuilder) replaceQuestionMarks(sql string, count int) string {
	result := sql
	for i := 1; i <= count; i++ {
		old := "?"
		new := fmt.Sprintf(":arg%d", i)
		result = strings.Replace(result, old, new, 1)
	}
	return result
}

// BuildJoinCondition builds a proper join condition from a string
func (qb *QueryBuilder) BuildJoinCondition(condition string) goqu.Expression {
	// Parse conditions like "table1.id = table2.foreign_id"
	// This is simplified - you'd want more robust parsing
	parts := strings.Split(condition, "=")
	if len(parts) == 2 {
		left := strings.TrimSpace(parts[0])
		right := strings.TrimSpace(parts[1])
		return goqu.I(left).Eq(goqu.I(right))
	}

	// Default to treating it as a literal expression
	return goqu.L(condition)
}
