package commentbus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"time"

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

// TestNewUserApprovalCommentHistorical creates comments distributed across a time range for seeding.
// daysBack specifies how many days of history to generate (30-180 days recommended for comments).
// Comments are evenly distributed across the time range.
func TestNewUserApprovalCommentHistorical(n int, daysBack int, commenterID, userID uuid.UUIDs) []NewUserApprovalComment {
	newUserApprovalComment := make([]NewUserApprovalComment, n)
	now := time.Now()

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		// Distribute evenly across the time range
		daysAgo := (i * daysBack) / n
		createdDate := now.AddDate(0, 0, -daysAgo)

		nac := NewUserApprovalComment{
			CommenterID: commenterID[i%len(commenterID)],
			UserID:      userID[i%len(userID)],
			Comment:     fmt.Sprintf("Comment%d", idx),
			CreatedDate: &createdDate,
		}

		newUserApprovalComment[i] = nac
	}

	return newUserApprovalComment
}

// TestSeedUserApprovalCommentHistorical seeds comments with historical date distribution.
func TestSeedUserApprovalCommentHistorical(ctx context.Context, n int, daysBack int, commenterID, userID uuid.UUIDs, api *Business) ([]UserApprovalComment, error) {
	newUserApprovalComments := TestNewUserApprovalCommentHistorical(n, daysBack, commenterID, userID)

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
