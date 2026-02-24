package putawaytaskapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/inventory/putawaytaskapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	putawaytaskapp *putawaytaskapp.App
}

func newAPI(putawaytaskapp *putawaytaskapp.App) *api {
	return &api{
		putawaytaskapp: putawaytaskapp,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app putawaytaskapp.NewPutAwayTask
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	task, err := api.putawaytaskapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return task
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app putawaytaskapp.UpdatePutAwayTask
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	taskID := web.Param(r, "task_id")
	parsed, err := uuid.Parse(taskID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	task, err := api.putawaytaskapp.Update(ctx, parsed, app)
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

	if err := api.putawaytaskapp.Delete(ctx, parsed); err != nil {
		return errs.NewError(err)
	}

	return nil
}

func (api *api) query(ctx context.Context, r *http.Request) web.Encoder {
	qp, err := parseQueryParams(r)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	tasks, err := api.putawaytaskapp.Query(ctx, qp)
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

	task, err := api.putawaytaskapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return task
}
