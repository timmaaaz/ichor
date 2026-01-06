package tablebuilder

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/Knetic/govaluate"
)

// Typed errors for expression evaluation
var (
	// ErrInvalidNumericArg is returned when a numeric function receives a non-numeric argument
	ErrInvalidNumericArg = errors.New("invalid numeric argument")

	// ErrInvalidDateFormat is returned when a date function receives an unparseable date string
	ErrInvalidDateFormat = errors.New("invalid date format")

	// ErrMissingArgument is returned when a function is called with too few arguments
	ErrMissingArgument = errors.New("missing required argument")

	// ErrNilArgument is returned when a function receives a nil argument where a value is required
	ErrNilArgument = errors.New("nil argument where value expected")
)

// Evaluator handles expression evaluation for computed columns
type Evaluator struct {
	cache        map[string]*govaluate.EvaluableExpression
	maxCacheSize int
	functions    map[string]govaluate.ExpressionFunction
}

// NewEvaluator creates a new expression evaluator
func NewEvaluator() *Evaluator {
	return &Evaluator{
		cache:        make(map[string]*govaluate.EvaluableExpression),
		maxCacheSize: 100,
		functions:    buildExpressionFunctions(),
	}
}

// buildExpressionFunctions returns custom functions for expression evaluation
func buildExpressionFunctions() map[string]govaluate.ExpressionFunction {
	return map[string]govaluate.ExpressionFunction{
		"ceil": func(args ...any) (any, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("%w: ceil requires 1 argument, got %d", ErrMissingArgument, len(args))
			}
			if args[0] == nil || isNilMarker(args[0]) {
				return nil, fmt.Errorf("%w: ceil argument must be numeric, got nil", ErrInvalidNumericArg)
			}
			val, err := toFloat64(args[0])
			if err != nil {
				return nil, fmt.Errorf("%w: ceil argument must be numeric, got %T", ErrInvalidNumericArg, args[0])
			}
			return math.Ceil(val), nil
		},
		"floor": func(args ...any) (any, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("%w: floor requires 1 argument, got %d", ErrMissingArgument, len(args))
			}
			if args[0] == nil || isNilMarker(args[0]) {
				return nil, fmt.Errorf("%w: floor argument must be numeric, got nil", ErrInvalidNumericArg)
			}
			val, err := toFloat64(args[0])
			if err != nil {
				return nil, fmt.Errorf("%w: floor argument must be numeric, got %T", ErrInvalidNumericArg, args[0])
			}
			return math.Floor(val), nil
		},
		"round": func(args ...any) (any, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("%w: round requires 1 argument, got %d", ErrMissingArgument, len(args))
			}
			if args[0] == nil || isNilMarker(args[0]) {
				return nil, fmt.Errorf("%w: round argument must be numeric, got nil", ErrInvalidNumericArg)
			}
			val, err := toFloat64(args[0])
			if err != nil {
				return nil, fmt.Errorf("%w: round argument must be numeric, got %T", ErrInvalidNumericArg, args[0])
			}
			return math.Round(val), nil
		},
		"now": func(args ...any) (any, error) {
			return float64(time.Now().Unix()), nil
		},
		"daysUntil": func(args ...any) (any, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("%w: daysUntil requires 1 argument", ErrMissingArgument)
			}
			if args[0] == nil || isNilMarker(args[0]) {
				return nil, fmt.Errorf("%w: daysUntil requires a date value", ErrNilArgument)
			}
			dateStr, ok := toString(args[0])
			if !ok {
				return nil, fmt.Errorf("%w: daysUntil argument must be string, got %T", ErrInvalidDateFormat, args[0])
			}
			t, err := parseDate(dateStr)
			if err != nil {
				return nil, fmt.Errorf("%w: %s", ErrInvalidDateFormat, err)
			}
			days := time.Until(t).Hours() / 24
			return math.Ceil(days), nil
		},
		"daysSince": func(args ...any) (any, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("%w: daysSince requires 1 argument", ErrMissingArgument)
			}
			if args[0] == nil || isNilMarker(args[0]) {
				return nil, fmt.Errorf("%w: daysSince requires a date value", ErrNilArgument)
			}
			dateStr, ok := toString(args[0])
			if !ok {
				return nil, fmt.Errorf("%w: daysSince argument must be string, got %T", ErrInvalidDateFormat, args[0])
			}
			t, err := parseDate(dateStr)
			if err != nil {
				return nil, fmt.Errorf("%w: %s", ErrInvalidDateFormat, err)
			}
			days := time.Since(t).Hours() / 24
			return math.Floor(days), nil
		},
		"isOverdue": func(args ...any) (any, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("%w: isOverdue requires 1 argument", ErrMissingArgument)
			}
			if args[0] == nil || isNilMarker(args[0]) {
				return nil, fmt.Errorf("%w: isOverdue requires a date value", ErrNilArgument)
			}
			dateStr, ok := toString(args[0])
			if !ok {
				return nil, fmt.Errorf("%w: isOverdue argument must be string, got %T", ErrInvalidDateFormat, args[0])
			}
			t, err := parseDate(dateStr)
			if err != nil {
				return nil, fmt.Errorf("%w: %s", ErrInvalidDateFormat, err)
			}
			return t.Before(time.Now()), nil
		},
		"hasValue": func(args ...any) (any, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("%w: hasValue requires 1 argument", ErrMissingArgument)
			}
			if args[0] == nil || isNilMarker(args[0]) {
				return false, nil
			}
			if s, ok := args[0].(string); ok {
				return s != "", nil
			}
			return true, nil
		},
	}
}

