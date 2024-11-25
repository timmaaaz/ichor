package approvalstatusbus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"

	"github.com/google/uuid"
)

// TestNewApprovalStatus is a helper method for testing.
func TestNewApprovalStatus(n int) []NewApprovalStatus {
	newApprovalStatus := make([]NewApprovalStatus, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		nas := NewApprovalStatus{
			IconID: uuid.New(),
			Name:   fmt.Sprintf("ApprovalStatus%d", idx),
		}

		newApprovalStatus[i] = nas
	}

	return newApprovalStatus
}

// TestSeedApprovalStatus is a helper method for testing.
func TestSeedApprovalStatus(ctx context.Context, n int, api *Business) ([]ApprovalStatus, error) {
	newApprovalStatuses := TestNewApprovalStatus(n)

	aprvlStatuses := make([]ApprovalStatus, len(newApprovalStatuses))
	for i, nc := range newApprovalStatuses {
		as, err := api.Create(ctx, nc)
		if err != nil {
			return nil, fmt.Errorf("seeding city: idx: %d : %w", i, err)
		}

		aprvlStatuses[i] = as
	}

	// sort cities by name
	sort.Slice(aprvlStatuses, func(i, j int) bool {
		return aprvlStatuses[i].Name <= aprvlStatuses[j].Name
	})

	return aprvlStatuses, nil
}
