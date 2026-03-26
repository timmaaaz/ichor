package picktaskapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/inventory/picktaskapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	picktaskapp *picktaskapp.App
}

func newAPI(picktaskapp *picktaskapp.App) *api {
	return &api{
		picktaskapp: picktaskapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app picktaskapp.NewPickTask
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	task, err := api.picktaskapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return task
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app picktaskapp.UpdatePickTask
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	taskID := web.Param(r, "task_id")
	parsed, err := uuid.Parse(taskID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	task, err := api.picktaskapp.Update(ctx, parsed, app)
	if err != nil {
		return errs.NewError(err)
	}

	return task
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	taskID := web.Param(r, "task_id")
	parsed, err := uuid.Parse(taskID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if err := api.picktaskapp.Delete(ctx, parsed); err != nil {
		return errs.NewError(err)
	}

	return nil
}

func (api *api) query(ctx context.Context, r *http.Request) web.Encoder {
	qp, err := parseQueryParams(r)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	tasks, err := api.picktaskapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return tasks
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	taskID := web.Param(r, "task_id")
	parsed, err := uuid.Parse(taskID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	task, err := api.picktaskapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return task
}
