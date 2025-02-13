package commentapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/users/status/commentapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	userapprovalcommentapp *commentapp.App
}

func newAPI(userapprovalcommentapp *commentapp.App) *api {
	return &api{
		userapprovalcommentapp: userapprovalcommentapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app commentapp.NewUserApprovalComment
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	status, err := api.userapprovalcommentapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return status
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app commentapp.UpdateUserApprovalComment
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	commentID := web.Param(r, "user_status_comment_id")
	parsed, err := uuid.Parse(commentID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	status, err := api.userapprovalcommentapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return status
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	commentStatusID := web.Param(r, "user_status_comment_id")

	parsed, err := uuid.Parse(commentStatusID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.userapprovalcommentapp.Delete(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return nil
}

func (api *api) query(ctx context.Context, r *http.Request) web.Encoder {
	qp, err := parseQueryParams(r)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	comments, err := api.userapprovalcommentapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return comments
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	commentID := web.Param(r, "user_status_comment_id")

	parsed, err := uuid.Parse(commentID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	comment, err := api.userapprovalcommentapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return comment
}
