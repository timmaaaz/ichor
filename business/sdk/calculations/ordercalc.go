// Package calculations provides financial calculation utilities for order processing.
//
// This package uses github.com/shopspring/decimal for arbitrary-precision decimal
// arithmetic. This is essential for financial calculations where floating-point
// precision issues (e.g., 0.1 + 0.2 = 0.30000000000000004) are unacceptable.
//
// Design Decision: Why decimal.Decimal instead of types.Money?
//
// The types.Money type in ordersbus is a string-based value object designed for:
//   - API boundaries (JSON serialization)
//   - Database storage (VARCHAR)
//   - Validation (format checking)
//
// It intentionally does NOT support arithmetic operations because:
//  1. Separation of concerns: Money represents values, this package computes them
//  2. Domain purity: The domain layer (ordersbus) stays focused on business rules
//  3. Flexibility: Calculations can be used across different domains
//
// The pattern is: Money (string) → decimal.Decimal (calculate) → string (store)
package calculations

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

// Set of error variables for calculation operations.
var (
	ErrInvalidQuantity     = errors.New("invalid quantity")
	ErrInvalidUnitPrice    = errors.New("invalid unit price")
	ErrInvalidDiscount     = errors.New("invalid discount")
	ErrInvalidDiscountType = errors.New("invalid discount type")
	ErrInvalidTaxRate      = errors.New("invalid tax rate")
	ErrInvalidShippingCost = errors.New("invalid shipping cost")
	ErrOrderTotalExceeded  = errors.New("order total exceeds maximum")
)

// Maximum bounds for financial values (security against overflow/abuse)
const (
	MaxQuantity     = 1000000
	MaxUnitPrice    = "1000000.00"
	MaxShippingCost = "100000.00"    // $100k max shipping
	MaxTaxRate      = "50.00"        // 50% max tax rate
	MaxOrderTotal   = "100000000.00" // $100M max order total
)

var (
	maxUnitPriceDecimal    = decimal.RequireFromString(MaxUnitPrice)
	maxShippingCostDecimal = decimal.RequireFromString(MaxShippingCost)
	maxTaxRateDecimal      = decimal.RequireFromString(MaxTaxRate)
	maxOrderTotalDecimal   = decimal.RequireFromString(MaxOrderTotal)
)

// LineItemInput represents the fields needed to calculate a line total.
type LineItemInput struct {
	Quantity     int
	UnitPrice    decimal.Decimal
	Discount     decimal.Decimal
	DiscountType string // "flat" or "percent"
}

// Validate checks that line item input is within acceptable bounds.
func (l LineItemInput) Validate() error {
	if l.Quantity < 0 {
		return fmt.Errorf("%w: %d (cannot be negative)", ErrInvalidQuantity, l.Quantity)
	}
	if l.Quantity > MaxQuantity {
		return fmt.Errorf("%w: %d exceeds max %d", ErrInvalidQuantity, l.Quantity, MaxQuantity)
	}
	if l.UnitPrice.LessThan(decimal.Zero) {
		return fmt.Errorf("%w: %s (cannot be negative)", ErrInvalidUnitPrice, l.UnitPrice)
	}
	if l.UnitPrice.GreaterThan(maxUnitPriceDecimal) {
		return fmt.Errorf("%w: %s exceeds max %s", ErrInvalidUnitPrice, l.UnitPrice, MaxUnitPrice)
	}
	if l.DiscountType != "flat" && l.DiscountType != "percent" && l.DiscountType != "" {
		return fmt.Errorf("%w: got %q (must be 'flat' or 'percent')", ErrInvalidDiscountType, l.DiscountType)
	}
	return nil
}

// OrderTotals holds the calculated totals for an order.
type OrderTotals struct {
	Subtotal  decimal.Decimal
	TaxAmount decimal.Decimal
	Total     decimal.Decimal
}

// CalculateLineTotal calculates the total for a single line item.
// Returns error if validation fails.
//
// Calculation logic:
//   - Flat discount: (quantity × unit_price) - discount, minimum 0
//   - Percent discount: (quantity × unit_price) × (1 - discount/100), clamped 0-100%
func CalculateLineTotal(item LineItemInput) (decimal.Decimal, error) {
	if err := item.Validate(); err != nil {
		return decimal.Zero, fmt.Errorf("invalid line item: %w", err)
	}

	gross := decimal.NewFromInt(int64(item.Quantity)).Mul(item.UnitPrice)

	discountType := item.DiscountType
	if discountType == "" {
		discountType = "flat" // default
	}

	switch discountType {
	case "percent":
		pct := item.Discount
		// Clamp percent discount to 0-100
		if pct.LessThan(decimal.Zero) {
			pct = decimal.Zero
		}
		if pct.GreaterThan(decimal.NewFromInt(100)) {
			pct = decimal.NewFromInt(100)
		}
		// gross × (1 - discount/100)
		multiplier := decimal.NewFromInt(1).Sub(pct.Div(decimal.NewFromInt(100)))
		return gross.Mul(multiplier).Round(2), nil

	default: // "flat"
		result := gross.Sub(item.Discount)
		// Prevent negative totals
		if result.LessThan(decimal.Zero) {
			return decimal.Zero, nil
		}
		return result.Round(2), nil
	}
}

// CalculateOrderTotals calculates subtotal, tax amount, and total for an order.
//
// Calculation logic:
//   - Subtotal = sum of all line totals
//   - TaxAmount = subtotal × (taxRate / 100), rounded to 2 decimals
//   - Total = subtotal + taxAmount + shippingCost, rounded to 2 decimals
func CalculateOrderTotals(lineItems []LineItemInput, taxRate, shippingCost decimal.Decimal) (OrderTotals, error) {
	// Validate tax rate bounds
	if taxRate.LessThan(decimal.Zero) {
		return OrderTotals{}, fmt.Errorf("%w: %s (cannot be negative)", ErrInvalidTaxRate, taxRate.String())
	}
	if taxRate.GreaterThan(maxTaxRateDecimal) {
		return OrderTotals{}, fmt.Errorf("%w: %s exceeds max %s%%", ErrInvalidTaxRate, taxRate.String(), MaxTaxRate)
	}

	// Validate shipping cost bounds
	if shippingCost.LessThan(decimal.Zero) {
		return OrderTotals{}, fmt.Errorf("%w: %s (cannot be negative)", ErrInvalidShippingCost, shippingCost.String())
	}
	if shippingCost.GreaterThan(maxShippingCostDecimal) {
		return OrderTotals{}, fmt.Errorf("%w: %s exceeds max $%s", ErrInvalidShippingCost, shippingCost.String(), MaxShippingCost)
	}

	subtotal := decimal.Zero

	for i, item := range lineItems {
		lineTotal, err := CalculateLineTotal(item)
		if err != nil {
			return OrderTotals{}, fmt.Errorf("line item %d: %w", i, err)
		}
		subtotal = subtotal.Add(lineTotal)
	}

	// Tax = subtotal × (taxRate / 100)
	taxAmount := subtotal.Mul(taxRate.Div(decimal.NewFromInt(100))).Round(2)

	// Total = subtotal + tax + shipping
	total := subtotal.Add(taxAmount).Add(shippingCost).Round(2)

	// Validate max order total
	if total.GreaterThan(maxOrderTotalDecimal) {
		return OrderTotals{}, fmt.Errorf("%w: %s exceeds max $%s", ErrOrderTotalExceeded, total.String(), MaxOrderTotal)
	}

	return OrderTotals{
		Subtotal:  subtotal,
		TaxAmount: taxAmount,
		Total:     total,
	}, nil
}
