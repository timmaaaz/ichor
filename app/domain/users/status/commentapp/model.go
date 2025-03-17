package commentapp

import (
	"encoding/json"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/users/status/commentbus"
	"github.com/timmaaaz/ichor/business/sdk/convert"
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
	dest := commentbus.NewUserApprovalComment{}

	err := convert.PopulateTypesFromStrings(app, &dest)
	return dest, err
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
