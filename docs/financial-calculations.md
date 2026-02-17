# Financial Calculations Guide

This guide covers decimal arithmetic patterns for financial calculations in Ichor.

## Why Decimal?

**Library**: `github.com/shopspring/decimal`

Go's `float64` has precision issues that cause errors in financial calculations:

```go
0.1 + 0.2 = 0.30000000000000004  // float64 - WRONG
0.1 + 0.2 = 0.3                   // decimal - CORRECT
```

In financial systems, these tiny errors accumulate and cause real problems:
- Orders might be off by pennies
- Tax calculations become inconsistent
- Audit trails become unreliable

## Separation of Concerns

| Layer | Type | Purpose |
|-------|------|---------|
| API/Storage | `types.Money` (string-based) | Safe storage, validation, serialization |
| Calculations | `decimal.Decimal` | Arithmetic operations, precision math |

## Why NOT Add Arithmetic to types.Money?

The `types.Money` type (e.g., in `ordersbus/types/money.go`) is a **value object** designed for:
- API boundaries (JSON serialization)
- Database storage (VARCHAR)
- Validation (format checking)

It intentionally does **NOT** support arithmetic operations because:
1. **Separation of concerns**: Money represents values, calculations package computes them
2. **Domain purity**: The domain layer (ordersbus) stays focused on business rules
3. **Flexibility**: Calculations can be used across different domains
4. **Dependency isolation**: Only the calculation package needs the decimal library

## Usage Pattern

```go
// Convert string/Money to decimal for calculations
unitPrice, _ := decimal.NewFromString(order.UnitPrice.Value())

// Perform calculations with precision
total := quantity.Mul(unitPrice).Round(2)

// Convert back to string for storage
order.Total = total.String()
```

## Package Location

Calculation helpers live in `business/sdk/calculations/` because:
- Pure business logic (no HTTP, no context, no transactions)
- Reusable across domains
- Follows existing pattern: `business/sdk/page`, `business/sdk/order`

## Common Operations

```go
import "github.com/shopspring/decimal"

// Parse from string
price, err := decimal.NewFromString("19.99")

// Arithmetic
subtotal := price.Mul(quantity)
withTax := subtotal.Mul(decimal.NewFromFloat(1.08))
discount := subtotal.Mul(decimal.NewFromFloat(0.10))
final := withTax.Sub(discount)

// Rounding
rounded := final.Round(2)  // Round to 2 decimal places

// Comparison
if price.GreaterThan(decimal.Zero) {
    // ...
}

// Convert to string for storage
priceStr := rounded.String()
```

## Best Practices

1. **Always use decimal for money math** - Never use float64 for financial calculations
2. **Round at the end** - Perform all calculations, then round the final result
3. **Store as strings** - Use VARCHAR in the database, not DECIMAL or FLOAT
4. **Validate early** - Use `types.Money` to validate format at API boundaries
5. **Calculate late** - Convert to `decimal.Decimal` only when you need arithmetic
