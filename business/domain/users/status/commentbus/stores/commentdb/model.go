package commentdb

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/users/status/commentbus"
)

type userApprovalComment struct {
	ID          uuid.UUID `db:"comment_id"`
	Comment     string    `db:"comment"`
	UserID      uuid.UUID `db:"user_id"`
	CommenterID uuid.UUID `db:"commenter_id"`
	CreatedDate time.Time `db:"created_date"`
}

func toDBUserApprovalComment(as commentbus.UserApprovalComment) userApprovalComment {
	return userApprovalComment{
		ID:          as.ID,
		Comment:     as.Comment,
		UserID:      as.UserID,
		CommenterID: as.CommenterID,
		CreatedDate: as.CreatedDate,
	}
}

func toBusUserApprovalComment(dbAS userApprovalComment) commentbus.UserApprovalComment {
	return commentbus.UserApprovalComment{
		ID:          dbAS.ID,
		Comment:     dbAS.Comment,
		CommenterID: dbAS.CommenterID,
		UserID:      dbAS.UserID,
		CreatedDate: dbAS.CreatedDate.Truncate(time.Second),
	}
}

func toBusUserApprovalComments(dbAS []userApprovalComment) []commentbus.UserApprovalComment {
	aprvlStatuses := make([]commentbus.UserApprovalComment, len(dbAS))
	for i, as := range dbAS {
		aprvlStatuses[i] = toBusUserApprovalComment(as)
	}

	return aprvlStatuses
}
