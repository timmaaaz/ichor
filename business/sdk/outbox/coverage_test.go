package outbox_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestCoverage_EveryCascadeBusEmits is the F2 coverage trip-wire (DESIGN §8).
//
// Invariant: any business/domain bus that fires a cascade delegate event
// (b.delegate.Call(ctx, …) / b.del.Call(ctx, …)) MUST also persist that event to
// the transactional outbox (b.outbox.Emit(ctx, …)). After the F2 cutover removes
// the delegate's cascade subscriber, a bus that still fires the delegate but forgot
// to emit would SILENTLY drop its cascade — this test turns that into a build-time
// failure instead of a lost cascade in production.
//
// It is the practical equivalent of "every domain in workflowdomains.Registrations()
// emits": every Registrations bus fires a delegate event, and the only domain buses
// that fire one without being cascade-registered are the explicit exclusions below.
//
// Scope is business/domain only. The two non-bus cascade sources — workflow.Business
// (allocation_results / M2, via its injected emitter) and the generic data handlers
// (Path C, via fireSynthesizedEvent) — emit through different seams and are covered by
// their own wiring + the F6 reliability suite; scanning them generically here would
// false-flag their legitimately emit-less rule-lifecycle delegate calls.
func TestCoverage_EveryCascadeBusEmits(t *testing.T) {
	t.Parallel()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// business/sdk/outbox/coverage_test.go -> business/domain
	domainDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "domain")
	if _, err := os.Stat(domainDir); err != nil {
		t.Fatalf("domain dir not found at %s: %v", domainDir, err)
	}

	// Buses that fire a delegate event but are NOT in workflowdomains.Registrations()
	// (so they never had a cascade subscriber and correctly do not emit to the outbox).
	excluded := map[string]bool{
		"productuombus":      true,
		"settingsbus":        true,
		"approvalrequestbus": true,
	}

	var missing []string
	checked := 0
	err := filepath.WalkDir(domainDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		src, rerr := os.ReadFile(path)
		if rerr != nil {
			return rerr
		}
		s := string(src)

		firesCascadeDelegate := strings.Contains(s, ".delegate.Call(ctx,") ||
			strings.Contains(s, ".del.Call(ctx,")
		if !firesCascadeDelegate {
			return nil
		}

		pkg := filepath.Base(filepath.Dir(path))
		if excluded[pkg] {
			return nil
		}

		checked++
		if !strings.Contains(s, ".outbox.Emit(ctx,") {
			rel := strings.TrimPrefix(path, filepath.Dir(domainDir)+string(filepath.Separator))
			missing = append(missing, rel)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking domain dir: %v", err)
	}

	// Floor guard (not an exact count — won't drift as buses are added): if the scan
	// matched far fewer cascade buses than the ~64 we know fire delegate events, the
	// detection pattern has silently stopped matching and this test is vacuously passing.
	if checked < 40 {
		t.Fatalf("coverage scan only matched %d cascade buses (expected ~64) — the "+
			"delegate-call detection pattern likely broke; this test would pass vacuously", checked)
	}

	if len(missing) > 0 {
		t.Fatalf("F2 coverage gap: these cascade buses fire a delegate event but do NOT "+
			"emit to the transactional outbox — they would silently drop their cascade after "+
			"the cutover removes the delegate subscriber. Add b.outbox.Emit (or add the package "+
			"to the documented exclusions if it is not in Registrations()):\n  %s",
			strings.Join(missing, "\n  "))
	}
}
