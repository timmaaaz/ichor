package directedworkapi

import (
	"context"
	"net/http"

	"github.com/timmaaaz/ichor/app/domain/floor/directedworkapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/mid"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	directedWorkApp *directedworkapp.App
}

func newAPI(directedWorkApp *directedworkapp.App) *api {
	return &api{
		directedWorkApp: directedWorkApp,
	}
}

// queryNext returns the single best next work item for the authenticated
// worker, or {"work_item": null} if nothing is directed.
func (a *api) queryNext(ctx context.Context, r *http.Request) web.Encoder {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return errs.New(errs.Unauthenticated, err)
	}

	item, err := a.directedWorkApp.QueryNext(ctx, userID)
	if err != nil {
		return errs.NewError(err)
	}

	return Response{WorkItem: item}
}
