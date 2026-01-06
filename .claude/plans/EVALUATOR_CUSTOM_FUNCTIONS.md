# Fix Expression Evaluation for Computed Columns

## Type Safety Approach

We will use **strict typing with typed errors**. This provides:
- Clear error messages for debugging (e.g., `ErrInvalidDateFormat`, `ErrInvalidNumericArg`)
- Consistency with the rest of the Ichor codebase which uses typed errors
- Safe fallback: evaluator logs errors and sets value to `nil` (no crash)

## Root Cause Analysis

The computed column expressions in the seed data are written in **JavaScript syntax**, but the evaluator uses **govaluate** - a Go expression evaluation library that does NOT support JavaScript constructs.

### Specific Errors:
1. **`Undefined function Math_ceil`** - govaluate has NO built-in functions (Math.ceil, Math.floor, Math.round, etc.)
2. **`Invalid token: '!=='`** - govaluate only supports `!=`, not JavaScript's `!==`
3. **`new Date()` fails** - govaluate doesn't support JavaScript constructors

### Problematic Expressions Found:

| Location | Expression | Issue |
|----------|-----------|-------|
| tables.go:497 | `Math.ceil((new Date(due_date) - new Date()) / ...)` | Math.ceil + new Date |
| tables.go:501 | `new Date(due_date) < new Date() && status_name !== 'Delivered'` | new Date + !== |
| tables.go:711 | `Math.round(rating * 2) / 2` | Math.round |
| tables.go:3355 | `actual_delivery_date ? 'delivered' : (new Date(...) < new Date() ? ...)` | new Date |
| tables.go:4115 | `Math.floor((new Date() - new Date(order_date)) / ...)` | Math.floor + new Date |
| tables.go:4470 | same as 3355 | new Date |
| model.go:747 | `date_hired ? Math.floor((new Date() - new Date(date_hired)) / ...) : null` | Math.floor + new Date |

## Solution: Add Custom Functions to govaluate

govaluate supports custom functions via `NewEvaluableExpressionWithFunctions()`. We need to:

1. **Add custom functions** to the evaluator for: `ceil`, `floor`, `round`, `now`, `parseDate`, `dateDiff`
2. **Transform expressions** to use the new function syntax instead of JavaScript syntax
3. **Fix operator syntax** (replace `!==` with `!=`)

## Implementation Plan

### Step 1: Update Evaluator with Custom Functions
**File**: `business/sdk/tablebuilder/evaluator.go`

Add a functions map with:
- `ceil(x)` - rounds up to nearest integer
- `floor(x)` - rounds down to nearest integer
- `round(x)` - rounds to nearest integer
- `now()` - returns current Unix timestamp in seconds
- `daysUntil(dateStr)` - returns days until given date (positive = future, negative = past)
- `daysSince(dateStr)` - returns days since given date (positive = past, negative = future)
- `isOverdue(dateStr)` - returns true if date is in the past
- `hasValue(val)` - returns true if value is not nil/empty (for null checks)

**Add typed errors:**
```go
import "errors"

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
```

**Changes to Evaluator struct:**
```go
type Evaluator struct {
    cache        map[string]*govaluate.EvaluableExpression
    maxCacheSize int
    functions    map[string]govaluate.ExpressionFunction  // ADD THIS
}
```

