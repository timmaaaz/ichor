package paperworkapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/app/domain/paperwork/paperworkapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/paperwork/paperworkbus"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	app *paperworkapp.App
}

func newAPI(app *paperworkapp.App) *api {
	return &api{app: app}
}

// pickSheet handles GET /v1/paperwork/pick-sheet?order_id=&zone=
func (api *api) pickSheet(ctx context.Context, r *http.Request) web.Encoder {
	orderID, err := uuid.Parse(r.URL.Query().Get("order_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, errors.New("order_id must be a valid uuid"))
	}
	req := paperworkapp.PickSheetRequest{
		OrderID: orderID,
		Zone:    r.URL.Query().Get("zone"),
	}
	if _, err := api.app.BuildPickSheet(ctx, req); err != nil {
		return mapErr(err)
	}
	return errs.Newf(errs.Internal, "unexpected nil error from build")
}

// receiveCover handles GET /v1/paperwork/receive-cover?po_id=
func (api *api) receiveCover(ctx context.Context, r *http.Request) web.Encoder {
	poID, err := uuid.Parse(r.URL.Query().Get("po_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, errors.New("po_id must be a valid uuid"))
	}
	req := paperworkapp.ReceiveCoverRequest{PurchaseOrderID: poID}
	if _, err := api.app.BuildReceiveCover(ctx, req); err != nil {
		return mapErr(err)
	}
	return errs.Newf(errs.Internal, "unexpected nil error from build")
}

// transferSheet handles GET /v1/paperwork/transfer-sheet?transfer_id=
func (api *api) transferSheet(ctx context.Context, r *http.Request) web.Encoder {
	transferID, err := uuid.Parse(r.URL.Query().Get("transfer_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, errors.New("transfer_id must be a valid uuid"))
	}
	req := paperworkapp.TransferSheetRequest{TransferID: transferID}
	if _, err := api.app.BuildTransferSheet(ctx, req); err != nil {
		return mapErr(err)
	}
	return errs.Newf(errs.Internal, "unexpected nil error from build")
}

// mapErr translates business-layer errors to HTTP-layer error responses.
// Phase 0g.B2 maps ErrNotImplemented → 501. B3 expands this with NotFound /
// InvalidArgument mappings as real handler bodies land.
func mapErr(err error) web.Encoder {
	if errors.Is(err, paperworkbus.ErrNotImplemented) {
		return errs.New(errs.Unimplemented, err)
	}
	return errs.New(errs.Internal, err)
}
