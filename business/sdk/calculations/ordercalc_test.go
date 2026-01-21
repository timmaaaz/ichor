package calculations

import (
	"errors"
	"testing"

	"github.com/shopspring/decimal"
)

func TestCalculateLineTotal_FlatDiscount(t *testing.T) {
	tests := []struct {
		name     string
		input    LineItemInput
		expected decimal.Decimal
		wantErr  bool
	}{
		{
			name: "basic flat discount",
			input: LineItemInput{
				Quantity:     10,
				UnitPrice:    decimal.NewFromFloat(5.00),
				Discount:     decimal.NewFromFloat(2.00),
				DiscountType: "flat",
			},
			expected: decimal.NewFromFloat(48.00), // (10 × 5) - 2 = 48
		},
		{
			name: "zero discount",
			input: LineItemInput{
				Quantity:     5,
				UnitPrice:    decimal.NewFromFloat(10.00),
				Discount:     decimal.Zero,
				DiscountType: "flat",
			},
			expected: decimal.NewFromFloat(50.00), // 5 × 10 = 50
		},
		{
			name: "discount larger than total - returns zero",
			input: LineItemInput{
				Quantity:     1,
				UnitPrice:    decimal.NewFromFloat(5.00),
				Discount:     decimal.NewFromFloat(10.00),
				DiscountType: "flat",
			},
			expected: decimal.Zero, // (1 × 5) - 10 = -5 → clamped to 0
		},
		{
			name: "empty discount type defaults to flat",
			input: LineItemInput{
				Quantity:     2,
				UnitPrice:    decimal.NewFromFloat(25.00),
				Discount:     decimal.NewFromFloat(5.00),
				DiscountType: "",
			},
			expected: decimal.NewFromFloat(45.00), // (2 × 25) - 5 = 45
		},
		{
			name: "decimal precision",
			input: LineItemInput{
				Quantity:     3,
				UnitPrice:    decimal.NewFromFloat(9.99),
				Discount:     decimal.NewFromFloat(1.50),
				DiscountType: "flat",
			},
			expected: decimal.NewFromFloat(28.47), // (3 × 9.99) - 1.50 = 28.47
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CalculateLineTotal(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateLineTotal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !result.Equal(tt.expected) {
				t.Errorf("CalculateLineTotal() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestCalculateLineTotal_PercentDiscount(t *testing.T) {
	tests := []struct {
		name     string
		input    LineItemInput
		expected decimal.Decimal
		wantErr  bool
	}{
		{
			name: "10% discount",
			input: LineItemInput{
				Quantity:     10,
				UnitPrice:    decimal.NewFromFloat(5.00),
				Discount:     decimal.NewFromFloat(10.00),
				DiscountType: "percent",
			},
			expected: decimal.NewFromFloat(45.00), // 50 × 0.90 = 45
		},
		{
			name: "0% discount",
			input: LineItemInput{
				Quantity:     5,
				UnitPrice:    decimal.NewFromFloat(10.00),
				Discount:     decimal.Zero,
				DiscountType: "percent",
			},
			expected: decimal.NewFromFloat(50.00), // 50 × 1.00 = 50
		},
		{
			name: "100% discount",
			input: LineItemInput{
				Quantity:     10,
				UnitPrice:    decimal.NewFromFloat(5.00),
				Discount:     decimal.NewFromFloat(100.00),
				DiscountType: "percent",
			},
			expected: decimal.Zero, // 50 × 0.00 = 0
		},
		{
			name: "discount > 100% clamped to 100%",
			input: LineItemInput{
				Quantity:     10,
				UnitPrice:    decimal.NewFromFloat(5.00),
				Discount:     decimal.NewFromFloat(150.00),
				DiscountType: "percent",
			},
			expected: decimal.Zero, // Clamped to 100% = 0
		},
		{
			name: "negative discount clamped to 0%",
			input: LineItemInput{
				Quantity:     10,
				UnitPrice:    decimal.NewFromFloat(5.00),
				Discount:     decimal.NewFromFloat(-10.00),
				DiscountType: "percent",
			},
			expected: decimal.NewFromFloat(50.00), // Treated as 0% discount
		},
		{
			name: "25% discount with rounding",
			input: LineItemInput{
				Quantity:     3,
				UnitPrice:    decimal.NewFromFloat(9.99),
				Discount:     decimal.NewFromFloat(25.00),
				DiscountType: "percent",
			},
			expected: decimal.NewFromFloat(22.48), // 29.97 × 0.75 = 22.4775 → 22.48
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CalculateLineTotal(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateLineTotal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !result.Equal(tt.expected) {
				t.Errorf("CalculateLineTotal() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestCalculateLineTotal_ItemizedDiscount(t *testing.T) {
	tests := []struct {
		name     string
		input    LineItemInput
		expected decimal.Decimal
		wantErr  bool
	}{
		{
			name: "basic itemized discount",
			input: LineItemInput{
				Quantity:     5,
				UnitPrice:    decimal.NewFromFloat(10.00),
				Discount:     decimal.NewFromFloat(2.50),
				DiscountType: "itemized",
			},
			expected: decimal.NewFromFloat(37.50), // (5 × 10) - (5 × 2.50) = 37.50
		},
		{
			name: "zero discount",
			input: LineItemInput{
				Quantity:     5,
				UnitPrice:    decimal.NewFromFloat(10.00),
				Discount:     decimal.Zero,
				DiscountType: "itemized",
			},
			expected: decimal.NewFromFloat(50.00), // (5 × 10) - (5 × 0) = 50
		},
		{
			name: "discount exceeds unit price - returns zero",
			input: LineItemInput{
				Quantity:     2,
				UnitPrice:    decimal.NewFromFloat(5.00),
				Discount:     decimal.NewFromFloat(10.00),
				DiscountType: "itemized",
			},
			expected: decimal.Zero, // (2 × 5) - (2 × 10) = -10 → clamped to 0
		},
		{
			name: "decimal precision",
			input: LineItemInput{
				Quantity:     3,
				UnitPrice:    decimal.NewFromFloat(9.99),
				Discount:     decimal.NewFromFloat(1.50),
				DiscountType: "itemized",
			},
			expected: decimal.NewFromFloat(25.47), // (3 × 9.99) - (3 × 1.50) = 25.47
		},
		{
			name: "single item",
			input: LineItemInput{
				Quantity:     1,
				UnitPrice:    decimal.NewFromFloat(100.00),
				Discount:     decimal.NewFromFloat(15.00),
				DiscountType: "itemized",
			},
			expected: decimal.NewFromFloat(85.00), // (1 × 100) - (1 × 15) = 85
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CalculateLineTotal(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateLineTotal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !result.Equal(tt.expected) {
				t.Errorf("CalculateLineTotal() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestCalculateLineTotal_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   LineItemInput
		wantErr bool
	}{
		{
			name: "negative quantity",
			input: LineItemInput{
				Quantity:     -1,
				UnitPrice:    decimal.NewFromFloat(10.00),
				Discount:     decimal.Zero,
				DiscountType: "flat",
			},
			wantErr: true,
		},
		{
			name: "quantity exceeds max",
			input: LineItemInput{
				Quantity:     MaxQuantity + 1,
				UnitPrice:    decimal.NewFromFloat(10.00),
				Discount:     decimal.Zero,
				DiscountType: "flat",
			},
			wantErr: true,
		},
		{
			name: "negative unit price",
			input: LineItemInput{
				Quantity:     10,
				UnitPrice:    decimal.NewFromFloat(-5.00),
				Discount:     decimal.Zero,
				DiscountType: "flat",
			},
			wantErr: true,
		},
		{
			name: "unit price exceeds max",
			input: LineItemInput{
				Quantity:     10,
				UnitPrice:    decimal.NewFromFloat(1000001.00),
				Discount:     decimal.Zero,
				DiscountType: "flat",
			},
			wantErr: true,
		},
		{
			name: "invalid discount type",
			input: LineItemInput{
				Quantity:     10,
				UnitPrice:    decimal.NewFromFloat(5.00),
				Discount:     decimal.Zero,
				DiscountType: "invalid",
			},
			wantErr: true,
		},
		{
			name: "zero quantity is valid",
			input: LineItemInput{
				Quantity:     0,
				UnitPrice:    decimal.NewFromFloat(10.00),
				Discount:     decimal.Zero,
				DiscountType: "flat",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CalculateLineTotal(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateLineTotal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCalculateOrderTotals(t *testing.T) {
	tests := []struct {
		name             string
		lineItems        []LineItemInput
		taxRate          decimal.Decimal
		shippingCost     decimal.Decimal
		expectedSubtotal decimal.Decimal
		expectedTax      decimal.Decimal
		expectedTotal    decimal.Decimal
		wantErr          bool
	}{
		{
			name: "basic order",
			lineItems: []LineItemInput{
				{Quantity: 2, UnitPrice: decimal.NewFromFloat(10.00), Discount: decimal.Zero, DiscountType: "flat"},
				{Quantity: 1, UnitPrice: decimal.NewFromFloat(25.00), Discount: decimal.Zero, DiscountType: "flat"},
			},
			taxRate:          decimal.NewFromFloat(8.5),
			shippingCost:     decimal.NewFromFloat(5.00),
			expectedSubtotal: decimal.NewFromFloat(45.00),  // 20 + 25 = 45
			expectedTax:      decimal.NewFromFloat(3.83),   // 45 × 0.085 = 3.825 → 3.83
			expectedTotal:    decimal.NewFromFloat(53.83),  // 45 + 3.83 + 5 = 53.83
		},
		{
			name: "mixed discount types",
			lineItems: []LineItemInput{
				{Quantity: 10, UnitPrice: decimal.NewFromFloat(5.00), Discount: decimal.NewFromFloat(2.00), DiscountType: "flat"},    // 48
				{Quantity: 10, UnitPrice: decimal.NewFromFloat(5.00), Discount: decimal.NewFromFloat(10.00), DiscountType: "percent"}, // 45
			},
			taxRate:          decimal.NewFromFloat(10.00),
			shippingCost:     decimal.Zero,
			expectedSubtotal: decimal.NewFromFloat(93.00),  // 48 + 45 = 93
			expectedTax:      decimal.NewFromFloat(9.30),   // 93 × 0.10 = 9.30
			expectedTotal:    decimal.NewFromFloat(102.30), // 93 + 9.30 + 0 = 102.30
		},
		{
			name:             "empty line items",
			lineItems:        []LineItemInput{},
			taxRate:          decimal.NewFromFloat(8.5),
			shippingCost:     decimal.NewFromFloat(5.00),
			expectedSubtotal: decimal.Zero,
			expectedTax:      decimal.Zero,
			expectedTotal:    decimal.NewFromFloat(5.00), // Only shipping
		},
		{
			name: "zero tax rate",
			lineItems: []LineItemInput{
				{Quantity: 5, UnitPrice: decimal.NewFromFloat(10.00), Discount: decimal.Zero, DiscountType: "flat"},
			},
			taxRate:          decimal.Zero,
			shippingCost:     decimal.NewFromFloat(10.00),
			expectedSubtotal: decimal.NewFromFloat(50.00),
			expectedTax:      decimal.Zero,
			expectedTotal:    decimal.NewFromFloat(60.00), // 50 + 0 + 10 = 60
		},
		{
			name: "zero shipping",
			lineItems: []LineItemInput{
				{Quantity: 5, UnitPrice: decimal.NewFromFloat(10.00), Discount: decimal.Zero, DiscountType: "flat"},
			},
			taxRate:          decimal.NewFromFloat(5.00),
			shippingCost:     decimal.Zero,
			expectedSubtotal: decimal.NewFromFloat(50.00),
			expectedTax:      decimal.NewFromFloat(2.50),  // 50 × 0.05 = 2.50
			expectedTotal:    decimal.NewFromFloat(52.50), // 50 + 2.50 + 0 = 52.50
		},
		{
			name: "single item with full discount",
			lineItems: []LineItemInput{
				{Quantity: 1, UnitPrice: decimal.NewFromFloat(10.00), Discount: decimal.NewFromFloat(100.00), DiscountType: "percent"},
			},
			taxRate:          decimal.NewFromFloat(10.00),
			shippingCost:     decimal.NewFromFloat(5.00),
			expectedSubtotal: decimal.Zero,
			expectedTax:      decimal.Zero,
			expectedTotal:    decimal.NewFromFloat(5.00), // Only shipping
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CalculateOrderTotals(tt.lineItems, tt.taxRate, tt.shippingCost)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateOrderTotals() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !result.Subtotal.Equal(tt.expectedSubtotal) {
				t.Errorf("Subtotal = %s, want %s", result.Subtotal, tt.expectedSubtotal)
			}
			if !result.TaxAmount.Equal(tt.expectedTax) {
				t.Errorf("TaxAmount = %s, want %s", result.TaxAmount, tt.expectedTax)
			}
			if !result.Total.Equal(tt.expectedTotal) {
				t.Errorf("Total = %s, want %s", result.Total, tt.expectedTotal)
			}
		})
	}
}

func TestCalculateOrderTotals_ValidationError(t *testing.T) {
	// Order with invalid line item should fail
	lineItems := []LineItemInput{
		{Quantity: 5, UnitPrice: decimal.NewFromFloat(10.00), Discount: decimal.Zero, DiscountType: "flat"},
		{Quantity: -1, UnitPrice: decimal.NewFromFloat(10.00), Discount: decimal.Zero, DiscountType: "flat"}, // Invalid
	}

	_, err := CalculateOrderTotals(lineItems, decimal.NewFromFloat(8.5), decimal.NewFromFloat(5.00))
	if err == nil {
		t.Error("CalculateOrderTotals() expected error for invalid line item, got nil")
	}
}

func TestLineItemInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   LineItemInput
		wantErr bool
	}{
		{
			name: "valid input",
			input: LineItemInput{
				Quantity:     10,
				UnitPrice:    decimal.NewFromFloat(5.00),
				Discount:     decimal.Zero,
				DiscountType: "flat",
			},
			wantErr: false,
		},
		{
			name: "valid with percent discount",
			input: LineItemInput{
				Quantity:     10,
				UnitPrice:    decimal.NewFromFloat(5.00),
				Discount:     decimal.NewFromFloat(15.00),
				DiscountType: "percent",
			},
			wantErr: false,
		},
		{
			name: "valid with itemized discount",
			input: LineItemInput{
				Quantity:     10,
				UnitPrice:    decimal.NewFromFloat(5.00),
				Discount:     decimal.NewFromFloat(1.00),
				DiscountType: "itemized",
			},
			wantErr: false,
		},
		{
			name: "valid with empty discount type",
			input: LineItemInput{
				Quantity:     10,
				UnitPrice:    decimal.NewFromFloat(5.00),
				Discount:     decimal.Zero,
				DiscountType: "",
			},
			wantErr: false,
		},
		{
			name: "valid at max quantity",
			input: LineItemInput{
				Quantity:     MaxQuantity,
				UnitPrice:    decimal.NewFromFloat(5.00),
				Discount:     decimal.Zero,
				DiscountType: "flat",
			},
			wantErr: false,
		},
		{
			name: "valid at max unit price",
			input: LineItemInput{
				Quantity:  1,
				UnitPrice: maxUnitPriceDecimal,
				Discount:  decimal.Zero,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCalculateOrderTotals_TaxRateBounds(t *testing.T) {
	items := []LineItemInput{
		{Quantity: 1, UnitPrice: decimal.NewFromFloat(10.00), DiscountType: "flat"},
	}

	tests := []struct {
		name      string
		taxRate   decimal.Decimal
		expectErr error
	}{
		{
			name:      "negative tax rate",
			taxRate:   decimal.NewFromFloat(-1),
			expectErr: ErrInvalidTaxRate,
		},
		{
			name:      "tax rate exceeds max 50%",
			taxRate:   decimal.NewFromFloat(51),
			expectErr: ErrInvalidTaxRate,
		},
		{
			name:      "tax rate at max 50% is valid",
			taxRate:   decimal.NewFromFloat(50),
			expectErr: nil,
		},
		{
			name:      "zero tax rate is valid",
			taxRate:   decimal.Zero,
			expectErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CalculateOrderTotals(items, tt.taxRate, decimal.Zero)
			if tt.expectErr != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectErr)
				} else if !errors.Is(err, tt.expectErr) {
					t.Errorf("expected error %v, got %v", tt.expectErr, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestCalculateOrderTotals_ShippingCostBounds(t *testing.T) {
	items := []LineItemInput{
		{Quantity: 1, UnitPrice: decimal.NewFromFloat(10.00), DiscountType: "flat"},
	}

	tests := []struct {
		name         string
		shippingCost decimal.Decimal
		expectErr    error
	}{
		{
			name:         "negative shipping cost",
			shippingCost: decimal.NewFromFloat(-1),
			expectErr:    ErrInvalidShippingCost,
		},
		{
			name:         "shipping cost exceeds max $100k",
			shippingCost: decimal.NewFromFloat(100001),
			expectErr:    ErrInvalidShippingCost,
		},
		{
			name:         "shipping cost at max $100k is valid",
			shippingCost: decimal.RequireFromString(MaxShippingCost),
			expectErr:    nil,
		},
		{
			name:         "zero shipping cost is valid",
			shippingCost: decimal.Zero,
			expectErr:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CalculateOrderTotals(items, decimal.Zero, tt.shippingCost)
			if tt.expectErr != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectErr)
				} else if !errors.Is(err, tt.expectErr) {
					t.Errorf("expected error %v, got %v", tt.expectErr, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestCalculateOrderTotals_MaxOrderTotal(t *testing.T) {
	// Create items that would exceed $100M total
	// MaxQuantity (1,000,000) × MaxUnitPrice ($1,000,000) = $1 trillion
	// This should exceed the $100M max order total
	items := []LineItemInput{
		{
			Quantity:     MaxQuantity,
			UnitPrice:    maxUnitPriceDecimal,
			Discount:     decimal.Zero,
			DiscountType: "flat",
		},
	}

	_, err := CalculateOrderTotals(items, decimal.Zero, decimal.Zero)
	if !errors.Is(err, ErrOrderTotalExceeded) {
		t.Errorf("expected ErrOrderTotalExceeded, got %v", err)
	}
}

func TestCalculateOrderTotals_OrderTotalAtMaxIsValid(t *testing.T) {
	// Create items that result in exactly $100M total
	// Use 100 items × $1M each = $100M
	items := make([]LineItemInput, 100)
	for i := range items {
		items[i] = LineItemInput{
			Quantity:     1,
			UnitPrice:    maxUnitPriceDecimal, // $1M each
			Discount:     decimal.Zero,
			DiscountType: "flat",
		}
	}

	totals, err := CalculateOrderTotals(items, decimal.Zero, decimal.Zero)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !totals.Total.Equal(maxOrderTotalDecimal) {
		t.Errorf("expected total %s, got %s", MaxOrderTotal, totals.Total)
	}
}

// TestSentinelErrorsCanBeCheckedInCalculations verifies that sentinel errors
// from validation can be checked using errors.Is() after wrapping.
func TestSentinelErrorsCanBeCheckedInCalculations(t *testing.T) {
	// Test ErrInvalidQuantity
	item := LineItemInput{
		Quantity:  -1,
		UnitPrice: decimal.NewFromFloat(10.00),
	}
	_, err := CalculateLineTotal(item)
	if !errors.Is(err, ErrInvalidQuantity) {
		t.Errorf("expected errors.Is(err, ErrInvalidQuantity) to be true, got false")
	}

	// Test ErrInvalidUnitPrice
	item = LineItemInput{
		Quantity:  1,
		UnitPrice: decimal.NewFromFloat(-10.00),
	}
	_, err = CalculateLineTotal(item)
	if !errors.Is(err, ErrInvalidUnitPrice) {
		t.Errorf("expected errors.Is(err, ErrInvalidUnitPrice) to be true, got false")
	}

	// Test ErrInvalidDiscountType
	item = LineItemInput{
		Quantity:     1,
		UnitPrice:    decimal.NewFromFloat(10.00),
		DiscountType: "invalid",
	}
	_, err = CalculateLineTotal(item)
	if !errors.Is(err, ErrInvalidDiscountType) {
		t.Errorf("expected errors.Is(err, ErrInvalidDiscountType) to be true, got false")
	}

	// Test ErrInvalidTaxRate
	items := []LineItemInput{{Quantity: 1, UnitPrice: decimal.NewFromFloat(10.00)}}
	_, err = CalculateOrderTotals(items, decimal.NewFromFloat(-1), decimal.Zero)
	if !errors.Is(err, ErrInvalidTaxRate) {
		t.Errorf("expected errors.Is(err, ErrInvalidTaxRate) to be true, got false")
	}

	// Test ErrInvalidShippingCost
	_, err = CalculateOrderTotals(items, decimal.Zero, decimal.NewFromFloat(-1))
	if !errors.Is(err, ErrInvalidShippingCost) {
		t.Errorf("expected errors.Is(err, ErrInvalidShippingCost) to be true, got false")
	}
}
