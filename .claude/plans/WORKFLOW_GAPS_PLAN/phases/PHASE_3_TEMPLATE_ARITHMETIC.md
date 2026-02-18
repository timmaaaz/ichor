# Phase 9: Add Arithmetic to Template Processor

**Category**: Backend
**Status**: Pending
**Dependencies**: None
**Effort**: Medium

---

## Overview

The current template processor in `business/sdk/workflow/template.go` supports:
- Variable substitution: `{{variable_name}}`
- Pipe filters: `{{total | currency:USD}}`, `{{date | formatDate:short}}`

It does **not** support arithmetic expressions. This means workflows cannot compute calculated fields:
- `{{quantity * unit_price}}` for line item totals
- `{{subtotal + tax_amount + shipping_cost}}` for order totals
- `{{reorder_quantity - available_quantity}}` for replenishment amounts

This phase adds `{{expr: ...}}` syntax to the template processor using a safe expression evaluator.

---

## Goals

1. Add `{{expr: <math expression>}}` syntax to TemplateProcessor
2. Variable names in expressions resolve from `ActionExecutionContext.RawData`
3. Support basic arithmetic: `+`, `-`, `*`, `/`, `%`, parentheses
4. Result is returned as a string (subject to subsequent pipe filters)
5. Fail-open: unknown variables or errors return the original expression string

---

## Library Decision

**Option A: `github.com/expr-lang/expr`**
- Purpose-built safe expression evaluator for Go
- ~1.5k stars, MIT license, well-maintained
- Supports typed environments (can restrict to float64 variables)
- No access to arbitrary Go functions, no reflection, no network

**Option B: Custom minimal parser**
- Write a simple recursive descent parser for `+`, `-`, `*`, `/`, `%`, parens
- No new dependencies
- ~100-150 lines, well-tested
- Limited to numeric arithmetic (no string ops, comparisons)

**Recommendation: Option B (custom parser)** — avoids a new dependency for a focused use case. The scope is intentionally limited to numeric arithmetic. A custom parser is ~100 lines and can be fully unit tested.

---

## Task Breakdown

### Task 1: Choose and Implement Expression Evaluator

If using `expr-lang/expr`:
```bash
go get github.com/expr-lang/expr
```

If using custom parser (recommended), create:

**New file**: `business/sdk/workflow/expr.go`

```go
package workflow

import (
    "fmt"
    "math"
    "strconv"
    "strings"
    "unicode"
)

// EvalExpr evaluates a simple arithmetic expression with variables from the given map.
// Supports: +, -, *, /, %, parentheses, float literals, and variable names.
// Variable names are resolved from the vars map (values must be numeric or string-numeric).
// On any error (unknown variable, division by zero, parse error), returns ("", err).
func EvalExpr(expr string, vars map[string]any) (float64, error) {
    p := &exprParser{input: strings.TrimSpace(expr), vars: vars}
    result, err := p.parseExpr()
    if err != nil {
        return 0, err
    }
    return result, nil
}

type exprParser struct {
    input string
    pos   int
    vars  map[string]any
}

func (p *exprParser) parseExpr() (float64, error) {
    return p.parseAddSub()
}

func (p *exprParser) parseAddSub() (float64, error) {
    left, err := p.parseMulDiv()
    if err != nil {
        return 0, err
    }
    for p.pos < len(p.input) {
        p.skipWhitespace()
        if p.pos >= len(p.input) {
            break
        }
        op := p.input[p.pos]
        if op != '+' && op != '-' {
            break
        }
        p.pos++
        right, err := p.parseMulDiv()
        if err != nil {
            return 0, err
        }
        if op == '+' {
            left += right
        } else {
            left -= right
        }
    }
    return left, nil
}

func (p *exprParser) parseMulDiv() (float64, error) {
    left, err := p.parseUnary()
    if err != nil {
        return 0, err
    }
    for p.pos < len(p.input) {
        p.skipWhitespace()
        if p.pos >= len(p.input) {
            break
        }
        op := p.input[p.pos]
        if op != '*' && op != '/' && op != '%' {
            break
        }
        p.pos++
        right, err := p.parseUnary()
        if err != nil {
            return 0, err
        }
        switch op {
        case '*':
            left *= right
        case '/':
            if right == 0 {
                return 0, fmt.Errorf("division by zero")
            }
            left /= right
        case '%':
            if right == 0 {
                return 0, fmt.Errorf("modulo by zero")
            }
            left = math.Mod(left, right)
        }
    }
    return left, nil
}

func (p *exprParser) parseUnary() (float64, error) {
    p.skipWhitespace()
    if p.pos < len(p.input) && p.input[p.pos] == '-' {
        p.pos++
        val, err := p.parsePrimary()
        return -val, err
    }
    return p.parsePrimary()
}

func (p *exprParser) parsePrimary() (float64, error) {
    p.skipWhitespace()
    if p.pos >= len(p.input) {
        return 0, fmt.Errorf("unexpected end of expression")
    }

    // Parenthesized expression
    if p.input[p.pos] == '(' {
        p.pos++
        val, err := p.parseExpr()
        if err != nil {
            return 0, err
        }
        p.skipWhitespace()
        if p.pos >= len(p.input) || p.input[p.pos] != ')' {
            return 0, fmt.Errorf("missing closing parenthesis")
        }
        p.pos++
        return val, nil
    }

    // Number literal
    if unicode.IsDigit(rune(p.input[p.pos])) || p.input[p.pos] == '.' {
        start := p.pos
        for p.pos < len(p.input) && (unicode.IsDigit(rune(p.input[p.pos])) || p.input[p.pos] == '.') {
            p.pos++
        }
        return strconv.ParseFloat(p.input[start:p.pos], 64)
    }

    // Variable name
    if unicode.IsLetter(rune(p.input[p.pos])) || p.input[p.pos] == '_' {
        start := p.pos
        for p.pos < len(p.input) && (unicode.IsLetter(rune(p.input[p.pos])) || unicode.IsDigit(rune(p.input[p.pos])) || p.input[p.pos] == '_') {
            p.pos++
        }
        name := p.input[start:p.pos]
        val, ok := p.vars[name]
        if !ok {
            return 0, fmt.Errorf("unknown variable %q", name)
        }
        return toFloat64(val)
    }

    return 0, fmt.Errorf("unexpected character %q at position %d", p.input[p.pos], p.pos)
}

func (p *exprParser) skipWhitespace() {
    for p.pos < len(p.input) && unicode.IsSpace(rune(p.input[p.pos])) {
        p.pos++
    }
}

// toFloat64 converts various types to float64 for arithmetic.
func toFloat64(v any) (float64, error) {
    switch n := v.(type) {
    case float64:
        return n, nil
    case float32:
        return float64(n), nil
    case int:
        return float64(n), nil
    case int64:
        return float64(n), nil
    case string:
        return strconv.ParseFloat(n, 64)
    default:
        return 0, fmt.Errorf("cannot convert %T to number", v)
    }
}
```

