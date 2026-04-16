package apitest

import (
	"context"
	"net/http/httptest"
	"sync"
	"testing"

	authbuild "github.com/timmaaaz/ichor/api/cmd/services/auth/build/all"
	ichorbuild "github.com/timmaaaz/ichor/api/cmd/services/ichor/build/all"
	"github.com/timmaaaz/ichor/api/sdk/http/mux"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

// RecPrinter is a thread-safe recording implementation of mux.LabelPrinter.
// Tests inject one via StartLabelTest, exercise label-print routes, and
// then assert against the recorded ZPL bytes — no TCP, no real printer.
type RecPrinter struct {
	mu    sync.Mutex
	calls [][]byte
	err   error
}

// NewRecPrinter constructs a fresh recording printer.
func NewRecPrinter() *RecPrinter {
	return &RecPrinter{}
}

// SendZPL satisfies mux.LabelPrinter. Each successful call appends a copy
// of the ZPL bytes to the call log so callers can assert on payload shape.
func (p *RecPrinter) SendZPL(_ context.Context, zpl []byte) error {
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

// SetError installs an error that all subsequent SendZPL calls will return.
// Use to exercise the printer-failure code path.
func (p *RecPrinter) SetError(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.err = err
}

// Calls returns a snapshot of the recorded ZPL payloads in dispatch order.
func (p *RecPrinter) Calls() [][]byte {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([][]byte, len(p.calls))
	for i, c := range p.calls {
		cp := make([]byte, len(c))
		copy(cp, c)
		out[i] = cp
	}
	return out
}

// Reset clears the call log and any installed error.
func (p *RecPrinter) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls = nil
	p.err = nil
}

// StartLabelTest mirrors StartTest but wires a RecPrinter into the label
// subsystem so print and render-print integration tests can assert ZPL
// dispatch without touching real hardware.
func StartLabelTest(t *testing.T, testName string) (*Test, *RecPrinter) {
	db := dbtest.NewDatabase(t, testName)

	ath, err := auth.New(auth.Config{
		Log:       db.Log,
		DB:        db.DB,
		KeyLookup: &KeyStore{},
	})
	if err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(mux.WebAPI(mux.Config{
		Log:  db.Log,
		Auth: ath,
		DB:   db.DB,
	}, authbuild.Routes()))

	authClient := authclient.New(db.Log, server.URL)

	printer := NewRecPrinter()

	mux := mux.WebAPI(mux.Config{
		Log:          db.Log,
		AuthClient:   authClient,
		DB:           db.DB,
		LabelPrinter: printer,
	}, ichorbuild.Routes())

	return New(db, ath, mux), printer
}
