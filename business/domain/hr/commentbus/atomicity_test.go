package commentbus_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/hr/commentbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
)

// Test_UserComment_TxRollback_RevertsUserStatus is the regression guard for the commentbus
// nested-write atomicity fix (found in code review of the formdata-atomic-submit change).
//
// commentbus.Create does TWO writes: it inserts the comment via its own storer AND calls
// userbus.SetUnderReview, which UPDATEs core.users (user.UserApprovalStatus -> "UNDER REVIEW").
// When commentbus is tx-bound via NewWithTx, the comment storer rides the tx — but the nested
// userbus must ALSO be rebound onto the same tx, or SetUnderReview's UPDATE autocommits on the
// base pool and survives a rollback. That split is the atomicity hole this guards: a rolled-back
// form submit that created a comment would otherwise leave the user's status change orphaned.
//
// The decisive assertion is the user's status after rollback (the comment rolls back either way
// since it rides commentbus's own storer):
//
//	RED  (nested userbus pool-bound): SetUnderReview's UPDATE commits on the pool -> after
//	     rollback the user is still "UNDER REVIEW" (status changed) -> fails.
//	GREEN (nested userbus tx-bound):  the UPDATE rides the tx -> rollback reverts it -> status
//	     unchanged -> passes.
func Test_UserComment_TxRollback_RevertsUserStatus(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_UserComment_TxRollback_RevertsUserStatus")
	ctx := context.Background()

	// Two real users: a target (gets commented on / set under review) and a commenter.
	users, err := userbus.TestSeedUsersWithNoFKs(ctx, 2, userbus.Roles.User, db.BusDomain.User)
	require.NoError(t, err, "seeding users")
	target := users[0]
	commenter := users[1]

	// Capture the target's approval status BEFORE the rolled-back submit.
	before, err := db.BusDomain.User.QueryByID(ctx, target.ID)
	require.NoError(t, err, "query target before")
	origStatus := before.UserApprovalStatus

	// Open a transaction, enroll it on ctx (outbox), and build a tx-bound commentbus — exactly
	// what formdataapp.UpsertFormData does for the hr.user_approval_comments registry write.
	tx, err := db.DB.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	require.NoError(t, err, "begin tx")
	defer tx.Rollback()
	txCtx := sqldb.WithTx(ctx, tx)

	cbTx, err := db.BusDomain.UserApprovalComment.NewWithTx(tx)
	require.NoError(t, err, "commentbus NewWithTx")

	created, err := cbTx.Create(txCtx, commentbus.NewUserApprovalComment{
		Comment:     "rollback regression",
		CommenterID: commenter.ID,
		UserID:      target.ID,
	})
	require.NoError(t, err, "create on tx (writes comment + sets user under review, both on the tx)")
	require.Equal(t, target.ID, created.UserID, "sanity: created comment is for the target user")

	// Roll the whole submit back, as a failing multi-entity form submit would.
	require.NoError(t, tx.Rollback(), "rollback")

	// DECISIVE: the user-status UPDATE must roll back with the tx.
	after, err := db.BusDomain.User.QueryByID(ctx, target.ID)
	require.NoError(t, err, "query target after")
	require.Equal(t, origStatus, after.UserApprovalStatus,
		"atomicity violated: commentbus.Create set the user UNDER REVIEW via a nested userbus that "+
			"did not ride the rolled-back tx, so the status write escaped to the base pool. "+
			"commentbus.NewWithTx must rebind the nested userbus onto the same tx.")

	// Corroborate: the comment write rolled back too (no comments seeded, so the table is empty).
	comments, err := db.BusDomain.UserApprovalComment.Query(ctx,
		commentbus.QueryFilter{}, order.NewBy(commentbus.OrderByID, order.ASC), page.MustParse("1", "100"))
	require.NoError(t, err, "query comments after")
	for _, c := range comments {
		require.NotEqual(t, created.ID, c.ID, "the comment created on the rolled-back tx must not be committed")
	}
}
