package commentapp

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/hr/commentbus"
)

type QueryParams struct {
	Page        string
	Rows        string
	OrderBy     string
	UserID      string
	CommenterID string
	Comment     string
	CreatedDate string
}

type UserApprovalComment struct {
	ID          string `json:"id"`
	CommenterID string `json:"commenter_id"`
	UserID      string `json:"user_id"`
	Comment     string `json:"comment"`
	CreatedDate string `json:"created_date"`
}

func (app UserApprovalComment) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppUserApprovalComment(bus commentbus.UserApprovalComment) UserApprovalComment {
	return UserApprovalComment{
		ID:          bus.ID.String(),
		CommenterID: bus.CommenterID.String(),
		UserID:      bus.UserID.String(),
		Comment:     bus.Comment,
		CreatedDate: bus.CreatedDate.String(),
	}
}

func ToAppUserApprovalComments(bus []commentbus.UserApprovalComment) []UserApprovalComment {
	app := make([]UserApprovalComment, len(bus))
	for i, v := range bus {
		app[i] = ToAppUserApprovalComment(v)
	}
	return app
}

// =============================================================================

type NewUserApprovalComment struct {
	Comment     string `json:"comment" validate:"required"`
	UserID      string `json:"user_id" validate:"required"`
	CommenterID string `json:"commenter_id" validate:"required"`
}

func (app *NewUserApprovalComment) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewUserApprovalComment) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "Validate: %s", err)
	}

	return nil
}

func toBusNewUserApprovalComment(app NewUserApprovalComment) (commentbus.NewUserApprovalComment, error) {
	userID, err := uuid.Parse(app.UserID)
	if err != nil {
		return commentbus.NewUserApprovalComment{}, errs.Newf(errs.InvalidArgument, "parse userID: %s", err)
	}

	commenterID, err := uuid.Parse(app.CommenterID)
	if err != nil {
		return commentbus.NewUserApprovalComment{}, errs.Newf(errs.InvalidArgument, "parse commenterID: %s", err)
	}

	bus := commentbus.NewUserApprovalComment{
		Comment:     app.Comment,
		UserID:      userID,
		CommenterID: commenterID,
	}
	return bus, nil
}

type UpdateUserApprovalComment struct {
	Comment *string `json:"comment" validate:"required"`
}

func (app *UpdateUserApprovalComment) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateUserApprovalComment) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusUpdateUserApprovalComment(app UpdateUserApprovalComment) commentbus.UpdateUserApprovalComment {
	dest := commentbus.UpdateUserApprovalComment{}
	dest.Comment = app.Comment

	return dest
}
