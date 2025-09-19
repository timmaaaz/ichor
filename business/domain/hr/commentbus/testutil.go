package commentbus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"

	"github.com/google/uuid"
)

func TestNewUserApprovalComment(n int, commenterID, userID uuid.UUIDs) []NewUserApprovalComment {
	newUserApprovalComment := make([]NewUserApprovalComment, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		nac := NewUserApprovalComment{
			CommenterID: commenterID[i%len(commenterID)],
			UserID:      userID[i%len(userID)],
			Comment:     fmt.Sprintf("Comment%d", idx),
		}

		newUserApprovalComment[i] = nac
	}

	return newUserApprovalComment
}

func TestSeedUserApprovalComment(ctx context.Context, n int, commenterID, userID uuid.UUIDs, api *Business) ([]UserApprovalComment, error) {
	newUserApprovalComments := TestNewUserApprovalComment(n, commenterID, userID)

	userApprovalComments := make([]UserApprovalComment, len(newUserApprovalComments))
	for i, nac := range newUserApprovalComments {
		ua, err := api.Create(ctx, nac)
		if err != nil {
			return nil, fmt.Errorf("seeding user approval comment: idx: %d : %w", i, err)
		}

		userApprovalComments[i] = ua
	}

	sort.Slice(userApprovalComments, func(i, j int) bool {
		return userApprovalComments[i].ID.String() < userApprovalComments[j].ID.String()
	})

	return userApprovalComments, nil

}
