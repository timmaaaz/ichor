package tablebuilder

import (
	"fmt"
	"strings"

	"github.com/Knetic/govaluate"
)

// Evaluator handles expression evaluation for computed columns
type Evaluator struct {
	cache        map[string]*govaluate.EvaluableExpression
	maxCacheSize int // ADD: Prevent unbounded cache growth
}

// NewEvaluator creates a new expression evaluator
func NewEvaluator() *Evaluator {
	return &Evaluator{
		cache:        make(map[string]*govaluate.EvaluableExpression),
		maxCacheSize: 100, // ADD: Set a reasonable limit
	}
}

// Evaluate evaluates an expression with the given row context
func (e *Evaluator) Evaluate(expression string, row TableRow) (interface{}, error) {
	// ADD: Basic safety checks
	if len(expression) > 500 {
		return nil, fmt.Errorf("expression too long")
	}

	// ADD: Cache management
	if len(e.cache) > e.maxCacheSize {
		e.cache = make(map[string]*govaluate.EvaluableExpression)
	}

	// Transform the expression to work with govaluate
	transformedExpr := e.transformExpression(expression, row)

	// Check cache for compiled expression
	expr, exists := e.cache[transformedExpr]
	if !exists {
		var err error
		expr, err = govaluate.NewEvaluableExpression(transformedExpr)
		if err != nil {
			return nil, fmt.Errorf("parse expression: %w", err)
		}
		e.cache[transformedExpr] = expr
	}

	// Build parameters from row data
	params := e.buildParameters(row)

	// Evaluate the expression
	result, err := expr.Evaluate(params)
	if err != nil {
		return nil, fmt.Errorf("evaluate expression: %w", err)
	}

	return result, nil
}

// KEEP ALL YOUR EXISTING METHODS AS-IS, just modify flattenValue:

// flattenValue flattens nested structures for evaluation
func (e *Evaluator) flattenValue(prefix string, value interface{}, params map[string]interface{}) {
	e.flattenValueWithDepth(prefix, value, params, 0)
}

// ADD: New method with depth protection
func (e *Evaluator) flattenValueWithDepth(prefix string, value interface{}, params map[string]interface{}, depth int) {
	// ADD: Depth protection
	if depth > 5 {
		return
	}

	// Clean the prefix
	prefix = strings.ReplaceAll(prefix, ".", "_")

	switch v := value.(type) {
	case map[string]any:
		// ADD: Limit map size
		count := 0
		for k, nested := range v {
			if count > 100 { // Reasonable limit
				break
			}
			newKey := prefix + "_" + k
			e.flattenValueWithDepth(newKey, nested, params, depth+1)
			count++
		}
	case []any:
		// ADD: Limit array processing
		maxItems := 10
		if len(v) < maxItems {
			maxItems = len(v)
		}
		for i := 0; i < maxItems; i++ {
			newKey := fmt.Sprintf("%s_%d", prefix, i)
			e.flattenValueWithDepth(newKey, v[i], params, depth+1)
		}
	default:
		// Simple value
		params[prefix] = value
	}
}

// KEEP ALL YOUR OTHER METHODS UNCHANGED

// transformExpression transforms JavaScript-style expressions to govaluate format
func (e *Evaluator) transformExpression(expr string, row TableRow) string {
	// Replace row.field with just field
	result := strings.ReplaceAll(expr, "row.", "")

	// Handle array access like products.product_costs[0].purchase_cost
	// Convert to products_product_costs_0_purchase_cost
	result = e.replaceArrayAccess(result)

	// Handle ternary operator (? :) -> convert to if-then-else
	result = e.replaceTernary(result)

	// Handle null checks
	result = strings.ReplaceAll(result, "null", "nil")

	return result
}

// replaceArrayAccess replaces array access notation with underscores
func (e *Evaluator) replaceArrayAccess(expr string) string {
	// Simple replacement - in production you'd want proper parsing
	result := expr

	// Replace [0] with _0
	for i := 0; i < 10; i++ {
		old := fmt.Sprintf("[%d]", i)
		new := fmt.Sprintf("_%d", i)
		result = strings.ReplaceAll(result, old, new)
	}

	// Replace dots with underscores for nested access
	// But preserve dots in numbers
	parts := strings.Fields(result)
	for i, part := range parts {
		if !isNumeric(part) {
			parts[i] = strings.ReplaceAll(part, ".", "_")
		}
	}

	return strings.Join(parts, " ")
}

// replaceTernary converts ternary operators to govaluate conditionals
func (e *Evaluator) replaceTernary(expr string) string {
	// Simple ternary replacement
	// condition ? true_value : false_value
	// becomes: (condition) && (true_value) || (false_value)

	if !strings.Contains(expr, "?") || !strings.Contains(expr, ":") {
		return expr
	}

	// This is simplified - in production you'd need proper parsing
	parts := strings.Split(expr, "?")
	if len(parts) != 2 {
		return expr
	}

	condition := strings.TrimSpace(parts[0])
	rest := parts[1]

	valueParts := strings.Split(rest, ":")
	if len(valueParts) != 2 {
		return expr
	}

	trueValue := strings.TrimSpace(valueParts[0])
	falseValue := strings.TrimSpace(valueParts[1])

	// Use govaluate's ternary syntax
	return fmt.Sprintf("(%s) ? (%s) : (%s)", condition, trueValue, falseValue)
}

// buildParameters builds parameter map from row data
func (e *Evaluator) buildParameters(row TableRow) map[string]interface{} {
	params := make(map[string]interface{})

	// Flatten nested structures
	for key, value := range row {
		e.flattenValue(key, value, params)
	}

	return params
}

// CompileExpressions pre-compiles a list of expressions
func (e *Evaluator) CompileExpressions(expressions []ComputedColumn) error {
	for _, col := range expressions {
		// Pre-transform and compile
		// We can't fully transform without row data, but we can validate syntax
		_, err := govaluate.NewEvaluableExpression(col.Expression)
		if err != nil {
			return fmt.Errorf("compile expression %s: %w", col.Name, err)
		}
	}
	return nil
}

// Helper function to check if a string is numeric
func isNumeric(s string) bool {
	// Simple check - in production use strconv.ParseFloat
	for _, ch := range s {
		if (ch < '0' || ch > '9') && ch != '.' && ch != '-' {
			return false
		}
	}
	return true
}
