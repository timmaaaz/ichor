package lottrackingsapi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/inventory/lottrackingsapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/app/sdk/mid"
	"github.com/timmaaaz/ichor/business/domain/config/settingsbus"
	"github.com/timmaaaz/ichor/foundation/web"
)

type qualityStatuses []string

func (qs qualityStatuses) Encode() ([]byte, string, error) {
	data, err := json.Marshal(qs)
	return data, "application/json", err
}

type api struct {
	lottrackingsapp *lottrackingsapp.App
	settingsBus     *settingsbus.Business
}

func newAPI(lotTrackingsApp *lottrackingsapp.App, settingsBus *settingsbus.Business) *api {
	return &api{
		lottrackingsapp: lotTrackingsApp,
		settingsBus:     settingsBus,
	}
}

func (api *api) create(ctx context.Context, r *http.Request) web.Encoder {
	var app lottrackingsapp.NewLotTrackings
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	lotTrackings, err := api.lottrackingsapp.Create(ctx, app)
	if err != nil {
		return errs.NewError(err)
	}

	return lotTrackings
}

func (api *api) update(ctx context.Context, r *http.Request) web.Encoder {
	var app lottrackingsapp.UpdateLotTrackings
	if err := web.Decode(r, &app); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	lotTrackingsID := web.Param(r, "lot_id")
	parsed, err := uuid.Parse(lotTrackingsID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	// If quality_status is being set to a restricted value, enforce the quarantine access setting.
	if app.QualityStatus != nil && (*app.QualityStatus == "quarantined" || *app.QualityStatus == "on_hold") {
		if api.settingsBus != nil {
			setting, err := api.settingsBus.QueryByKey(ctx, "inventory.quarantine_access")
			if err == nil {
				var accessPolicy string
				if jsonErr := json.Unmarshal(setting.Value, &accessPolicy); jsonErr == nil && accessPolicy == "supervisor_only" {
					claims := mid.GetClaims(ctx)
					if !hasAdminRole(claims.Roles) {
						return errs.Newf(errs.PermissionDenied, "quarantine access requires supervisor role")
					}
				}
			}
		}
	}

	lotTrackings, err := api.lottrackingsapp.Update(ctx, app, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return lotTrackings
}

func hasAdminRole(roles []string) bool {
	for _, r := range roles {
		if r == "ADMIN" {
			return true
		}
	}
	return false
}

func (api *api) delete(ctx context.Context, r *http.Request) web.Encoder {
	lotTrackingsID := web.Param(r, "lot_id")
	parsed, err := uuid.Parse(lotTrackingsID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	err = api.lottrackingsapp.Delete(ctx, parsed)
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

	lotTrackingss, err := api.lottrackingsapp.Query(ctx, qp)
	if err != nil {
		return errs.NewError(err)
	}

	return lotTrackingss
}

func (api *api) queryByID(ctx context.Context, r *http.Request) web.Encoder {
	lotTrackingsID := web.Param(r, "lot_id")
	parsed, err := uuid.Parse(lotTrackingsID)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	lotTrackings, err := api.lottrackingsapp.QueryByID(ctx, parsed)
	if err != nil {
		return errs.NewError(err)
	}

	return lotTrackings
}

func (api *api) queryQualityStatuses(ctx context.Context, r *http.Request) web.Encoder {
	return qualityStatuses{"good", "on_hold", "quarantined", "released", "expired"}
}
