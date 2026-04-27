package paperworkapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/app/domain/paperwork/paperworkapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/foundation/web"
)

type api struct {
	app *paperworkapp.App
}

func newAPI(app *paperworkapp.App) *api {
	return &api{app: app}
}

// pdfResponse implements web.Encoder for raw PDF bytes with
// Content-Type: application/pdf. Phase 0g.B2 never reaches the success path
// because the bus returns ErrNotImplemented; B3 fills in real bodies and
// these handlers stream PDFs through this encoder unchanged.
type pdfResponse struct {
	bytes []byte
}

func (p pdfResponse) Encode() ([]byte, string, error) {
	return p.bytes, "application/pdf", nil
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
	pdf, err := api.app.BuildPickSheet(ctx, req)
	if err != nil {
		return errs.NewError(err)
	}
	return pdfResponse{bytes: pdf}
}

// receiveCover handles GET /v1/paperwork/receive-cover?po_id=
func (api *api) receiveCover(ctx context.Context, r *http.Request) web.Encoder {
	poID, err := uuid.Parse(r.URL.Query().Get("po_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, errors.New("po_id must be a valid uuid"))
	}
	req := paperworkapp.ReceiveCoverRequest{PurchaseOrderID: poID}
	pdf, err := api.app.BuildReceiveCover(ctx, req)
	if err != nil {
		return errs.NewError(err)
	}
	return pdfResponse{bytes: pdf}
}

// transferSheet handles GET /v1/paperwork/transfer-sheet?transfer_id=
func (api *api) transferSheet(ctx context.Context, r *http.Request) web.Encoder {
	transferID, err := uuid.Parse(r.URL.Query().Get("transfer_id"))
	if err != nil {
		return errs.New(errs.InvalidArgument, errors.New("transfer_id must be a valid uuid"))
	}
	req := paperworkapp.TransferSheetRequest{TransferID: transferID}
	pdf, err := api.app.BuildTransferSheet(ctx, req)
	if err != nil {
		return errs.NewError(err)
	}
	return pdfResponse{bytes: pdf}
}
