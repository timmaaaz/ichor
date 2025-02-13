package commentbus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/users/status/commentbus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/unitest"
)

func Test_UserCommentStatus(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_UserCommentStatus")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	unitest.Run(t, query(db.BusDomain, sd), "query")
	unitest.Run(t, create(db.BusDomain, sd), "create")
	unitest.Run(t, update(db.BusDomain, sd), "update")
	unitest.Run(t, delete(db.BusDomain, sd), "delete")
}

// =============================================================================

func insertSeedData(busDomain dbtest.BusDomain) (unitest.SeedData, error) {
	ctx := context.Background()

	users, err := userbus.TestSeedUsersWithNoFKs(ctx, 10, userbus.Roles.User, busDomain.User)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	userIDs := make([]uuid.UUID, 0, len(users))
	for _, u := range users {
		userIDs = append(userIDs, u.ID)
	}

	ac, err := commentbus.TestSeedUserApprovalComment(ctx, 10, userIDs[:5], userIDs[5:], busDomain.UserApprovalComment)
	if err != nil {
		return unitest.SeedData{}, fmt.Errorf("seeding approval comments : %w", err)
	}

	return unitest.SeedData{
		UserApprovalComment: ac,
	}, nil
}

func query(busdomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "query",
			ExpResp: []commentbus.UserApprovalComment{
				{ID: sd.UserApprovalComment[0].ID, UserID: sd.UserApprovalComment[0].UserID, CommenterID: sd.UserApprovalComment[0].CommenterID, Comment: sd.UserApprovalComment[0].Comment, CreatedDate: sd.UserApprovalComment[0].CreatedDate},
				{ID: sd.UserApprovalComment[1].ID, UserID: sd.UserApprovalComment[1].UserID, CommenterID: sd.UserApprovalComment[1].CommenterID, Comment: sd.UserApprovalComment[1].Comment, CreatedDate: sd.UserApprovalComment[1].CreatedDate},
				{ID: sd.UserApprovalComment[2].ID, UserID: sd.UserApprovalComment[2].UserID, CommenterID: sd.UserApprovalComment[2].CommenterID, Comment: sd.UserApprovalComment[2].Comment, CreatedDate: sd.UserApprovalComment[2].CreatedDate},
				{ID: sd.UserApprovalComment[3].ID, UserID: sd.UserApprovalComment[3].UserID, CommenterID: sd.UserApprovalComment[3].CommenterID, Comment: sd.UserApprovalComment[3].Comment, CreatedDate: sd.UserApprovalComment[3].CreatedDate},
				{ID: sd.UserApprovalComment[4].ID, UserID: sd.UserApprovalComment[4].UserID, CommenterID: sd.UserApprovalComment[4].CommenterID, Comment: sd.UserApprovalComment[4].Comment, CreatedDate: sd.UserApprovalComment[4].CreatedDate},
			},
			ExcFunc: func(ctx context.Context) any {
				aprvlStatuses, err := busdomain.UserApprovalComment.Query(ctx, commentbus.QueryFilter{}, order.NewBy(commentbus.OrderByID, order.ASC), page.MustParse("1", "5"))
				if err != nil {
					return err
				}
				return aprvlStatuses
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func create(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {

	var commenterID = sd.UserApprovalComment[1].CommenterID
	var userID = sd.UserApprovalComment[3].CommenterID

	table := []unitest.Table{
		{
			Name: "create",
			ExpResp: commentbus.UserApprovalComment{
				Comment:     "Comment created",
				CommenterID: commenterID,
				UserID:      userID,
			},
			ExcFunc: func(ctx context.Context) any {
				aprvlStatus, err := busDomain.UserApprovalComment.Create(ctx, commentbus.NewUserApprovalComment{
					Comment:     "Comment created",
					CommenterID: commenterID,
					UserID:      userID,
				})
				if err != nil {
					return err
				}
				return aprvlStatus
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(commentbus.UserApprovalComment)
				if !exists {
					return fmt.Sprintf("got is not an approval status %v", got)
				}

				expResp := exp.(commentbus.UserApprovalComment)
				expResp.ID = gotResp.ID
				expResp.CreatedDate = gotResp.CreatedDate

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func update(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name: "update",
			ExpResp: commentbus.UserApprovalComment{
				ID:          sd.UserApprovalComment[0].ID,
				Comment:     "New Comment",
				CommenterID: sd.UserApprovalComment[0].CommenterID,
				UserID:      sd.UserApprovalComment[0].UserID,
				CreatedDate: sd.UserApprovalComment[0].CreatedDate,
			},
			ExcFunc: func(ctx context.Context) any {
				aprvlStatus, err := busDomain.UserApprovalComment.Update(ctx, sd.UserApprovalComment[0], commentbus.UpdateUserApprovalComment{
					Comment: dbtest.StringPointer("New Comment"),
				})
				if err != nil {
					return err
				}
				return aprvlStatus
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(commentbus.UserApprovalComment)
				if !exists {
					return fmt.Sprintf("got is not an approval status %v", got)
				}
				expResp := exp.(commentbus.UserApprovalComment)
				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func delete(busDomain dbtest.BusDomain, sd unitest.SeedData) []unitest.Table {
	table := []unitest.Table{
		{
			Name:    "delete",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				err := busDomain.UserApprovalComment.Delete(ctx, sd.UserApprovalComment[0])
				return err
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
