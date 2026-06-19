package outbox_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/timmaaaz/ichor/business/sdk/workflowdomains"
)

// TestCoverage_EveryCascadeBusEmits is the F2 coverage trip-wire (DESIGN §8 unit #17).
//
// Invariant: any business/domain bus that fires a cascade delegate event
// (b.delegate.Call(ctx, …) / b.del.Call(ctx, …)) MUST also persist that event to the
// transactional outbox (b.outbox.Emit(ctx, …)). After the F2 cutover removed the delegate's
// cascade subscriber, a bus that still fires the delegate but forgot to emit would SILENTLY
// drop its cascade — this turns that into a build-time failure instead of a lost cascade.
//
// The test is driven off workflowdomains.Registrations() — the authoritative cascade domain
// set the outbox routes on — rather than a frozen magic count. Two guards:
//
//  1. Coverage: every delegate-firing business/domain bus package also emits, except the
//     documented exclusions (buses that fire a delegate event but are NOT cascade-registered).
//  2. Registry floor: the number of bus packages that emit must be at least the number of
//     schema-qualified Registrations() domains. This kills the old frozen floor AND makes the
//     exclusions unable to hide a registered miss — if a REGISTERED domain's bus stops emitting,
//     the emit count drops below the registry count and this fails regardless of `excluded`.
//
// Scope is business/domain only. The two non-bus cascade sources — workflow.Business
// (allocation_results, via its injected emitter) and the generic data handlers (Path C, via
// fireSynthesizedEvent) — emit through different seams (both are schema-"" Registrations
// entries) and are covered by their own wiring + the reliability suite.
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

	// Buses that fire a cascade delegate event but are NOT in workflowdomains.Registrations()
	// (no cascade subscriber → correctly emit-less). Drift here cannot vacuously hide a real
	// miss: the registry floor below fails if any REGISTERED domain stops emitting, excluded or not.
	excluded := map[string]bool{
		"productuombus":      true,
		"settingsbus":        true,
		"approvalrequestbus": true,
	}

	// Detection is per-package, not per-file: a bus may fire its delegate call and its outbox
	// Emit from different files in the same package.
	firing := map[string]bool{}   // packages firing a cascade delegate event
	emitting := map[string]bool{} // packages emitting to the transactional outbox
	wrapped := map[string]bool{}  // packages that wrap their outbox writes in WriteAtomic/WriteAtomicVoid
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
		pkg := filepath.Base(filepath.Dir(path))

		if strings.Contains(s, ".outbox.Emit(ctx,") {
			emitting[pkg] = true
		}
		if strings.Contains(s, ".delegate.Call(ctx,") || strings.Contains(s, ".del.Call(ctx,") {
			firing[pkg] = true
		}
		if strings.Contains(s, "outbox.WriteAtomic(") || strings.Contains(s, "outbox.WriteAtomicVoid(") {
			wrapped[pkg] = true
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking domain dir: %v", err)
	}

	// Guard 1 — coverage: every delegate-firing cascade bus must also emit (minus exclusions).
	var missing []string
	for pkg := range firing {
		if excluded[pkg] || emitting[pkg] {
			continue
		}
		missing = append(missing, pkg)
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Fatalf("F2 coverage gap: these business/domain cascade buses fire a delegate event but do "+
			"NOT emit to the transactional outbox — their cascade would silently vanish. Add b.outbox.Emit "+
			"(or add the package to the documented exclusions if it is not in Registrations()):\n  %s",
			strings.Join(missing, "\n  "))
	}

	// Guard 2 — registry floor: every schema-qualified Registrations() entry is a business/domain
	// cascade domain, so at least that many bus packages must emit. Tying the floor to the registry
	// (instead of a frozen number) means it tracks growth, fails if emit detection silently breaks,
	// and cannot pass vacuously when `excluded` drifts to mask a registered miss.
	wantCascadeDomains := 0
	for _, r := range workflowdomains.Registrations() {
		if r.Schema != "" {
			wantCascadeDomains++
		}
	}
	if len(emitting) < wantCascadeDomains {
		t.Fatalf("coverage found %d business/domain buses emitting to the outbox, but "+
			"workflowdomains.Registrations() declares %d schema-qualified cascade domains — either the "+
			"emit detection broke (vacuous-pass risk) or a registered domain's bus stopped emitting",
			len(emitting), wantCascadeDomains)
	}

	// Guard 3 — atomicity: every emitting bus must wrap its outbox writes in
	// outbox.WriteAtomic or outbox.WriteAtomicVoid. A plain b.outbox.Emit call outside
	// the atomic wrapper means the outbox row is written without the application tx, so
	// a subsequent tx rollback silently drops the cascade. FF#2 wrapped them all; this
	// guard ensures future buses don't regress.
	var unwrapped []string
	for pkg := range emitting {
		if excluded[pkg] || wrapped[pkg] {
			continue
		}
		unwrapped = append(unwrapped, pkg)
	}
	sort.Strings(unwrapped)
	if len(unwrapped) > 0 {
		t.Fatalf("FF#2 atomicity gap: these cascade buses emit to the outbox but do NOT wrap "+
			"their writes in outbox.WriteAtomic/WriteAtomicVoid — a simple-write emit failure would "+
			"silently lose the cascade. Wrap each emitting method:\n  %s", strings.Join(unwrapped, "\n  "))
	}
}
