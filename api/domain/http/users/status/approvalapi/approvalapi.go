package approvalapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/users/status/approvalapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	userapprovalstatusapp *approvalapp.App
}

func newAPI(userapprovalstatusapp *approvalapp.App) *api {
	return &api{
		userapprovalstatusapp: userapprovalstatusapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app approvalapp.NewUserApprovalStatus
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	status, err := api.userapprovalstatusapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return status
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app approvalapp.UpdateUserApprovalStatus
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	statusID := web.Param(r, "user_approval_status_id")
	parsed, err := uuid.Parse(statusID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	status, err := api.userapprovalstatusapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return status
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	approvalStatusID := web.Param(r, "user_approval_status_id")

	parsed, err := uuid.Parse(approvalStatusID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.userapprovalstatusapp.Delete(ctx, parsed)
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

	statuses, err := api.userapprovalstatusapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return statuses
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	aprvlStatusID := web.Param(r, "user_approval_status_id")

	parsed, err := uuid.Parse(aprvlStatusID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	aprvlStatus, err := api.userapprovalstatusapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return aprvlStatus
}