// toFloat64 converts any to float64 with strict type checking
func toFloat64(v any) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case nil:
		return 0, fmt.Errorf("nil value")
	default:
		return 0, fmt.Errorf("unsupported type %T", v)
	}
}

// toString converts any to string with type checking
func toString(v any) (string, bool) {
	if v == nil {
		return "", false
	}
	switch val := v.(type) {
	case string:
		return val, true
	case time.Time:
		return val.Format(time.RFC3339), true
	default:
		// Try fmt.Sprint as last resort
		return fmt.Sprintf("%v", v), true
	}
}

// parseDate parses a date string in multiple formats
func parseDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, fmt.Errorf("empty date string")
	}

	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse date '%s', expected formats: RFC3339, YYYY-MM-DD, YYYY-MM-DD HH:MM:SS", s)
}

// Evaluate evaluates an expression with the given row context
func (e *Evaluator) Evaluate(expression string, row TableRow) (any, error) {
	// Basic safety checks
	if len(expression) > 500 {
		return nil, fmt.Errorf("expression too long")
	}

	// Cache management
	if len(e.cache) > e.maxCacheSize {
		e.cache = make(map[string]*govaluate.EvaluableExpression)
	}

	// Transform the expression to work with govaluate
	transformedExpr := e.transformExpression(expression, row)

	// Check cache for compiled expression
	expr, exists := e.cache[transformedExpr]
	if !exists {
		var err error
		expr, err = govaluate.NewEvaluableExpressionWithFunctions(transformedExpr, e.functions)
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

// flattenValue flattens nested structures for evaluation
func (e *Evaluator) flattenValue(prefix string, value any, params map[string]any) {
	e.flattenValueWithDepth(prefix, value, params, 0)
}

// flattenValueWithDepth recursively flattens nested structures with depth protection
func (e *Evaluator) flattenValueWithDepth(prefix string, value any, params map[string]any, depth int) {
	// Depth protection
	if depth > 5 {
		return
	}

	// Clean the prefix
	prefix = strings.ReplaceAll(prefix, ".", "_")

	// Handle nil values with a marker since govaluate doesn't pass nil to functions
	if value == nil {
		params[prefix] = nilValue
		return
	}

	switch v := value.(type) {
	case map[string]any:
		// Limit map size
		count := 0
		for k, nested := range v {
			if count > 100 {
				break
			}
			newKey := prefix + "_" + k
			e.flattenValueWithDepth(newKey, nested, params, depth+1)
			count++
		}
	case []any:
		// Limit array processing
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

// transformExpression transforms JavaScript-style expressions to govaluate format
func (e *Evaluator) transformExpression(expr string, row TableRow) string {
	result := expr

	// Fix JavaScript strict inequality -> Go inequality
	result = strings.ReplaceAll(result, "!==", "!=")
	result = strings.ReplaceAll(result, "===", "==")

	// Replace Math functions with our custom functions
	result = strings.ReplaceAll(result, "Math.ceil", "ceil")
	result = strings.ReplaceAll(result, "Math.floor", "floor")
	result = strings.ReplaceAll(result, "Math.round", "round")

	// Replace row.field with just field
	result = strings.ReplaceAll(result, "row.", "")

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

// nilMarker is a special value used to represent nil in govaluate expressions.
// govaluate doesn't pass nil values as function arguments, so we use this marker
// and check for it in our custom functions.
type nilMarker struct{}

var nilValue = nilMarker{}

// isNilMarker checks if a value is our nil marker
func isNilMarker(v any) bool {
	_, ok := v.(nilMarker)
	return ok
}

// buildParameters builds parameter map from row data
func (e *Evaluator) buildParameters(row TableRow) map[string]any {
	params := make(map[string]any)

	// Add nil as a reserved parameter so expressions can use "nil" as a literal value
	// This allows ternary expressions like: hasValue(x) ? x : nil
	params["nil"] = nil

	// Flatten nested structures
	for key, value := range row {
		e.flattenValue(key, value, params)
	}

	return params
}

// CompileExpressions pre-compiles a list of expressions
func (e *Evaluator) CompileExpressions(expressions []ComputedColumn) error {
	for _, col := range expressions {
		// Pre-transform and compile with functions
		_, err := govaluate.NewEvaluableExpressionWithFunctions(col.Expression, e.functions)
		if err != nil {
			return fmt.Errorf("compile expression %s: %w", col.Name, err)
		}
	}
	return nil
}

// isNumeric checks if a string is numeric
func isNumeric(s string) bool {
	for _, ch := range s {
		if (ch < '0' || ch > '9') && ch != '.' && ch != '-' {
			return false
		}
	}
	return true
}
