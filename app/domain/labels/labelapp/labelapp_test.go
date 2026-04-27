package labelapp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/labels/labelapp"
	"github.com/timmaaaz/ichor/business/domain/labels/labelbus"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

// --- fakes ---

type memStorer struct {
	mu   sync.Mutex
	rows map[uuid.UUID]labelbus.LabelCatalog
}

func newMemStorer() *memStorer {
	return &memStorer{rows: map[uuid.UUID]labelbus.LabelCatalog{}}
}

func (s *memStorer) NewWithTx(tx sqldb.CommitRollbacker) (labelbus.Storer, error) {
	return s, nil
}

func (s *memStorer) Create(ctx context.Context, lc labelbus.LabelCatalog) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rows[lc.ID] = lc
	return nil
}

func (s *memStorer) Update(ctx context.Context, lc labelbus.LabelCatalog) error {
	return nil
}
func (s *memStorer) Delete(ctx context.Context, lc labelbus.LabelCatalog) error {
	return nil
}
func (s *memStorer) Query(ctx context.Context, _ labelbus.QueryFilter, _ order.By, _ page.Page) ([]labelbus.LabelCatalog, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]labelbus.LabelCatalog, 0, len(s.rows))
	for _, v := range s.rows {
		out = append(out, v)
	}
	return out, nil
}
func (s *memStorer) Count(ctx context.Context, _ labelbus.QueryFilter) (int, error) {
	return len(s.rows), nil
}
func (s *memStorer) QueryByID(ctx context.Context, id uuid.UUID) (labelbus.LabelCatalog, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	lc, ok := s.rows[id]
	if !ok {
		return labelbus.LabelCatalog{}, labelbus.ErrNotFound
	}
	return lc, nil
}
func (s *memStorer) QueryByCode(ctx context.Context, code string) (labelbus.LabelCatalog, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, v := range s.rows {
		if v.Code == code {
			return v, nil
		}
	}
	return labelbus.LabelCatalog{}, labelbus.ErrNotFound
}

type recPrinter struct {
	mu    sync.Mutex
	calls [][]byte
	err   error
}

func (p *recPrinter) SendZPL(ctx context.Context, zpl []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.err != nil {
		return p.err
	}
	cp := make([]byte, len(zpl))
	copy(cp, zpl)
	p.calls = append(p.calls, cp)
	return nil
}

// --- helpers ---

func newApp(t *testing.T) (*labelapp.App, *memStorer, *recPrinter) {
	t.Helper()
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })
	store := newMemStorer()
	printer := &recPrinter{}
	bus := labelbus.NewBusiness(log, delegate.New(log), store, printer)
	return labelapp.NewApp(bus), store, printer
}

// --- tests ---

func Test_Print_CatalogLabel_SingleCopy(t *testing.T) {
	app, store, printer := newApp(t)

	id := uuid.New()
	_ = store.Create(context.Background(), labelbus.LabelCatalog{
		ID: id, Code: "TOTE-001", Type: labelbus.TypeContainer,
	})

	err := app.Print(context.Background(), labelapp.PrintRequest{LabelID: id.String()})
	if err != nil {
		t.Fatalf("Print: %v", err)
	}
	if len(printer.calls) != 1 {
		t.Fatalf("expected 1 SendZPL call, got %d", len(printer.calls))
	}
	if !bytes.Contains(printer.calls[0], []byte("TOTE-001")) {
		t.Fatalf("expected ZPL to contain TOTE-001, got: %s", printer.calls[0])
	}
}

func Test_Print_CatalogLabel_MultipleCopies(t *testing.T) {
	app, store, printer := newApp(t)

	id := uuid.New()
	_ = store.Create(context.Background(), labelbus.LabelCatalog{
		ID: id, Code: "LOC-A1", Type: labelbus.TypeLocation,
	})

	err := app.Print(context.Background(), labelapp.PrintRequest{LabelID: id.String(), Copies: 3})
	if err != nil {
		t.Fatalf("Print: %v", err)
	}
	if len(printer.calls) != 3 {
		t.Fatalf("expected 3 SendZPL calls, got %d", len(printer.calls))
	}
}

func Test_Print_NotFound(t *testing.T) {
	app, _, printer := newApp(t)

	err := app.Print(context.Background(), labelapp.PrintRequest{LabelID: uuid.New().String()})
	if err == nil {
		t.Fatal("expected error for missing label, got nil")
	}
	if len(printer.calls) != 0 {
		t.Fatalf("printer should not be called on not-found, got %d calls", len(printer.calls))
	}
}

func Test_Print_BadUUID(t *testing.T) {
	app, _, _ := newApp(t)
	if err := app.Print(context.Background(), labelapp.PrintRequest{LabelID: "not-a-uuid"}); err == nil {
		t.Fatal("expected error for bad uuid, got nil")
	}
}

func Test_Print_PrinterError(t *testing.T) {
	app, store, printer := newApp(t)
	printer.err = errors.New("boom")

	id := uuid.New()
	_ = store.Create(context.Background(), labelbus.LabelCatalog{
		ID: id, Code: "X", Type: labelbus.TypeContainer,
	})

	if err := app.Print(context.Background(), labelapp.PrintRequest{LabelID: id.String()}); err == nil {
		t.Fatal("expected printer error to surface, got nil")
	}
}

func Test_RenderPrint_Receiving(t *testing.T) {
	app, _, printer := newApp(t)

	payload := map[string]any{
		"productName": "Widget",
		"sku":         "SKU-1",
		"upc":         "012345678905",
		"lotNumber":   nil,
		"expiryDate":  nil,
		"quantity":    10,
		"poNumber":    "PO-42",
	}
	raw, _ := json.Marshal(payload)

	err := app.RenderPrint(context.Background(), labelapp.RenderPrintRequest{
		Type:    labelbus.TypeReceiving,
		Payload: raw,
		Copies:  2,
	})
	if err != nil {
		t.Fatalf("RenderPrint: %v", err)
	}
	if len(printer.calls) != 2 {
		t.Fatalf("expected 2 SendZPL calls, got %d", len(printer.calls))
	}
	if !strings.Contains(string(printer.calls[0]), "PO-42") {
		t.Fatalf("expected ZPL to contain PO-42, got: %s", printer.calls[0])
	}
}

func Test_RenderPrint_BadType(t *testing.T) {
	app, _, _ := newApp(t)
	err := app.RenderPrint(context.Background(), labelapp.RenderPrintRequest{
		Type:    "nonsense",
		Payload: json.RawMessage(`{}`),
	})
	if err == nil {
		t.Fatal("expected error for unknown type, got nil")
	}
}
