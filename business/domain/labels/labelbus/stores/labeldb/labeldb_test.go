package labeldb_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/labels/labelbus"
	"github.com/timmaaaz/ichor/business/domain/labels/labelbus/stores/labeldb"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

func Test_LabelCatalog_CreateQueryByCode(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_LabelCatalog_CreateQueryByCode")

	store := labeldb.NewStore(db.Log, db.DB)

	ctx := context.Background()

	lc := labelbus.LabelCatalog{
		ID:          uuid.New(),
		Code:        "LOC-A-01",
		Type:        labelbus.TypeLocation,
		EntityRef:   "warehouse-a/aisle-01",
		PayloadJSON: `{"zone":"A"}`,
		CreatedDate: time.Now(),
	}

	if err := store.Create(ctx, lc); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := store.QueryByCode(ctx, lc.Code)
	if err != nil {
		t.Fatalf("querybycode: %v", err)
	}

	if got.ID != lc.ID {
		t.Errorf("ID: got %s, want %s", got.ID, lc.ID)
	}
	if got.Code != lc.Code {
		t.Errorf("Code: got %q, want %q", got.Code, lc.Code)
	}
	if got.Type != lc.Type {
		t.Errorf("Type: got %q, want %q", got.Type, lc.Type)
	}
	if got.EntityRef != lc.EntityRef {
		t.Errorf("EntityRef: got %q, want %q", got.EntityRef, lc.EntityRef)
	}
	if got.PayloadJSON != lc.PayloadJSON {
		t.Errorf("PayloadJSON: got %q, want %q", got.PayloadJSON, lc.PayloadJSON)
	}
}

func Test_LabelCatalog_DuplicateCode(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_LabelCatalog_DuplicateCode")

	store := labeldb.NewStore(db.Log, db.DB)
	ctx := context.Background()

	lc := labelbus.LabelCatalog{
		ID:          uuid.New(),
		Code:        "DUPE-01",
		Type:        labelbus.TypeTote,
		PayloadJSON: `{}`,
		CreatedDate: time.Now(),
	}
	if err := store.Create(ctx, lc); err != nil {
		t.Fatalf("create first: %v", err)
	}

	lc2 := lc
	lc2.ID = uuid.New()
	err := store.Create(ctx, lc2)
	if err == nil {
		t.Fatalf("expected duplicate-code error, got nil")
	}
	// Business sentinel should be wrapped.
	if !isUniqueCode(err) {
		t.Fatalf("expected ErrUniqueCode in chain, got %v", err)
	}
}

func isUniqueCode(err error) bool {
	for e := err; e != nil; {
		if e == labelbus.ErrUniqueCode {
			return true
		}
		type unwrapper interface{ Unwrap() error }
		u, ok := e.(unwrapper)
		if !ok {
			return false
		}
		e = u.Unwrap()
	}
	return false
}