**Add function builder with strict type checking:**
```go
func buildExpressionFunctions() map[string]govaluate.ExpressionFunction {
    return map[string]govaluate.ExpressionFunction{
        "ceil": func(args ...interface{}) (interface{}, error) {
            if len(args) != 1 {
                return nil, fmt.Errorf("%w: ceil requires 1 argument, got %d", ErrMissingArgument, len(args))
            }
            val, err := toFloat64(args[0])
            if err != nil {
                return nil, fmt.Errorf("%w: ceil argument must be numeric, got %T", ErrInvalidNumericArg, args[0])
            }
            return math.Ceil(val), nil
        },
        "floor": func(args ...interface{}) (interface{}, error) {
            if len(args) != 1 {
                return nil, fmt.Errorf("%w: floor requires 1 argument, got %d", ErrMissingArgument, len(args))
            }
            val, err := toFloat64(args[0])
            if err != nil {
                return nil, fmt.Errorf("%w: floor argument must be numeric, got %T", ErrInvalidNumericArg, args[0])
            }
            return math.Floor(val), nil
        },
        "round": func(args ...interface{}) (interface{}, error) {
            if len(args) != 1 {
                return nil, fmt.Errorf("%w: round requires 1 argument, got %d", ErrMissingArgument, len(args))
            }
            val, err := toFloat64(args[0])
            if err != nil {
                return nil, fmt.Errorf("%w: round argument must be numeric, got %T", ErrInvalidNumericArg, args[0])
            }
            return math.Round(val), nil
        },
        "now": func(args ...interface{}) (interface{}, error) {
            return float64(time.Now().Unix()), nil
        },
        "daysUntil": func(args ...interface{}) (interface{}, error) {
            if len(args) != 1 {
                return nil, fmt.Errorf("%w: daysUntil requires 1 argument", ErrMissingArgument)
            }
            if args[0] == nil {
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
        "daysSince": func(args ...interface{}) (interface{}, error) {
            if len(args) != 1 {
                return nil, fmt.Errorf("%w: daysSince requires 1 argument", ErrMissingArgument)
            }
            if args[0] == nil {
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
        "isOverdue": func(args ...interface{}) (interface{}, error) {
            if len(args) != 1 {
                return nil, fmt.Errorf("%w: isOverdue requires 1 argument", ErrMissingArgument)
            }
            if args[0] == nil {
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
        "hasValue": func(args ...interface{}) (interface{}, error) {
            if len(args) != 1 {
                return nil, fmt.Errorf("%w: hasValue requires 1 argument", ErrMissingArgument)
            }
            if args[0] == nil {
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
// Note: govaluate uses interface{} in its API, but we use 'any' in our helpers
// since they're equivalent (any is an alias for interface{} in Go 1.18+)
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
```

**Update NewEvaluator:**
```go
func NewEvaluator() *Evaluator {
    return &Evaluator{
        cache:        make(map[string]*govaluate.EvaluableExpression),
        maxCacheSize: 100,
        functions:    buildExpressionFunctions(),  // ADD THIS
    }
}
```

**Update Evaluate() to use functions:**
```go
// Change line 43 from:
expr, err = govaluate.NewEvaluableExpression(transformedExpr)
// To:
expr, err = govaluate.NewEvaluableExpressionWithFunctions(transformedExpr, e.functions)
```

### Step 2: Update Expression Transformation
**File**: `business/sdk/tablebuilder/evaluator.go`

Update `transformExpression()`:
```go
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
    result = e.replaceArrayAccess(result)

    // Handle ternary operator
    result = e.replaceTernary(result)

    // Handle null checks
    result = strings.ReplaceAll(result, "null", "nil")

    return result
}
```

### Step 3: Update Seed Data Expressions
**Files**:
- `business/sdk/dbtest/seedmodels/tables.go`
- `business/sdk/dbtest/seedmodels/model.go`

| Old Expression | New Expression |
|---------------|----------------|
| `Math.ceil((new Date(due_date) - new Date()) / (1000 * 60 * 60 * 24))` | `daysUntil(due_date)` |
| `new Date(due_date) < new Date() && status_name !== 'Delivered'` | `isOverdue(due_date) && status_name != 'Delivered'` |
| `Math.round(rating * 2) / 2` | `round(rating * 2) / 2` |
| `Math.floor((new Date() - new Date(order_date)) / (1000 * 60 * 60 * 24))` | `daysSince(order_date)` |
| `actual_delivery_date ? 'delivered' : (new Date(expected_delivery_date) < new Date() ? 'overdue' : 'pending')` | `hasValue(actual_delivery_date) ? 'delivered' : (isOverdue(expected_delivery_date) ? 'overdue' : 'pending')` |
| `date_hired ? Math.floor((new Date() - new Date(date_hired)) / (1000 * 60 * 60 * 24)) : null` | `hasValue(date_hired) ? daysSince(date_hired) : nil` |

## Files to Modify

1. **business/sdk/tablebuilder/evaluator.go** - Add custom functions and update transformation
2. **business/sdk/dbtest/seedmodels/tables.go** - Update 5 computed column expressions (lines 497, 501, 711, 3355, 4115, 4470)
3. **business/sdk/dbtest/seedmodels/model.go** - Update 1 computed column expression (line 747)

