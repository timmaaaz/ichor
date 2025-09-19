package approvalbus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"

	"github.com/google/uuid"
)

func TestNewUserApprovalStatus(n int) []NewUserApprovalStatus {
	newUserApprovalStatus := make([]NewUserApprovalStatus, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		nas := NewUserApprovalStatus{
			Name:   fmt.Sprintf("UserApprovalStatus%d", idx),
			IconID: uuid.New(),
		}

		newUserApprovalStatus[i] = nas
	}

	return newUserApprovalStatus
}

func TestSeedUserApprovalStatus(ctx context.Context, n int, api *Business) ([]UserApprovalStatus, error) {
	newUserApprovalStatuses := TestNewUserApprovalStatus(n)

	userApprovalStatuses := make([]UserApprovalStatus, len(newUserApprovalStatuses))

	for i, nas := range newUserApprovalStatuses {
		userApprovalStatus, err := api.Create(ctx, nas)
		if err != nil {
			return nil, fmt.Errorf("seeding user approval status: idx: %d : %w", i, err)
		}

		userApprovalStatuses[i] = userApprovalStatus
	}

	sort.Slice(userApprovalStatuses, func(i, j int) bool {
		return userApprovalStatuses[i].Name > userApprovalStatuses[j].Name
	})

	return userApprovalStatuses, nil
}
