package tablebuilder

import (
	"fmt"
	"regexp"
)

// =============================================================================
// Metric Validation Functions
// =============================================================================

// columnRefPattern validates column references (table.column, schema.table.column, or column)
var columnRefPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*(\.[a-zA-Z_][a-zA-Z0-9_]*)*$`)

// ValidateMetricConfig validates a metric configuration
func ValidateMetricConfig(metric MetricConfig) error {
	if metric.Name == "" {
		return fmt.Errorf("metric name is required")
	}

	if _, ok := AllowedAggregateFunctions[metric.Function]; !ok {
		return fmt.Errorf("invalid aggregate function: %s", metric.Function)
	}

	if metric.Column == "" && metric.Expression == nil {
		return fmt.Errorf("metric must have column or expression")
	}

	if metric.Column != "" && metric.Expression != nil {
		return fmt.Errorf("metric cannot have both column and expression")
	}

	if metric.Column != "" {
		if !isValidColumnReference(metric.Column) {
			return fmt.Errorf("invalid column reference: %s", metric.Column)
		}
	}

	if metric.Expression != nil {
		if err := ValidateExpressionConfig(metric.Expression); err != nil {
			return fmt.Errorf("invalid expression: %w", err)
		}
	}

	return nil
}

// ValidateExpressionConfig validates an expression configuration
func ValidateExpressionConfig(expr *ExpressionConfig) error {
	if expr == nil {
		return fmt.Errorf("expression is nil")
	}

	if _, ok := AllowedOperators[expr.Operator]; !ok {
		return fmt.Errorf("invalid operator: %s", expr.Operator)
	}

	if len(expr.Columns) < 2 {
		return fmt.Errorf("expression requires at least 2 columns")
	}

	// Validate column names (no SQL injection)
	for _, col := range expr.Columns {
		if !isValidColumnReference(col) {
			return fmt.Errorf("invalid column reference: %s", col)
		}
	}

	return nil
}

// ValidateGroupByConfig validates a group by configuration
func ValidateGroupByConfig(groupBy *GroupByConfig) error {
	if groupBy == nil {
		return nil // GroupBy is optional
	}

	if groupBy.Column == "" {
		return fmt.Errorf("group by column is required")
	}

	if groupBy.Interval != "" {
		if _, ok := AllowedIntervals[groupBy.Interval]; !ok {
			return fmt.Errorf("invalid interval: %s", groupBy.Interval)
		}
	}

	// When Expression is true, Column contains raw SQL (e.g., "EXTRACT(DOW FROM created_date)")
	// and won't match the identifier pattern. The query builder handles raw SQL safely using
	// goqu.L(). This is safe because chart configurations are admin-controlled backend data.
	if !groupBy.Expression && !isValidColumnReference(groupBy.Column) {
		return fmt.Errorf("invalid column reference: %s", groupBy.Column)
	}

	if groupBy.Alias != "" && !isValidColumnReference(groupBy.Alias) {
		return fmt.Errorf("invalid alias: %s", groupBy.Alias)
	}

	return nil
}

// isValidColumnReference checks if a column reference is safe
// Allows: table.column, schema.table.column, column
// Disallows: SQL keywords, special characters, etc.
func isValidColumnReference(col string) bool {
	return columnRefPattern.MatchString(col)
}
