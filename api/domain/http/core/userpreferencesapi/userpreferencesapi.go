package userpreferencesapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/core/userpreferencesapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	userPreferencesApp *userpreferencesapp.App
}

func newAPI(userPreferencesApp *userpreferencesapp.App) *api {
	return &api{
		userPreferencesApp: userPreferencesApp,
	}
}

func (api *api) set(ctx context.Context, r *http.Request) web.Encoder {
	var app userpreferencesapp.NewUserPreference
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	userID, err := uuid.Parse(web.Param(r, "user_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	key := web.Param(r, "key")

	pref, err := api.userPreferencesApp.Set(ctx, userID, key, app)
	if err != nil {
		return errs.NewError(err)
	}

	return pref
}

func (api *api) get(ctx context.Context, r *http.Request) web.Encoder {
	userID, err := uuid.Parse(web.Param(r, "user_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	key := web.Param(r, "key")

	pref, err := api.userPreferencesApp.Get(ctx, userID, key)
	if err != nil {
		return errs.NewError(err)
	}

	return pref
}

func (api *api) getAll(ctx context.Context, r *http.Request) web.Encoder {
	userID, err := uuid.Parse(web.Param(r, "user_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	prefs, err := api.userPreferencesApp.GetAll(ctx, userID)
	if err != nil {
		return errs.NewError(err)
	}

	return prefs
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	userID, err := uuid.Parse(web.Param(r, "user_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	key := web.Param(r, "key")

	if err := api.userPreferencesApp.Delete(ctx, userID, key); err != nil {
		return errs.NewError(err)
	}

	return nil
}
