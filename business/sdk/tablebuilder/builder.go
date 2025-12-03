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
	// Check if this is a metric/chart query
	if len(ds.Metrics) > 0 {
		return qb.buildMetricQuery(ds, params, isPrimary)
	}

	var from string
	if ds.Schema != "" {
		from = strings.Join([]string{ds.Schema, ds.Source}, ".")
	} else {
		from = ds.Source
	}

	// Start with the base table/view
	query := qb.dialect.From(from)

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
		limit := params.Rows
		if limit == 0 && ds.Rows > 0 {
			limit = ds.Rows
		}
		if limit > 0 {
			offset := (params.Page - 1) * limit
			query = query.Limit(uint(limit)).Offset(uint(offset))
		}
	} else if ds.Rows > 0 {
		query = query.Limit(uint(ds.Rows))
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
	var from string
	if ds.Schema != "" {
		from = strings.Join([]string{ds.Schema, ds.Source}, ".")
	} else {
		from = ds.Source
	}

	query := qb.dialect.From(from)

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
		// Use alias if present, otherwise use table name
		tableRef := getTableOrAlias(ft)

		for _, col := range ft.Columns {
			colExpr := goqu.T(tableRef).Col(col.Name)

			if col.Alias != "" {
				cols = append(cols, colExpr.As(col.Alias))
			} else {
				cols = append(cols, colExpr)
			}
		}

		// Recursively add nested foreign table columns
		if len(ft.ForeignTables) > 0 {
			cols = append(cols, qb.buildForeignColumns(ft.ForeignTables, tableRef)...)
		}
	}

	return cols
}

