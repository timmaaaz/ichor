package paymenttermbus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
)

// TestNewPaymentTerms is a helper method for testing.
func TestNewPaymentTerms(n int) []NewPaymentTerm {
	newPaymentTerms := make([]NewPaymentTerm, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		na := NewPaymentTerm{
			Name:        fmt.Sprintf("PaymentTerm%d", idx),
			Description: fmt.Sprintf("PaymentTerm%d Description", idx),
		}

		newPaymentTerms[i] = na
	}

	return newPaymentTerms
}

// TestSeedPaymentTerms is a helper method for testing.
func TestSeedPaymentTerms(ctx context.Context, n int, api *Business) ([]PaymentTerm, error) {
	newPaymentTerms := TestNewPaymentTerms(n)

	paymentTerms := make([]PaymentTerm, len(newPaymentTerms))
	for i, na := range newPaymentTerms {
		paymentTerm, err := api.Create(ctx, na)
		if err != nil {
			return nil, fmt.Errorf("seeding payment term: idx: %d : %w", i, err)
		}

		paymentTerms[i] = paymentTerm
	}

	// sort payment terms by name
	sort.Slice(paymentTerms, func(i, j int) bool {
		return paymentTerms[i].Name <= paymentTerms[j].Name
	})

	return paymentTerms, nil
}
