package temporal

import (
	"context"
	"time"

	"github.com/timmaaaz/ichor/foundation/logger"
)

// StaleExecutionReaper deletes orphaned StatusPending execution records older than a cutoff.
// Satisfied by workflow/stores/workflowdb.Store.
type StaleExecutionReaper interface {
	ReapStaleExecutions(ctx context.Context, cutoff time.Time) (int64, error)
}

// ExecutionReaperConfig tunes the reaper. Zero-value fields fall back to defaults.
type ExecutionReaperConfig struct {
	Interval time.Duration // how often to sweep (default 1h)
	Window   time.Duration // how old a pending row must be before reaping (default 24h)
}

func (c ExecutionReaperConfig) withDefaults() ExecutionReaperConfig {
	if c.Interval <= 0 {
		c.Interval = time.Hour
	}
	if c.Window <= 0 {
		c.Window = 24 * time.Hour
	}
	return c
}

// ExecutionReaper periodically sweeps orphaned StatusPending automation_executions rows — the
// crash-safe backstop for the rare case where a process dies after the trigger writes the
// pending record (CreateExecution) but before ExecuteWorkflow returns, so fast-follow #3's
// delete-on-error never ran. It mirrors the cascade Relay's reap ticker (relay.go) and is
// SERVER-ONLY (started by the composition root next to the relay); the worker runs neither.
// The Window guards against reaping a legitimately slow-to-start pending row (the only pending
// window is the brief gap before MarkExecutionRunning fires the first activity).
type ExecutionReaper struct {
	log   *logger.Logger
	store StaleExecutionReaper
	cfg   ExecutionReaperConfig
}

// NewExecutionReaper constructs an ExecutionReaper.
func NewExecutionReaper(log *logger.Logger, store StaleExecutionReaper, cfg ExecutionReaperConfig) *ExecutionReaper {
	return &ExecutionReaper{
		log:   log,
		store: store,
		cfg:   cfg.withDefaults(),
	}
}

// Run sweeps every cfg.Interval until ctx is cancelled. Intended to be launched in a goroutine
// by the composition root. Returns ctx.Err() when stopped.
func (r *ExecutionReaper) Run(ctx context.Context) error {
	r.log.Info(ctx, "execution reaper starting", "interval", r.cfg.Interval, "window", r.cfg.Window)

	t := time.NewTicker(r.cfg.Interval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			r.log.Info(ctx, "execution reaper stopping", "reason", ctx.Err())
			return ctx.Err()
		case <-t.C:
			n, err := r.store.ReapStaleExecutions(ctx, time.Now().Add(-r.cfg.Window))
			if err != nil {
				r.log.Error(ctx, "execution reaper: reap failed", "error", err)
				continue
			}
			if n > 0 {
				r.log.Info(ctx, "execution reaper: reaped stale pending executions", "count", n)
			}
		}
	}
}