## Tests to Add

### New Test File: `business/sdk/tablebuilder/evaluator_test.go`

```go
package tablebuilder_test

import (
    "testing"
    "time"

    "github.com/timmaaaz/ichor/business/sdk/tablebuilder"
)

func Test_EvaluatorFunctions(t *testing.T) {
    t.Parallel()

    eval := tablebuilder.NewEvaluator()

    t.Run("ceil function", func(t *testing.T) {
        row := tablebuilder.TableRow{"value": 3.2}
        result, err := eval.Evaluate("ceil(value)", row)
        if err != nil {
            t.Fatalf("evaluate failed: %v", err)
        }
        if result != 4.0 {
            t.Errorf("ceil(3.2) = %v, want 4", result)
        }
    })

    t.Run("floor function", func(t *testing.T) {
        row := tablebuilder.TableRow{"value": 3.8}
        result, err := eval.Evaluate("floor(value)", row)
        if err != nil {
            t.Fatalf("evaluate failed: %v", err)
        }
        if result != 3.0 {
            t.Errorf("floor(3.8) = %v, want 3", result)
        }
    })

    t.Run("round function", func(t *testing.T) {
        row := tablebuilder.TableRow{"value": 3.5}
        result, err := eval.Evaluate("round(value)", row)
        if err != nil {
            t.Fatalf("evaluate failed: %v", err)
        }
        if result != 4.0 {
            t.Errorf("round(3.5) = %v, want 4", result)
        }
    })

    t.Run("now function", func(t *testing.T) {
        row := tablebuilder.TableRow{}
        result, err := eval.Evaluate("now()", row)
        if err != nil {
            t.Fatalf("evaluate failed: %v", err)
        }
        nowUnix := float64(time.Now().Unix())
        resultFloat := result.(float64)
        // Allow 2 second tolerance
        if resultFloat < nowUnix-2 || resultFloat > nowUnix+2 {
            t.Errorf("now() = %v, want ~%v", resultFloat, nowUnix)
        }
    })

    t.Run("daysUntil with future date", func(t *testing.T) {
        futureDate := time.Now().AddDate(0, 0, 5).Format("2006-01-02")
        row := tablebuilder.TableRow{"due_date": futureDate}
        result, err := eval.Evaluate("daysUntil(due_date)", row)
        if err != nil {
            t.Fatalf("evaluate failed: %v", err)
        }
        resultFloat := result.(float64)
        if resultFloat < 4 || resultFloat > 6 {
            t.Errorf("daysUntil(5 days from now) = %v, want ~5", resultFloat)
        }
    })

    t.Run("daysUntil with past date returns negative", func(t *testing.T) {
        pastDate := time.Now().AddDate(0, 0, -3).Format("2006-01-02")
        row := tablebuilder.TableRow{"due_date": pastDate}
        result, err := eval.Evaluate("daysUntil(due_date)", row)
        if err != nil {
            t.Fatalf("evaluate failed: %v", err)
        }
        resultFloat := result.(float64)
        if resultFloat > -2 || resultFloat < -4 {
            t.Errorf("daysUntil(3 days ago) = %v, want ~-3", resultFloat)
        }
    })

    t.Run("daysSince with past date", func(t *testing.T) {
        pastDate := time.Now().AddDate(0, 0, -5).Format("2006-01-02")
        row := tablebuilder.TableRow{"order_date": pastDate}
        result, err := eval.Evaluate("daysSince(order_date)", row)
        if err != nil {
            t.Fatalf("evaluate failed: %v", err)
        }
        resultFloat := result.(float64)
        if resultFloat < 4 || resultFloat > 6 {
            t.Errorf("daysSince(5 days ago) = %v, want ~5", resultFloat)
        }
    })

    t.Run("isOverdue with past date returns true", func(t *testing.T) {
        pastDate := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
        row := tablebuilder.TableRow{"due_date": pastDate}
        result, err := eval.Evaluate("isOverdue(due_date)", row)
        if err != nil {
            t.Fatalf("evaluate failed: %v", err)
        }
        if result != true {
            t.Errorf("isOverdue(yesterday) = %v, want true", result)
        }
    })

    t.Run("isOverdue with future date returns false", func(t *testing.T) {
        futureDate := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
        row := tablebuilder.TableRow{"due_date": futureDate}
        result, err := eval.Evaluate("isOverdue(due_date)", row)
        if err != nil {
            t.Fatalf("evaluate failed: %v", err)
        }
        if result != false {
            t.Errorf("isOverdue(tomorrow) = %v, want false", result)
        }
    })

    t.Run("hasValue with non-nil value returns true", func(t *testing.T) {
        row := tablebuilder.TableRow{"delivery_date": "2025-01-15"}
        result, err := eval.Evaluate("hasValue(delivery_date)", row)
        if err != nil {
            t.Fatalf("evaluate failed: %v", err)
        }
        if result != true {
            t.Errorf("hasValue('2025-01-15') = %v, want true", result)
        }
    })

    t.Run("hasValue with nil returns false", func(t *testing.T) {
        row := tablebuilder.TableRow{"delivery_date": nil}
        result, err := eval.Evaluate("hasValue(delivery_date)", row)
        if err != nil {
            t.Fatalf("evaluate failed: %v", err)
        }
        if result != false {
            t.Errorf("hasValue(nil) = %v, want false", result)
        }
    })

    t.Run("hasValue with empty string returns false", func(t *testing.T) {
        row := tablebuilder.TableRow{"delivery_date": ""}
        result, err := eval.Evaluate("hasValue(delivery_date)", row)
        if err != nil {
            t.Fatalf("evaluate failed: %v", err)
        }
        if result != false {
            t.Errorf("hasValue('') = %v, want false", result)
        }
    })
}

func Test_EvaluatorTransformations(t *testing.T) {
    t.Parallel()

    eval := tablebuilder.NewEvaluator()

    t.Run("transforms !== to !=", func(t *testing.T) {
        row := tablebuilder.TableRow{"status": "Pending"}
        result, err := eval.Evaluate("status !== 'Delivered'", row)
        if err != nil {
            t.Fatalf("evaluate failed: %v", err)
        }
        if result != true {
            t.Errorf("status !== 'Delivered' with status='Pending' = %v, want true", result)
        }
    })

    t.Run("transforms === to ==", func(t *testing.T) {
        row := tablebuilder.TableRow{"status": "Delivered"}
        result, err := eval.Evaluate("status === 'Delivered'", row)
        if err != nil {
            t.Fatalf("evaluate failed: %v", err)
        }
        if result != true {
            t.Errorf("status === 'Delivered' with status='Delivered' = %v, want true", result)
        }
    })

    t.Run("transforms Math.ceil to ceil", func(t *testing.T) {
        row := tablebuilder.TableRow{"value": 3.2}
        result, err := eval.Evaluate("Math.ceil(value)", row)
        if err != nil {
            t.Fatalf("evaluate failed: %v", err)
        }
        if result != 4.0 {
            t.Errorf("Math.ceil(3.2) = %v, want 4", result)
        }
    })

    t.Run("transforms Math.floor to floor", func(t *testing.T) {
        row := tablebuilder.TableRow{"value": 3.8}
        result, err := eval.Evaluate("Math.floor(value)", row)
        if err != nil {
            t.Fatalf("evaluate failed: %v", err)
        }
        if result != 3.0 {
            t.Errorf("Math.floor(3.8) = %v, want 3", result)
        }
    })

    t.Run("transforms Math.round to round", func(t *testing.T) {
        row := tablebuilder.TableRow{"value": 3.5}
        result, err := eval.Evaluate("Math.round(value)", row)
        if err != nil {
            t.Fatalf("evaluate failed: %v", err)
        }
        if result != 4.0 {
            t.Errorf("Math.round(3.5) = %v, want 4", result)
        }
    })
}

func Test_EvaluatorComplexExpressions(t *testing.T) {
    t.Parallel()

    eval := tablebuilder.NewEvaluator()

    t.Run("isOverdue combined with status check", func(t *testing.T) {
        pastDate := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

        // Overdue and not delivered -> true
        row := tablebuilder.TableRow{
            "due_date":    pastDate,
            "status_name": "Pending",
        }
        result, err := eval.Evaluate("isOverdue(due_date) && status_name != 'Delivered'", row)
        if err != nil {
            t.Fatalf("evaluate failed: %v", err)
        }
        if result != true {
            t.Errorf("isOverdue && not delivered = %v, want true", result)
        }

        // Overdue but delivered -> false
        row["status_name"] = "Delivered"
        result, err = eval.Evaluate("isOverdue(due_date) && status_name != 'Delivered'", row)
        if err != nil {
            t.Fatalf("evaluate failed: %v", err)
        }
        if result != false {
            t.Errorf("isOverdue && delivered = %v, want false", result)
        }
    })

    t.Run("ternary with hasValue for delivery status", func(t *testing.T) {
        pastDate := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

        // Has delivery date -> delivered
        row := tablebuilder.TableRow{
            "actual_delivery_date":     "2025-01-10",
            "expected_delivery_date":   pastDate,
        }
        result, err := eval.Evaluate("hasValue(actual_delivery_date) ? 'delivered' : (isOverdue(expected_delivery_date) ? 'overdue' : 'pending')", row)
        if err != nil {
            t.Fatalf("evaluate failed: %v", err)
        }
        if result != "delivered" {
            t.Errorf("with delivery date = %v, want 'delivered'", result)
        }

        // No delivery date, past expected -> overdue
        row["actual_delivery_date"] = nil
        result, err = eval.Evaluate("hasValue(actual_delivery_date) ? 'delivered' : (isOverdue(expected_delivery_date) ? 'overdue' : 'pending')", row)
        if err != nil {
            t.Fatalf("evaluate failed: %v", err)
        }
        if result != "overdue" {
            t.Errorf("no delivery, past expected = %v, want 'overdue'", result)
        }

        // No delivery date, future expected -> pending
        futureDate := time.Now().AddDate(0, 0, 5).Format("2006-01-02")
        row["expected_delivery_date"] = futureDate
        result, err = eval.Evaluate("hasValue(actual_delivery_date) ? 'delivered' : (isOverdue(expected_delivery_date) ? 'overdue' : 'pending')", row)
        if err != nil {
            t.Fatalf("evaluate failed: %v", err)
        }
        if result != "pending" {
            t.Errorf("no delivery, future expected = %v, want 'pending'", result)
        }
    })

    t.Run("round rating expression", func(t *testing.T) {
        row := tablebuilder.TableRow{"rating": 3.7}
        result, err := eval.Evaluate("round(rating * 2) / 2", row)
        if err != nil {
            t.Fatalf("evaluate failed: %v", err)
        }
        // 3.7 * 2 = 7.4, round = 7, / 2 = 3.5
        if result != 3.5 {
            t.Errorf("round(3.7 * 2) / 2 = %v, want 3.5", result)
        }
    })

    t.Run("daysSince with ternary for tenure", func(t *testing.T) {
        pastDate := time.Now().AddDate(0, 0, -30).Format("2006-01-02")

        // Has date_hired -> show days
        row := tablebuilder.TableRow{"date_hired": pastDate}
        result, err := eval.Evaluate("hasValue(date_hired) ? daysSince(date_hired) : nil", row)
        if err != nil {
            t.Fatalf("evaluate failed: %v", err)
        }
        resultFloat := result.(float64)
        if resultFloat < 29 || resultFloat > 31 {
            t.Errorf("daysSince(30 days ago) = %v, want ~30", resultFloat)
        }

        // No date_hired -> nil
        row["date_hired"] = nil
        result, err = eval.Evaluate("hasValue(date_hired) ? daysSince(date_hired) : nil", row)
        if err != nil {
            t.Fatalf("evaluate failed: %v", err)
        }
        if result != nil {
            t.Errorf("daysSince with nil date_hired = %v, want nil", result)
        }
    })
}

func Test_EvaluatorDateFormats(t *testing.T) {
    t.Parallel()

    eval := tablebuilder.NewEvaluator()

    testCases := []struct {
        name   string
        format string
    }{
        {"RFC3339", time.Now().AddDate(0, 0, 5).Format(time.RFC3339)},
        {"RFC3339 with Z", time.Now().AddDate(0, 0, 5).Format("2006-01-02T15:04:05Z")},
        {"datetime with space", time.Now().AddDate(0, 0, 5).Format("2006-01-02 15:04:05")},
        {"date only", time.Now().AddDate(0, 0, 5).Format("2006-01-02")},
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            row := tablebuilder.TableRow{"date": tc.format}
            result, err := eval.Evaluate("daysUntil(date)", row)
            if err != nil {
                t.Fatalf("evaluate failed for format %s: %v", tc.name, err)
            }
            resultFloat := result.(float64)
            if resultFloat < 4 || resultFloat > 6 {
                t.Errorf("daysUntil with format %s = %v, want ~5", tc.name, resultFloat)
            }
        })
    }
}

func Test_EvaluatorTypeErrors(t *testing.T) {
    t.Parallel()

    eval := tablebuilder.NewEvaluator()

    t.Run("ceil with non-numeric returns ErrInvalidNumericArg", func(t *testing.T) {
        row := tablebuilder.TableRow{"value": "not-a-number"}
        _, err := eval.Evaluate("ceil(value)", row)
        if err == nil {
            t.Fatal("expected error for non-numeric ceil argument")
        }
        if !errors.Is(err, tablebuilder.ErrInvalidNumericArg) {
            t.Errorf("expected ErrInvalidNumericArg, got: %v", err)
        }
    })

    t.Run("ceil with nil returns ErrInvalidNumericArg", func(t *testing.T) {
        row := tablebuilder.TableRow{"value": nil}
        _, err := eval.Evaluate("ceil(value)", row)
        if err == nil {
            t.Fatal("expected error for nil ceil argument")
        }
        if !errors.Is(err, tablebuilder.ErrInvalidNumericArg) {
            t.Errorf("expected ErrInvalidNumericArg, got: %v", err)
        }
    })

    t.Run("daysUntil with nil returns ErrNilArgument", func(t *testing.T) {
        row := tablebuilder.TableRow{"date": nil}
        _, err := eval.Evaluate("daysUntil(date)", row)
        if err == nil {
            t.Fatal("expected error for nil date")
        }
        if !errors.Is(err, tablebuilder.ErrNilArgument) {
            t.Errorf("expected ErrNilArgument, got: %v", err)
        }
    })

    t.Run("daysUntil with invalid date returns ErrInvalidDateFormat", func(t *testing.T) {
        row := tablebuilder.TableRow{"date": "not-a-date"}
        _, err := eval.Evaluate("daysUntil(date)", row)
        if err == nil {
            t.Fatal("expected error for invalid date format")
        }
        if !errors.Is(err, tablebuilder.ErrInvalidDateFormat) {
            t.Errorf("expected ErrInvalidDateFormat, got: %v", err)
        }
    })

    t.Run("isOverdue with invalid date returns ErrInvalidDateFormat", func(t *testing.T) {
        row := tablebuilder.TableRow{"date": "January 5th, 2025"}
        _, err := eval.Evaluate("isOverdue(date)", row)
        if err == nil {
            t.Fatal("expected error for invalid date format")
        }
        if !errors.Is(err, tablebuilder.ErrInvalidDateFormat) {
            t.Errorf("expected ErrInvalidDateFormat, got: %v", err)
        }
    })

    t.Run("daysSince with empty string returns ErrInvalidDateFormat", func(t *testing.T) {
        row := tablebuilder.TableRow{"date": ""}
        _, err := eval.Evaluate("daysSince(date)", row)
        if err == nil {
            t.Fatal("expected error for empty date string")
        }
        if !errors.Is(err, tablebuilder.ErrInvalidDateFormat) {
            t.Errorf("expected ErrInvalidDateFormat, got: %v", err)
        }
    })

    t.Run("error message contains helpful context", func(t *testing.T) {
        row := tablebuilder.TableRow{"date": "bad-date"}
        _, err := eval.Evaluate("daysUntil(date)", row)
        if err == nil {
            t.Fatal("expected error")
        }
        errMsg := err.Error()
        if !strings.Contains(errMsg, "RFC3339") && !strings.Contains(errMsg, "YYYY-MM-DD") {
            t.Errorf("error message should mention expected formats: %v", errMsg)
        }
    })
}

func Test_EvaluatorEdgeCases(t *testing.T) {
    t.Parallel()

    eval := tablebuilder.NewEvaluator()

    t.Run("arithmetic with dates", func(t *testing.T) {
        row := tablebuilder.TableRow{"value": 100.0}
        result, err := eval.Evaluate("value * 2 + ceil(value / 3)", row)
        if err != nil {
            t.Fatalf("evaluate failed: %v", err)
        }
        // 100 * 2 + ceil(33.33) = 200 + 34 = 234
        if result != 234.0 {
            t.Errorf("100 * 2 + ceil(100/3) = %v, want 234", result)
        }
    })

    t.Run("nested ternary", func(t *testing.T) {
        row := tablebuilder.TableRow{"status": "pending", "priority": "high"}
        result, err := eval.Evaluate("status == 'complete' ? 'done' : (priority == 'high' ? 'urgent' : 'normal')", row)
        if err != nil {
            t.Fatalf("evaluate failed: %v", err)
        }
        if result != "urgent" {
            t.Errorf("nested ternary = %v, want 'urgent'", result)
        }
    })
}
```