### Task 2: Add {{expr: ...}} to TemplateProcessor

**File**: `business/sdk/workflow/template.go`

Add a pattern to detect `{{expr: ...}}` before the regular `{{variable}}` pattern:

```go
var exprPattern = regexp.MustCompile(`\{\{expr:\s*([^}]+)\}\}`)
```

In the `Process()` method (or wherever `{{variable}}` substitution is done), first handle expr blocks:

```go
// Replace {{expr: ...}} blocks first (before regular {{variable}} substitution)
result = exprPattern.ReplaceAllStringFunc(result, func(match string) string {
    // Extract expression from {{expr: <expression>}}
    submatches := exprPattern.FindStringSubmatch(match)
    if len(submatches) < 2 {
        return match
    }
    expression := strings.TrimSpace(submatches[1])

    val, err := EvalExpr(expression, tp.context) // tp.context is the map[string]any
    if err != nil {
        tp.log.Warn(context.Background(), "template expr eval failed",
            "expr", expression, "error", err)
        return match // fail-open: return original {{expr: ...}} string
    }

    // Format the result (strip trailing zeros for clean output)
    if val == float64(int64(val)) {
        return strconv.FormatInt(int64(val), 10)
    }
    return strconv.FormatFloat(val, 'f', -1, 64)
})
```

### Task 3: Unit Tests

**File**: `business/sdk/workflow/template_test.go` (or `expr_test.go`)

```go
func TestEvalExpr(t *testing.T) {
    vars := map[string]any{
        "quantity":   float64(5),
        "unit_price": float64(10.5),
        "subtotal":   float64(52.5),
        "tax_rate":   float64(0.1),
    }

    tests := []struct{
        expr    string
        want    float64
        wantErr bool
    }{
        {"quantity * unit_price", 52.5, false},
        {"subtotal * (1 + tax_rate)", 57.75, false},
        {"quantity + 3", 8, false},
        {"unknown_var", 0, true},
        {"quantity / 0", 0, true},
        {"(2 + 3) * 4", 20, false},
    }
    // ...
}

func TestTemplateExprSubstitution(t *testing.T) {
    // Test {{expr: quantity * unit_price}} in template processor
}
```

---

## Validation

```bash
go build ./...
go test ./business/sdk/workflow/... -run TestEvalExpr
go test ./business/sdk/workflow/... -run TestTemplate

# Integration: create a workflow action with {{expr: ...}} in a config value and verify execution
```

---

## Gotchas

- **Pattern detection order**: `{{expr: ...}}` must be processed before `{{variable}}` substitution to prevent the variable pattern from partially consuming `{{expr: quantity * unit_price}}`.
- **Integer vs float output**: `5 * 10 = 50.0` but should display as `50`, not `50.000000`. Use the `val == float64(int64(val))` check to format as integer when appropriate.
- **Nested pipes**: `{{expr: quantity * unit_price | currency:USD}}` should work — the expr block produces a float string, then the pipe filter formats it. This only works if the expr pattern doesn't consume the `| currency:USD` part. Adjust the regex to stop at `|` if needed: `\{\{expr:\s*([^|}]+)\s*(\|[^}]*)?\}\}`.
- **Variable names with dots**: `{{item.quantity}}` is a nested access pattern. The arithmetic evaluator uses simple identifier names (`quantity`). For nested access, the caller should use `lookup_entity` first to flatten the structure into the context.
- **Security**: The custom parser only handles numbers, arithmetic operators, and variable names. It cannot access OS functions, execute code, or cause side effects. This is intentionally more restrictive than a full expression language like `expr-lang`.