// applyForeignTableJoins applies joins for foreign table relationships
func (qb *QueryBuilder) applyForeignTableJoins(query *goqu.SelectDataset, foreignTables []ForeignTable) *goqu.SelectDataset {
	for _, ft := range foreignTables {
		// Use alias if present, otherwise use table name
		tableRef := getTableOrAlias(ft)

		// Parse the from and to relationships
		fromParts := strings.Split(ft.RelationshipFrom, ".")
		toParts := strings.Split(ft.RelationshipTo, ".")

		var joinCond goqu.Expression
		if len(fromParts) == 2 && len(toParts) == 2 {
			// Standard table.column format
			// Use tableRef (alias or table) for the right side of the condition
			joinCond = goqu.T(fromParts[0]).Col(fromParts[1]).Eq(
				goqu.T(toParts[0]).Col(toParts[1]),
			)
		} else if len(fromParts) == 1 && len(toParts) == 2 {
			// From is just column name (assume base table)
			// Use tableRef for the right side
			joinCond = goqu.C(fromParts[0]).Eq(
				goqu.T(toParts[0]).Col(toParts[1]),
			)
		} else if len(fromParts) == 2 && len(toParts) == 1 {
			// To is just column name (assume foreign table, use alias if present)
			joinCond = goqu.T(fromParts[0]).Col(fromParts[1]).Eq(
				goqu.T(tableRef).Col(toParts[0]),
			)
		} else {
			// Fallback to simple column names
			joinCond = goqu.C(ft.RelationshipFrom).Eq(goqu.C(ft.RelationshipTo))
		}

		// Build the join table expression with alias support
		// Use exp.Expression to handle both IdentifierExpression and AliasedExpression
		var joinTable exp.Expression
		if ft.Schema != "" {
			if ft.Alias != "" {
				// Schema with alias: schema.table AS alias
				joinTable = goqu.S(ft.Schema).Table(ft.Table).As(ft.Alias)
			} else {
				// Schema without alias: schema.table
				joinTable = goqu.S(ft.Schema).Table(ft.Table)
			}
		} else {
			if ft.Alias != "" {
				// No schema with alias: table AS alias
				joinTable = goqu.T(ft.Table).As(ft.Alias)
			} else {
				// No schema, no alias: table
				joinTable = goqu.T(ft.Table)
			}
		}

		// Apply the join based on type
		switch strings.ToLower(ft.JoinType) {
		case "left":
			query = query.LeftJoin(joinTable, goqu.On(joinCond))
		case "right":
			query = query.RightJoin(joinTable, goqu.On(joinCond))
		case "full":
			query = query.FullJoin(joinTable, goqu.On(joinCond))
		default: // "inner" or empty
			query = query.InnerJoin(joinTable, goqu.On(joinCond))
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

		var joinTable exp.IdentifierExpression
		if join.Schema != "" {
			joinTable = goqu.S(join.Schema).Table(join.Table)
		} else {
			joinTable = goqu.T(join.Table)
		}
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

// getTableOrAlias returns the alias if present, otherwise the table name
func getTableOrAlias(ft ForeignTable) string {
	if ft.Alias != "" {
		return ft.Alias
	}
	return ft.Table
}

// =============================================================================
// Metric Query Builder (for charts with aggregations)
// =============================================================================

// buildMetricQuery builds aggregation queries for charts
func (qb *QueryBuilder) buildMetricQuery(ds *DataSource, params QueryParams, isPrimary bool) (string, map[string]interface{}, error) {
	// 1. Validate all metrics
	for _, metric := range ds.Metrics {
		if err := ValidateMetricConfig(metric); err != nil {
			return "", nil, fmt.Errorf("invalid metric %q: %w", metric.Name, err)
		}
	}

	// Validate all group by configs if present
	for i, groupBy := range ds.GroupBy {
		if err := ValidateGroupByConfig(&groupBy); err != nil {
			return "", nil, fmt.Errorf("invalid group by %d: %w", i, err)
		}
	}

	// Build FROM clause
	var from string
	if ds.Schema != "" {
		from = strings.Join([]string{ds.Schema, ds.Source}, ".")
	} else {
		from = ds.Source
	}

	query := qb.dialect.From(from)

	// 2. Build SELECT with aggregates
	var selectCols []interface{}

	// Add all group by columns to select if present
	for _, groupBy := range ds.GroupBy {
		groupBySelectExpr, err := qb.buildGroupBySelectExpression(&groupBy)
		if err != nil {
			return "", nil, fmt.Errorf("build group by select: %w", err)
		}
		selectCols = append(selectCols, groupBySelectExpr)
	}

	// Add metric expressions
	for _, metric := range ds.Metrics {
		metricExpr, err := qb.buildMetricExpression(metric)
		if err != nil {
			return "", nil, fmt.Errorf("build metric %q: %w", metric.Name, err)
		}
		selectCols = append(selectCols, metricExpr)
	}

	query = query.Select(selectCols...)

	// 3. Apply joins from foreign tables (needed for expressions referencing other tables)
	query = qb.applyForeignTableJoins(query, ds.Select.ForeignTables)

	// 4. Apply explicit joins
	query = qb.applyJoins(query, ds.Joins)

	// 5. Apply filters
	query = qb.applyFilters(query, ds.Filters, params)

	// 6. Build GROUP BY if present
	if len(ds.GroupBy) > 0 {
		groupByExprs, err := qb.buildGroupByClauses(ds.GroupBy)
		if err != nil {
			return "", nil, fmt.Errorf("build group by clauses: %w", err)
		}
		query = query.GroupBy(groupByExprs...)
	}

	// 7. Apply sorting (only for primary data source)
	if isPrimary {
		query = qb.applySorting(query, ds.Sort, params)
	}

	// 8. Apply row limit if set
	if ds.Rows > 0 {
		query = query.Limit(uint(ds.Rows))
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

// buildMetricExpression safely builds a metric SQL expression
func (qb *QueryBuilder) buildMetricExpression(metric MetricConfig) (interface{}, error) {
	var innerExpr string

	if metric.Column != "" {
		// Simple column reference
		innerExpr = metric.Column
	} else if metric.Expression != nil {
		// Build arithmetic expression from columns
		expr, err := qb.buildArithmeticExpression(metric.Expression)
		if err != nil {
			return nil, err
		}
		innerExpr = expr
	} else {
		return nil, fmt.Errorf("metric must have column or expression")
	}

	// Wrap with aggregate function
	sqlFunc, ok := AllowedAggregateFunctions[metric.Function]
	if !ok {
		return nil, fmt.Errorf("invalid aggregate function: %s", metric.Function)
	}

	var sqlExpr string
	switch metric.Function {
	case "count_distinct":
		// COUNT(DISTINCT column)
		sqlExpr = fmt.Sprintf("COUNT(DISTINCT %s)", innerExpr)
	default:
		// SUM(expr), AVG(expr), etc.
		sqlExpr = fmt.Sprintf("%s(%s)", sqlFunc, innerExpr)
	}

	// Add alias using goqu.L which returns LiteralExpression that has .As()
	return goqu.L(sqlExpr).As(goqu.C(metric.Name)), nil
}

// buildArithmeticExpression builds a safe arithmetic expression from columns
func (qb *QueryBuilder) buildArithmeticExpression(expr *ExpressionConfig) (string, error) {
	operator, ok := AllowedOperators[expr.Operator]
	if !ok {
		return "", fmt.Errorf("invalid operator: %s", expr.Operator)
	}

	// Build expression like "col1 * col2" or "col1 + col2 + col3"
	var parts []string
	for _, col := range expr.Columns {
		if !isValidColumnReference(col) {
			return "", fmt.Errorf("invalid column reference: %s", col)
		}
		parts = append(parts, col)
	}

	return strings.Join(parts, " "+operator+" "), nil
}

// buildGroupBySelectExpression builds the SELECT expression for group by column
func (qb *QueryBuilder) buildGroupBySelectExpression(groupBy *GroupByConfig) (interface{}, error) {
	// For SQL expressions, alias is required
	if groupBy.Expression {
		if groupBy.Alias == "" {
			return nil, fmt.Errorf("alias required for GROUP BY expression: %s", groupBy.Column)
		}
		// Use raw SQL expression
		return goqu.L(groupBy.Column).As(goqu.C(groupBy.Alias)), nil
	}

	alias := groupBy.Alias
	if alias == "" {
		// Use column name as alias if not specified
		parts := strings.Split(groupBy.Column, ".")
		alias = parts[len(parts)-1]
	}

	if groupBy.Interval != "" {
		// Time-based grouping: DATE_TRUNC('month', column) AS alias
		interval, ok := AllowedIntervals[groupBy.Interval]
		if !ok {
			return nil, fmt.Errorf("invalid interval: %s", groupBy.Interval)
		}
		expr := goqu.L(fmt.Sprintf("DATE_TRUNC('%s', %s)", interval, groupBy.Column))
		return expr.As(goqu.C(alias)), nil
	}

	// Categorical grouping: column AS alias
	expr := goqu.L(groupBy.Column)
	return expr.As(goqu.C(alias)), nil
}

// buildGroupByClause builds the GROUP BY clause expression
func (qb *QueryBuilder) buildGroupByClause(groupBy *GroupByConfig) (interface{}, error) {
	// For SQL expressions, use the expression directly in GROUP BY
	if groupBy.Expression {
		return goqu.L(groupBy.Column), nil
	}

	if groupBy.Interval != "" {
		// Time-based grouping: GROUP BY DATE_TRUNC('month', column)
		interval, ok := AllowedIntervals[groupBy.Interval]
		if !ok {
			return nil, fmt.Errorf("invalid interval: %s", groupBy.Interval)
		}
		return goqu.L(fmt.Sprintf("DATE_TRUNC('%s', %s)", interval, groupBy.Column)), nil
	}

	// Categorical grouping: GROUP BY column
	return goqu.L(groupBy.Column), nil
}

// buildGroupByClauses builds multiple GROUP BY clause expressions
func (qb *QueryBuilder) buildGroupByClauses(groupBys []GroupByConfig) ([]interface{}, error) {
	var exprs []interface{}

	for _, groupBy := range groupBys {
		expr, err := qb.buildGroupByClause(&groupBy)
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, expr)
	}

	return exprs, nil
}
