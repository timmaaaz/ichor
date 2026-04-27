// Package paperworkbus_test verifies the paperwork business slice scaffold.
package paperworkbus_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/paperwork/paperworkbus"
)

func TestBootstrap(t *testing.T) {
	t.Parallel()

	bus := paperworkbus.NewBusiness(nil)
	if bus == nil {
		t.Fatal("NewBusiness returned nil")
	}

	ctx := context.Background()

	if _, err := bus.BuildPickSheet(ctx, paperworkbus.PickSheetRequest{OrderID: uuid.New()}); !errors.Is(err, paperworkbus.ErrNotImplemented) {
		t.Errorf("BuildPickSheet: want ErrNotImplemented, got %v", err)
	}
	if _, err := bus.BuildReceiveCover(ctx, paperworkbus.ReceiveCoverRequest{PurchaseOrderID: uuid.New()}); !errors.Is(err, paperworkbus.ErrNotImplemented) {
		t.Errorf("BuildReceiveCover: want ErrNotImplemented, got %v", err)
	}
	if _, err := bus.BuildTransferSheet(ctx, paperworkbus.TransferSheetRequest{TransferID: uuid.New()}); !errors.Is(err, paperworkbus.ErrNotImplemented) {
		t.Errorf("BuildTransferSheet: want ErrNotImplemented, got %v", err)
	}
}
