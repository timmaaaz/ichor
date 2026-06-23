package executionapp

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
)

type fakeReranner struct {
	newID uuid.UUID
	err   error
	gotID uuid.UUID
}

func (f *fakeReranner) RerunExecution(_ context.Context, id uuid.UUID) (uuid.UUID, error) {
	f.gotID = id
	return f.newID, f.err
}

func TestRerun_Success(t *testing.T) {
	orig := uuid.New()
	fresh := uuid.New()
	app := NewApp(&fakeReranner{newID: fresh})

	resp, err := app.Rerun(context.Background(), orig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.OriginalExecutionID != orig || resp.NewExecutionID != fresh {
		t.Fatalf("resp = %+v", resp)
	}
}

func TestRerun_NotRerunnable_FailedPrecondition(t *testing.T) {
	app := NewApp(&fakeReranner{err: temporal.ErrExecutionNotRerunnable})
	_, err := app.Rerun(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error")
	}
	// Assert errs.FailedPrecondition code via errors.As on *errs.Error.
	var appErr *errs.Error
	if errors.As(err, &appErr) {
		if appErr.Code != errs.FailedPrecondition {
			t.Fatalf("expected FailedPrecondition code, got %v", appErr.Code)
		}
	} else {
		// Fallback: message substring check
		if !strings.Contains(err.Error(), "re-run") {
			t.Fatalf("err = %v", err)
		}
	}
}

func TestRerun_NotFound(t *testing.T) {
	app := NewApp(&fakeReranner{err: errors.New("not found test")})
	_, err := app.Rerun(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRerun_NilRerunner(t *testing.T) {
	app := NewApp(nil)
	_, err := app.Rerun(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error when rerunner is nil")
	}
	var appErr *errs.Error
	if !errors.As(err, &appErr) || appErr.Code != errs.Internal {
		t.Fatalf("expected Internal error, got %v", err)
	}
}