### Integration Test Updates

Add test case to existing `Test_TableBuilder` in `business/sdk/tablebuilder/tablebuilder_test.go`:

```go
// Add this test case to verify computed columns with date functions work
func computedColumnsWithDateFunctionsTest(ctx context.Context, store *tablebuilder.Store) {
    config := &tablebuilder.Config{
        Title:         "Orders with Date Calculations",
        WidgetType:    "table",
        Visualization: "table",
        DataSource: []tablebuilder.DataSource{
            {
                Type:   "view",
                Source: "orders_base",
                Schema: "sales",
                Select: tablebuilder.SelectConfig{
                    Columns: []tablebuilder.ColumnDefinition{
                        {Name: "orders_id", TableColumn: "orders.id"},
                        {Name: "orders_due_date", Alias: "due_date", TableColumn: "orders.due_date"},
                        {Name: "order_fulfillment_statuses_name", Alias: "status_name", TableColumn: "order_fulfillment_statuses.name"},
                    },
                    ClientComputedColumns: []tablebuilder.ComputedColumn{
                        {
                            Name:       "days_until_due",
                            Expression: "daysUntil(due_date)",
                        },
                        {
                            Name:       "is_overdue",
                            Expression: "isOverdue(due_date) && status_name != 'Delivered'",
                        },
                    },
                },
                Rows: 10,
            },
        },
    }

    params := tablebuilder.QueryParams{Page: 1, Rows: 10}
    result, err := store.FetchTableData(ctx, config, params)
    if err != nil {
        log.Printf("Error fetching data with date functions: %v", err)
        return
    }

    // Verify computed columns exist and have values
    for i, row := range result.Data {
        if i >= 3 {
            break
        }
        daysUntil := row["days_until_due"]
        isOverdue := row["is_overdue"]

        // Both should be non-nil
        if daysUntil == nil {
            log.Printf("Row %d: days_until_due is nil", i)
        }
        if isOverdue == nil {
            log.Printf("Row %d: is_overdue is nil", i)
        }
    }

    fmt.Printf("\n=== Computed Columns with Date Functions Test ===\n")
    fullJSON, _ := json.MarshalIndent(result, "", "  ")
    fmt.Printf("Full JSON result:\n%s\n\n", fullJSON)
}
```

## Testing Plan

After implementation:
1. Run `make test` to verify no regressions
2. Run `go test -v ./business/sdk/tablebuilder/... -run Test_Evaluator` to verify new function tests
3. Load http://localhost:5173/sales and verify:
   - No expression errors in logs
   - Computed columns display correct values
   - `days_until_due` shows correct day counts
   - `is_overdue` shows correct boolean values

## Summary of Changes

| File | Changes |
|------|---------|
| `business/sdk/tablebuilder/evaluator.go` | Add custom functions, update transformations |
| `business/sdk/tablebuilder/evaluator_test.go` | NEW - Unit tests for all functions |
| `business/sdk/dbtest/seedmodels/tables.go` | Update 6 expression strings |
| `business/sdk/dbtest/seedmodels/model.go` | Update 1 expression string |
| `business/sdk/tablebuilder/tablebuilder_test.go` | Add integration test for date functions |
